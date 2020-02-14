# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0
#
# Makefile for application


# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Releases should modify and double check these vars.
REGISTRY ?= gcr.io/$(shell gcloud config get-value project)
IMAGE_NAME ?= application-controller
CONTROLLER_IMG ?= $(REGISTRY)/$(IMAGE_NAME)
TAG ?= dev
ARCH ?= amd64
ALL_ARCH = amd64 arm arm64 ppc64le s390x

# Directories.
TOOLS_DIR := $(shell pwd)/hack/tools
TOOLBIN := $(TOOLS_DIR)/bin

# Allow overriding manifest generation destination directory
MANIFEST_ROOT ?= config
CRD_ROOT ?= $(MANIFEST_ROOT)/crd/bases
WEBHOOK_ROOT ?= $(MANIFEST_ROOT)/webhook
RBAC_ROOT ?= $(MANIFEST_ROOT)/rbac
COVER_FILE ?= cover.out


.DEFAULT_GOAL := all
.PHONY: all
all: generate license fix vet fmt manifests test lint misspell tidy bin/manager

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
.PHONY: e2e-test
e2e-test: generate fmt vet manifests $(TOOLBIN)/kind
	BIN=$(TOOLBIN) ./e2e/test_e2e.sh

## --------------------------------------
## Build and run
## --------------------------------------

# Build manager binary
bin/manager: main.go generate fmt vet manifests
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: generate fmt vet manifests
	go run ./main.go

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
## Deploying
## --------------------------------------

# Install CRDs into a cluster
.PHONY: install
install: $(TOOLBIN)/kustomize
	$(TOOLBIN)/kustomize build config/crd| kubectl apply -f -

# Uninstall CRDs from a cluster
.PHONY: uninstall
uninstall: $(TOOLBIN)/kustomize
	$(TOOLBIN)/kustomize build config/crd| kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy: $(TOOLBIN)/kustomize
	cd config/manager && $(TOOLBIN)/kustomize edit set image controller=$(CONTROLLER_IMG)-$(ARCH):$(TAG)
	$(TOOLBIN)/kustomize build config/default | kubectl apply -f -

# unDeploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: undeploy
undeploy: $(TOOLBIN)/kustomize
	$(TOOLBIN)/kustomize build config/default | kubectl delete -f -

# Deploy wordpress
.PHONY: deploy-wordpress
deploy-wordpress: $(TOOLBIN)/kustomize
	mkdir -p /tmp/data1 /tmp/data2
	$(TOOLBIN)/kustomize build docs/examples/wordpress | kubectl apply -f -

# Uneploy wordpress
.PHONY: undeploy-wordpress
undeploy-wordpress: $(TOOLBIN)/kustomize
	$(TOOLBIN)/kustomize build docs/examples/wordpress | kubectl delete -f -
	# kubectl delete pvc --all
	# sudo rm -fr /tmp/data1 /tmp/data2

## --------------------------------------
## Generating
## --------------------------------------

.PHONY: generate
generate: ## Generate code
	$(MAKE) generate-go
	$(MAKE) manifests


# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: $(TOOLBIN)/controller-gen
	$(TOOLBIN)/controller-gen \
		$(CRD_OPTIONS) \
		rbac:roleName=manager-role \
		paths=./... \
		output:crd:artifacts:config=$(CRD_ROOT) \
		output:crd:dir=$(CRD_ROOT) \
		output:webhook:dir=$(WEBHOOK_ROOT) \
		webhook

.PHONY: generate-go
generate-go: $(TOOLBIN)/controller-gen $(TOOLBIN)/conversion-gen  $(TOOLBIN)/mockgen
	go generate ./api/... ./controllers/...
	$(TOOLBIN)/controller-gen \
		paths=./api/v1beta1/... \
		object:headerFile=./hack/boilerplate.go.txt

## --------------------------------------
## Docker
## --------------------------------------

.PHONY: docker-build
docker-build: test $(TOOLBIN)/kustomize ## Build the docker image for controller-manager
	docker build --network=host --pull --build-arg ARCH=$(ARCH) . -t $(CONTROLLER_IMG)-$(ARCH):$(TAG)
	@echo "updating kustomize image patch file for manager resource"
	cd config/manager && $(TOOLBIN)/kustomize edit set image controller=$(CONTROLLER_IMG)-$(ARCH):$(TAG)

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(CONTROLLER_IMG)-$(ARCH):$(TAG)

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
