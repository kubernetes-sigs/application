# Kubernetes Applications

Kubernetes has many primitives for managing workloads (e.g. Pods, ReplicaSets, Deployments, DaemonSets and 
StatefulSets), storage (e.g. PersistentVolumeClaims and PersistentVolumes), and networking (e.g. Services, 
Headless Services, and Ingresses). When these primitives are aggregated to provide a service to an end user or to 
another system, the whole becomes something more than the individual parts. Instead of a set of loosely coupled 
workloads and their corresponding storage and networking, we have an application. 

To address these issues, we are developing the Application CRD (Custom Resource Definition). This Kind can be used 
by tools to communicate that the applications they create are more than just a loosely coupled set of API objects. It 
is our hope that broad adoption of this resource will promote interoperability in the ecosystem and reduce 
fragmentation.

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

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](http://slack.k8s.io/)
- [Mailing List](https://groups.google.com/d/forum/k8s-app-extension)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
