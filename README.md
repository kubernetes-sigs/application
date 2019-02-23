# Kubernetes Applications

> Kubernetes is an open-source system for automating deployment, scaling, and management of containerized applications.

The above description, from the [Kubernetes homepage](https://kubernetes.io/), is centered on containerized _applications_. Yet, the Kubernetes metadata, objects, and visualizations (e.g., within Dashboard) are focused on container infrastructure rather than the applications themselves.

The Application CRD [(Custom Resource Definition)](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#customresourcedefinitions) and [Controller](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#custom-controllers) in this project aim to change that in a way that's interoperable between many supporting tools.

**It provides:**

* The ability to describe an applications metadata (e.g., that an application like WordPress is running)
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

## Application Objects

After creating the Application CRD, you can create Application objects. An Application object provides a way for you to aggregate individual Kubernetes components (e.g. Services, Deployments,
StatefulSets, Ingresses, CRDs), and manage them as a group. It provides UIs with a resource that allows for the
aggregation and display of all the components in the Application.

<table>
    <tr>
        <th>Field</th>
        <th>Type</th>
        <th>Description</th>
    </tr>
    <tr>
        <td>spec.descriptor.type</td>
        <td>string</td>
        <td>The type of the application (e.g. WordPress, MySQL, Cassandra). You can have many applications of different
        names in the same namespace. The type field is used to indicate that they are all the same type of application.
        </td>
    </tr>
    <tr>
        <td>spec.componentKinds</td>
        <td>[]<a href=https://kubernetes.io/docs/reference/using-api/api-overview/#api-groups> GroupKind </a> </td>
        <td>This array of GroupKinds is used to indicate the types of resources that the application is composed of. As
        an example an Application that has a service and a deployment would set this field to
        <i>[{"group":"core","kind": "Service"},{"group":"apps","kind":"Deployment"}]</i></td>
    </tr>
    <tr>
        <td>spec.selector</td>
        <td><a href=https://kubernetes.io/docs/concepts/overview/working-with-objects/labels>LabelSelector</a></td>
        <td>The selector is used to match resources that belong to the Application. All of the applications
        resources should have labels such that they match this selector. Users should use the
        <i>app.kubernetes.io/name</i> label on all components of the Application and set the selector to
        match this label. For instance, <i>{"matchLabels": [{"app.kubernetes.io/name": "my-cool-app"}]}</i> should be
        used as the selector for an Application named "my-cool-app", and each component should contain a label that
        matches.</td>
    </tr>
    <tr>
        <td>spec.addOwnerRef</td>
        <td>bool</td>
        <td>Flag controlling if the matched resources need to be adopted by the Application object. When adopting, an <a href=https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/#owners-and-dependents>OwnerRef</a> to the Application object is inserted into the matched objects <i>.metadata.[]OwnerRefs</i>.
	The injected OwnerRef has <i>blockOwnerDeletion</i> set to True and <i>controller</i> set to False.
        </td>
    </tr>
    <tr>
        <td>spec.descriptor.version</td>
        <td>string</a></td>
        <td>A version indicator for the application (e.g. 5.7 for MySQL version 5.7).</td>
    </tr>
    <tr>
        <td>spec.descriptor.description</td>
        <td>string</a></td>
        <td>A short, human readable textual description of the Application.</td>
    </tr>
    <tr>
        <td>spec.descriptor.icons</td>
        <td>[]ImageSpec</a></td>
        <td>A list of icons for an application. Icon information includes the source, size, and mime type.</td>
    </tr>
    <tr>
        <td>spec.descriptor.maintainers</td>
        <td>[]ContactData</a></td>
        <td>A list of the maintainers of the Application. Each maintainer has a name, email, and URL. This
        field is meant for the distributors of the Application to indicate their identity and contact information.</td>
    </tr>
    <tr>
        <td>spec.descriptor.owners</td>
        <td>[]ContactData</td>
        <td>A list of the operational owners of the application. This field is meant to be left empty by the
        distributors of application, and set by the installer to indicate who should be contacted in the event of a
        planned or unplanned disruption to the Application</td>
    </tr>
    <tr>
        <td>spec.descriptor.keywords</td>
        <td>array string</td>
        <td>A list of keywords that identify the application.</td>
    </tr>
    <tr>
        <td>spec.info</td>
        <td>[]InfoItem</td>
        <td>Info contains human readable key-value pairs for the Application.</td>
    </tr>
    <tr>
        <td>spec.descriptor.links</td>
        <td>[]Link</td>
        <td>Links are a list of descriptive URLs intended to be used to surface additional documentation,
        dashboards, etc.</td>
    </tr>
    <tr>
        <td>spec.descriptor.notes</td>
        <td>string</td>
        <td>Notes contain human readable snippets intended as a quick start for the users of the
        Application. They may be plain text or <a href="spec.commonmark.org">CommonMark</a> markdown.</td>
    </tr>
    <tr>
        <td>spec.assemblyPhase</td>
        <td>string: "Pending", "Succeeded" or "Failed"</td>
        <td>The installer can set this field to indicate that the
        application's components are still being deployed
        ("Pending") or all are deployed already ("Succeeded"). When the
        application cannot be successfully assembled, the installer can set this
        field to "Failed".</td>
    </tr>
</table>

## Building

This project uses the [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) tool for code generation.
kubebuilder provides the same code generation features (and a bit more) for Custom Resource Definitions and Extension
API Servers that are provided by the Kubernetes project. Installing kubebuilder is not needed to build the project, but
it is needed to do things like adding additional resources. 

To generate the manifests and run unit tests:
```commandline
make
```

### Controller

The controller doesn't do much at the moment. However, if you'd like to build an image you'll need to install Docker and
golang 1.9 or greater. To build the controller into an image named ```image``` use the following command.

```commandline
make docker-build IMG=<image>
make docker-push IMG=<image>
```

## Installing the CRD

To install the crd and the controller, just run:

```commandline
make deploy
```

This will install the controller into the application-system namespace and with the default RBAC permissions.

There is also a sample Application CR in the config/samples folder.


## Using the Application CRD

The application CRD can be used both via manifests and programmatically.

### Manifests

The docs directory contains a [manifest](docs/example.yaml) that shows how to you can integrate the Application CRD
with a WordPress deployment.

The Application object shown below declares that the Application is a WordPress installation that uses StatefulSets
and Services. It also contains some other relevant metadata described above.

```yaml
apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: "wordpress-01"
  labels:
    app.kubernetes.io/name: "wordpress-01"
    app.kubernetes.io/version: "3"
spec:
  selector:
    matchLabels:
     app.kubernetes.io/name: "wordpress-01"
  componentKinds:
    - group: core
      kind: Service
    - group: apps
      kind: Deployment
    - group: apps
      kind: StatefulSet
  assemblyPhase: "Pending"
  descriptor:
    version: "4.9.4"
    description: "WordPress is open source software you can use to create a beautiful website, blog, or app."
    icons:
      - src: "https://example.com/wordpress.png"
        type: "image/png"
    type: "wordpress"
    maintainers:
      - name: Kenneth Owens
        email: kow3ns@github.com
    owners:
      - "Kenneth Owens kow3ns@github.com"
    keywords:
      - "cms"
      - "blog"
      - "wordpress"
    links:
      - description: About
        url: "https://wordpress.org/"
      - description: Web Server Dashboard
        url: "https://metrics/internal/wordpress-01/web-app"
      - description: Mysql Dashboard
        url: "https://metrics/internal/wordpress-01/mysql"
```

Notice that each Service and StatefulSet is labeled such that Application's Selector matches the labels.

```yaml
app.kubernetes.io/name: "wordpress-01"
```

The additional labels on the Applications components come from the recommended application labels and annotations.

You can use the standard `kubectl` verbs (e.g. `get`, `apply`, `create`, `delete`, `list`, `watch`) to interact with
an Application specified in a manifest.

### Programmatically

Kubebuilder provides a client to get, create, update and delete resources and this also works for application resources.
This is well documented in the kubebuilder book: https://book.kubebuilder.io/

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

## Contributing

Go to the [CONTRIBUTING.md](CONTRIBUTING.md) documentation

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

* [Slack](http://slack.k8s.io/)
* [Mailing List](https://groups.google.com/d/forum/k8s-app-extension)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
