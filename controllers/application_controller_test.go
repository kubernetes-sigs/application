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

package controllers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	appv1beta1 "sigs.k8s.io/application/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

const timeout = time.Second * 30

var _ = Describe("Application Reconciler", func() {
	var stopMgr chan struct{}
	var mgrStopped *sync.WaitGroup
	var recFn reconcile.Reconciler
	var requests chan reconcile.Request
	var ctx context.Context
	var applicationReconciler *ApplicationReconciler
	var labelSet1 = map[string]string{"foo": "bar"}
	var labelSet2 = map[string]string{"baz": "qux"}
	var namespace1 = metav1.NamespaceDefault
	var namespace2 = "default2"
	var deployment *apps.Deployment
	var statefulSet *apps.StatefulSet
	var service *core.Service

	BeforeEach(func() {
		// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
		// channel when it is finished.
		mgr, err := manager.New(cfg, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		c = mgr.GetClient()

		applicationReconciler = NewReconciler(mgr)
		logger := applicationReconciler.Log.WithValues("application", metav1.NamespaceDefault+"/application")
		ctx = context.WithValue(context.Background(), loggerCtxKey, logger)
		recFn, requests = SetupTestReconcile(applicationReconciler)
		Expect(CreateController("app", mgr, recFn)).NotTo(HaveOccurred())

		stopMgr, mgrStopped = StartTestManager(mgr)
	})

	AfterEach(func() {
		close(stopMgr)
		mgrStopped.Wait()
	})

	Describe("fetchComponentListResources", func() {
		It("should fetch corresponding components with matched labels within a namespace", func() {
			var objs []runtime.Object = nil
			deployment = createDeployment(labelSet1, namespace1)
			service = createService(labelSet1, namespace1)
			statefulSet = createStatefulSet(labelSet1, namespace1)
			objs = append(objs, deployment)
			objs = append(objs, service)
			objs = append(objs, statefulSet)
			objs = append(objs, createPod(labelSet2, namespace2))
			objs = append(objs, createDaemonSet(labelSet2, namespace2))
			objs = append(objs, createReplicaSet(labelSet2, namespace2))
			objs = append(objs, CreatePersistentVolumeClaim(labelSet2, namespace2))
			objs = append(objs, createPodDisruptionBudget(labelSet2, namespace2))

			for _, obj := range objs {
				err := c.Create(ctx, obj)
				Expect(err).NotTo(HaveOccurred())
			}

			groupKinds := []metav1.GroupKind{
				{
					Group: "apps",
					Kind:  "StatefulSet",
				},
				{
					Group: "apps",
					Kind:  "Deployment",
				},
				{
					Group: "apps",
					Kind:  "ReplicaSet",
				},
				{
					Group: "apps",
					Kind:  "DaemonSet",
				},
				{
					Group: "batch",
					Kind:  "Job",
				},
				{
					Group: "", // NOTICE: The group of Service should be empty, instead of "core"
					Kind:  "Service",
				},
				{
					Group: "", // NOTICE: The group of PVC should be empty, instead of "core"
					Kind:  "PersistentVolumeClaim",
				},
				{
					Group: "", // NOTICE: The group of Pod should be empty, instead of "core"
					Kind:  "Pod",
				},
				{
					Group: "policy",
					Kind:  "PodDisruptionBudget",
				},
			}

			ns1List, err := applicationReconciler.fetchComponentListResources(ctx, groupKinds, metav1.SetAsLabelSelector(labelSet1), namespace1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(ns1List)).To(Equal(3))
			Expect(componentKinds(ns1List)).To(ConsistOf("StatefulSet", "Deployment", "Service"))

			ns2List, err := applicationReconciler.fetchComponentListResources(ctx, groupKinds, metav1.SetAsLabelSelector(labelSet2), namespace2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(ns2List)).To(Equal(5))
			Expect(componentKinds(ns2List)).To(ConsistOf("ReplicaSet", "DaemonSet", "PersistentVolumeClaim", "Pod", "PodDisruptionBudget"))

		})

		It("should fetch components when version is included in the group", func() {
			groupKinds := []metav1.GroupKind{
				{
					Group: "apps/v1",
					Kind:  "Deployment",
				},
				{
					Group: "/v1",
					Kind:  "Service",
				},
			}

			ns1List, err := applicationReconciler.fetchComponentListResources(ctx, groupKinds, metav1.SetAsLabelSelector(labelSet1), metav1.NamespaceDefault)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(ns1List)).To(Equal(2))
			Expect(componentKinds(ns1List)).To(ConsistOf("Deployment", "Service"))

		})
	})

	Describe("setOwnerRefForResources", func() {
		var resource = &unstructured.Unstructured{}
		resource.SetGroupVersionKind(schema.GroupVersionKind{
			Group: "apps",
			Kind:  "StatefulSet",
		})
		var key types.NamespacedName
		var resources []*unstructured.Unstructured
		var uid types.UID = "old-uid"
		var newUid types.UID = "new-uid"
		var ownerRef = metav1.OwnerReference{
			APIVersion: "app.k8s.io/v1beta1",
			Kind:       "Application",
			Name:       "application-foo",
			UID:        uid,
		}

		It("should append new ownerReference to the resources", func() {
			key = types.NamespacedName{
				Name:      statefulSet.Name,
				Namespace: metav1.NamespaceDefault,
			}
			resources = append(resources, resource)

			err := c.Get(ctx, key, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(BeEmpty())

			err = applicationReconciler.setOwnerRefForResources(ctx, ownerRef, resources)
			Expect(err).NotTo(HaveOccurred())
			err = c.Get(ctx, key, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(HaveLen(1))
			Expect(resource.GetOwnerReferences()).To(ContainElement(ownerRef))
		})

		It("should update existing ownerReference with new UID", func() {
			err := c.Get(ctx, key, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(HaveLen(1))
			Expect(resource.GetOwnerReferences()[0].UID).To(Equal(uid))

			ownerRef.UID = newUid
			err = applicationReconciler.setOwnerRefForResources(ctx, ownerRef, resources)
			Expect(err).NotTo(HaveOccurred())
			err = c.Get(ctx, key, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(HaveLen(1))
			Expect(resource.GetOwnerReferences()[0].UID).To(Equal(newUid))
		})

		It("should NOT update identical ownerReference", func() {
			err := c.Get(ctx, key, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(HaveLen(1))
			Expect(resource.GetOwnerReferences()[0].UID).To(Equal(newUid))

			err = applicationReconciler.setOwnerRefForResources(ctx, ownerRef, resources)
			Expect(err).NotTo(HaveOccurred())
			err = c.Get(ctx, key, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.GetOwnerReferences()).To(HaveLen(1))
			Expect(resource.GetOwnerReferences()[0].UID).To(Equal(newUid))
		})
	})

	Describe("Application Reconciler", func() {

		It("should receive a request when an application instance is created", func() {
			instance := &appv1beta1.Application{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"}, Spec: appv1beta1.ApplicationSpec{}}

			// Create the Application object and expect the Reconcile and Deployment to be created
			err := c.Create(ctx, instance)
			// The instance object may not be a valid object because it might be missing some required fields.
			// Please modify the instance object by adding required fields and then remove the following if statement.
			if apierrors.IsInvalid(err) {
				fmt.Printf("failed to create object, got an invalid object error: %v\n", err)
				return
			}
			Expect(err).NotTo(HaveOccurred())
			defer c.Delete(ctx, instance)
			var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
		})

		It("should update the application status, as well as the components' ownerReference", func() {
			application := &appv1beta1.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "application-01",
					Namespace: metav1.NamespaceDefault,
					Labels:    labelSet1,
				},
				Spec: appv1beta1.ApplicationSpec{
					Selector: &metav1.LabelSelector{MatchLabels: labelSet1},
					ComponentGroupKinds: []metav1.GroupKind{
						{
							Group: "apps",
							Kind:  "Deployment",
						},
						{
							Group: "", // NOTICE: The group of Service should be empty, instead of "core"
							Kind:  "Service",
						},
					},
					AddOwnerRef: true,
				}}

			Expect(deployment.OwnerReferences).To(BeNil())
			Expect(service.OwnerReferences).To(BeNil())

			err := c.Create(ctx, application)
			Expect(err).NotTo(HaveOccurred())
			waitForComponentsAddedToStatus(ctx, application, deployment.Name, service.Name)

			wait.PollImmediate(time.Second, timeout, func() (bool, error) {
				fetchUpdatedDeployment(ctx, deployment)
				fetchUpdatedService(ctx, service)
				if len(deployment.OwnerReferences) == 1 && len(service.OwnerReferences) == 1 {
					return true, nil
				}
				return false, nil
			})

			Expect(deployment.OwnerReferences[0].Name).To(Equal(application.Name))
			Expect(service.OwnerReferences[0].Name).To(Equal(application.Name))

		})
	})

})

func fetchUpdatedDeployment(ctx context.Context, deployment *apps.Deployment) {
	key := types.NamespacedName{
		Name:      deployment.Name,
		Namespace: deployment.Namespace,
	}
	err := c.Get(ctx, key, deployment)
	Expect(err).NotTo(HaveOccurred())
}

func fetchUpdatedService(ctx context.Context, service *core.Service) {
	key := types.NamespacedName{
		Name:      service.Name,
		Namespace: service.Namespace,
	}
	err := c.Get(ctx, key, service)
	Expect(err).NotTo(HaveOccurred())
}

func waitForComponentsAddedToStatus(ctx context.Context, app *appv1beta1.Application, expectedNames ...string) {
	key := types.NamespacedName{
		Name:      app.Name,
		Namespace: app.Namespace,
	}
	wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		names, err := applicationStatusComponentNames(ctx, app, key)
		if err != nil {
			return false, err
		}
		if len(names) < len(expectedNames) {
			return false, nil
		}
		Expect(names).Should(ConsistOf(expectedNames))
		return true, nil
	})
}

func applicationStatusComponentNames(ctx context.Context, app *appv1beta1.Application, key types.NamespacedName) ([]string, error) {
	var names = make([]string, 0)
	if err := c.Get(ctx, key, app); err != nil {
		return names, err
	}
	Expect(app.Status.ComponentList).NotTo(BeNil())
	for _, component := range app.Status.ComponentList.Objects {
		names = append(names, component.Name)
	}
	return names, nil
}

func componentKinds(list []*unstructured.Unstructured) []string {
	var kinds []string = nil
	for _, l := range list {
		kinds = append(kinds, l.GetKind())
	}
	return kinds
}

func objectMeta(t string, labels map[string]string, ns string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s", t, uuid.New()),
		Namespace: ns,
		Labels:    labels,
	}
}

func podTemplateSpec(labels map[string]string, ns string) core.PodTemplateSpec {
	return core.PodTemplateSpec{
		ObjectMeta: objectMeta("pod-template", labels, ns),
		Spec: core.PodSpec{
			RestartPolicy: core.RestartPolicyAlways,
			DNSPolicy:     core.DNSClusterFirst,
			Containers:    []core.Container{{Name: "abc", Image: "image", ImagePullPolicy: "IfNotPresent"}},
		},
	}
}

func createStatefulSet(labels map[string]string, ns string) *apps.StatefulSet {
	podLabels := map[string]string{"xxx": "yyy"}

	return &apps.StatefulSet{
		ObjectMeta: objectMeta("statefulset", labels, ns),
		Spec: apps.StatefulSetSpec{
			PodManagementPolicy: apps.OrderedReadyPodManagement,
			Selector:            &metav1.LabelSelector{MatchLabels: podLabels},
			Template:            podTemplateSpec(podLabels, ns),
			UpdateStrategy:      apps.StatefulSetUpdateStrategy{Type: apps.RollingUpdateStatefulSetStrategyType},
		},
	}
}

func createDeployment(labels map[string]string, ns string) *apps.Deployment {
	podLabels := map[string]string{"xxx": "yyy"}
	return &apps.Deployment{
		ObjectMeta: objectMeta("deployment", labels, ns),
		Spec: apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: podTemplateSpec(podLabels, ns),
		},
	}
}

