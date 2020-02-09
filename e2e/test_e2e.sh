#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0



set -o errexit
set -o nounset
set -o pipefail

K8S_VERSION="v1.16.4"

export KUBECONFIG="$(${BIN}/kind get kubeconfig-path --name="kind")"
# image supported version are listed at https://hub.docker.com/r/kindest/node/tags
${BIN}/kind create cluster -v 4 --retain --wait=1m --config e2e/kind-config.yaml --image=kindest/node:$K8S_VERSION

# remove running containers on exit
function cleanup() {
    ${BIN}/kind delete cluster
}

trap cleanup EXIT

go test -v ./e2e/main_test.go
