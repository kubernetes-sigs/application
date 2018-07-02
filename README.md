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

## What is the Application CRD?

The Application CRD provides a way for you to aggregate individual Kubernetes components (e.g. Services, Deployments,
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
        names in the same namespace. They type field is used to indicate that they are all the same type of application.
        </td>
    </tr>
    <tr>
        <td>spec.componentKinds</td>
        <td>[]<a href=https://kubernetes.io/docs/reference/api-overview/#api-groups> GroupKind </a> </td>
        <td>This array of GroupKinds is used to indicate the types of resources that the application is composed of. As
        an example an Application that has a service and a deployment would set this field to
        <i>[{"group":"","kind": "Service"},{"group":"apps","kind":"StatefulSet"}]</i></td>
    </tr>
    <tr>
        <td>spec.selector</td>
        <td><a href=https://kubernetes.io/docs/concepts/overview/working-with-objects/labels>LabelSelector</a></td>
        <td>The selector is used to match resources that belong to the Application. All of the applications
        resources should be labels such that they match this selector. Users should use the
        <i>app.kubernetes.io/name</i> label on all components of the Application and set the selector to
        match this label. For instance, <i>{"matchLables": [{"app.kubernetes.io/name": "my-cool-app"}]}</i> should be
        used as the selector for an Application named "my-cool-app", and each component should contain a label that
        matches.</td>
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
        <td>Info contains human readable key,value pairs for the Application.</td>
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
        <td>Notes contain a human readable snippets intended as a quick start for the users of the
        Application.</td>
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
API Servers that are provided by the Kubernetes project. In order to build the source, you need to download and install
the latest release of kubebuilder per the instructions there.

### Controller

The controller doesn't do much at the moment. However, if you'd like to build it you'll need to install Docker and
golang 1.9 or greater. To build the controller into an image named ```image``` use the following command.

```commandline
docker build -t <image> -f Dockerfile.controller .
```

## Installing the CRD

In order to install the CRD you will either need to use kubectl or you will need to call against the Kubernetes CRD
API directly. An example [manifest](hack/install.yaml) is supplied in the hack directory. You can use the following
command to install the CRD (where ```manifest`` is the manifest containing the CRD declaration).

```commandline
kubectl apply -f <manifest>
```

## Generating an Installation Manifest

When the CRD is installed as above, you need to ensure that the correct RBAC configuration is applied prior to
installation. You can use `kubebuilder create config` to generate a manifest that is configured to create the
requisite RBAC permissions, CRD, and controller StatefulSet in the supplied namespace. The command below will generate
a manifest that can be applied to create all of the necessary components the `image` as the controller
image and `namespace` as the namespace. Note that, if you would like to remove the controller from the configuration
you can delete the generated StatefulSet from the manifest, and, while you must specify a controller image, you need
can supply any string if you do not wish to install the controller when the manifest is applied (i.e. you intend to
delete the StatefulSet from the generated manifest). Work is in progress to generate a controllerless configuration.

```commandline
kubebuilder create config --controller-image <image>  --name <namespace>
```

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

kubebuilder creates a Kubernetes [ClientSet](pkg/client/clientset/versioned/clientset.go) for the Application object.
You can create a new client using either a rest.Config or a rest.Interface as below.

```go
client,err := clientset.NewForConfig(config)

client := clientset.New(ri)
```

Once you've created a client you can interact with Applications via the structs declared in
[types.go](pkg/apis/app/v1beta1/application_types.go). For instance to retrieve an application you can used the code
below.

```go
app, err := client.AppV1Aplha1().Applications("my-namespace").Get("my-app",v1.GetOptions{})
if err != nil {
        handleError(err)
}
```

The other standard client operations are supported. The interface is described
[here](pkg/client/clientset/versioned/typed/app/v1beta1/application.go).

## Contributing

Go to the [CONTRIBUTING.md](CONTRIBUTING.md) documentation

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

* [Slack](http://slack.k8s.io/)
* [Mailing List](https://groups.google.com/d/forum/k8s-app-extension)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
