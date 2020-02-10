// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kstatus "sigs.k8s.io/kustomize/kstatus/status"
)

// Constants defining labels
const (
	StatusReady      = "Ready"
	StatusInProgress = "InProgress"
	StatusDisabled   = "Disabled"
)

func status(u *unstructured.Unstructured) (string, error) {
	s, err := kstatus.Compute(u)
	if err != nil {
		return "", err
	}
	if s.Status == kstatus.CurrentStatus {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}
