# Inventory Object Proposal

## Motivation
The proposal is to seperate the Application CRD's object selector and group kinds into a separate object. Separating the resource grouping and Application info allows us to reuse resoruce grouping (inventory object) in other controllers and tools. The Inventory object is a descriptive object and needs no controller and status field.

## Summary

- Introduce an Inventory object and a data format that is used to group objects
- Allow selector and groupKind based grouping
- Allow explicit object list based grouping
- Reference array of Inventory objects in Application object

## Proposal

We propose to introduce a separate Inventory/Grouping object. It could be a new API Type or use ConfigMap to achieve the same. 
There are multiple ways of grouping objects:
1. (existing) Selectors + List of GroupKinds
2. Explicity list of GroupKindNamespaceName for cross namespace grouping
3. Explicity list of GroupKindName for grouping within a namespace

The Application object would have a new field `.spec.inventory` which would point to a list of grouping objects.
```golang

ApplicationSpec {
      ... 
      Inventory []GroupingObject
      ...

}

// GroupingObjectType is a string
type GroupingObjectType string

// Constants for info type
const (
	ConfigMapGroupingObject         GroupingObjectType = "ConfigMap"
	ClusterConfigMapGroupingObject  GroupingObjectType = "ClusterConfigMap"
	InventoryGroupingObject         GroupingObjectType = "Inventory"
)
GroupingObject {
     Type GroupingObjectType
     Reference corev1.ObjectReference `json:",inline"`
}

```

### Using Inventory API Type

A new inventory type is introduced called `Inventory`. This requires an administrator to install the CRD in the cluster. This hinders windespread adoption.

```golang

type InventorySpec {
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
	ComponentGroupKinds []metav1.GroupKind `json:"componentKinds,omitempty"`
	Objects []corev1.ObjectReference `json:",inline"`
}

```

### Using ConfigMap

ConfigMap is ubiquitous in all clusters. No new CRDs need to be installed which require admin priveleges. 
Kustomize has a well defined inventory format that relies on annotations. We could reuse it. It only supports explicit list of objects.
