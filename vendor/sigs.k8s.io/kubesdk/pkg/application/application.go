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

package application

import (
	app "github.com/kubernetes-sigs/application/pkg/apis/app/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kubesdk/pkg/component"
	"sigs.k8s.io/kubesdk/pkg/resource"
)

// Application obj to attach methods
type Application struct {
	app.Application
}

// SetSelector attaches selectors to Application object
func (a *Application) SetSelector(labels map[string]string) *Application {
	a.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
	return a
}

// SetName sets name
func (a *Application) SetName(value string) *Application {
	a.ObjectMeta.Name = value
	return a
}

// SetNamespace asets namespace
func (a *Application) SetNamespace(value string) *Application {
	a.ObjectMeta.Namespace = value
	return a
}

// AddLabels adds more labels
func (a *Application) AddLabels(value component.KVMap) *Application {
	value.Merge(a.ObjectMeta.Labels)
	a.ObjectMeta.Labels = value
	return a
}

// Observable returns resource object
func (a *Application) Observable() *resource.Observable {
	return &resource.Observable{
		Obj:     &a.Application,
		ObjList: &app.ApplicationList{},
		Labels:  a.GetLabels(),
	}
}

// Object returns resource object
func (a *Application) Object() *resource.Object {
	return &resource.Object{
		Lifecycle: resource.LifecycleManaged,
		Obj:       &a.Application,
		ObjList:   &app.ApplicationList{},
	}
}

// AddToScheme return AddToScheme of application crd
func AddToScheme(sb *runtime.SchemeBuilder) {
	*sb = append(*sb, app.AddToScheme)
}

// SetComponentGK attaches component GK to Application object
func (a *Application) SetComponentGK(bag *resource.ObjectBag) *Application {
	a.Spec.ComponentGroupKinds = []metav1.GroupKind{}
	gkmap := map[schema.GroupKind]struct{}{}
	for _, obj := range bag.Items() {
		if obj.ObjList != nil {
			ro := obj.Obj.(runtime.Object)
			gk := ro.GetObjectKind().GroupVersionKind().GroupKind()
			if _, ok := gkmap[gk]; !ok {
				gkmap[gk] = struct{}{}
				mgk := metav1.GroupKind{
					Group: gk.Group,
					Kind:  gk.Kind,
				}
				a.Spec.ComponentGroupKinds = append(a.Spec.ComponentGroupKinds, mgk)
			}
		}
	}
	return a
}
