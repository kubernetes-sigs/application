


// Api versions allow the api contract for a resource to be changed while keeping
// backward compatibility by support multiple concurrent versions
// of the same resource

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=github.com/kubernetes-sigs/apps_application/pkg/apis/application
// +k8s:defaulter-gen=TypeMeta
// +groupName=application.k8s.io
package v1alpha1 // import "github.com/kubernetes-sigs/apps_application/pkg/apis/application/v1alpha1"
