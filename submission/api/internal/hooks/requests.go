package hooks

import (
	"context"
	"fmt"
	"net/http"
	"os"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	batchv1type "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1type "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
)

const maxRequests = 3

func AddHooks(ctx context.Context, app *pocketbase.PocketBase, clientset *kubernetes.Clientset) {
	app.OnRecordBeforeCreateRequest("requests").Add(func(e *core.RecordCreateEvent) error {
		return validateParticipantCreateRequest(ctx, app, e, clientset.BatchV1())
	})
	app.OnRecordAfterCreateRequest("requests").Add(func(e *core.RecordCreateEvent) error {
		return createK8sParticipant(ctx, e, clientset.CoreV1(), clientset.BatchV1())
	})
}

// validateParticipantCreateRequest ensures that completed requests are removed
// from the database and that the new request does not violate per user limits.
func validateParticipantCreateRequest(
	ctx context.Context,
	app *pocketbase.PocketBase,
	e *core.RecordCreateEvent,
	bv1 batchv1type.BatchV1Interface) error {
	// use the authenticated user to assign the requester field
	userID := e.Record.GetString("requester")
	if userID == "" {
		var err error
		userID, err = findUserID(e.HttpContext)
		if err != nil {
			return err
		}
		e.Record.Set("requester", userID)
	}

	// clean up previous requests and count active requests
	dao := app.Dao()
	requests, err := dao.FindRecordsByExpr("requests", dbx.HashExp{"requester": userID})
	if err != nil {
		return apis.NewApiError(http.StatusInternalServerError,
			"unable to determine status",
			map[string]validation.Error{
				"requests": validation.NewError("internal_error", err.Error()),
			},
		)
	}
	activeRequests, err := processRequests(ctx, dao, requests, bv1)
	if err != nil {
		return apis.NewApiError(http.StatusInternalServerError,
			"unable to determine the status of the authenticated users existing requests",
			map[string]validation.Error{
				"requests": validation.NewError("internal_error", err.Error()),
			},
		)
	}

	// limit the number of concurrent active requests
	if activeRequests >= maxRequests {
		return apis.NewApiError(
			http.StatusForbidden,
			"no more active requests allowed",
			map[string]validation.Error{
				"limit": validation.NewError("active_request_limit",
					"only 3 active participant requests are allowed concurrently"),
			},
		)
	}

	return nil
}

func findUserID(e echo.Context) (string, error) {
	var userID string
	ri := apis.RequestInfo(e)
	if ri.Admin != nil {
		userID = ri.Admin.Id
	}
	if ri.AuthRecord != nil {
		userID = ri.AuthRecord.Id
	}
	if userID == "" {
		return "", apis.NewBadRequestError(
			"Invalid request",
			map[string]validation.Error{
				"requester": validation.NewError("invalid_requester",
					"unable to determine the requester from the authenticated user"),
			},
		)
	}
	return userID, nil
}

// processRequests deletes previously completed requests and returns the active count
// NOTE: the list requests output will be stale until that route is also updated.
func processRequests(
	ctx context.Context,
	dao *daos.Dao,
	requests []*models.Record,
	bv1 batchv1type.BatchV1Interface) (int, error) {
	activeRequests := 0
	for _, r := range requests {
		namespace := r.GetId()
		jobList, err := bv1.Jobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return 0, err
		}
		for _, job := range jobList.Items {
			if job.Status.StartTime != nil {
				if job.Status.Active != 0 {
					activeRequests++
				} else {
					err = dao.DeleteRecord(r)
					if err != nil {
						return 0, err
					}
				}
			}
		}
	}
	return activeRequests, nil
}

// createK8sParticipant creates a namespace and job to create a Zoom participant.
func createK8sParticipant(
	ctx context.Context,
	e *core.RecordCreateEvent,
	cv1 corev1type.CoreV1Interface,
	bv1 batchv1type.BatchV1Interface) error {
	namespace := e.Record.GetId()
	ns, err := cv1.Namespaces().Create(ctx, generateNamespaceSpec(namespace), metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Error occurred while creating namespace %s: %s", ns.Name, err.Error())
		return apis.NewApiError(http.StatusInternalServerError,
			"unable to process request",
			map[string]validation.Error{
				"namespace": validation.NewError("internal_error", err.Error()),
			},
		)
	}

	klog.Infof("Namespace %s is successfully created", namespace)

	meetingID := e.Record.GetString("meeting_id")
	meetingPasscode := e.Record.GetString("meeting_passcode")
	count := e.Record.GetInt("count")
	duration := e.Record.GetInt("seconds")
	jobSpec := generateJobSpec(namespace, meetingID, meetingPasscode, int32(count), duration)
	job, err := bv1.Jobs(namespace).Create(ctx, jobSpec, metav1.CreateOptions{})
	if err != nil {
		klog.Errorf("Error occurred while creating job %s: %s", job.Name, err.Error())
		return apis.NewApiError(http.StatusInternalServerError,
			"unable to process request",
			map[string]validation.Error{
				"job": validation.NewError("internal_error", err.Error()),
			},
		)
	}

	klog.Infof("Job for namespace %s with count %d is successfully created", namespace, count)
	return nil
}

func generateNamespaceSpec(namespace string) *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
}

func generateJobSpec(namespace, meetingID, meetingPasscode string, count int32, duration int) *batchv1.Job {
	partTag := os.Getenv("PARTICIPANT_TAG")
	dockerRepo := os.Getenv("DOCKER_REPO")

	const jobName = "participant"
	const meetingIDEnvVar = "ZOOM_MEETING_ID"
	const meetingPasscodeEnvVar = "ZOOM_MEETING_PASSCODE"
	const meetingDurationEnvVar = "ZOOM_SESSION_LENGTH_SECONDS"

	image := fmt.Sprintf("%s/participant:%s", dockerRepo, partTag)
	completionMode := batchv1.IndexedCompletion

	env := []corev1.EnvVar{
		{Name: meetingIDEnvVar, Value: meetingID},
		{Name: meetingPasscodeEnvVar, Value: meetingPasscode},
		{Name: meetingDurationEnvVar, Value: fmt.Sprintf("%d", duration)},
	}

	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Completions:    &count,
			Parallelism:    &count,
			CompletionMode: &completionMode,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  jobName,
							Image: image,
							Env:   env,
						},
					},
				},
			},
		},
	}
}
