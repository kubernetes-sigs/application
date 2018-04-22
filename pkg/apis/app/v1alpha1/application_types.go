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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplicationSpec defines the specification for an Application.
type ApplicationSpec struct {
	// Type is the type of the application (e.g. WordPress, MySQL, Cassandra).
	Type string `json:"type,omitempty"`

	// ComponentGroupKinds is a list of Kinds for Application's components (e.g. Deployments, Pods, Services, CRDs). It
	// can be used in conjunction with the Application's Selector to list or watch the Applications components.
	ComponentGroupKinds []metav1.GroupKind `json:"componentKinds,omitempty"`

	// Selector is a label query over kinds that created by the application. It must match the component objects' labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Version is an optional version indicator for the Application.
	Version string `json:"version,omitempty"`

	// Description is a brief string description of the Application.
	Description string `json:"description,omitempty"`

	// Maintainers is an optional list of maintainers of the application. The maintainers in this list maintain the
	// the source code, images, and package for the application.
	Maintainers []Maintainer `json:"maintainers,omitempty"`

	// Owners is an optional list of the owners of the installed application. The owners of the application should be
	// contacted in the event of a planned or unplanned disruption affecting the application.
	Owners []string `json:"owners,omitempty"`

	// Keywords is an optional list of key words associated with the application (e.g. MySQL, RDBMS, database).
	Keywords []string `json:"keywords,omitempty"`

	// Info contains human readable key,value pairs for the Application.
	Info []InfoItem `json:"info,omitempty"`

	// Links are a list of descriptive URLs intended to be used to surface additional documentation, dashboards, etc.
	Links []Link `json:"links,omitempty"`

	// Notes contain a human readable snippets intended as a quick start for the users of the Application.
	Notes string `json:"notes,omitempty"`

	// AssemblyPhase represents the current phase of the application's assembly.
	// An empty value is equivalent to "Succeeded".
	AssemblyPhase ApplicationAssemblyPhase `json:"assemblyPhase,omitempty"`
}

// ApplicationStatus defines controllers the observed state of Application
type ApplicationStatus struct {
	// ObservedGeneration is used by the Application Controller to report the last Generation of an Application
	// that it has observed.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// Maintainer contains information about an individual or organization that maintains the source code, images, and
// package for an Application. An Application can have more than one maintainer.
type Maintainer struct {
	// Name is the descriptive name of the maintainer.
	Name string `json:"name,omitempty"`

	// Url could typically be a website address.
	Url string `json:"url,omitempty"`

	// Email is the email address.
	Email string `json:"email,omitempty"`
}

// Link contains information about an URL to surface documentation, dashboards, etc.
type Link struct {
	// Description is human readable content explaining the purpose of the link.
	Description string `json:"description,omitempty"`

	// Url typically points at a website address.
	Url string `json:"url,omitempty"`
}

// InfoItem is a human readable key,value pair containing important information about how to access the Application.
type InfoItem struct {
	// Name is a human readable title for this piece of information.
	Name string `json:"name,omitempty"`

	// Value is human readable content.
	Value string `json:"value,omitempty"`
}

type ApplicationAssemblyPhase string

const (
	// Used to indicate that not all of application's components
	// have been deployed yet.
	Pending ApplicationAssemblyPhase = "Pending"
	// Used to indicate that all of application's components
	// have alraedy been deployed.
	Succeeded = "Succeeded"
	// Used to indicate that deployment of application's components
	// failed. Some components might be present, but deployment of
	// the remaining ones will not be re-attempted.
	Failed = "Failed"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Application
// +k8s:openapi-gen=true
// +resource:path=applications
// The Application object acts as an aggregator for components that comprise an Application. Its
// Spec.ComponentGroupKinds indicate the GroupKinds of the components the comprise the Application. Its Spec. Selector
// is used to list and watch those components. All components of an Application should be labeled such the Application's
// Spec. Selector matches.
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// The specification object for the Application.
	Spec ApplicationSpec `json:"spec,omitempty"`
	// The status object for the Application.
	Status ApplicationStatus `json:"status,omitempty"`
}
