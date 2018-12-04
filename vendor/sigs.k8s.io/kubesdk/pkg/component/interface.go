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

package component

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kubesdk/pkg/resource"
)

// Handle is an interface for operating on logical Components of a CR
type Handle interface {
	ExpectedResources(rsrc interface{}, labels map[string]string, aggregated *resource.ObjectBag) (*resource.ObjectBag, error)
	Observables(scheme *runtime.Scheme, rsrc interface{}, labels map[string]string, expected *resource.ObjectBag) []resource.Observable
	Mutate(rsrc interface{}, status interface{}, expected, observed *resource.ObjectBag) (*resource.ObjectBag, error)
	Differs(expected metav1.Object, observed metav1.Object) bool
	UpdateComponentStatus(rsrc, status interface{}, reconciled []metav1.Object, err error)
	Finalize(rsrc, status interface{}, observed *resource.ObjectBag) error
}
