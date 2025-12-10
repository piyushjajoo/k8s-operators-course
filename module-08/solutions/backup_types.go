// Solution: Backup Types from Module 8
// This file contains the complete API type definitions for the Backup resource.
// Use kubebuilder to scaffold the API first, then replace the generated types with these.
//
// Scaffold with:
//   kubebuilder create api --group backup --version v1 --kind Backup --resource --controller

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupSpec defines the desired state of Backup
type BackupSpec struct {
	// DatabaseRef references the Database to backup
	// +kubebuilder:validation:Required
	DatabaseRef corev1.LocalObjectReference `json:"databaseRef"`

	// Schedule is the cron schedule for automated backups (optional)
	// If not specified, backup is a one-time operation
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Retention is the number of backups to retain
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=5
	// +optional
	Retention int `json:"retention,omitempty"`

	// StorageLocation is where to store the backup (e.g., s3://bucket/path)
	// +optional
	StorageLocation string `json:"storageLocation,omitempty"`
}

// BackupStatus defines the observed state of Backup
type BackupStatus struct {
	// Phase is the current backup phase
	// +kubebuilder:validation:Enum=Pending;InProgress;Completed;Failed
	Phase string `json:"phase,omitempty"`

	// BackupTime is when the backup was created
	BackupTime *metav1.Time `json:"backupTime,omitempty"`

	// BackupLocation is where the backup is stored
	BackupLocation string `json:"backupLocation,omitempty"`

	// LastScheduledTime is when the last scheduled backup was triggered
	LastScheduledTime *metav1.Time `json:"lastScheduledTime,omitempty"`

	// BackupCount is the number of successful backups stored
	BackupCount int `json:"backupCount,omitempty"`

	// Conditions represent the latest observations of the Backup's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Database",type="string",JSONPath=".spec.databaseRef.name"
// +kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule"
// +kubebuilder:printcolumn:name="Last Backup",type="date",JSONPath=".status.backupTime"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Backup is the Schema for the backups API.
// It manages backup operations for Database resources.
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSpec   `json:"spec,omitempty"`
	Status BackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackupList contains a list of Backup
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backup{}, &BackupList{})
}

