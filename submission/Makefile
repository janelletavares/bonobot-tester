DOCKER_REPO ?= janelletavares
PARTICIPANT_TAG ?= latest

docker_build:
	docker build participant -f participant/Dockerfile -t $(DOCKER_REPO)/participant:$(PARTICIPANT_TAG)

docker_run:
	docker run --rm -it \
    --env HOSTNAME=$(HOSTNAME) \
    --env ZOOM_MEETING_ID=$(ZOOM_MEETING_ID) \
    --env ZOOM_MEETING_PASSCODE=$(ZOOM_MEETING_PASSCODE) \
    --env ZOOM_SESSION_LENGTH_SECONDS=$(ZOOM_SESSION_LENGTH_SECONDS) \
    $(DOCKER_REPO)/participant:$(PARTICIPANT_TAG)

kind_create_cluster:
	scripts/create_kind_cluster_with_registry.sh

kind_delete_cluster:
	kind delete cluster --name zoom-participant-factory

kind_push_image:
	kind load docker-image $(DOCKER_REPO)/participant:$(PARTICIPANT_TAG) --name zoom-participant-factory

kind_new_participant:
	envsubst < yaml/participant.yaml | kubectl apply -f -

kind_delete_participant:
	kubectl delete job participant
