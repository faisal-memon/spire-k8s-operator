package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type v1Object = metav1.Object
type runtimeObject = runtime.Object

type CommonSpiffeId interface {
	v1Object
	runtimeObject
	GetStatus() *SpiffeIdStatus
}

type Selector struct {
	// Pod label names/values to match for this spiffe ID
	// To match, pods must be in the same namespace as this ID resource.
	PodLabel map[string]string `json:"podLabel,omitempty"`
	PodName  string            `json:"podName,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	ServiceAccount string `json:"serviceAccount,omitempty"`
	Arbitrary []string `json:"arbitrary,omitempty"`
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

func (in *ClusterSpiffeId) GetStatus() *SpiffeIdStatus {
	return &in.Status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterSpiffeIdList contains a list of ClusterSpiffeId
type ClusterSpiffeIdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterSpiffeId `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpiffeId is the Schema for the spiffeids API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=spiffeids,scope=Namespaced
type SpiffeId struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpiffeIdSpec   `json:"spec,omitempty"`
	Status SpiffeIdStatus `json:"status,omitempty"`


}

func (in *SpiffeId) GetStatus() *SpiffeIdStatus {
	return &in.Status
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpiffeIdList contains a list of spiffeId
type SpiffeIdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpiffeId `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterSpiffeId{}, &ClusterSpiffeIdList{})
	SchemeBuilder.Register(&SpiffeId{}, &SpiffeIdList{})
}
