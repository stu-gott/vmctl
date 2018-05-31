# Overview

Technically: Controlling KubeVirt VMs from a pod

Logically: Ability to leverge Kubernetes workload controllers with KubeVirt VMs

# Try

1. Deploy KubeVirt demo on minikube
2. Deploy `deployment.yaml`

```
kubectl create rolebinding default-admin --clusterrole=cluster-admin --serviceaccount=default:default --namespace default
```
