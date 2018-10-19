/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2018 Red Hat, Inc.
 *
 */

package main

import (
	goflag "flag"
	"fmt"
	flag "github.com/spf13/pflag"
	"os"
	"os/signal"
	"syscall"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/core"

	"io/ioutil"
	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/log"
	"kubevirt.io/kubevirt/pkg/service"
)

const (
	hostOverride = ""
	podNamePath  = "/etc/podinfo/name"
	podResource  = "pods"
)

type vmCtlApp struct {
	prototypeVMName string
	prototypeNS     string
	namespace       string
	hostOverride    string
}

var _ service.Service = &vmCtlApp{}

func cleanup(virtCli kubecli.KubevirtClient, namespace string, vmName string) {
	logger := log.DefaultLogger()
	deleteOptions := &k8smetav1.DeleteOptions{}
	err := virtCli.VirtualMachine(namespace).Delete(vmName, deleteOptions)
	if err != nil {
		logger.Errorf("Unable to delete VM: %s/%s", namespace, vmName)
	} else {
		logger.Infof("VM deleted: %s", vmName)
	}
}

func deriveVM(vm *v1.VirtualMachine, nodeName string) *v1.VirtualMachine {
	instanceName := fmt.Sprintf("%s-%s", vm.GetName(), nodeName)

	newVM := &v1.VirtualMachine{}

	spec := vm.Spec.DeepCopy()
	newVM.Spec = *spec

	newVM.ObjectMeta.Name = instanceName
	newVM.Spec.Running = true
	if vm.Spec.Template == nil {
		newVM.Spec.Template = &v1.VirtualMachineInstanceTemplateSpec{}
	}
	if newVM.Spec.Template.Spec.NodeSelector == nil {
		newVM.Spec.Template.Spec.NodeSelector = map[string]string{}
	}
	newVM.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"] = nodeName

	return newVM
}

func getNodeName(virtCli kubecli.KubevirtClient, namespace string) (string, error) {
	podName, err := ioutil.ReadFile(podNamePath)
	if err != nil {
		return "", fmt.Errorf("Unable to find pod name: %v", err)
	}

	kubeClient := virtCli.CoreV1()
	getOptions := k8smetav1.GetOptions{}

	pod, err := kubeClient.Pods(namespace).Get(string(podName), getOptions)
	if err != nil {
		return "", fmt.Errorf("Unable to look get pod: %v", err)
	}

	return pod.Spec.NodeName, nil
}

func (app *vmCtlApp) Run() {
	logger := log.DefaultLogger()

	virtCli, err := kubecli.GetKubevirtClient()
	if err != nil {
		logger.Reason(err).Errorf("Unable to get KubeVirt client")
		return
	}

	hostname := app.hostOverride
	if hostname == "" {
		hostname, err = getNodeName(virtCli, app.namespace)
		if err != nil {
			panic(err)
		}
	}

	logger.Infof("Running on node: %s", hostname)

	getOptions := &k8smetav1.GetOptions{}
	prototypeNS := app.prototypeNS
	if prototypeNS == "" {
		prototypeNS = app.namespace
	}
	vm, err := virtCli.VirtualMachine(prototypeNS).Get(app.prototypeVMName, getOptions)
	if err != nil {
		logger.Reason(err).Errorf("Unable to fetch prototype VM")
		return
	}

	newVM := deriveVM(vm, hostname)
	_, err = virtCli.VirtualMachine(app.namespace).Create(newVM)
	if err != nil {
		logger.Reason(err).Errorf("Unable to create VM")
		return
	} else {
		defer cleanup(virtCli, app.namespace, newVM.GetName())
	}

	logger.Object(newVM).Infof("Virtual Machine launched")

	// wait forever
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
}

func (app *vmCtlApp) AddFlags() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	flag.StringVar(&app.namespace, "namespace", core.NamespaceDefault, "Namespace to create VirtualMachine in.")
	flag.StringVar(&app.prototypeNS, "proto-namespace", "", "Namespace of prototype VirtualMachine. Defaults to <namespace>")

	flag.StringVar(&app.hostOverride, "hostname-override", hostOverride,
		"Name under which the node is registered in Kubernetes, where this vmctl instance is running on.")
}

func main() {
	app := &vmCtlApp{}
	service.Setup(app)
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Prototype VM name is required\n")
		flag.Usage()
		os.Exit(1)
	} else {
		app.prototypeVMName = flag.Arg(0)
		app.Run()
	}
	os.Exit(0)
}
