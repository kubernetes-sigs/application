// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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
	crdv080Path     = "resources/app-crd-v0.8.0.yaml"
	testCrdPath     = "resources/withcrd/test_crd.yaml"
	applicationPath = "resources/withcrd/base/application.yaml"
	waitTimeout     = time.Second * 120
	pullPeriod      = time.Second * 2
	syncPeriod      = "2"
)

var _ = Describe("Application CRD e2e", func() {
	s := scheme.Scheme
	_ = appv1beta1.AddToScheme(s)

	crdv080, err := testutil.ParseCRDYaml(crdv080Path)
	if err != nil {
		log.Fatal("Unable to parse CRD YAML with version 0.8.0", err)
	}

	crd, err := testutil.ParseCRDYaml(crdPath)
	if err != nil {
		log.Fatal("Unable to parse CRD YAML", err)
	}

	testcrd, err := testutil.ParseCRDYaml(testCrdPath)
	if err != nil {
		log.Fatal("Unable to parse test CRD YAML", err)
	}

	config, err := getClientConfig()
	if err != nil {
		log.Fatal("Unable to get client configuration", err)
	}

	extClient, err := apiextcs.NewForConfig(config)
	if err != nil {
		log.Fatal("Unable to construct extensions client", err)
	}

	var managerStdout bytes.Buffer
	var managerStderr bytes.Buffer
	managerCmd := exec.Command("../bin/kube-app-manager", "--sync-period", syncPeriod)
	managerCmd.Stdout = &managerStdout
	managerCmd.Stderr = &managerStderr

	It("should create CRD v0.8.0", func() {
		err = testutil.CreateOrUpdateCRD(extClient, crdv080)
		Expect(err).NotTo(HaveOccurred())
		err = testutil.WaitForCRDOrDie(extClient, crdv080.Name)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should create the wordpress application", func() {
		err = applyKustomize("../docs/examples/wordpress")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should update the CRD v0.8.0 to new CRD", func() {
		err = testutil.CreateOrUpdateCRD(extClient, crd)
		Expect(err).NotTo(HaveOccurred())
		err = testutil.WaitForCRDOrDie(extClient, crd.Name)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should create test CRD", func() {
		err = testutil.CreateOrUpdateCRD(extClient, testcrd)
		Expect(err).NotTo(HaveOccurred())
		err = testutil.WaitForCRDOrDie(extClient, testcrd.Name)
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

	It("should start the controller", func() {
		err = managerCmd.Start()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should create the test application with custom resources", func() {
		err = applyKustomize("resources/withcrd/overlays/working")
		Expect(err).NotTo(HaveOccurred())
	})

	It("should update wordpress-01 status", func() {
		kubeClient := getKubeClientOrDie(config, s)
		application := &appv1beta1.Application{}
		objectKey := types.NamespacedName{
			Namespace: metav1.NamespaceDefault,
			Name:      "wordpress-01",
		}
		waitForApplicationStatusToHaveNComponents(kubeClient, objectKey, application, 5, 5)
		Expect(application.Status.ObservedGeneration).To(BeNumerically("<=", 5))
		Expect(application.Status.ComponentList.Objects).To(HaveLen(5))
	})

	It("should update ok-withcrd status", func() {
		kubeClient := getKubeClientOrDie(config, s)
		application := &appv1beta1.Application{}
		objectKey := types.NamespacedName{
			Namespace: metav1.NamespaceDefault,
			Name:      "ok-withcrd",
		}
		waitForApplicationStatusToHaveNComponents(kubeClient, objectKey, application, 7, 7)
		Expect(application.Status.ObservedGeneration).To(BeNumerically("<=", 7))
		Expect(application.Status.ComponentList.Objects).To(HaveLen(7))
	})

	It("should add ownerReference to components", func() {
		kubeClient := getKubeClientOrDie(config, s)
		matchingLabels := map[string]string{"app.kubernetes.io/name": "wordpress-01"}

		list := &unstructured.UnstructuredList{}
		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group: "",
			Kind:  "Service",
			Version: "v1",
		})
		validateComponentOwnerReferences(kubeClient, list, matchingLabels, "wordpress-01")

		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group: "apps",
			Kind:  "StatefulSet",
			Version: "v1",
		})
		validateComponentOwnerReferences(kubeClient, list, matchingLabels, "wordpress-01")

		matchingLabels = map[string]string{"app.kubernetes.io/name": "test-01"}
		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group: "test.crd.com",
			Kind:  "TestCRD",
			Version: "v1",
		})
		validateComponentOwnerReferences(kubeClient, list, matchingLabels, "test-application-01")
	})

	It("should mark the application not-ready if not all components are ready", func() {
		err = applyKustomize("resources/withcrd/overlays/broken")
		Expect(err).NotTo(HaveOccurred())

		kubeClient := getKubeClientOrDie(config, s)
		application := &appv1beta1.Application{}
		objectKey := types.NamespacedName{
			Namespace: metav1.NamespaceDefault,
			Name:      "nok-withcrd",
		}
		waitForApplicationStatusToHaveNComponents(kubeClient, objectKey, application, 6, 7)
		Expect(application.Status.ObservedGeneration).To(BeNumerically("<=", 7))
		Expect(hasConditionTypeStatusAndReason(application.Status.Conditions, appv1beta1.StatusReady, corev1.ConditionFalse, "ComponentsNotReady")).To(BeTrue())
	})

	It("should stop the controller", func() {
		err = managerCmd.Process.Signal(os.Interrupt)
		_, _ = io.Copy(os.Stderr, &managerStderr)
		_, _ = io.Copy(os.Stdout, &managerStdout)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should delete application CRD", func() {
		err = testutil.DeleteCRD(extClient, crd.Name)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should delete test CRD", func() {
		err = testutil.DeleteCRD(extClient, testcrd.Name)
		Expect(err).NotTo(HaveOccurred())
	})
})

