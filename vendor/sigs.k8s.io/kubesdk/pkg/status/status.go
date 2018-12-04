/*
Copyright 2018 Google LLC
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

package status

import (
	//"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

func (s *Statefulset) update(rsrc *appsv1.StatefulSet) string {
	s.Replicas = rsrc.Status.Replicas
	s.ReadyReplicas = rsrc.Status.ReadyReplicas
	s.CurrentReplicas = rsrc.Status.CurrentReplicas
	if rsrc.Status.ReadyReplicas == *rsrc.Spec.Replicas && rsrc.Status.CurrentReplicas == *rsrc.Spec.Replicas {
		return StatusReady
	}
	return StatusInProgress
}

func (s *ObjectStatus) update(rsrc metav1.Object) {
	ro := rsrc.(runtime.Object)
	gvk := ro.GetObjectKind().GroupVersionKind()
	s.Link = rsrc.GetSelfLink()
	s.Name = rsrc.GetName()
	s.Group = gvk.GroupVersion().String()
	s.Kind = gvk.GroupKind().Kind
	s.Status = StatusReady
}

// Pdb is a generic status holder for pdb
func (s *Pdb) update(rsrc *policyv1.PodDisruptionBudget) string {
	s.CurrentHealthy = rsrc.Status.CurrentHealthy
	s.DesiredHealthy = rsrc.Status.DesiredHealthy
	if s.CurrentHealthy >= s.DesiredHealthy {
		return StatusReady
	}
	return StatusInProgress
}

// ResetComponentList - reset component list objects
func (m *Meta) ResetComponentList() {
	m.ComponentList.Objects = []ObjectStatus{}
}

// UpdateStatus the component status
func (m *Meta) UpdateStatus(rsrcs []metav1.Object, err error) {
	var ready = true
	for _, r := range rsrcs {
		os := ObjectStatus{}
		os.update(r)
		switch r.(type) {
		case *appsv1.StatefulSet:
			os.ExtendedStatus.STS = &Statefulset{}
			os.Status = os.ExtendedStatus.STS.update(r.(*appsv1.StatefulSet))
		case *policyv1.PodDisruptionBudget:
			os.ExtendedStatus.PDB = &Pdb{}
			os.Status = os.ExtendedStatus.PDB.update(r.(*policyv1.PodDisruptionBudget))
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
		m.SetCondition(ConditionError, ConditionTrue, "ErrorSeen", err.Error())
	}
}

func (m *Meta) addCondition(ctype ConditionType, status corev1.ConditionStatus, reason, message string) {
	now := metav1.Now()
	c := &Condition{
		Type:               ctype,
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Status:             status,
		Reason:             reason,
		Message:            message,
	}
	//fmt.Printf(" <>>>>> adding ocndition: %s\n", ctype)
	m.Conditions = append(m.Conditions, *c)
}

// SetCondition updates or creates a new condition
func (m *Meta) SetCondition(ctype ConditionType, status corev1.ConditionStatus, reason, message string) {
	var c *Condition
	for i := range m.Conditions {
		if m.Conditions[i].Type == ctype {
			c = &m.Conditions[i]
		}
	}
	if c == nil {
		m.addCondition(ctype, status, reason, message)
	} else {
		// check message ?
		if c.Status == status && c.Reason == reason && c.Message == message {
			return
		}
		now := metav1.Now()
		c.LastUpdateTime = now
		if c.Status != status {
			c.LastTransitionTime = now
		}
		c.Status = status
		c.Reason = reason
		c.Message = message
	}
}

// RemoveCondition removes the condition with the provided type.
func (m *Meta) RemoveCondition(ctype ConditionType) {
	for i, c := range m.Conditions {
		if c.Type == ctype {
			m.Conditions[i] = m.Conditions[len(m.Conditions)-1]
			m.Conditions = m.Conditions[:len(m.Conditions)-1]
			break
		}
	}
}

// GetCondition get existing condition
func (m *Meta) GetCondition(ctype ConditionType) *Condition {
	for i := range m.Conditions {
		if m.Conditions[i].Type == ctype {
			return &m.Conditions[i]
		}
	}
	return nil
}

// IsConditionTrue - if condition is true
func (m *Meta) IsConditionTrue(ctype ConditionType) bool {
	if c := m.GetCondition(ctype); c != nil {
		return c.Status == ConditionTrue
	}
	return false
}

// IsReady returns true if ready condition is set
func (m *Meta) IsReady() bool { return m.IsConditionTrue(ConditionReady) }

// IsNotReady returns true if ready condition is set
func (m *Meta) IsNotReady() bool { return !m.IsConditionTrue(ConditionReady) }

// ConditionReason - return condition reason
func (m *Meta) ConditionReason(ctype ConditionType) string {
	if c := m.GetCondition(ctype); c != nil {
		return c.Reason
	}
	return ""
}

// Ready - shortcut to set ready contition to true
func (m *Meta) Ready(reason, message string) {
	m.SetCondition(ConditionReady, ConditionTrue, reason, message)
}

// NotReady - shortcut to set ready contition to false
func (m *Meta) NotReady(reason, message string) {
	m.SetCondition(ConditionReady, ConditionFalse, reason, message)
}

// SetError - shortcut to set error condition
func (m *Meta) SetError(reason, message string) {
	m.SetCondition(ConditionError, ConditionTrue, reason, message)
}

// ClearError - shortcut to set error condition
func (m *Meta) ClearError() {
	m.SetCondition(ConditionError, ConditionFalse, "NoError", "No error seen")
}

// Settled - shortcut to set Settled contition to true
func (m *Meta) Settled(reason, message string) {
	m.SetCondition(ConditionSettled, ConditionTrue, reason, message)
}

// NotSettled - shortcut to set Settled contition to false
func (m *Meta) NotSettled(reason, message string) {
	m.SetCondition(ConditionSettled, ConditionFalse, reason, message)
}

// EnsureCondition useful for adding default conditions
func (m *Meta) EnsureCondition(ctype ConditionType) {
	if c := m.GetCondition(ctype); c != nil {
		return
	}
	m.addCondition(ctype, ConditionUnknown, ConditionInit, "Not Observed")
}

// EnsureStandardConditions - helper to inject standard conditions
func (m *Meta) EnsureStandardConditions() {
	m.EnsureCondition(ConditionReady)
	m.EnsureCondition(ConditionSettled)
	m.EnsureCondition(ConditionError)
}
