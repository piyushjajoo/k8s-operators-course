// Solution: Validating Webhook from Module 5
// This implements custom validation for the Database resource

package v1

import (
    "fmt"
    "strings"

    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
    "sigs.k8s.io/controller-runtime/pkg/webhook"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var databaseLog = logf.Log.WithName("database-resource")

func (r *Database) SetupWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManager().
        For(r).
        Complete()
}

//+kubebuilder:webhook:path=/validate-database-example-com-v1-database,mutating=false,failurePolicy=fail,sideEffects=None,groups=database.example.com,resources=databases,verbs=create;update,versions=v1,name=vdatabase.kb.io

var _ webhook.Validator = &Database{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateCreate() (admission.Warnings, error) {
    databaseLog.Info("validate create", "name", r.Name)

    var errors []string

    // Validate image is PostgreSQL
    if !strings.Contains(r.Spec.Image, "postgres") {
        errors = append(errors, fmt.Sprintf("spec.image: must be a PostgreSQL image, got '%s'. Valid examples: postgres:14, postgres:13", r.Spec.Image))
    }

    // Validate replicas and storage relationship
    if r.Spec.Replicas != nil && *r.Spec.Replicas > 5 {
        if r.Spec.Storage.Size == "10Gi" {
            errors = append(errors, fmt.Sprintf("spec.storage.size: when replicas > 5, storage must be >= 50Gi, got '%s'", r.Spec.Storage.Size))
        }
    }

    // Validate database name format
    if len(r.Spec.DatabaseName) > 63 {
        errors = append(errors, fmt.Sprintf("spec.databaseName: must be <= 63 characters, got %d", len(r.Spec.DatabaseName)))
    }

    if len(errors) > 0 {
        return nil, fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }

    return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
    databaseLog.Info("validate update", "name", r.Name)

    oldDB := old.(*Database)
    var errors []string

    // Prevent reducing storage size
    oldSize := parseStorageSize(oldDB.Spec.Storage.Size)
    newSize := parseStorageSize(r.Spec.Storage.Size)

    if newSize < oldSize {
        errors = append(errors, fmt.Sprintf("spec.storage.size: cannot reduce storage from %s to %s", oldDB.Spec.Storage.Size, r.Spec.Storage.Size))
    }

    // Prevent changing database name
    if oldDB.Spec.DatabaseName != r.Spec.DatabaseName {
        errors = append(errors, fmt.Sprintf("spec.databaseName: cannot change from %s to %s", oldDB.Spec.DatabaseName, r.Spec.DatabaseName))
    }

    if len(errors) > 0 {
        return nil, fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }

    return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Database) ValidateDelete() (admission.Warnings, error) {
    databaseLog.Info("validate delete", "name", r.Name)
    // Add any deletion validation logic
    return nil, nil
}

// Helper function to parse storage size (simplified)
func parseStorageSize(size string) int64 {
    // In production, use proper parsing
    // This is a simplified example
    if strings.HasSuffix(size, "Gi") {
        // Parse number and convert
        return 0 // Implement proper parsing
    }
    return 0
}

