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

package application

import (
	"log"

	"github.com/najena/kubebuilder/pkg/builders"

	"github.com/kubernetes-sigs/apps_application/pkg/apis/application/v1alpha1"
	listers "github.com/kubernetes-sigs/apps_application/pkg/client/listers_generated/application/v1alpha1"
	"github.com/kubernetes-sigs/apps_application/pkg/controller/sharedinformers"
)

// EDIT THIS FILE!
// Created by "kubebuilder create resource" for you to implement controller logic for the Application resource API

// Reconcile handles enqueued messages
func (c *ApplicationControllerImpl) Reconcile(u *v1alpha1.Application) error {
	// INSERT YOUR CODE HERE - implement controller logic to reconcile observed and desired state of the object
	log.Printf("Running reconcile Application for %s\n", u.Name)
	return nil
}

// +controller:group=application,version=v1alpha1,kind=Application,resource=applications
type ApplicationControllerImpl struct {
	builders.DefaultControllerFns

	// lister indexes properties about Application
	lister listers.ApplicationLister
}

// Init initializes the controller and is called by the generated code
// Register watches for additional resource types here.
func (c *ApplicationControllerImpl) Init(arguments sharedinformers.ControllerInitArguments) {
	// INSERT YOUR CODE HERE - add logic for initializing the controller as needed

	// Use the lister for indexing applications labels
	c.lister = arguments.GetSharedInformers().Factory.Application().V1alpha1().Applications().Lister()

	// To watch other resource types, uncomment this function and replace Foo with the resource name to watch.
	// Must define the func FooToApplication(i interface{}) (string, error) {} that returns the Application
	// "namespace/name"" to reconcile in response to the updated Foo
	// Note: To watch Kubernetes resources, you must also update the StartAdditionalInformers function in
	// pkg/controllers/sharedinformers/informers.go
	//
	// arguments.Watch("ApplicationFoo",
	//     arguments.GetSharedInformers().Factory.Bar().V1beta1().Bars().Informer(),
	//     c.FooToApplication)
}

func (c *ApplicationControllerImpl) Get(namespace, name string) (*v1alpha1.Application, error) {
	return c.lister.Applications(namespace).Get(name)
}
