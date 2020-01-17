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

package component_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/application/pkg/component"
)

type Filer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FilerSpec   `json:"spec,omitempty"`
	Status FilerStatus `json:"status,omitempty"`
}

type FilerSpec struct {
	Image    string `json:"image,omitempty"`
	Version  string `json:"version,omitempty"`
	Replicas int    `json:"replicas,omitempty"`
}

type FilerStatus struct {
	Health         string `json:"health,omitempty"`
	ActiveReplicas int    `json:"active,omitempty"`
}

type FilerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Filer `json:"items"`
}

var FilerObject = Filer{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "mlfiler",
		Namespace: "mlwork",
	},
	Spec: FilerSpec{
		Image:    "gcr.io/project/image",
		Version:  "v1.0",
		Replicas: 3,
	},
	Status: FilerStatus{
		Health:         "ok",
		ActiveReplicas: 2,
	},
}

var _ = Describe("Resource", func() {
	var c = component.Component{
		Handle: nil,
		Name:   "base",
		CR:     &FilerObject,
	}

	BeforeEach(func() {
	})

	Describe("Component Labels", func() {
		It("Getting resource labels from component interface works", func(done Done) {
			labels := c.Labels()
			Expect(labels[component.LabelCR]).To(Equal("component_test.Filer"))
			Expect(labels[component.LabelCRName]).To(Equal("mlfiler"))
			Expect(labels[component.LabelCRNamespace]).To(Equal("mlwork"))
			Expect(labels[component.LabelComponent]).To(Equal("base"))
			close(done)
		})
	})
})
