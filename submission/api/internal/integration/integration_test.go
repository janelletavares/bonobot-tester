package integration

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tokens"
)

const testDataDir = "../../test_pb_data"

func TestRequestsPermissions(t *testing.T) {
	recordToken, err := generateRecordToken("users", "abc@example.com")
	if err != nil {
		t.Fatal(err)
	}

	adminToken, err := generateAdminToken("admin@example.com")
	if err != nil {
		t.Fatal(err)
	}

	setupTestApp := func(t *testing.T) *tests.TestApp {
		testApp, err := tests.NewTestApp(testDataDir)
		if err != nil {
			t.Fatal(err)
		}

		return testApp
	}
	url := "/api/collections/requests/records"
	request := `{"meeting_id":"123","meeting_passcode":"def","count":"1","duration":"10"}`

	scenarios := []tests.ApiScenario{
		{
			Name:            "try to list requests as guest (aka. no Authorization header)",
			Method:          http.MethodGet,
			Url:             url,
			ExpectedStatus:  200,
			ExpectedContent: []string{`"totalItems":0`},
			ExpectedEvents:  map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "try to list requests as authenticated app user",
			Method: http.MethodGet,
			Url:    url,
			RequestHeaders: map[string]string{
				"Authorization": recordToken,
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"totalItems":1`}, // should only see one created by this user
			ExpectedEvents:  map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "try to list requests as authenticated admin",
			Method: http.MethodGet,
			Url:    url,
			RequestHeaders: map[string]string{
				"Authorization": adminToken,
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"totalItems":2`},
			ExpectedEvents:  map[string]int{"OnRecordsListRequest": 1},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:            "try to create a request as guest (aka. no Authorization header)",
			Method:          http.MethodPost,
			Url:             url,
			Body:            strings.NewReader(request),
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "try to create a request as authenticated app user",
			Method: http.MethodPost,
			Url:    url,
			RequestHeaders: map[string]string{
				"Authorization": recordToken,
			},
			Body:            strings.NewReader(request),
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			TestAppFactory:  setupTestApp,
		},
		{
			Name:   "try create a request as authenticated admin",
			Method: http.MethodPost,
			Url:    url,
			RequestHeaders: map[string]string{
				"Authorization": adminToken,
			},
			Body:            strings.NewReader(request),
			ExpectedStatus:  200,
			ExpectedContent: []string{`"created"`},
			ExpectedEvents: map[string]int{
				"OnModelAfterCreate":          1,
				"OnModelBeforeCreate":         1,
				"OnRecordAfterCreateRequest":  1,
				"OnRecordBeforeCreateRequest": 1,
			},
			TestAppFactory: setupTestApp,
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func generateAdminToken(email string) (string, error) {
	app, err := tests.NewTestApp(testDataDir)
	if err != nil {
		return "", err
	}
	defer app.Cleanup()

	admin, err := app.Dao().FindAdminByEmail(email)
	if err != nil {
		return "", err
	}

	return tokens.NewAdminAuthToken(app, admin)
}

func generateRecordToken(collectionNameOrID string, email string) (string, error) {
	app, err := tests.NewTestApp(testDataDir)
	if err != nil {
		return "", err
	}
	defer app.Cleanup()

	record, err := app.Dao().FindAuthRecordByEmail(collectionNameOrID, email)
	if err != nil {
		return "", err
	}

	return tokens.NewRecordAuthToken(app, record)
}
