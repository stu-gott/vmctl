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

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/core"

	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/service"
)

const hostOverride = ""

type vmCtlApp struct {
	prototypeVMName string
	prototypeNS     string
	namespace       string
	hostOverride    string
}

var _ service.Service = &vmCtlApp{}

func cleanup(virtCli kubecli.KubevirtClient, namespace string, vmName string) {
	deleteOptions := &k8smetav1.DeleteOptions{}
	err := virtCli.VirtualMachine(namespace).Delete(vmName, deleteOptions)
	if err != nil {
		panic(fmt.Errorf("unable to delete VM: %s/%s", namespace, vmName))
	}
}

func deriveVM(vm *v1.VirtualMachine, nodeName string) *v1.VirtualMachine {
	instanceName := fmt.Sprintf("%s-%s", vm.GetName(), nodeName)

	newVM := vm.DeepCopy()
	newVM.ObjectMeta.OwnerReferences = nil
	newVM.Status = v1.VirtualMachineStatus{}

	newVM.ObjectMeta.Name = instanceName
	newVM.Spec.Running = true
	if vm.Spec.Template == nil {
		newVM.Spec.Template = &v1.VirtualMachineInstanceTemplateSpec{}
	}
	newVM.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"] = nodeName

	return newVM
}

func (app *vmCtlApp) Run() {
	hostname := app.hostOverride
	if hostname == "" {
		defaultHostName, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		hostname = defaultHostName
	}

	virtCli, err := kubecli.GetKubevirtClient()
	if err != nil {
		panic(fmt.Errorf("unable to get kubevirt client: %v", err))
	}

	getOptions := &k8smetav1.GetOptions{}
	vm, err := virtCli.VirtualMachine(app.prototypeNS).Get(app.prototypeVMName, getOptions)
	if err != nil {
		panic(fmt.Errorf("unable to fetch prototype vm: %v", err))
	}

	newVM := deriveVM(vm, hostname)
	_, err = virtCli.VirtualMachine(app.namespace).Create(newVM)
	if err != nil {
		panic(fmt.Errorf("unable to create vm: %v", err))
	} else {
		defer cleanup(virtCli, app.namespace, newVM.GetName())
	}

	// wait forever
	stop := make(chan struct{})
	<-stop
}

func (app *vmCtlApp) AddFlags() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	flag.StringVar(&app.prototypeNS, "prototype-ns", core.NamespaceDefault, "Namespace of prototype VirtualMachine")
	flag.StringVar(&app.namespace, "namespace", core.NamespaceDefault, "Namespace to create VirtualMachine in")
	flag.StringVar(&app.hostOverride, "hostname-override", hostOverride,
		"Name under which the node is registered in Kubernetes, where this vmctl instance is running on")
}

func main() {
	app := &vmCtlApp{}
	service.Setup(app)
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Prototype vm name is required\n")
		flag.Usage()
		os.Exit(1)
	} else {
		app.prototypeVMName = flag.Arg(0)
		app.Run()
	}
	os.Exit(0)
}
