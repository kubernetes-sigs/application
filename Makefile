# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# Makefile for application


# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Releases should modify and double check these vars.
VERSION ?= dev
ARCH ?= amd64
ALL_ARCH = amd64 arm arm64 ppc64le s390x
IMAGE_NAME ?= kube-app-manager
ifeq ($(ARCH), amd64)
TAG ?= $(VERSION)
else
TAG ?= $ARCH-$(VERSION)
endif
# GCLOUD_REGISTRY ?= gcr.io/$(shell gcloud config get-value project)
# REGISTRY = $(GCLOUD_REGISTRY)

REGISTRY ?= quay.io/kubernetes-sigs
CONTROLLER_IMG ?= $(REGISTRY)/$(IMAGE_NAME):$(TAG)

# Directories.
TOOLS_DIR := $(shell pwd)/hack/tools
TOOLBIN := $(TOOLS_DIR)/bin

# Allow overriding manifest generation destination directory
MANIFEST_ROOT ?= config
CRD_ROOT ?= $(MANIFEST_ROOT)/crd/bases
WEBHOOK_ROOT ?= $(MANIFEST_ROOT)/webhook
RBAC_ROOT ?= $(MANIFEST_ROOT)/rbac
COVER_FILE ?= cover.out

VERSIONS := dev v0.8.1
.DEFAULT_GOAL := all
.PHONY: all
all: generate fix vet fmt manifests test lint license misspell tidy bin/kube-app-manager


## --------------------------------------
## Tooling Binaries
## --------------------------------------

$(TOOLBIN)/controller-gen:
	GOBIN=$(TOOLBIN) GO111MODULE=on go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5

$(TOOLBIN)/golangci-lint:
	GOBIN=$(TOOLBIN) GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.23.6

$(TOOLBIN)/mockgen:
	GOBIN=$(TOOLBIN) GO111MODULE=on go get github.com/golang/mock/mockgen@v1.3.1

$(TOOLBIN)/conversion-gen:
	GOBIN=$(TOOLBIN) GO111MODULE=on go get k8s.io/code-generator/cmd/conversion-gen@v0.17.0

$(TOOLBIN)/kubebuilder $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver $(TOOLBIN)/kubectl:
	cd $(TOOLS_DIR); ./install_kubebuilder.sh

$(TOOLBIN)/kustomize:
	cd $(TOOLS_DIR); ./install_kustomize.sh

$(TOOLBIN)/kind:
	GOBIN=$(TOOLBIN) GO111MODULE=on go get sigs.k8s.io/kind@v0.6.0

$(TOOLBIN)/addlicense:
	GOBIN=$(TOOLBIN) GO111MODULE=on go get github.com/google/addlicense

$(TOOLBIN)/misspell:
	GOBIN=$(TOOLBIN) GO111MODULE=on go get github.com/client9/misspell/cmd/misspell@v0.3.4

.PHONY: install-tools
install-tools: \
	$(TOOLBIN)/controller-gen \
	$(TOOLBIN)/golangci-lint \
	$(TOOLBIN)/mockgen \
	$(TOOLBIN)/conversion-gen \
	$(TOOLBIN)/kubebuilder \
	$(TOOLBIN)/kustomize \
	$(TOOLBIN)/addlicense \
	$(TOOLBIN)/misspell \
	$(TOOLBIN)/kind

## --------------------------------------
## Tests
## --------------------------------------

# Run tests
.PHONY: test
test: $(TOOLBIN)/etcd $(TOOLBIN)/kube-apiserver $(TOOLBIN)/kubectl
	TEST_ASSET_KUBECTL=$(TOOLBIN)/kubectl \
	TEST_ASSET_KUBE_APISERVER=$(TOOLBIN)/kube-apiserver \
	TEST_ASSET_ETCD=$(TOOLBIN)/etcd \
	go test -v ./api/... ./controllers/... -coverprofile $(COVER_FILE)

# Run e2e-tests
K8S_VERSION := "v1.16.4"

.PHONY: e2e-setup
e2e-setup: $(TOOLBIN)/kind
	KUBECONFIG=$(shell $(TOOLBIN)/kind get kubeconfig-path --name="kind") \
	$(TOOLBIN)/kind create cluster \
	 -v 4 --retain --wait=1m \
	 --config e2e/kind-config.yaml \
	 --image=kindest/node:$(K8S_VERSION)

.PHONY: e2e-cleanup
e2e-cleanup: $(TOOLBIN)/kind
	$(TOOLBIN)/kind delete cluster

.PHONY: e2e-test
e2e-test: generate fmt vet manifests $(TOOLBIN)/kind $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	go test -v ./e2e/main_test.go

.PHONY: local-e2e-test
local-e2e-test: e2e-setup e2e-test e2e-cleanup

## --------------------------------------
## Build and run
## --------------------------------------

# Build kube-app-kube-app-manager binary
bin/kube-app-manager: main.go generate fmt vet manifests
	go build -o bin/kube-app-manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: runbg
runbg: bin/kube-app-manager
	bin/kube-app-manager --metrics-addr ":8083" >& kube-app-manager.log & echo $$! > kube-app-manager.pid

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: bin/kube-app-manager
	bin/kube-app-manager

# Debug using the configured Kubernetes cluster in ~/.kube/config
.PHONY: debug
debug: generate fmt vet manifests
	dlv debug ./main.go


