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

	"k8s.io/kubernetes/pkg/apis/core"

	"kubevirt.io/vmctl/pkg/vmctl"
)

const (
	hostOverride = ""
)

func main() {
	var prototypeNS string
	var namespace string
	var nodeName string

	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.StringVar(&namespace, "namespace", core.NamespaceDefault, "Namespace to create VirtualMachine in.")
	flag.StringVar(&prototypeNS, "proto-namespace", "", "Namespace of prototype VirtualMachine. Defaults to <namespace>")

	flag.StringVar(&nodeName, "hostname-override", hostOverride,
		"Name under which the node is registered in Kubernetes, where this vmctl instance is running on.")

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Prototype VM name is required\n")
		flag.Usage()
		os.Exit(1)
	} else {
		app := vmctl.NewVmctlApp(flag.Arg(0), prototypeNS, namespace, nodeName)
		app.Run()
	}
	os.Exit(0)
}
