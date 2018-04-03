


package application_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

	"github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    . "github.com/kubernetes-sigs/application/pkg/apis/app/v1alpha1"
    . "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned/typed/app/v1alpha1"
)

// EDIT THIS FILE!
// Created by "kubebuilder create resource" for you to implement controller logic tests

var _ = Describe("Application controller", func() {
    var instance Application
    var expectedKey types.ReconcileKey
    var client ApplicationInterface

    BeforeEach(func() {
        instance = Application{}
        instance.Name = "instance-1"
        expectedKey = types.ReconcileKey{
            Namespace: "default",
            Name: "instance-1",
        }
    })

    AfterEach(func() {
        client.Delete(instance.Name, &metav1.DeleteOptions{})
    })

    Describe("when creating a new object", func() {
        It("invoke the reconcile method", func() {
            after := make(chan struct{})
            ctrl.AfterReconcile = func(key types.ReconcileKey, err error) {
                defer func() {
                    // Recover in case the key is reconciled multiple times
                    defer func() { recover() }()
                    close(after)
                }()
                defer GinkgoRecover()
                Expect(key).To(Equal(expectedKey))
                Expect(err).ToNot(HaveOccurred())
            }

            // Create the instance
            client = cs.AppV1alpha1().Applications("default")
            _, err := client.Create(&instance)
            Expect(err).ShouldNot(HaveOccurred())

            // Wait for reconcile to happen
            Eventually(after, "10s", "100ms").Should(BeClosed())

            // INSERT YOUR CODE HERE - test conditions post reconcile
        })
    })
})
