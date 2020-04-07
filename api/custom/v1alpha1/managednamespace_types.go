/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ManagedNamespaceSpec defines the desired state of ManagedNamespace
type ManagedNamespaceSpec struct {
	namespace.Namespace `json:",inline"`
}

// ManagedNamespaceStatus defines the observed state of ManagedNamespace
type ManagedNamespaceStatus struct {
	//State of the resource
	State State `json:"state,omitempty"`
	//RetryCount in case of error
	RetryCount int `json:"retryCount"`
	//ErrorDescription in case of error
	ErrorDescription string `json:"errorDescription,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="current state of the managed namespace"
// +kubebuilder:printcolumn:name="RetryCount",type="integer",JSONPath=".status.retryCount",description="Retry count"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="time passed since managed namespace created"
// +kubebuilder:resource:path=managednamespaces,scope=Namespaced,shortName=mns,singular=managednamespace
// ManagedNamespace is the Schema for the managednamespaces API
type ManagedNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedNamespaceSpec   `json:"spec,omitempty"`
	Status ManagedNamespaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ManagedNamespaceList contains a list of ManagedNamespace
type ManagedNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedNamespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedNamespace{}, &ManagedNamespaceList{})
}
