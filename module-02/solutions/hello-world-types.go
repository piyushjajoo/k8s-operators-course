// Solution: Complete HelloWorld API Types from Module 2
// This defines the Custom Resource structure

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelloWorldSpec defines the desired state of HelloWorld
type HelloWorldSpec struct {
	// Message is the message to display
	Message string `json:"message,omitempty"`

	// Count is the number of times to display the message
	Count int32 `json:"count,omitempty"`
}

// HelloWorldStatus defines the observed state of HelloWorld
type HelloWorldStatus struct {
	// Phase represents the current phase
	Phase string `json:"phase,omitempty"`

	// ConfigMapCreated indicates if the ConfigMap was created
	ConfigMapCreated bool `json:"configMapCreated,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HelloWorld is the Schema for the helloworlds API
type HelloWorld struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelloWorldSpec   `json:"spec,omitempty"`
	Status HelloWorldStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HelloWorldList contains a list of HelloWorld
type HelloWorldList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelloWorld `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HelloWorld{}, &HelloWorldList{})
}
