#!/bin/bash -x

[[ -f bin/kustomize ]] && exit 0

mkdir -p ./bin
curl -L https://github.com/kubernetes-sigs/kustomize/releases/download/v3.2.0/kustomize_3.2.0_linux_amd64 -o ./bin/kustomize
chmod +x ./bin/kustomize
