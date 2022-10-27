[![Build Status](https://travis-ci.org/kubernetes-sigs/application.svg?branch=master)](https://travis-ci.org/kubernetes-sigs/application "Travis")
[![Go Report Card](https://goreportcard.com/badge/sigs.k8s.io/application)](https://goreportcard.com/report/sigs.k8s.io/application)

# Kubernetes Applications

> Kubernetes is an open-source system for automating deployment, scaling, and management of containerized applications.

The above description, from the [Kubernetes homepage](https://kubernetes.io/), is centered on containerized _applications_. Yet, the Kubernetes metadata, objects, and visualizations (e.g., within Dashboard) are focused on container infrastructure rather than the applications themselves.

The Application CRD [(Custom Resource Definition)](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#customresourcedefinitions) and [Controller](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#custom-controllers) in this project aim to change that in a way that's interoperable between many supporting tools.

**It provides:**

* The ability to describe an application's metadata (e.g., that an application like WordPress is running)
* A point to connect the infrastructure, such as Deployments, to as a root object. This is useful for tying things together and even cleanup (i.e., garbage collection)
* Information for supporting applications to help them query and understand the objects supporting an application
* Application level health checks

**This can be used by:**

* Application operators who want to center what they operate on applications
* Tools, such as Helm, that center their package releases on application installations can do so in a way that's interoperable with other tools (e.g., Dashboard)
* Dashboards that want to visualize the applications in addition to or instead of an infrastructure view

## Goals

1. Provide a standard API for creating, viewing, and managing applications in Kubernetes.
1. Provide a CLI implementation, via kubectl, that interacts with the Application API.
1. Provide installation status and garbage collection for applications.
1. Provide a standard way for applications to surface a basic health check to the UIs.
1. Provide an explicit mechanism for applications to declare dependencies on another application.
1. Promote interoperability among ecosystem tools and UIs by creating a standard that tools MAY implement.
1. Promote the use of common labels and annotations for Kubernetes Applications.

## Non-Goals

1. Create a standard that all tools MUST implement.
1. Provide a way for UIs to surface metrics from an application.

## Application API

Refer [API doc](docs/api.md).
For an example look at [wordpress application](docs/examples/wordpress/application.yaml)

## Quickstart

Refer [Quickstart Guide](docs/quickstart.md)

## Development

Refer [Development Guide](docs/develop.md)

## Contributing

Go to the [CONTRIBUTING.md](CONTRIBUTING.md) documentation

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

* [Slack](http://slack.k8s.io/)
* [Mailing List](https://groups.google.com/d/forum/k8s-app-extension)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

## Releasing

Refer [Releasing Guide](docs/release.md)
