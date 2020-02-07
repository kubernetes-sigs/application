#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

source ./hack/scripts/common.sh


K8S_VERSION="v1.16.4"

fetch_kb_tools
install_kind
setup_envs

export KUBECONFIG="$(kind get kubeconfig-path --name="kind")"
# You can use --image flag to specify the cluster version you want, e.g --image=kindest/node:v1.13.6, the supported version are listed at https://hub.docker.com/r/kindest/node/tags
kind create cluster -v 4 --retain --wait=1m --config e2e/kind-config.yaml --image=kindest/node:$K8S_VERSION

# remove running containers on exit
function cleanup() {
#    make undeploy-wordpress
#    make undeploy
    kind delete cluster
}

trap cleanup EXIT

go test -v ./e2e/main_test.go

# test the wordpress example
# Disable E2E test for the wordpress example because the controller image is not available to the travis build

#export REGISTRY="gcr.io/application-crd"
#export TAG="travis-build"
#make docker-build -o test
#make docker-push
#make deploy
#make deploy-wordpress

#go test -v ./e2e/wordpress_test.go