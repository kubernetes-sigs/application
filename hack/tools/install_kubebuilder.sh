#!/usr/bin/env bash
# Copyright 2020 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


source ./common.sh

version=2.3.1

header_text "Checking for bin/kubebuilder"
[[ -f bin/kubebuilder ]] && exit 0

header_text "Installing bin/kubebuilder"
mkdir -p ./bin
curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${version}/kubebuilder_${version}_${os}_${arch}.tar.gz"

tar -zxvf kubebuilder_${version}_${os}_${arch}.tar.gz
mv kubebuilder_${version}_${os}_${arch}/bin/* bin

rm kubebuilder_${version}_${os}_${arch}.tar.gz
rm -r kubebuilder_${version}_${os}_${arch}
