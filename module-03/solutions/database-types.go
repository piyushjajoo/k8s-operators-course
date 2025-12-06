// Solution: Complete Database API Types from Module 3
// This defines the Database Custom Resource structure

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
	// Image is the PostgreSQL image to use
	// +kubebuilder:validation:Required
	// +kubebuilder:default="postgres:14"
	Image string `json:"image"`

	// Replicas is the number of database replicas
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Storage is the storage configuration
	Storage StorageSpec `json:"storage"`

	// Resources are the resource requirements
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// DatabaseName is the name of the database to create
	// +kubebuilder:validation:Required
	DatabaseName string `json:"databaseName"`

	// Username is the database user
	// +kubebuilder:validation:Required
	Username string `json:"username"`
}

// StorageSpec defines storage configuration
type StorageSpec struct {
	// Size is the storage size (e.g., "10Gi")
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[0-9]+(Gi|Mi)$`
	Size string `json:"size"`

	// StorageClass is the storage class to use
	StorageClass string `json:"storageClass,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
	// Phase is the current phase
	// +kubebuilder:validation:Enum=Pending;Creating;Ready;Failed
	Phase string `json:"phase,omitempty"`

	// Ready indicates if the database is ready
	Ready bool `json:"ready,omitempty"`

	// Endpoint is the database endpoint
	Endpoint string `json:"endpoint,omitempty"`

	// SecretName is the name of the Secret containing database credentials
	SecretName string `json:"secretName,omitempty"`

	// Conditions represent the latest observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Database is the Schema for the databases API
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec,omitempty"`
	Status DatabaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
