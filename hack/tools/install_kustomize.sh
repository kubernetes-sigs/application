#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


source ./common.sh

version=3.2.0

header_text "Checking for bin/kustomize"
[[ -f bin/kustomize ]] && exit 0

header_text "Installing for bin/kustomize"
mkdir -p ./bin
curl -L https://github.com/kubernetes-sigs/kustomize/releases/download/v${version}/kustomize_${version}_${os}_${arch} -o ./bin/kustomize
chmod +x ./bin/kustomize
