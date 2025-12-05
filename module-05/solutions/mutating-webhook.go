// Solution: Mutating Webhook from Module 5
// This implements defaulting for the Database resource

package v1

import (
    "sigs.k8s.io/controller-runtime/pkg/webhook"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var databaseLog = logf.Log.WithName("database-resource")

// +kubebuilder:webhook:path=/mutate-database-example-com-v1-database,mutating=true,failurePolicy=fail,sideEffects=None,groups=database.example.com,resources=databases,verbs=create;update,versions=v1,name=mdatabase.kb.io

var _ webhook.Defaulter = &Database{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Database) Default() {
    databaseLog.Info("default", "name", r.Name)

    // Set defaults based on namespace
    if r.Namespace == "production" {
        // Production defaults
        if r.Spec.Image == "" {
            r.Spec.Image = "postgres:14" // Stable version
        }
        if r.Spec.Replicas == nil {
            replicas := int32(3) // More replicas
            r.Spec.Replicas = &replicas
        }
    } else {
        // Development defaults
        if r.Spec.Image == "" {
            r.Spec.Image = "postgres:latest"
        }
        if r.Spec.Replicas == nil {
            replicas := int32(1)
            r.Spec.Replicas = &replicas
        }
    }

    // Common defaults
    if r.Spec.Storage.StorageClass == "" {
        r.Spec.Storage.StorageClass = "standard"
    }

    // Add labels (idempotent)
    if r.Labels == nil {
        r.Labels = make(map[string]string)
    }
    if _, exists := r.Labels["managed-by"]; !exists {
        r.Labels["managed-by"] = "database-operator"
    }

    // Add annotations (idempotent)
    if r.Annotations == nil {
        r.Annotations = make(map[string]string)
    }
    if _, exists := r.Annotations["database.example.com/version"]; !exists {
        r.Annotations["database.example.com/version"] = "v1"
    }
}

