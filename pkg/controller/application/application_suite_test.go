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

package application_test

import (
	"testing"

	"github.com/najena/kubebuilder/pkg/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"

	"github.com/kubernetes-sigs/apps_application/pkg/apis"
	"github.com/kubernetes-sigs/apps_application/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-sigs/apps_application/pkg/controller/application"
	"github.com/kubernetes-sigs/apps_application/pkg/controller/sharedinformers"
)

var testenv *test.TestEnvironment
var config *rest.Config
var cs *clientset.Clientset
var shutdown chan struct{}
var controller *application.ApplicationController
var si *sharedinformers.SharedInformers

func TestApplication(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Application Suite", []Reporter{test.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	testenv = &test.TestEnvironment{CRDs: apis.APIMeta.GetCRDs()}
	var err error
	config, err = testenv.Start()
	Expect(err).NotTo(HaveOccurred())
	cs = clientset.NewForConfigOrDie(config)

	shutdown = make(chan struct{})
	si = sharedinformers.NewSharedInformers(config, shutdown)
	controller = application.NewApplicationController(config, si)
	controller.Run(shutdown)
})

var _ = AfterSuite(func() {
	testenv.Stop()
})
