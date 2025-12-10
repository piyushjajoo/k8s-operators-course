// Solution: Restore Types from Module 8
// This file contains the complete API type definitions for the Restore resource.
// Use kubebuilder to scaffold the API first, then replace the generated types with these.
//
// Scaffold with (same group as Database and Backup):
//   kubebuilder create api --group database --version v1 --kind Restore --resource --controller

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestoreSpec defines the desired state of Restore
type RestoreSpec struct {
	// DatabaseRef references the Database to restore to
	// +kubebuilder:validation:Required
	DatabaseRef corev1.LocalObjectReference `json:"databaseRef"`

	// BackupRef references the Backup to restore from
	// +kubebuilder:validation:Required
	BackupRef corev1.LocalObjectReference `json:"backupRef"`
}

// RestoreStatus defines the observed state of Restore
type RestoreStatus struct {
	// Phase is the current restore phase
	// +kubebuilder:validation:Enum=Pending;InProgress;Completed;Failed
	Phase string `json:"phase,omitempty"`

	// RestoreTime is when the restore completed
	RestoreTime *metav1.Time `json:"restoreTime,omitempty"`

	// Message provides additional information about the current phase
	Message string `json:"message,omitempty"`

	// Conditions represent the latest observations of the Restore's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Database",type="string",JSONPath=".spec.databaseRef.name"
// +kubebuilder:printcolumn:name="Backup",type="string",JSONPath=".spec.backupRef.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Restore is the Schema for the restores API.
// It restores a Database from a Backup.
type Restore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RestoreSpec   `json:"spec,omitempty"`
	Status RestoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RestoreList contains a list of Restore
type RestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Restore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Restore{}, &RestoreList{})
}

