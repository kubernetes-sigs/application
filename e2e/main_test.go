/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	appv1beta1 "sigs.k8s.io/application/api/v1beta1"
	"sigs.k8s.io/application/e2e/testutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("/workspace/_artifacts/junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Application Type Suite", []Reporter{junitReporter})
}

func getClientConfig() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", path.Join(os.Getenv("HOME"), ".kube/config"))
}

func getKubeClientOrDie(config *rest.Config, s *runtime.Scheme) client.Client {
	c, err := client.New(config, client.Options{Scheme: s})
	if err != nil {
		panic(err)
	}
	return c
}

const (
	crdPath         = "../config/crd/bases/app.k8s.io_applications.yaml"
	applicationPath = "../config/samples/app_v1beta1_application.yaml"
)

var _ = Describe("Application CRD should install correctly", func() {
	s := scheme.Scheme
	_ = appv1beta1.AddToScheme(s)

	crd, err := testutil.ParseCRDYaml(crdPath)
	if err != nil {
		log.Fatal("Unable to parse CRD YAML", err)
	}

	config, err := getClientConfig()
	if err != nil {
		log.Fatal("Unable to get client configuration", err)
	}

	extClient, err := apiextcs.NewForConfig(config)
	if err != nil {
		log.Fatal("Unable to construct extensions client", err)
	}

	It("should create CRD", func() {
		err = testutil.CreateCRD(extClient, crd)
		Expect(err).NotTo(HaveOccurred())
		err = testutil.WaitForCRDOrDie(extClient, crd.Name)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should register an application", func() {
		client := getKubeClientOrDie(config, s) //Make sure to create the client after CRD has been created.
		err = testutil.CreateApplication(client, "default", applicationPath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should delete application", func() {
		client := getKubeClientOrDie(config, s)
		err = testutil.DeleteApplication(client, "default", applicationPath)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should delete application CRD", func() {
		err = testutil.DeleteCRD(extClient, crd.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})
