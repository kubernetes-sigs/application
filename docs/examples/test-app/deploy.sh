#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

source ../../../hack/tools/common.sh

# Cleanup on exit
function cleanup() {
  header_text "Cleanup"
  #  kubectl delete -f application.yaml
  #  kubectl delete -f test_crd.yaml
  #  make undeploy
}

# Check if component's ownerReferences is not found
function ownerReferencesNotFound() {
  if [[ "$1" == *"kind: Application"* ]] && [[ "$1" == *"name: test-application-01"* ]]; then
    echo "$2 ownerReferences updated"
    return 1
  else
    echo "$2 ownerReferences was not updated"
    return 0
  fi
}

header_text "Runn test on a Kubernetes cluster"

header_text "Deploy the application CRD with the controller"
#make deploy

header_text "Install the test CRD"
crd_output=$(kubectl apply -f test_crd.yaml)
if [ "$crd_output" != "customresourcedefinition.apiextensions.k8s.io/testcrds.test.crd.com created" ] &&
  [ "$crd_output" != "customresourcedefinition.apiextensions.k8s.io/testcrds.test.crd.com configured" ]; then
  echo "failed to install test CRD"
  echo "$crd_output"
  cleanup
  exit 1
else
  echo "$crd_output"
fi

header_text "Create the application object with all compoents"
app_output=$(kubectl apply -f application.yaml)
if [[ "$app_output" != *"application.app.k8s.io/test-application-01 created"* ]] &&
  [[ "$app_output" != *"application.app.k8s.io/test-application-01 configured"* ]] &&
  [[ "$app_output" != *"application.app.k8s.io/test-application-01 unchanged"* ]]; then
  echo "failed to install application"
  echo "$app_output"
  cleanup
  exit 1
else
  echo "$app_output"
fi

header_text "Verify application's status/components shows all components"
# Timeout waiting after 150 seconds
total=150
counter=$total
updated=false
while [ $counter -gt 0 ]; do
  app=$(kubectl get application test-application-01 -o yaml)
  if [[ "$app" == *"/apis/apps/v1/namespaces/default/deployments/test-hello"* ]] &&
    [[ "$app" == *"/apis/batch/v1/namespaces/default/jobs/pi"* ]] &&
    [[ "$app" == *"/api/v1/namespaces/default/services/test-webserver-svc"* ]] &&
    [[ "$app" == *"/api/v1/namespaces/default/configmaps/test-configmap"* ]] &&
    [[ "$app" == *"/apis/test.crd.com/v1/namespaces/default/testcrds/testcrd-sample"* ]] &&
    [[ "$app" == *"/apis/test.crd.com/v1/namespaces/default/testcrds/testcrd-sample-2"* ]]; then
    echo "All components are appended to application status"
    counter=0
    updated=true
  else
    echo "Still waiting for controller to reconcile... ($counter/$total)"
    sleep 1
    ((counter--))
  fi
done

if [ "$updated" == "false" ]; then
  echo "Timeout waiting for application status to be updated"
  cleanup
  exit 1
fi

header_text "Verify the application is added to all components' ownerReferences"
if ownerReferencesNotFound "$(kubectl get deployment test-hello -o yaml)" "deployment/test-hello" ||
  ownerReferencesNotFound "$(kubectl get job pi -o yaml)" "job/pi" ||
  ownerReferencesNotFound "$(kubectl get service test-webserver-svc -o yaml)" "service/test-webserver-svc" ||
  ownerReferencesNotFound "$(kubectl get configmap test-configmap -o yaml)" "configmap/test-configmap" ||
  ownerReferencesNotFound "$(kubectl get testcrd testcrd-sample -o yaml)" "testcrd/testcrd-sample" ||
  ownerReferencesNotFound "$(kubectl get testcrd testcrd-sample-2 -o yaml)" "testcrd/testcrd-sample-2"; then
  cleanup
  exit 1
fi

trap cleanup EXIT