#!/bin/bash

kubectl delete -f ./pod.yaml
go build ./k8saadmsi.go
docker build -t trstringer/k8saadmsi:latest .
docker push trstringer/k8saadmsi:latest
kubectl apply -f ./pod.yaml
