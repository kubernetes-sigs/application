// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"io"
	"os"
	"time"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// CreateOrUpdateCRD - create or update application CRD
func CreateOrUpdateCRD(kubeClient apiextcs.Interface, crd *apiextensions.CustomResourceDefinition) error {

	currentCrd, err := kubeClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), crd.Name, metav1.GetOptions{})

	if err == nil {
		// CustomResourceDefinition already exists -> Update

		// Bypass the 'metadata.resourceVersion: Invalid value: 0x0: must be specified for an update' error
		crd.ResourceVersion = currentCrd.ResourceVersion
		_, err = kubeClient.ApiextensionsV1().CustomResourceDefinitions().Update(context.TODO(), crd, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

	} else {
		// CustomResourceDefinition doesn't exist -> Create
		_, err = kubeClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// WaitForCRDOrDie - wait for CRD conditions to be set
func WaitForCRDOrDie(kubeClient apiextcs.Interface, name string) error {
	err := wait.PollImmediate(2*time.Second, 20*time.Second, func() (bool, error) {
		crd, err := kubeClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return establishedCondition(crd.Status.Conditions), nil
	})
	return err
}

func establishedCondition(conditions []apiextensions.CustomResourceDefinitionCondition) bool {
	for _, condition := range conditions {
		if condition.Type == apiextensions.Established && condition.Status == apiextensions.ConditionTrue {
			return true
		}
	}
	return false
}

// DeleteCRD - Delete CRD from cluster
func DeleteCRD(kubeClient apiextcs.Interface, crdName string) error {
	err := kubeClient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.TODO(), crdName, metav1.DeleteOptions{})
	return err
}

// ParseCRDYaml - load crd from file
func ParseCRDYaml(relativePath string) (*apiextensions.CustomResourceDefinition, error) {
	var manifest *os.File
	var err error

	var crd apiextensions.CustomResourceDefinition
	if manifest, err = PathToOSFile(relativePath); err != nil {
		return nil, err
	}

	decoder := yaml.NewYAMLOrJSONDecoder(manifest, 100)
	for {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err != nil {
			// this would indicate it's malformed YAML.
			break
		}

		if out.GetKind() == "CustomResourceDefinition" {
			var marshaled []byte
			marshaled, err = out.MarshalJSON()
			_ = json.Unmarshal(marshaled, &crd)
			break
		}
	}

	if err != io.EOF && err != nil {
		return nil, err
	}
	return &crd, nil
}
