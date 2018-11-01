## Background
Application CRD has a list of GroupKinds and a LabelSelector to logically group live objects in the cluster. The Application CRD also provides a way to attach descriptive metadata to the logical grouping. 

## What does this solve
This PR proposes to enhance Application CRD controller to provide status tracking for the objects that match the selectors. This dramatically simplifies the status tracking requirements for a client that uses Application CRD. 

## Proposal
The proposal is to add a reconciler that creates watches for the GroupKinds that the application CRD refers to. The reconciler then inspects the matching objects and updates the aggregate status in the Application CRD.

### New fields in Application CRD
We propose to use Conditions to aggregate status of the matching objects.
These conditions could be added to `.status.conditions`
```
"Ready"   //Aggregate readiness
"Settled" //Observed generation is latest, the controller has acted on latest version of spec
```
In addition we could have the matching objects status also recorded in `.status.components`.  The motivation behind this is to provide individual component status as well for clients that need more than just the aggregate status. Well known object types can be aggregated to a higher degree of fidelity. For other types and custom resources we can propose standardizing `Ready` condition that could then be aggregated.

```
package application
// ApplicationStatus defines controllers the observed state of Application
type ApplicationStatus struct {
	// ObservedGeneration is used by the Application Controller to report the last Generation of an Application
	// that it has observed.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions - Ensure Ready,Settled conditions are set
	Conditions []status.Condition   `json:"conditions,omitempty"`
	// List of subcomponent status
	Resources status.ResourceStatus `json:",inline"`
}

--------------------

package status
// ResourceStatus is a generic status holder for the top level resource
type ResourceStatus struct {
	// Object status array for all matching objects
	Objects []ObjectStatus `json:"objects,omitempty"`
}

// ObjectStatus is a generic status holder for objects
type ObjectStatus struct {
	// Link to object
	Link string `json:"link,omitempty"`
	// Name of object
	Name string `json:"name,omitempty"`
	// Kind of object 
	Kind string `json:"kind,omitempty"`
	// Object group
	Group string `json:"kind,omitempty"`
	// Status. Values: InProgress, Ready, Unknown
	Status string `json:"status,omitempty"`
	// progress is a fuzzy indicator. Interpret as a percentage (0-100)
	// eg: for statefulsets, progress = 100*readyreplicas/replicas
	Progress int32 `json:"progress"`
}

type Condition struct {
...
}

``` 

### Logic for computing object status of well known objects
The logic to compute the readiness for the objects is based on the resource group-kind.
1. Statefulset is Ready if `.status.readyreplicas == .spec.replicas && .status.currentreplicas == .spec.replicas`
2. Service is Ready always (could not derive from `.status.loadbalancer`)
3. PodDisruptionBudget is Ready if .`status.currenthealthy >= .status.desiredhealthy`

The long term plan is for these logic to be added to core-controllers and them to update the `.status.conditions.[Ready]` for the individual objects.

### Status for other objects
We propose looking for `.status.conditions.[Ready]` for non well known group-kinds and using that for aggregation.
If the object does not have `.status.conditions.[Ready]`, we propose to mark that objects status as "Unknown". We will revisit this after feedback from users.

### Aggregating status
Application CRD's `.status.conditions.[Ready]` is derived by a simple ANDing of the matching object status computed under `.status.components`. 

### Example output
```
status:
  objects:
    - progress: 100
      link: /apis/apps/v1/namespaces/default/statefulsets/esbasic-di
      name: esbasic-di
      group: apps/v1
      kind: statefulset
      status: Ready
    - progress: 100
      link: /apis/apps/v1/namespaces/default/statefulsets/esbasic-m
      name: esbasic-m
      group: apps/v1
      kind: statefulset
      status: Ready
    - progress: 100
      link: /apis/apps/v1/namespaces/default/statefulsets/esbasic-metrics
      name: esbasic-metrics
      group: apps/v1
      kind: statefulset
      status: Ready
    - progress: 100
      link: /api/v1/namespaces/default/services/esbasic-di
      name: esbasic-di
      group: v1
      kind: service
      status: Ready
    - progress: 100
      link: /api/v1/namespaces/default/services/esbasic-m
      name: esbasic-m
      group: v1
      kind: service
      status: Ready
  conditions:
  - lastTransitionTime: 2018-10-29T23:30:03Z
    lastUpdateTime: 2018-10-29T23:30:03Z
    message: All components ready
    reason: ComponentsReady
    status: "True"
    type: Ready
  - lastTransitionTime: 2018-10-29T23:25:49Z
    lastUpdateTime: 2018-10-29T23:25:49Z
    message: Observed revision: 1
    reason: ReconcileSuccess
    status: "True"
    type: Settled
```
