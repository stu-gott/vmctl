[![Build Status](https://travis-ci.com/fabiand/vmctl.svg?branch=master)](https://travis-ci.com/fabiand/vmctl)
[![Docker Repository on Quay](https://quay.io/repository/fabiand/vmctl/status "Docker Repository on Quay")](https://quay.io/repository/fabiand/vmctl)

# Overview

Technically: Controlling KubeVirt VMs from a pod

Logically: Ability to leverge Kubernetes workload controllers with KubeVirt VMs

[![asciicast](https://asciinema.org/a/184816.png)](https://asciinema.org/a/184816)

## Idea

Let a pod control a VM - not directly (as in creating a qemu process), but
indirectly by talking to KubeVirt.

The `vmctl` pod works with _Virtual Machines_ (previousl offline virtual
machines).

## Notes

- The pod only acts as a "remote control" to start and stop virtual machine
  _instances_
- The pod can be simple
- All VM configurations are done on the corresponding _virtual machine_

# Try

## Overview

1. Deploy KubeVirt demo on minikube
2. Deploy `deployment.yaml`
3. Scale the deployment up

## Step by step

### Deploy the KubeVirt demo

Follow this guide to setup the base demo:

https://github.com/kubevirt/demo

### Create deployment

```bash
$ kubectl apply \
  -f https://raw.githubusercontent.com/fabiand/vmctl/master/manifests/deployment.yaml

$ kubectl get deployments
NAME      DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
vmctl     0         0         0            0           3s

```

### Scale the deployment

Now you are ready to scale the deployment, and indirectly the number of VMs:

```bash
$ kubectl scale --replicas=1 deployment/vmctl
deployment.extensions "vmctl" scaled
```

This is just like scaling any other deployment.

You can now check that the scaling really happened:

```bash
$ kubectl get deployments
NAME      DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
vmctl     1         1         1            1           3m
```

And you can also look at the VM instances in to see that really VMs were
spawned:

```bash
$ kubectl get pods,vms
NAME                         READY     STATUS    RESTARTS   AGE
virt-launcher-testvm-j7f8k   2/2       Running   0          39m
vmctl-58ff778cc4-wskgs       1/1       Running   0          23s

NAME                            AGE
testvm                          39m
testvm-vmctl-58ff778cc4-wskgs   20s
```

Scaling down just works as expected:

```bash
$ kubectl scale --replicas=0 deployment/vmctl
deployment.extensions "vmctl" scaled

$ kubectl get deployments
NAME      DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
vmctl     0         0         0            0           3h

$ kubectl get pods,vms
NAME                         READY     STATUS        RESTARTS   AGE
virt-launcher-testvm-j7f8k   2/2       Running       0          41m

NAME      AGE
testvm    41m
```

And the instance is gone again.

> **Note**: One VM will always be defined, as it act's as the prototype
> for the VMs any `vmctl` pod is creating.
