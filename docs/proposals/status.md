# Status

## Background
Application CRD has a list of GroupKinds and a LabelSelector to logically group live objects in the cluster. The Application CRD also provides a way to attach descriptive metadata to the logical grouping. 

## Objective
- Enhance Application CRD controller to provide status tracking for the objects that matches its selectors. 
- Simplify the status tracking requirements for a consumer of Application CRD. 
- Status aggregation should aim to reduce the entropy of information that the consumers of Application CRD's `.status` would otherwise have to deal with. If a consumer needs a more detailed status of the underlying components, they could query the underlying components status.
- Propose a minimal set of standardized Conditions for all resources to implement.

## Scope
The scope is limited to the objects matching the application crd selectors. The scope is also limited to using the information present in the `.status` field for kubernetes objects. External systems to augment health (black/white box) is not part of this scope.

## Proposal
The proposal is to add a reconciler to the Application CRD controller, that creates watches for the GroupKinds that the application CRD refers to. The reconciler then inspects the matching objects and updates the aggregate status in the Application CRD. Lowering information entropy and reducing the amount of interpertation to be done by the consumers of status is important for the adoption.

### Fields
Fields are a structured way to capture the status of resources. As part of this proposal, no standardized fields are proposed. The authors opinion is it does not provide additional information than what conditions provide. If the community strongly feels about standardizing some fields for in-tree and custom resources, it can be considered. 

