package v1alpha1

import (
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceTemplateSpec defines the spec for NamespaceTemplate
type NamespaceTemplateSpec struct {
	namespace.NamespaceTemplate `json:",inline"`
}

// NamespaceTemplateStatus defines the status for NamespaceTemplate resource
type NamespaceTemplateStatus struct {
}

// +kubebuilder:object:generate=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=namespacetemplate,scope=Cluster,shortName=nt,singular=namespacetemplate
// NamespaceTemplate is the Schema for the namespacetemplate API
type NamespaceTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespaceTemplateSpec   `json:"spec,omitempty"`
	Status NamespaceTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:object:generate=true
// NamespaceTemplateList contains a list of NamespaceTemplate
type NamespaceTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespaceTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NamespaceTemplate{}, &NamespaceTemplateList{})
}
