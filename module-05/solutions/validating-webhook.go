// Solution: Validating Webhook from Module 5
// This implements custom validation for the Database resource
// Location: internal/webhook/v1/database_webhook.go

package v1

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

var databaselog = logf.Log.WithName("database-resource")

// SetupDatabaseWebhookWithManager registers the webhook for Database in the manager.
func SetupDatabaseWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&databasev1.Database{}).
		WithValidator(&DatabaseCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-database-example-com-v1-database,mutating=false,failurePolicy=fail,sideEffects=None,groups=database.example.com,resources=databases,verbs=create;update,versions=v1,name=vdatabase-v1.kb.io,admissionReviewVersions=v1

// DatabaseCustomValidator struct is responsible for validating the Database resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type DatabaseCustomValidator struct {
	// Add more fields as needed for validation
}

var _ webhook.CustomValidator = &DatabaseCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Database.
func (v *DatabaseCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	database, ok := obj.(*databasev1.Database)
	if !ok {
		return nil, fmt.Errorf("expected a Database object but got %T", obj)
	}
	databaselog.Info("Validation for Database upon creation", "name", database.GetName())

	var errors []string

	// Validate image is PostgreSQL
	if !strings.Contains(database.Spec.Image, "postgres") {
		errors = append(errors, fmt.Sprintf("spec.image: must be a PostgreSQL image, got '%s'. Valid examples: postgres:14, postgres:13", database.Spec.Image))
	}

	// Validate replicas and storage relationship
	if database.Spec.Replicas != nil && *database.Spec.Replicas > 5 {
		if database.Spec.Storage.Size == "10Gi" {
			errors = append(errors, fmt.Sprintf("spec.storage.size: when replicas > 5, storage must be >= 50Gi, got '%s'", database.Spec.Storage.Size))
		}
	}

	// Validate database name format
	if len(database.Spec.DatabaseName) > 63 {
		errors = append(errors, fmt.Sprintf("spec.databaseName: must be <= 63 characters, got %d", len(database.Spec.DatabaseName)))
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Database.
func (v *DatabaseCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	database, ok := newObj.(*databasev1.Database)
	if !ok {
		return nil, fmt.Errorf("expected a Database object for the newObj but got %T", newObj)
	}
	oldDB, ok := oldObj.(*databasev1.Database)
	if !ok {
		return nil, fmt.Errorf("expected a Database object for the oldObj but got %T", oldObj)
	}
	databaselog.Info("Validation for Database upon update", "name", database.GetName())

	var errors []string

	// Prevent reducing storage size
	oldSize := parseStorageSize(oldDB.Spec.Storage.Size)
	newSize := parseStorageSize(database.Spec.Storage.Size)

	if newSize < oldSize {
		errors = append(errors, fmt.Sprintf("spec.storage.size: cannot reduce storage from %s to %s", oldDB.Spec.Storage.Size, database.Spec.Storage.Size))
	}

	// Prevent changing database name
	if oldDB.Spec.DatabaseName != database.Spec.DatabaseName {
		errors = append(errors, fmt.Sprintf("spec.databaseName: cannot change from %s to %s", oldDB.Spec.DatabaseName, database.Spec.DatabaseName))
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Database.
func (v *DatabaseCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	database, ok := obj.(*databasev1.Database)
	if !ok {
		return nil, fmt.Errorf("expected a Database object but got %T", obj)
	}
	databaselog.Info("Validation for Database upon deletion", "name", database.GetName())

	// Add any deletion validation logic
	return nil, nil
}

// Helper function to parse storage size (e.g., "10Gi" -> 10)
func parseStorageSize(size string) int64 {
	if strings.HasSuffix(size, "Gi") {
		num := strings.TrimSuffix(size, "Gi")
		val, err := strconv.ParseInt(num, 10, 64)
		if err != nil {
			return 0
		}
		return val
	}
	if strings.HasSuffix(size, "Mi") {
		num := strings.TrimSuffix(size, "Mi")
		val, err := strconv.ParseInt(num, 10, 64)
		if err != nil {
			return 0
		}
		// Convert Mi to Gi equivalent (1/1024)
		return val / 1024
	}
	return 0
}
