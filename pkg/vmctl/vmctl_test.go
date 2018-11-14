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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"testing"

	k8sv1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/golang/mock/gomock"
	"kubevirt.io/kubevirt/pkg/api/v1"
	"kubevirt.io/kubevirt/pkg/kubecli"
	"kubevirt.io/kubevirt/pkg/log"
	"os"
	"path"
)

var _ = Describe("Vmctl", func() {
	Context("Given a VirtualMachine", func() {
		var protoVm *v1.VirtualMachine
		var vmName string
		var workDir string
		var ctrl *gomock.Controller
		var virtClient *kubecli.MockKubevirtClient
		var fakeClientSet *fake.Clientset
		var podName string
		var nodeName string
		var namespace string
		var mockVmInterface *kubecli.MockVirtualMachineInterface

		BeforeEach(func() {
			var err error
			protoVm, vmName = NewRandomVM()

			workDir, err = ioutil.TempDir("", "vmctl-test-")
			if err != nil {
				panic(err)
			}

			ctrl = gomock.NewController(GinkgoT())
			virtClient = kubecli.NewMockKubevirtClient(ctrl)
			mockVmInterface = kubecli.NewMockVirtualMachineInterface(ctrl)
			fakeClientSet = fake.NewSimpleClientset()

			podName = fmt.Sprintf("vmctl-pod-%s", NewRandomString(6))
			nodeName = fmt.Sprintf("vmctl-node-%s", NewRandomString(6))
			namespace = "vmctl-default-test"
		})

		AfterEach(func() {
			os.RemoveAll(workDir)
		})

		It("Should successfully derive a new VM", func() {
			newVm := deriveVM(protoVm, podName, nodeName)
			Expect(newVm.Spec.Running).To(BeTrue(),
				"Derived VM should always be running")

			derivedVmName := fmt.Sprintf("%s-%s", vmName, podName)
			Expect(newVm.ObjectMeta.Name).To(Equal(derivedVmName),
				"Expected derived VM name to be <vmName>-<podName>")

			Expect(newVm.Spec.Template.Spec.NodeSelector).ToNot(BeNil())
			nodeSelector, ok := newVm.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"]
			Expect(ok).To(BeTrue())
			Expect(nodeSelector).To(Equal(nodeName),
				"Expected node selector to match node name")
		})

		It("Should successfully read pod name", func() {
			podInfoPath := path.Join(workDir, "podname")
			err := ioutil.WriteFile(podInfoPath, []byte(podName), 0644)
			Expect(err).ToNot(HaveOccurred())

			newPodName, err := getPodName(podInfoPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(newPodName).To(Equal(podName))
		})

		It("Should derive pod node name from pod", func() {
			fakeCoreV1 := fakeClientSet.CoreV1()
			pod := &k8sv1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name: podName,
				},
				Spec: k8sv1.PodSpec{
					NodeName: nodeName,
				},
			}
			fakeCoreV1.Pods(namespace).Create(pod)
			virtClient.EXPECT().CoreV1().Return(fakeCoreV1).AnyTimes()

			newNodeName, err := getPodNodeName(virtClient, namespace, podName)
			Expect(err).ToNot(HaveOccurred())
			Expect(newNodeName).To(Equal(nodeName))
		})

		It("Should raise error if no pod found", func() {
			fakeCoreV1 := fakeClientSet.CoreV1()
			virtClient.EXPECT().CoreV1().Return(fakeCoreV1).AnyTimes()

			_, err := getPodNodeName(virtClient, namespace, podName)
			Expect(err).To(HaveOccurred())
		})

		It("Should clean up vm", func() {
			deleteOptions := &k8smetav1.DeleteOptions{}

			virtClient.EXPECT().VirtualMachine(namespace).Return(mockVmInterface).AnyTimes()
			mockVmInterface.EXPECT().Delete(vmName, deleteOptions).Return(nil).AnyTimes()

			cleanup(virtClient, namespace, vmName)
		})

		It("Should clean up vm on error", func() {
			deleteOptions := &k8smetav1.DeleteOptions{}

			err := fmt.Errorf("fake error for testing")
			virtClient.EXPECT().VirtualMachine(namespace).Return(mockVmInterface).AnyTimes()
			mockVmInterface.EXPECT().Delete(vmName, deleteOptions).Return(err).AnyTimes()

			cleanup(virtClient, namespace, vmName)
		})

		It("Should run without error", func() {
			// Set up mocks for the cleanup function
			deleteOptions := &k8smetav1.DeleteOptions{}
			virtClient.EXPECT().VirtualMachine(namespace).Return(mockVmInterface).AnyTimes()
			mockVmInterface.EXPECT().Delete(gomock.Any(), deleteOptions).Return(nil).AnyTimes()

			// set up mocks for getting the Pod's node name
			fakeCoreV1 := fakeClientSet.CoreV1()
			pod := &k8sv1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name: podName,
				},
				Spec: k8sv1.PodSpec{
					NodeName: nodeName,
				},
			}
			fakeCoreV1.Pods(namespace).Create(pod)
			virtClient.EXPECT().CoreV1().Return(fakeCoreV1).AnyTimes()

			// set up mocks for getting the Prototype VirtualMachine
			getOptions := &k8smetav1.GetOptions{}
			mockVmInterface.EXPECT().Get(vmName, getOptions).Return(protoVm, nil).AnyTimes()

			// set up mocks for creating the derived VirtualMachine
			mockVmInterface.EXPECT().Create(gomock.Any()).Return(nil, nil).AnyTimes()

			// set up fake pod name
			podInfoPath := path.Join(workDir, "podname")
			err := ioutil.WriteFile(podInfoPath, []byte(podName), 0644)
			Expect(err).ToNot(HaveOccurred())

			// By closing the stop channel, the Run function will end immediately
			// after setting everything up.
			stop := make(chan struct{})
			close(stop)

			app := NewVmctlApp(virtClient, vmName, "", namespace, nodeName)
			app.namePath = podInfoPath
			err = app.Run(stop)
			Expect(err).ToNot(HaveOccurred())
		})

		It("It should fail if pod name can't be determined", func() {
			// define pod name path, but don't create the file
			podInfoPath := path.Join(workDir, "podname")

			// By closing the stop channel, the Run function will end immediately
			// after setting everything up.
			stop := make(chan struct{})
			close(stop)

			app := NewVmctlApp(virtClient, vmName, "", namespace, "")
			app.namePath = podInfoPath
			err := app.Run(stop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find pod name"))
		})

		It("It should fail if node name can't be determined", func() {
			podInfoPath := path.Join(workDir, "podname")
			err := ioutil.WriteFile(podInfoPath, []byte(podName), 0644)
			Expect(err).ToNot(HaveOccurred())

			// set up core client, but don't add the pod
			fakeCoreV1 := fakeClientSet.CoreV1()
			virtClient.EXPECT().CoreV1().Return(fakeCoreV1).AnyTimes()

			// By closing the stop channel, the Run function will end immediately
			// after setting everything up.
			stop := make(chan struct{})
			close(stop)

			app := NewVmctlApp(virtClient, vmName, "", namespace, "")
			app.namePath = podInfoPath
			err = app.Run(stop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to get pod"))
		})

		It("Should fail if prototype VM can't be found", func() {
			// Set up mocks for the cleanup function
			deleteOptions := &k8smetav1.DeleteOptions{}
			virtClient.EXPECT().VirtualMachine(namespace).Return(mockVmInterface).AnyTimes()
			mockVmInterface.EXPECT().Delete(gomock.Any(), deleteOptions).Return(nil).AnyTimes()

			// set up mocks for getting the Pod's node name
			fakeCoreV1 := fakeClientSet.CoreV1()
			pod := &k8sv1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name: podName,
				},
				Spec: k8sv1.PodSpec{
					NodeName: nodeName,
				},
			}
			fakeCoreV1.Pods(namespace).Create(pod)
			virtClient.EXPECT().CoreV1().Return(fakeCoreV1).AnyTimes()

			// set up mocks for getting the Prototype VirtualMachine
			getOptions := &k8smetav1.GetOptions{}
			errorMsg := fmt.Sprintf("fake get vm error: %s", NewRandomString(6))
			err := fmt.Errorf(errorMsg)
			mockVmInterface.EXPECT().Get(vmName, getOptions).Return(nil, err).AnyTimes()

			// set up fake pod name
			podInfoPath := path.Join(workDir, "podname")
			err = ioutil.WriteFile(podInfoPath, []byte(podName), 0644)
			Expect(err).ToNot(HaveOccurred())

			// By closing the stop channel, the Run function will end immediately
			// after setting everything up.
			stop := make(chan struct{})
			close(stop)

			app := NewVmctlApp(virtClient, vmName, "", namespace, nodeName)
			app.namePath = podInfoPath
			err = app.Run(stop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMsg))
		})

		It("Should fail if new VM can't be created", func() {
			// Set up mocks for the cleanup function
			deleteOptions := &k8smetav1.DeleteOptions{}
			virtClient.EXPECT().VirtualMachine(namespace).Return(mockVmInterface).AnyTimes()
			mockVmInterface.EXPECT().Delete(gomock.Any(), deleteOptions).Return(nil).AnyTimes()

			// set up mocks for getting the Pod's node name
			fakeCoreV1 := fakeClientSet.CoreV1()
			pod := &k8sv1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name: podName,
				},
				Spec: k8sv1.PodSpec{
					NodeName: nodeName,
				},
			}
			fakeCoreV1.Pods(namespace).Create(pod)
			virtClient.EXPECT().CoreV1().Return(fakeCoreV1).AnyTimes()

			// set up mocks for getting the Prototype VirtualMachine
			getOptions := &k8smetav1.GetOptions{}
			mockVmInterface.EXPECT().Get(vmName, getOptions).Return(protoVm, nil).AnyTimes()

			// set up mocks for creating the derived VirtualMachine
			errorMsg := fmt.Sprintf("fake create vm error: %s", NewRandomString(6))
			err := fmt.Errorf(errorMsg)
			mockVmInterface.EXPECT().Create(gomock.Any()).Return(nil, err).AnyTimes()

			// set up fake pod name
			podInfoPath := path.Join(workDir, "podname")
			err = ioutil.WriteFile(podInfoPath, []byte(podName), 0644)
			Expect(err).ToNot(HaveOccurred())

			// By closing the stop channel, the Run function will end immediately
			// after setting everything up.
			stop := make(chan struct{})
			close(stop)

			app := NewVmctlApp(virtClient, vmName, "", namespace, nodeName)
			app.namePath = podInfoPath
			err = app.Run(stop)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(errorMsg))
		})
	})
})

// numBytes will be base64 encoded, so basically resulting
// string will be 4/3 as long as the bytes its derived from
func NewRandomString(numBytes int) string {
	data := make([]byte, numBytes)
	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func NewRandomVM() (*v1.VirtualMachine, string) {
	var vmName string
	suffix := NewRandomString(6)
	vmName = fmt.Sprintf("vmctl-test-%s", suffix)
	return kubecli.NewMinimalVM(vmName), vmName
}

func TestTemplate(t *testing.T) {
	log.Log.SetIOWriter(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vmctl")
}
