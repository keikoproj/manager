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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterSpec defines the desired state of Cluster

type ClusterSpec struct {
	//Name contains cluster name
	Name string `json:"name"`
	//Type contains kubernetes cluster installation type. ex: AWS, GCP
	// +optional
	Cloud string `json:"type,omitempty"`

	//Config contains info to connect to the target cluster
	//This is same as config struct in https://github.com/kubernetes/client-go/blob/master/rest/config.go
	// but have to define it again here with whatever we need
	// +optional
	Config Config `json:"config,omitempty"`
}

// Config holds the common attributes that can be passed to a Kubernetes client on
// initialization.
// +optional
type Config struct {
	// Host must be a host string, a host:port pair, or a URL to the base of the apiserver.
	// If a URL is given then the (optional) Path of that URL represents a prefix that must
	// be appended to all request URIs used to access the apiserver. This allows a frontend
	// proxy to easily relocate all of the apiserver endpoints.
	Host string `json:"host"`

	// Server requires Basic authentication
	// +optional
	Username string `json:"username,omitempty"`
	// password contains basic auth password
	// +optional
	Password string `json:"password,omitempty"`

	// Secret containing a BearerToken.
	// If set, The last successfully read value takes precedence over BearerToken.
	// +optional
	BearerTokenSecret string `json:"bearerTokenSecret,omitempty"`

	// TLSClientConfig contains settings to enable transport layer security
	// +optional
	TLSClientConfig `json:"tlsClientConfig,omitempty"`
}

// TLSClientConfig contains settings to enable transport layer security
// +optional
type TLSClientConfig struct {
	// Server should be accessed without verifying the TLS certificate. For testing only.
	// +optional
	Insecure bool `json:"inSecure,omitempty"`
	// ServerName is passed to the server for SNI and is used in the client to check server
	// ceritificates against. If ServerName is empty, the hostname used to contact the
	// server is used.
	// +optional
	ServerName string `json:"serverName,omitempty"`

	// CertData holds PEM-encoded bytes (typically read from a client certificate file).
	// CertData takes precedence over CertFile
	// +optional
	CertData []byte `json:"certData,omitempty"`
	// KeyData holds PEM-encoded bytes (typically read from a client certificate key file).
	// KeyData takes precedence over KeyFile
	// +optional
	KeyData []byte `json:"keyData,omitempty"`
	// CAData holds PEM-encoded bytes (typically read from a root certificates bundle).
	// CAData takes precedence over CAFile
	// +optional
	CAData []byte `json:"caData,omitempty"`

	// NextProtos is a list of supported application level protocols, in order of preference.
	// Used to populate tls.Config.NextProtos.
	// To indicate to the server http/1.1 is preferred over http/2, set to ["http/1.1", "h2"] (though the server is free to ignore that preference).
	// To use only http/1.1, set to ["http/1.1"].
	// +optional
	NextProtos []string `json:"nextProtos,omitempty"`
}

type State string

const (
	Ready                State = "Ready"
	Warning              State = "Warning"
	Error                State = "Error"
	PolicyNotAllowed     State = "PolicyNotAllowed"
	RolesMaxLimitReached State = "RolesMaxLimitReached"
)

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	//State of the resource
	State State `json:"state,omitempty"`
	//RetryCount in case of error
	RetryCount int `json:"retryCount"`
	//ErrorDescription in case of error
	ErrorDescription string `json:"errorDescription,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusters,scope=Namespaced,shortName=cl,singular=cluster
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state",description="current state of the target cluster"
// +kubebuilder:printcolumn:name="RetryCount",type="integer",JSONPath=".status.retryCount",description="Retry count"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="time passed since managed cluster registration"
// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