func createDaemonSet(labels map[string]string, ns string) *apps.DaemonSet {
	return &apps.DaemonSet{
		ObjectMeta: objectMeta("daemonset", labels, ns),
		Spec: apps.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: podTemplateSpec(labels, ns),
			UpdateStrategy: apps.DaemonSetUpdateStrategy{
				Type: apps.OnDeleteDaemonSetStrategyType,
			},
		},
	}
}

func createReplicaSet(labels map[string]string, ns string) *apps.ReplicaSet {
	return &apps.ReplicaSet{
		ObjectMeta: objectMeta("replicaset", labels, ns),
		Spec: apps.ReplicaSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: podTemplateSpec(labels, ns),
		},
	}
}

func CreatePersistentVolumeClaim(labels map[string]string, ns string) *core.PersistentVolumeClaim {
	return &core.PersistentVolumeClaim{
		ObjectMeta: objectMeta("pvc", labels, ns),
		Spec: core.PersistentVolumeClaimSpec{
			Selector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "key2",
						Operator: "Exists",
					},
				},
			},
			AccessModes: []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
				core.ReadOnlyMany,
			},
			Resources: core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceStorage: resource.MustParse("10G"),
				},
			},
		},
	}
}

func createPod(labels map[string]string, ns string) *core.Pod {
	return &core.Pod{
		ObjectMeta: objectMeta("pod", labels, ns),
		Spec: core.PodSpec{
			Volumes:       []core.Volume{{Name: "vol", VolumeSource: core.VolumeSource{EmptyDir: &core.EmptyDirVolumeSource{}}}},
			Containers:    []core.Container{{Name: "ctr", Image: "image", ImagePullPolicy: "IfNotPresent", TerminationMessagePolicy: "File"}},
			RestartPolicy: core.RestartPolicyAlways,
			DNSPolicy:     core.DNSClusterFirst,
		},
	}
}

func createPodDisruptionBudget(labels map[string]string, ns string) *policy.PodDisruptionBudget {
	maxUnavailable := intstr.FromString("10%")
	return &policy.PodDisruptionBudget{
		ObjectMeta: objectMeta("pdb", labels, ns),
		Spec: policy.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailable,
		},
	}
}

func createService(labels map[string]string, ns string) *core.Service {
	serviceIPFamily := core.IPv4Protocol
	return &core.Service{
		ObjectMeta: objectMeta("service", labels, ns),
		Spec: core.ServiceSpec{
			SessionAffinity: "None",
			Type:            core.ServiceTypeClusterIP,
			Ports:           []core.ServicePort{{Name: "p", Protocol: "TCP", Port: 8675, TargetPort: intstr.FromInt(8675)}},
			IPFamily:        &serviceIPFamily,
		},
	}
}
