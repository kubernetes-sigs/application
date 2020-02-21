# Development Guide

## Prerequisites

### Tools
- make
- [go](https://golang.org/dl/) version v1.13+.
- [docker](https://docs.docker.com/install/) version 17.03+.

### Other tools
This repo uses other tools for development, building and testing.
These tools are installed with `make install-tools`:
- controller-gen
- golangci-lint
- mockgen
- conversion-gen
- kubebuilder
- [kustomize](https://github.com/kubernetes-sigs/kustomize)
- addlicense
- misspell
- [kind](https://github.com/kubernetes-sigs/kind)

### Cluster
- Access to a Kubernetes v1.11.3+ cluster.

## Development

### Fork and Clone

Fork [Application Repo](https://github.com/kubernetes-sigs/application).
Then clone your fork locally.

```bash
mkdir -p $GOPATH/src/sigs.k8s.io
cd $GOPATH/src/sigs.k8s.io

GITHUBID=<githubid>
git clone git@github.com:${GITHUBID}/application.git $GOPATH/src/sigs.k8s.io/application
```

### Cluster access
For running e2e tests and development testing you need access to a cluster. You could create a cluster with your cloud provider and ensure the `kubeconfig` points to the cluster.

##### Local cluster
For local testing you could create a `kind`.

```bash
# this created a kind cluster and updates kubeconfig to point to it
make e2e-setup
```
### Building the controller binary

```bash
# make or make all will build
make

# individual run make targets
#
# generate code
make generate

# create manifests
make manifests

# Inject license header to all generated files
make license

# building the kube-app-manager
make bin/kube-app-manager
```

### Running tests
After making changes use these targets to test locally.

##### unit tests
```bash
# running unitests
make test
```

##### e2e tests
```bash
# running e2e tests

# If your kubeconfig points to your test cluster skip this step
# This will create a kind cluster for testing
make e2e-setup

# run e2e tests
make e2e-test

# Tear down kind cluster
make e2e-cleanup
```

### Building controller docker
To build the controller into an image named `image` use the following command.

NOTE:
`CONTROLLER_IMG` is optional. The default value is `gcr.io/$(shell gcloud config get-value project)/application-controller`

```commandline
make docker-build CONTROLLER_IMG=<image>
```

To push the controller image, run:
```commandline
make docker-push CONTROLLER_IMG=<image>
```

### Installing CRD in the cluster
Once kubeconfig is setup with a cluster.
```bash
make install
```
### Deploying the controller in cluster

- This will install the controller into the application-system namespace and with the default RBAC permissions.
- It will also install the Application CRD.
- Ensure the docker image is built and pushed first.

```commandline
make deploy CONTROLLER_IMG=<image>
```

## Using the Application CRD

The application CRD can be used both via manifests and programmatically.

### Manifests

The docs directory contains an example [manifest](docs/examples/wordpress/application.yaml) that shows how to you can integrate the Application CRD with a [WordPress deployment](docs/examples/wordpress).

The Application object uses StatefulSets and Services. It also contains some other relevant metadata describing wordpress application. Notice that each Service and StatefulSet is labeled such that Application's Selector matches the labels. The additional labels on the Applications components come from the recommended application labels and annotations.

```bash
# Deploying the example

make deploy-wordpress
kubectl get application

# cleanup
make undeploy-wordpress
```
### Programmatically

Kubebuilder provides a client to get, create, update and delete resources and this also works for application resources. This is documented in the kubebuilder book: https://book.kubebuilder.io/

Create a client:
```go
kubeClient, err := client.New(config)
```

Get an application resource:
```go
object := &applicationsv1beta1.Application{}
objectKey := types.NamespacedName{
    Namespace: "namespace",
    Name: "name",
}
err = kubeClient.Get(context.TODO(), objectKey, object)
```

Create a new application resource:
```go
app := &applicationsv1beta1.Application{
	...
}
err = kubeClient.Create(context.TODO(), app)
```
