/*
Copyright 2018 The Kubernetes Authors
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

package v1beta1

import (
	appsv1 "k8s.io/api/apps/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Constants defining labels
const (
	StatusReady      = "Ready"
	StatusInProgress = "InProgress"
	StatusDisabled   = "Disabled"
)

func (s *ObjectStatus) update(rsrc metav1.Object) {
	ro := rsrc.(runtime.Object)
	gvk := ro.GetObjectKind().GroupVersionKind()
	s.Link = rsrc.GetSelfLink()
	s.Name = rsrc.GetName()
	s.Group = gvk.GroupVersion().String()
	s.Kind = gvk.GroupKind().Kind
	s.Status = StatusReady
}

// ResetComponentList - reset component list objects
func (m *ApplicationStatus) ResetComponentList() {
	m.ComponentList.Objects = []ObjectStatus{}
}

// UpdateStatus the component status
func (m *ApplicationStatus) UpdateStatus(rsrcs []metav1.Object, err error) {
	var ready = true
	for _, r := range rsrcs {
		os := ObjectStatus{}
		os.update(r)
		switch r.(type) {
		case *appsv1.StatefulSet:
			os.Status = stsStatus(r.(*appsv1.StatefulSet))
		case *policyv1.PodDisruptionBudget:
			os.Status = pdbStatus(r.(*policyv1.PodDisruptionBudget))
		}
		m.ComponentList.Objects = append(m.ComponentList.Objects, os)
	}
	for _, os := range m.ComponentList.Objects {
		if os.Status != StatusReady {
			ready = false
		}
	}

	if ready {
		m.Ready("ComponentsReady", "all components ready")
	} else {
		m.NotReady("ComponentsNotReady", "some components not ready")
	}
	if err != nil {
		m.SetCondition(Error, "ErrorSeen", err.Error())
	}
}

// Resource specific logic -----------------------------------

// Statefulset
func stsStatus(rsrc *appsv1.StatefulSet) string {
	if rsrc.Status.ReadyReplicas == *rsrc.Spec.Replicas && rsrc.Status.CurrentReplicas == *rsrc.Spec.Replicas {
		return StatusReady
	}
	return StatusInProgress
}

// PodDisruptionBudget
func pdbStatus(rsrc *policyv1.PodDisruptionBudget) string {
	if rsrc.Status.CurrentHealthy >= rsrc.Status.DesiredHealthy {
		return StatusReady
	}
	return StatusInProgress
}
