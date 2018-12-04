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
	appv1beta1 "github.com/kubernetes-sigs/application/pkg/apis/app/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	reconciler "sigs.k8s.io/kubesdk/pkg/genericreconciler"
	kbc "sigs.k8s.io/kubesdk/pkg/kbcontroller"
)

// Add creates a new Application Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return kbc.CreateController("airflowcluster", mgr, &appv1beta1.Application{}, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	r := &reconciler.Reconciler{
		Manager: mgr, // why do we need manager ?
		Handle:  &appv1beta1.Application{},
	}
	r.Init()
	return r
}