func validateComponentOwnerReferences(kubeClient client.Client, list *unstructured.UnstructuredList, matchedingLabels map[string]string, ownerName string) {
	err := wait.PollImmediate(pullPeriod, waitTimeout, func() (bool, error) {

		log.Println("Pulling the component with Kind = ", list.GetKind())
		if err := kubeClient.List(context.TODO(), list, client.InNamespace(metav1.NamespaceDefault), client.MatchingLabels(matchedingLabels)); err != nil {
			return false, nil
		}

		for _, item := range list.Items {
			if item.GetOwnerReferences() == nil || len(item.GetOwnerReferences()) < 1 || item.GetOwnerReferences()[0].Name != ownerName {
				log.Println("Component ownerReferences has NOT been updated yet")
				return false, nil
			}
		}
		log.Println("Component ownerReferences has been updated successfully")
		return true, nil
	})
	Expect(err).NotTo(HaveOccurred())
}

func waitForApplicationStatusToHaveNComponents(kubeClient client.Client, key client.ObjectKey, app *appv1beta1.Application, ready int, total int) {
	err := wait.PollImmediate(pullPeriod, waitTimeout, func() (bool, error) {
		log.Println("Pulling the application status")
		if err := kubeClient.Get(context.TODO(), key, app); err != nil {
			return false, nil
		}

		if app.Status.ComponentList.Objects != nil {
			if len(app.Status.ComponentList.Objects) == total && app.Status.Conditions != nil && app.Status.ComponentsReady == fmt.Sprintf("%d/%d", ready, total) {
				log.Println("Application status has been updated successfully")
				return true, nil
			}
			log.Printf("Application status ready components: %s", app.Status.ComponentsReady)
			return false, nil
		}
		log.Println("Application status has NOT been updated yet")
		return false, nil
	})
	Expect(err).NotTo(HaveOccurred())
}

func applyKustomize(path string) error {
	var err error
	var kubectlOP bytes.Buffer
	var kubectlError bytes.Buffer
	var kustError bytes.Buffer

	kustomize := exec.Command("../hack/tools/bin/kustomize", "build", path)
	kubectl := exec.Command("../hack/tools/bin/kubectl", "apply", "-f", "-")

	r, w := io.Pipe()
	kustomize.Stdout = w
	kustomize.Stderr = &kustError
	kubectl.Stdin = r
	kubectl.Stderr = &kubectlError
	kubectl.Stdout = &kubectlOP

	err = kustomize.Start()
	if err != nil {
		return err
	}
	err = kubectl.Start()
	if err != nil {
		return err
	}
	err = kustomize.Wait()
	if err != nil {
		_, _ = io.Copy(os.Stdout, &kustError)
		return err
	}
	w.Close()
	err = kubectl.Wait()
	if err != nil {
		_, _ = io.Copy(os.Stdout, &kubectlError)
		return err
	}
	_, _ = io.Copy(os.Stdout, &kubectlOP)

	return nil
}

func hasConditionTypeStatusAndReason(conditions []appv1beta1.Condition, t appv1beta1.ConditionType, s corev1.ConditionStatus, r string) bool {
	for _, condition := range conditions {
		if condition.Type == t && condition.Status == s && condition.Reason == r {
			return true
		}
	}
	return false
}
