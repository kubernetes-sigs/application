
# Define Docker related variables. Releases should modify and double check these vars.
REGISTRY ?= gcr.io/$(shell gcloud config get-value project)
IMAGE_NAME ?= application-controller
CONTROLLER_IMG ?= $(REGISTRY)/$(IMAGE_NAME)
TAG ?= dev
ARCH ?= amd64
ALL_ARCH = amd64 arm arm64 ppc64le s390x

.DEFAULT_GOAL := all

.PHONY: test manager run debug install deploy manifests fmt vet generate docker-build docker-push

# Directories.
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
BIN_DIR := bin

# Binaries.
CONTROLLER_GEN := $(TOOLS_BIN_DIR)/controller-gen
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint
MOCKGEN := $(TOOLS_BIN_DIR)/mockgen
CONVERSION_GEN := $(TOOLS_BIN_DIR)/conversion-gen
KUBEBUILDER := $(TOOLS_BIN_DIR)/kubebuilder
KUSTOMIZE := $(TOOLS_BIN_DIR)/kustomize

# Allow overriding manifest generation destination directory
MANIFEST_ROOT ?= config
CRD_ROOT ?= $(MANIFEST_ROOT)/crds
WEBHOOK_ROOT ?= $(MANIFEST_ROOT)/webhook
RBAC_ROOT ?= $(MANIFEST_ROOT)/rbac
COVER_FILE ?= cover.out

## --------------------------------------
## Tooling Binaries
## --------------------------------------

.PHONY: $(CONTOLLER_GEN)
$(CONTROLLER_GEN): $(TOOLS_DIR)/go.mod # Build controller-gen from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/controller-gen sigs.k8s.io/controller-tools/cmd/controller-gen

.PHONY: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(TOOLS_DIR)/go.mod # Build golangci-lint from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/golangci-lint github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: $(MOCKGEN)
$(MOCKGEN): $(TOOLS_DIR)/go.mod # Build mockgen from tools folder.
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/mockgen github.com/golang/mock/mockgen

.PHONY: $(CONVERSION_GEN)
$(CONVERSION_GEN): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); go build -tags=tools -o $(BIN_DIR)/conversion-gen k8s.io/code-generator/cmd/conversion-gen

.PHONY: $(KUBEBUILDER)
$(KUBEBUILDER): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); ./install_kubebuilder.sh

.PHONY: $(KUSTOMIZE)
$(KUSTOMIZE): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR); ./install_kustomize.sh

## --------------------------------------
## Linting
## --------------------------------------


all: test manager

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile $(COVER_FILE)

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager ./cmd/manager/main.go

# Run using the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Debug using the configured Kubernetes cluster in ~/.kube/config
debug: generate fmt vet
	dlv debug cmd/manager/main.go

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

## --------------------------------------
## Deploying
## --------------------------------------

# Install CRDs into a cluster
install: $(KUSTOMIZE)
	kubectl apply -k config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: $(KUSTOMIZE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -


# unDeploy controller in the configured Kubernetes cluster in ~/.kube/config
undeploy: $(KUSTOMIZE)
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Deploy wordpress
deploy-wordpress: $(KUSTOMIZE)
	mkdir -p /tmp/data1 /tmp/data2
	$(KUSTOMIZE) build examples/wordpress | kubectl apply -f -


# Uneploy wordpress
undeploy-wordpress: $(KUSTOMIZE)
	$(KUSTOMIZE) build examples/wordpress | kubectl delete -f -
	# kubectl delete pvc --all
	# sudo rm -fr /tmp/data1 /tmp/data2

## --------------------------------------
## Generating
## --------------------------------------

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: $(CONTROLLER_GEN) ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		paths=./pkg/apis/... \
		crd:trivialVersions=true \
		output:crd:dir=$(CRD_ROOT) \
		output:webhook:dir=$(WEBHOOK_ROOT) \
		webhook

.PHONY: generate
generate: ## Generate code
	$(MAKE) generate-go
	$(MAKE) manifests

# Generate code
.PHONY: generate-go
generate-go: $(CONTROLLER_GEN) $(MOCKGEN) $(CONVERSION_GEN) $(KUBEBUILDER) $(KUSTOMIZE) ## Runs Go related generate targets
	go generate ./pkg/... ./cmd/...
	$(CONTROLLER_GEN) \
		paths=./pkg/apis/app/v1beta1/... \
		object:headerFile=./hack/boilerplate.go.txt

## --------------------------------------
## Docker
## --------------------------------------

.PHONY: docker-build
docker-build: ## Build the docker image for controller-manager
	docker build --network=host --pull --build-arg ARCH=$(ARCH) . -t $(CONTROLLER_IMG)-$(ARCH):$(TAG)
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${CONTROLLER_IMG}-$(ARCH):$(TAG)"'@' ./config/default/manager_image_patch.yaml

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(CONTROLLER_IMG)-$(ARCH):$(TAG)

.PHONY: clean
clean:
	go clean --cache
	rm -f $(COVER_FILE)
	rm -f $(TOOLS_BIN_DIR)/kustomize
	rm -f $(TOOLS_BIN_DIR)/goimports
	rm -f $(TOOLS_BIN_DIR)/golangci-lint
	rm -f $(TOOLS_BIN_DIR)/controller-gen
	rm -f $(TOOLS_BIN_DIR)/conversion-gen
	rm -f $(TOOLS_BIN_DIR)/etcd
	rm -f $(TOOLS_BIN_DIR)/kube-apiserver
	rm -f $(TOOLS_BIN_DIR)/kubebuilder
	rm -f $(TOOLS_BIN_DIR)/kubectl
	rm -f $(TOOLS_BIN_DIR)/kustomize
	rm -f $(TOOLS_BIN_DIR)/mockgen
