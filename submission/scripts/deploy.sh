#!/bin/sh

envsubst < yaml/api.yaml | kubectl apply -f -
kubectl port-forward deployment/participant-api 8080:80
