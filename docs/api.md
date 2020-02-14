## Application Object

An Application object provides a way for you to aggregate individual Kubernetes components (e.g. Services, Deployments, StatefulSets, Ingresses, CRDs), and manage them as a group. It provides tooling and UI with a resource that allows for the aggregation and display of all the components in the Application.

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

