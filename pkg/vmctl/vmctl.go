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

package vmctl

import (
	"fmt"
	"io/ioutil"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/log"
)

const (
	podNamePath = "/etc/podinfo/name"
)

type vmctlApp struct {
	prototypeVMName string
	prototypeNS     string
	namespace       string
	hostOverride    string
	client          kubecli.KubevirtClient
	namePath        string
}

func NewVmctlApp(virtCli kubecli.KubevirtClient, prototypeVMName string, prototypeNamespace string, namespace string, hostOverride string) *vmctlApp {
	return &vmctlApp{
		prototypeVMName: prototypeVMName,
		prototypeNS:     prototypeNamespace,
		namespace:       namespace,
		hostOverride:    hostOverride,
		client:          virtCli,
		namePath:        podNamePath,
	}
}

func cleanup(virtCli kubecli.KubevirtClient, namespace string, vmName string) {
	logger := log.DefaultLogger()
	deleteOptions := &k8smetav1.DeleteOptions{}
	err := virtCli.VirtualMachine(namespace).Delete(vmName, deleteOptions)
	if err != nil {
		logger.Reason(err).Errorf("Unable to delete VM: %s/%s", namespace, vmName)
	} else {
		logger.Infof("VM deleted: %s/%s", namespace, vmName)
	}
}

func deriveVM(vm *v1.VirtualMachine, podName string, nodeName string) *v1.VirtualMachine {
	instanceName := fmt.Sprintf("%s-%s", vm.GetName(), podName)

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

func getPodName(namePath string) (string, error) {
	podName, err := ioutil.ReadFile(namePath)
	if err != nil {
		return "", fmt.Errorf("unable to find pod name: %v", err)
	}
	return string(podName), nil
}

func getPodNodeName(virtCli kubecli.KubevirtClient, namespace string, podName string) (string, error) {
	kubeClient := virtCli.CoreV1()
	getOptions := k8smetav1.GetOptions{}

	pod, err := kubeClient.Pods(namespace).Get(podName, getOptions)
	if err != nil {
		return "", fmt.Errorf("unable to get pod: %v", err)
	}

	return pod.Spec.NodeName, nil
}

func (app *vmctlApp) Run(stop chan struct{}) error {
	logger := log.DefaultLogger()

	podName, err := getPodName(app.namePath)
	if err != nil {
		logger.Reason(err).Errorf("Unable to get pod name")
		return err
	}

	nodeName := app.hostOverride
	if nodeName == "" {
		nodeName, err = getPodNodeName(app.client, app.namespace, podName)
		if err != nil {
			logger.Reason(err).Errorf("Unable to get node name")
			return err
		}
	}

	logger.Infof("Running on node: %s", nodeName)

	getOptions := &k8smetav1.GetOptions{}
	prototypeNS := app.prototypeNS
	if prototypeNS == "" {
		prototypeNS = app.namespace
	}
	vm, err := app.client.VirtualMachine(prototypeNS).Get(app.prototypeVMName, getOptions)
	if err != nil {
		logger.Reason(err).Errorf("Unable to fetch prototype VM")
		return err
	}

	newVM := deriveVM(vm, podName, nodeName)
	_, err = app.client.VirtualMachine(app.namespace).Create(newVM)
	if err != nil {
		logger.Reason(err).Errorf("Unable to create VM")
		return err
	}

	logger.Object(newVM).Infof("Virtual Machine launched")

	// wait forever
	<-stop
	cleanup(app.client, app.namespace, newVM.GetName())

	return nil
}
