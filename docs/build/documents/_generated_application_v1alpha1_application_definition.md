## Application v1alpha1 application

Group        | Version     | Kind
------------ | ---------- | -----------
application | v1alpha1 | Application



Application The Application object acts as an aggregator for components that comprise an Application. Its Spec.ComponentGroupKinds indicate the GroupKinds of the components the comprise the Application. Its Spec. Selector is used to list and watch those components. All components of an Application should be labeled such the Application's Spec. Selector matches.

<aside class="notice">
Appears In:

<ul> 
<li><a href="#applicationlist-v1alpha1-application">ApplicationList application/v1alpha1</a></li>
</ul></aside>

Field        | Description
------------ | -----------
apiVersion <br /> *string*    | APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources
kind <br /> *string*    | Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
metadata <br /> *[ObjectMeta](#objectmeta-v1-meta)*    | 
spec <br /> *[ApplicationSpec](#applicationspec-v1alpha1-application)*    | The specification object for the Application.
status <br /> *[ApplicationStatus](#applicationstatus-v1alpha1-application)*    | The status object for the Application.

