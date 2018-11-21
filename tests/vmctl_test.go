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

package tests


import (
	"os"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubetests "kubevirt.io/kubevirt/tests"
	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/vmctl/pkg/vmctl"
	"kubevirt.io/kubevirt/pkg/kubecli"
)

var _ = Describe("Vmctl", func() {
	Context("Given an existing VM", func() {
		var workDir string
		var prototypeVM v1.VirtualMachine
		var virtClient kubecli.KubevirtClient
		var nodeName string
		var stopChan chan struct{}

		BeforeEach(func() {
			kubetests.BeforeTestCleanup()

			vmi := kubetests.NewRandomVMIWithEphemeralDisk(kubetests.RegistryDiskFor(kubetests.RegistryDiskAlpine))
			prototypeVM = kubetests.NewRandomVirtualMachine(vmi, false)

			var err error
			workDir, err = ioutil.TempDir("", "vmctl-test-")
			if err != nil {
				panic(err)
			}
			virtClient, err := kubecli.GetKubevirtClient()
			Expect(err).ToNot(HaveOccurred())

			nodeName, err = GetFirstClusterNodeName(virtClient)
			Expect(err).ToNot(HaveOccurred(), "Error looking up node name")
			stopChan = make(chan struct{})
		})

		AfterEach(func() {
			os.RemoveAll(workDir)
			close(stopChan)
		})

		It("Should derive a vm when invoked", func() {
			app := vmctl.NewVmctlApp(virtClient, prototypeVM.GetName(), prototypeVM.Namespace, kubetests.NamespaceTestDefault, nodeName)
			app.Run(stopChan)

		})
	})
})

func GetFirstClusterNodeName(virtClient kubecli.KubevirtClient) (string, error) {
	coreClient := virtClient.CoreV1()
	listOptions := k8smetav1.ListOptions{}
	nodeList, err := coreClient.Nodes().List(listOptions)
	if err != nil {
		return "", err
	}
	if len(nodeList.Items) == 0 {
		return "", fmt.Errorf("cluster has no nodes")
	}
	node := nodeList.Items[0]
	return node.GetName(), nil
}