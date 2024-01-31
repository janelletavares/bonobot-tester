# Zoom Participant Factory

## Behaviors
* Create Kubernetes Pods that join a Zoom meeting without audio or video
* A single user may not request more than 15 participants in one request
* A single user may not have more than 3 active requests concurrently

## Prerequisites
* Python 3.8
* Docker
* Kind
* Kubectl

## /participant
This directory contains a Python script that connects to a Zoom meeting using Selenium. It also contains a Dockerfile such that this script can be run inside a container.

### Usage
```shell

export HOSTNAME=participant-A # used as Zoom display name
export ZOOM_MEETING_ID=
export ZOOM_MEETING_PASSCODE=
export ZOOM_SESSION_LENGTH_SECONDS=30

# manually
python3 participant.py

# Docker
export DOCKER_REPO=localhost:5001|dockerhub|ECR
make docker_build
make docker_run
```

## Local Cluster

To see the Python script connecting to a Zoom meeting inside Kubernetes pods:

```shell
export DOCKER_REPO=localhost:5001|dockerhub|ECR
make kind_create_cluster
make docker_build
make kind_push_image
make kind_new_participant
kubectl get jobs -A
kubectl get pods -A
kubectl logs ...
make kind_delete_participant
make kind_delete_cluster
```

To see the API managing requests for new participants:

```shell
make kind_create_cluster
make docker_build
make kind_push_image

cd api
export DOCKER_REPO=localhost:5001|dockerhub|ECR
export PARTICIPANT_TAG=latest

source scripts/convenience.sh
make api_run
export API=localhost:8090
// log in to Admin UI
// create user

export API_EMAIL=user@example.com
export API_PASSWORD=abc

get_new_token

export API_TOKEN=...

export ZOOM_MEETING_ID=123
export ZOOM_MEETING_PASSCODE=xyz

create_participant_request $ZOOM_MEETING_ID $ZOOM_MEETING_PASSCODE 20 2

kubectl get jobs -A
kubectl get pods -A
kubectl logs ...

cd ../
make kind_delete_cluster
```


## Demo Video

Click the image below

[![Zoom Participant Factory Demo Video](./images/participants.png)](https://drive.google.com/file/d/11rK04qPdiUtScPDvcP3tykUNEqBBg9lA/view?usp=sharing)