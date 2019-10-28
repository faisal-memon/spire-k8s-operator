package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Selector struct {
	// Pod label names/values to match for this spiffe ID
	// To match, pods must be in the same namespace as this ID resource.
	PodLabel map[string]string `json:"podLabel"`
	PodName  string            `json:"podName"`
}

// SpiffeIdSpec defines the desired state of SpiffeId
// +k8s:openapi-gen=true
type SpiffeIdSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// The Spiffe ID to create
	SpiffeId string `json:"spiffeId"`

	// Selectors to match for this ID
	Selector Selector `json:"selector"`
}

// SpiffeIdStatus defines the observed state of SpiffeId
// +k8s:openapi-gen=true
type SpiffeIdStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// The spire Entry ID created for this Spiffe ID
	EntryId string `json:"entryId"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterSpiffeId is the Schema for the spiffeids API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusterspiffeids,scope=Cluster
type ClusterSpiffeId struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpiffeIdSpec   `json:"spec,omitempty"`
	Status SpiffeIdStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterSpiffeIdList contains a list of ClusterSpiffeId
type ClusterSpiffeIdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSpiffeId `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterSpiffeId{}, &ClusterSpiffeIdList{})
}
