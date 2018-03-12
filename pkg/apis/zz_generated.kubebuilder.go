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
package apis

import (
	"github.com/kubernetes-sigs/apps_application/pkg/apis/application"
	applicationv1alpha1 "github.com/kubernetes-sigs/apps_application/pkg/apis/application/v1alpha1"
	"github.com/najena/kubebuilder/pkg/builders"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MetaData struct{}

var APIMeta = MetaData{}

// GetAllApiBuilders returns all known APIGroupBuilders
// so they can be registered with the apiserver
func (MetaData) GetAllApiBuilders() []*builders.APIGroupBuilder {
	return []*builders.APIGroupBuilder{
		GetApplicationAPIBuilder(),
	}
}

// GetCRDs returns all the CRDs for known resource types
func (MetaData) GetCRDs() []v1beta1.CustomResourceDefinition {
	return []v1beta1.CustomResourceDefinition{
		applicationv1alpha1.ApplicationCRD,
	}
}

func (MetaData) GetRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{"application.k8s.io"},
			Resources: []string{"*"},
			Verbs:     []string{"*"},
		},
	}
}

func (MetaData) GetGroupVersions() []schema.GroupVersion {
	return []schema.GroupVersion{
		{
			Group:   "application.k8s.io",
			Version: "v1alpha1",
		},
	}
}

var applicationApiGroup = builders.NewApiGroupBuilder(
	"application.k8s.io",
	"github.com/kubernetes-sigs/apps_application/pkg/apis/application").
	WithUnVersionedApi(application.ApiVersion).
	WithVersionedApis(
		applicationv1alpha1.ApiVersion,
	).
	WithRootScopedKinds()

func GetApplicationAPIBuilder() *builders.APIGroupBuilder {
	return applicationApiGroup
}
