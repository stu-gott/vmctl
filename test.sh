#!/bin/bash

set -e

kubectl apply -f demo/manifests/vm.yaml

kubectl create rolebinding default-admin \
  --clusterrole=cluster-admin \
  --serviceaccount=default:default \
  --namespace default || :

kubectl apply \
  -f manifests/deployment.yaml

kubectl get deployment | grep vmctl

for SCALE in 1 2 3;
do
  kubectl scale --replicas=$SCALE deployment/vmctl
  kubectl get deployments | egrep "vmctl\s+$SCALE"

  kubectl wait --for condition=available deployment vmctl

  kubectl get vm,vmi

  sleep 4.2  # Give some time to schedule the VMI

  test $(kubectl get vmi | grep testvm-vmctl | wc -l ) -eq $SCALE
done