## --------------------------------------
## Code maintenance
## --------------------------------------

.PHONY: fmt
fmt:
	go fmt ./api/... ./controllers/...

.PHONY: vet
vet:
	go vet ./api/... ./controllers/...

.PHONY: fix
fix:
	go fix ./api/... ./controllers/...

.PHONY: license
license: $(TOOLBIN)/addlicense
	$(TOOLBIN)/addlicense  -y $(shell date +"%Y") -c "The Kubernetes Authors." -f LICENSE_TEMPLATE .

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint: $(TOOLBIN)/golangci-lint
	$(TOOLBIN)/golangci-lint run ./...

.PHONY: misspell
misspell: $(TOOLBIN)/misspell
	$(TOOLBIN)/misspell ./**

.PHONY: misspell-fix
misspell-fix: $(TOOLBIN)/misspell
	$(TOOLBIN)/misspell -w ./**


## --------------------------------------
## Deploy all (CRDs + Controller)
## --------------------------------------

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy: $(TOOLBIN)/kubectl
	$(TOOLBIN)/kubectl apply -f deploy/kube-app-manager-aio_$(VERSION).yaml

# unDeploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: undeploy
undeploy: $(TOOLBIN)/kubectl
		$(TOOLBIN)/kubectl delete -f deploy/kube-app-manager-aio_$(VERSION).yaml

## --------------------------------------
## Deploy CRDs only
## --------------------------------------
# Install CRDs into a cluster,
.PHONY: install
deploy-crd: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	$(TOOLBIN)/kustomize build config/crd| $(TOOLBIN)/kubectl apply -f -

# Uninstall CRDs from a cluster
.PHONY: uninstall
undeploy-crd: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	$(TOOLBIN)/kustomize build config/crd| $(TOOLBIN)/kubectl delete -f -

## --------------------------------------
## Deploy demo
## --------------------------------------

# Deploy wordpress
.PHONY: deploy-wordpress
deploy-wordpress: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	mkdir -p /tmp/data1 /tmp/data2
	$(TOOLBIN)/kustomize build docs/examples/wordpress | $(TOOLBIN)/kubectl apply -f -

# Uneploy wordpress
.PHONY: undeploy-wordpress
undeploy-wordpress: $(TOOLBIN)/kustomize $(TOOLBIN)/kubectl
	$(TOOLBIN)/kustomize build docs/examples/wordpress | $(TOOLBIN)/kubectl delete -f -
	# $(TOOLBIN)/kubectl delete pvc --all
	# sudo rm -fr /tmp/data1 /tmp/data2

## --------------------------------------
## Generating
## --------------------------------------

.PHONY: generate
generate: ## Generate code
	$(MAKE) generate-go
	$(MAKE) manifests
	$(MAKE) generate-resources

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: $(TOOLBIN)/controller-gen
	$(TOOLBIN)/controller-gen \
		$(CRD_OPTIONS) \
		rbac:roleName=kube-app-manager-role \
		paths=./... \
		output:crd:artifacts:config=$(CRD_ROOT) \
		output:crd:dir=$(CRD_ROOT) \
		output:webhook:dir=$(WEBHOOK_ROOT) \
		webhook

.PHONY: generate-resources
generate-resources: $(TOOLBIN)/kustomize
	$(TOOLBIN)/kustomize build config/default/$(VERSION) -o deploy/kube-app-manager-aio_$(VERSION).yaml

.PHONY: generate-resources-%
generate-resources-%:
	$(MAKE) VERSION=$* generate-resources

.PHONY: generate-go
generate-go: $(TOOLBIN)/controller-gen $(TOOLBIN)/conversion-gen  $(TOOLBIN)/mockgen
	go generate ./api/... ./controllers/...
	$(TOOLBIN)/controller-gen \
		paths=./api/v1beta1/... \
		object:headerFile=./hack/boilerplate.go.txt

## --------------------------------------
## Docker
## --------------------------------------
.PHONY: set-image
set-image: $(TOOLBIN)/kustomize
	@echo "updating kustomize image patch file for kube-app-manager resource"
	cd config/kube-app-manager && $(TOOLBIN)/kustomize edit set image kube-app-manager=$(CONTROLLER_IMG)

.PHONY: docker-build
docker-build: set-image test $(TOOLBIN)/kustomize ## Build the docker image for kube-app-manager
	docker build --network=host --pull --build-arg ARCH=$(ARCH) . -t $(CONTROLLER_IMG)

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(CONTROLLER_IMG)

.PHONY: clean
clean:
	go clean --cache
	rm -f $(COVER_FILE)
	rm -f $(TOOLBIN)/kustomize
	rm -f $(TOOLBIN)/goimports
	rm -f $(TOOLBIN)/golangci-lint
	rm -f $(TOOLBIN)/controller-gen
	rm -f $(TOOLBIN)/conversion-gen
	rm -f $(TOOLBIN)/etcd
	rm -f $(TOOLBIN)/kube-apiserver
	rm -f $(TOOLBIN)/kubebuilder
	rm -f $(TOOLBIN)/addlicense
	rm -f $(TOOLBIN)/kubectl
	rm -f $(TOOLBIN)/kustomize
	rm -f $(TOOLBIN)/misspell
	rm -f $(TOOLBIN)/mockgen
