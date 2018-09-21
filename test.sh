#!/bin/bash

kubectl create rolebinding default-admin \
  --clusterrole=cluster-admin \
  --serviceaccount=default:default \
  --namespace default

kubectl apply \
  -f manifests/deployment.yaml


kubectl get deployment | grep vmctl

kubectl scale --replicas=1 deployment/vmctl
kubectl get deployments | egrep "vmctl\s+1"
