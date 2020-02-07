/*
Copyright 2020 The Kubernetes Authors.

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
	"context"
	"log"
	"os"
	"path"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	appv1beta1 "sigs.k8s.io/application/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestWordPress(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("/workspace/_artifacts/junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Application Type Suite", []Reporter{junitReporter})
}

func clientConfig() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", path.Join(os.Getenv("HOME"), ".kube/config"))
}

func getClientOrDie(config *rest.Config, s *runtime.Scheme) client.Client {
	c, err := client.New(config, client.Options{Scheme: s})
	if err != nil {
		panic(err)
	}
	return c
}

var _ = Describe("Application status should be updated", func() {
	s := scheme.Scheme
	_ = appv1beta1.AddToScheme(s)

	config, err := clientConfig()
	if err != nil {
		log.Fatal("Unable to get client configuration", err)
	}

	It("should update application status", func() {
		kubeClient := getClientOrDie(config, s)
		application := &appv1beta1.Application{}
		objectKey := types.NamespacedName{
			Namespace: metav1.NamespaceDefault,
			Name:      "wordpress-01",
		}
		waitForApplicationStatusUpdated(kubeClient, objectKey, application)
		Expect(application.Status.ObservedGeneration).To(BeNumerically("<=", 5))
		Expect(application.Status.ComponentList.Objects).To(HaveLen(5))
	})

	It("should add ownerReference to components", func() {
		kubeClient := getClientOrDie(config, s)
		matchingLabels := map[string]string{"app.kubernetes.io/name": "wordpress-01"}

		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group: "",
			Kind:  "Service",
		})
		validateComponentOwnerReferences(kubeClient, list, matchingLabels)

		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group: "apps",
			Kind:  "StatefulSet",
		})
		validateComponentOwnerReferences(kubeClient, list, matchingLabels)
	})
})

func validateComponentOwnerReferences(kubeClient client.Client, list *unstructured.UnstructuredList, matchedingLabels map[string]string) {
	componentsUpdated := false
	_ = wait.PollImmediate(time.Second, time.Second*30, func() (bool, error) {

		if err := kubeClient.List(context.TODO(), list, client.InNamespace(metav1.NamespaceDefault), client.MatchingLabels(matchedingLabels)); err != nil {
			return false, err
		}

		updated := true
		for _, item := range list.Items {
			if item.GetOwnerReferences() == nil || len(item.GetOwnerReferences()) < 1 || item.GetOwnerReferences()[0].Name != "wordpress-01" {
				updated = false
			}
		}
		componentsUpdated = updated
		return updated, nil
	})
	Expect(componentsUpdated).To(BeTrue())
}

func waitForApplicationStatusUpdated(kubeClient client.Client, key client.ObjectKey, app *appv1beta1.Application) {
	_ = wait.PollImmediate(time.Second, time.Second*30, func() (bool, error) {
		if err := kubeClient.Get(context.TODO(), key, app); err != nil {
			return false, err
		}

		if app.Status.ComponentList.Objects != nil && len(app.Status.ComponentList.Objects) == 5 && app.Status.Conditions != nil {
			return true, nil
		}
		return false, nil
	})
}
