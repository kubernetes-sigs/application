// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"fmt"
	"io"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	applicationsv1beta1 "sigs.k8s.io/application/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateApplication -  create application object
func CreateApplication(kubeClient client.Client, ns string, relativePath string) error {
	app, err := parseApplicationYaml(relativePath)
	if err != nil {
		return err
	}
	app.Namespace = ns

	object := &applicationsv1beta1.Application{}
	objectKey := types.NamespacedName{
		Namespace: ns,
		Name:      app.Name,
	}
	err = kubeClient.Get(context.TODO(), objectKey, object)

	if err == nil {
		// Application already exists -> Update
		err = kubeClient.Update(context.TODO(), app)
		if err != nil {
			return err
		}
	} else {
		// Application doesn't exist -> Create
		fmt.Printf("Creating new Application\n")
		err = kubeClient.Create(context.TODO(), app)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteApplication - delete application object
func DeleteApplication(kubeClient client.Client, ns string, relativePath string) error {
	app, err := parseApplicationYaml(relativePath)
	if err != nil {
		return err
	}

	object := &applicationsv1beta1.Application{}
	objectKey := types.NamespacedName{
		Namespace: ns,
		Name:      app.Name,
	}
	err = kubeClient.Get(context.TODO(), objectKey, object)
	if err != nil {
		return err
	}

	return kubeClient.Delete(context.TODO(), object)
}

func parseApplicationYaml(relativePath string) (*applicationsv1beta1.Application, error) {
	var manifest *os.File
	var err error

	var app applicationsv1beta1.Application
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

		if out.GetKind() == "Application" {
			var marshaled []byte
			marshaled, err = out.MarshalJSON()
			_ = json.Unmarshal(marshaled, &app)
			break
		}
	}

	if err != io.EOF && err != nil {
		return nil, err
	}
	return &app, nil
}
