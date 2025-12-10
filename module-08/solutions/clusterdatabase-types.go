// Solution: ClusterDatabase API Types from Module 8
// This defines a Cluster-Scoped Database Custom Resource structure
// Unlike the namespace-scoped Database, ClusterDatabase manages databases across all namespaces

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterDatabaseSpec defines the desired state of ClusterDatabase
type ClusterDatabaseSpec struct {
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

	// TargetNamespace is the namespace where resources will be created
	// Required for cluster-scoped resources to know where to create managed resources
	// +kubebuilder:validation:Required
	TargetNamespace string `json:"targetNamespace"`

	// Tenant identifies which tenant owns this database (for multi-tenancy)
	// +optional
	Tenant string `json:"tenant,omitempty"`
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

// ClusterDatabaseStatus defines the observed state of ClusterDatabase
type ClusterDatabaseStatus struct {
	// Phase is the current phase
	// +kubebuilder:validation:Enum=Pending;Creating;Ready;Failed
	Phase string `json:"phase,omitempty"`

	// Ready indicates if the database is ready
	Ready bool `json:"ready,omitempty"`

	// Endpoint is the database endpoint
	Endpoint string `json:"endpoint,omitempty"`

	// SecretName is the name of the Secret containing database credentials
	SecretName string `json:"secretName,omitempty"`

	// TargetNamespace shows where the managed resources were created
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// Conditions represent the latest observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".spec.targetNamespace"
// +kubebuilder:printcolumn:name="Tenant",type="string",JSONPath=".spec.tenant"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ClusterDatabase is the Schema for the clusterdatabases API
// It is cluster-scoped and manages databases across namespaces
type ClusterDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterDatabaseSpec   `json:"spec,omitempty"`
	Status ClusterDatabaseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterDatabaseList contains a list of ClusterDatabase
type ClusterDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterDatabase{}, &ClusterDatabaseList{})
}

