package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Selector struct {
	PodLabel map[string]string `json:"podLabel"`
}

// SpiffeIdSpec defines the desired state of SpiffeId
// +k8s:openapi-gen=true
type SpiffeIdSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	SpiffeId string `json:"spiffeId"`
	Selector Selector `json:"selector"`
}

// SpiffeIdStatus defines the observed state of SpiffeId
// +k8s:openapi-gen=true
type SpiffeIdStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	EntryId string `json:"entryId"`
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

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SpiffeIdList contains a list of SpiffeId
type SpiffeIdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpiffeId `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SpiffeId{}, &SpiffeIdList{})
}
