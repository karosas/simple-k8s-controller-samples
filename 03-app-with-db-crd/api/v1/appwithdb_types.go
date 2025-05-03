package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AppWithDbSpec defines the desired state of AppWithDb.
type AppWithDbSpec struct {
	Image string `json:"image"`
}

// AppWithDbStatus defines the observed state of AppWithDb.
type AppWithDbStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type AppWithDb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppWithDbSpec   `json:"spec,omitempty"`
	Status AppWithDbStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type AppWithDbList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppWithDb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppWithDb{}, &AppWithDbList{})
}
