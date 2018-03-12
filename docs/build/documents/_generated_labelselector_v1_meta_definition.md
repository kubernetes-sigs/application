## LabelSelector v1 meta

Group        | Version     | Kind
------------ | ---------- | -----------
meta | v1 | LabelSelector



A label selector is a label query over a set of resources. The result of matchLabels and matchExpressions are ANDed. An empty label selector matches all objects. A null label selector matches no objects.

<aside class="notice">
Appears In:

<ul> 
<li><a href="#applicationspec-v1alpha1-application">ApplicationSpec application/v1alpha1</a></li>
</ul></aside>

Field        | Description
------------ | -----------
matchExpressions <br /> *[LabelSelectorRequirement](#labelselectorrequirement-v1-meta) array*    | matchExpressions is a list of label selector requirements. The requirements are ANDed.
matchLabels <br /> *object*    | matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