### Conditions
[Conditions](https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#typical-status-properties) are used to describe the status of resources in an extensible way. Out of tree controllers support a variety of conditions. There is no consistent usage across the core controller. This is exacerbated with custom controllers. To tackle this diversity and its impact in status computation for Application CRD, we need a short term and a long term solution. Short term solution is to do aggregation of diverse core resources in the Application CRD controller. Long term is to identify best practices and standardize conditions and propagate them to the core controllers thereby making Application CRD aggregation simpler.

#### Standard Conditions
These conditions should be added to `.status.conditions` for Application CRD and subsequently to other resources:
##### `Ready`
Readiness as perceived by the resource controller. This means the controller deems the underlying resources to be ready for consumption to the best of its knowledge. It coule be implemented as a low pass filter to filter out transient or persistent failures such as restarts of underlying pods.

##### `Settled`
Would indicate whether controller acted on latest revision of `.spec` and the underlying resources conform to the latest revision of the `.spec`

##### `Error`
Would capture Errors seen in last reconcilation loop.

#### Other Conditions from core controllers
These are not as universal for all kubernetes objects as `Ready`, `Settled`, `Error` and provide yet another dimension of `Ready` and `Settled` condition. 

##### `Progressing`
Deployment uses this to indicate there is an update to the underlying ReplicaSet that is being acted upon. This is a inversion of the `Settled` condition.

##### `Available`
Deployment uses this to indicate that underlying pods have been `Ready` for a preconfigured amount of time. This is not a strong heuristic for actual availability but rather a low pass filter that fails when pods are restarting continiously.

#### Other Conditions
##### `Disruption`
`Disruption` is a special case of `Error` where we may know what caused the error when `Ready` is false. It is useful for some resources and could be standardized as an optional condition. 

### Aggregating status
#### Custom Resources
For custom resources, look for `Ready`, `Error` and `Settled`. If the object does not have the necessary conditions mark that objects status as "Unknown" and would not be used for computing the aggregate status. This would cover the case for nested Applciation CRDs as well. Based on feedback from users and reviewers this stance could be updated.

### Core controller resources
The logic to compute the readiness for the objects is based on the resource group-kind.
1. Statefulset is Ready if `.status.readyreplicas == .spec.replicas && .status.currentreplicas == .spec.replicas`
2. Service is Ready always (could not derive from `.status.loadbalancer`)
3. PodDisruptionBudget is Ready if .`status.currenthealthy >= .status.desiredhealthy`
... (todo other core resources)

The computed status conditions are injected as conditions in application's `status.components`. The long term plan is for this logic to be added to core-controllers and them to update the `.status.conditions` for the individual objects. 

### Aggregating status
Application CRD's `.status.conditions.[Ready]` is derived by a simple ANDing of the matching object status `status.conditions.[Ready]` computed under `.status.components`. `Settled` is implemented by just the Application CRD controller and is not a aggregated value. `Error` condition aggregates errors from underlying components.

## API
### Conditions API
```go
package status

// Constants
const (
	// Ready => Resource's controller considers this resource Ready
	ConditionReady = "Ready"
	// Error => last observed error as part of reconciliation by controller
	ConditionError = "Error"
	// Settled => observed generation == generation && controller is done satisfying the intent in .spec
	// This is a high fidelity condition that is sensitive to .spec mutations
	ConditionSettled = "Settled"
	// Cleanup => it is set to track finalizer failures
	ConditionCleanup = "Cleanup"
	
	// Availability via black/whitebox probes
	// TrafficReady => pod sees external traffic (blackbox/whitebox readiness - out of scope)
	ConditionTrafficReady = "TrafficReady"
	// Qualified => functionally tested (blackbox/whitebox readiness - out of scope)
	ConditionQualified = "Qualified"
	
	// Scale related condition
	// 
	ConditionScale "Scale"
	
	// Condition status
	ConditionTrue    = "True"
	ConditionFalse   = "False"
	ConditionUnknown = "Unknown"
)

// ConditionType encodes information on the condition
type ConditionType string

// Condition describes the state of an object at a certain point.
// +k8s:deepcopy-gen=true
type Condition struct {
	// Type of condition.
	Type ConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=StatefulSetConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
	// Last time the condition was probed
	// +optional
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
}
```

### API Component list
Matching objects' status would be recorded in `.status.components`.  The motivation behind this is to provide individual component status for clients that need more than just the aggregate status. Well known object types can be aggregated to a higher degree of fidelity. For other types and custom resources we can propose using standardized conditions.

```go
package status

// ComponentList is a generic status holder for the top level resource
// +k8s:deepcopy-gen=true
type ComponentList struct {
	// Object status array for all matching objects
	Objects []ObjectStatus `json:"components,omitempty"`
}

// ObjectStatus is a generic status holder for objects
// +k8s:deepcopy-gen=true
type ObjectStatus struct {
	// Link to object
	Link string `json:"link,omitempty"`
	// Name of object
	Name string `json:"name,omitempty"`
	// Kind of object
	Kind string `json:"kind,omitempty"`
	// Object group
	Group string `json:"group,omitempty"`
        // Conditions represents the latest state of the object
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// ExtendedStatus adds Kind specific status information for well known types
	ExtendedStatus `json:",inline,omitempty"`
}

// ExtendedStatus is a holder of additional status for well known types
// +k8s:deepcopy-gen=true
type ExtendedStatus struct {
	// StatefulSet status
	STS *Statefulset `json:"sts,omitempty"`
	// PDB status
	PDB *Pdb `json:"pdb,omitempty"`
}

// Statefulset is a generic status holder for stateful-set
// +k8s:deepcopy-gen=true
type Statefulset struct {
	// Replicas defines the no of MySQL instances desired
	Replicas int32 `json:"replicas"`
	// ReadyReplicas defines the no of MySQL instances that are ready
	ReadyReplicas int32 `json:"readycount"`
	// CurrentReplicas defines the no of MySQL instances that are created
	CurrentReplicas int32 `json:"currentcount"`
	// progress is a fuzzy indicator. Interpret as a percentage (0-100)
	// eg: for statefulsets, progress = 100*readyreplicas/replicas
	Progress int32 `json:"progress"`
}

// Pdb is a generic status holder for pdb
type Pdb struct {
	// currentHealthy
	CurrentHealthy int32 `json:"currenthealthy"`
	// desiredHealthy
	DesiredHealthy int32 `json:"desiredhealthy"`
}
``` 

### API Status meta
Meta is an aggregate of the common status fields that should be inlined in the resource status struct.

```go
package status

// Meta is a generic set of fields for status objects
// +k8s:deepcopy-gen=true
type Meta struct {
	// ObservedGeneration is the most recent generation observed. It corresponds to the
	// Object's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`
	// Conditions represents the latest state of the object
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=conditions"`
	// Resources embeds a list of object statuses
	// +optional
	ComponentList `json:",inline,omitempty"`
}
``` 

### API Application status
```go
package application
// ApplicationStatus defines controllers the observed state of Application
type ApplicationStatus struct {
	status.Meta `json:",inline"`	
}
```


### Example
```
status:
  components:
  - link: /apis/apps/v1/namespaces/default/statefulsets/basic-di
    name: basic-di
    status: Ready
    sts:
      currentcount: 3
      progress: 0
      readycount: 3
      replicas: 3
  - link: /api/v1/namespaces/default/services/basic-di
    name: basic-di
    status: Ready
  - link: /api/v1/namespaces/default/configmaps/basic-di
    name: basic-di
    status: Ready
  - link: /apis/apps/v1/namespaces/default/statefulsets/basic-m
    name: basic-m
    status: Ready
    sts:
      currentcount: 3
      progress: 0
      readycount: 3
      replicas: 3
  - link: /api/v1/namespaces/default/services/basic-m
    name: basic-m
    status: Ready
  - link: /api/v1/namespaces/default/configmaps/basic-m
    name: basic-m
    status: Ready
  - link: /api/v1/namespaces/default/services/basic-discovery
    name: basic-discovery
    status: Ready
  - link: /api/v1/namespaces/default/services/basic-master
    name: basic-master
    status: Ready
  - link: /api/v1/namespaces/default/services/basic-data
    name: basic-data
    status: Ready
  - link: /api/v1/namespaces/default/services/basic-ingest
    name: basic-ingest
    status: Ready
  - link: /api/v1/namespaces/default/services/basic-metrics
    name: basic-metrics
    status: Ready
  - link: /apis/apps/v1/namespaces/default/statefulsets/basic-metrics
    name: basic-metrics
    status: Ready
    sts:
      currentcount: 1
      progress: 0
      readycount: 1
      replicas: 1
  - link: /apis/app.k8s.io/v1beta1/namespaces/default/applications/basic
    name: basic
    status: Ready
  conditions:
  - lastTransitionTime: 2018-11-05T18:00:02Z
    lastUpdateTime: 2018-11-05T18:00:02Z
    message: all components ready
    reason: ComponentsReady
    status: "True"
    type: Ready
  - lastTransitionTime: 2018-11-05T08:22:08Z
    lastUpdateTime: 2018-11-05T08:22:08Z
    message: Converged to version: 323221
    reason: Settled
    status: "True"
    type: Settled
  - lastTransitionTime: 2018-11-05T08:22:12Z
    lastUpdateTime: 2018-11-05T08:22:12Z
    message: 'Operation cannot be fulfilled on statefulsets.apps "basic-metrics":
      the object has been modified; please apply your changes to the latest version
      and try again'
    reason: ErrorSeen
    status: "False"
    type: Error
```
