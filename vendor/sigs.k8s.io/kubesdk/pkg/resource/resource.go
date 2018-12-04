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

package resource

import (
	"bufio"
	"bytes"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"reflect"
	"strings"
	"text/template"
)

// objFromReader reads Object from []byte spec
func objFromReader(b *bufio.Reader, data interface{}, list metav1.ListInterface) (*Object, error) {
	var exdoc bytes.Buffer
	r := yaml.NewYAMLReader(b)
	doc, err := r.Read()
	if err == nil {
		tmpl, e := template.New("tmpl").Parse(string(doc))
		err = e
		if err == nil {
			err = tmpl.Execute(&exdoc, data)
			if err == nil {
				d := scheme.Codecs.UniversalDeserializer()
				obj, _, e := d.Decode(exdoc.Bytes(), nil, nil)
				err = e
				if err == nil {
					return &Object{
						Obj:       obj.DeepCopyObject().(metav1.Object),
						ObjList:   list,
						Lifecycle: LifecycleManaged,
					}, nil
				}
			}
		}
	}
	return nil, err
}

// ObjFromString populates Object from string spec
func ObjFromString(spec string, values interface{}, list metav1.ListInterface) (*Object, error) {
	return objFromReader(bufio.NewReader(strings.NewReader(spec)), values, list)
}

// ObjFromFile populates Object from file
func ObjFromFile(path string, values interface{}, list metav1.ListInterface) (*Object, error) {
	f, err := os.Open(path)
	if err == nil {
		return objFromReader(bufio.NewReader(f), values, list)
	}
	return nil, err
}

// ObservablesFromObjects returns ObservablesFromObjects
func ObservablesFromObjects(scheme *runtime.Scheme, bag *ObjectBag, labels map[string]string) []Observable {
	var gk schema.GroupKind
	var observables []Observable
	gkmap := map[schema.GroupKind]struct{}{}
	for _, obj := range bag.Items() {
		if obj.ObjList != nil {
			ro := obj.Obj.(runtime.Object)
			kinds, _, err := scheme.ObjectKinds(ro)
			if err == nil {
				// Expect only 1 kind.  If there is more than one kind this is probably an edge case such as ListOptions.
				if len(kinds) != 1 {
					err = fmt.Errorf("Expected exactly 1 kind for Object %T, but found %s kinds", ro, kinds)

				}
			}
			// Cache the Group and Kind for the OwnerType
			if err == nil {
				gk = schema.GroupKind{Group: kinds[0].Group, Kind: kinds[0].Kind}
			} else {
				gk = ro.GetObjectKind().GroupVersionKind().GroupKind()
			}
			if _, ok := gkmap[gk]; !ok {
				gkmap[gk] = struct{}{}
				observable := Observable{
					ObjList: obj.ObjList,
					Labels:  labels,
				}
				observables = append(observables, observable)
			}
		} else {
			observable := Observable{
				Obj: obj.Obj,
			}
			observables = append(observables, observable)
		}
	}
	return observables
}

// Add adds to the Object bag
func (b *ObjectBag) Add(objs ...Object) {
	b.objects = append(b.objects, objs...)
}

// Items get items from the Object bag
func (b *ObjectBag) Items() []Object {
	return b.objects
}

// Get returns an item which matched the kind and name
func (b *ObjectBag) Get(inobj metav1.Object) metav1.Object {
	for _, obj := range b.Items() {
		otype := reflect.TypeOf(obj.Obj).String()
		intype := reflect.TypeOf(inobj).String()
		if otype == intype && obj.Obj.GetName() == inobj.GetName() && obj.Obj.GetNamespace() == inobj.GetNamespace() {
			return obj.Obj
		}
	}
	return nil
}

// Validate validates the LocalObjectReference
func (s *LocalObjectReference) Validate(fp *field.Path, sfield string, errs field.ErrorList, required bool) field.ErrorList {
	fp = fp.Child(sfield)
	if s == nil {
		if required {
			errs = append(errs, field.Required(fp, "Required "+sfield+" missing"))
		}
		return errs
	}

	if s.Name == "" {
		errs = append(errs, field.Required(fp.Child("name"), "name is required"))
	}
	return errs
}
