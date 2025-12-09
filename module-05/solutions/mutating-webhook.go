// Solution: Mutating Webhook from Module 5
// This implements defaulting for the Database resource
// Location: internal/webhook/v1/database_webhook.go (add to existing file)

package v1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	databasev1 "github.com/example/postgres-operator/api/v1"
)

var databaselog = logf.Log.WithName("database-resource")

// +kubebuilder:webhook:path=/mutate-database-example-com-v1-database,mutating=true,failurePolicy=fail,sideEffects=None,groups=database.example.com,resources=databases,verbs=create;update,versions=v1,name=mdatabase-v1.kb.io,admissionReviewVersions=v1

// DatabaseCustomDefaulter struct is responsible for setting default values on the Database resource.
type DatabaseCustomDefaulter struct {
	// Add fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &DatabaseCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type Database.
func (d *DatabaseCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	database, ok := obj.(*databasev1.Database)
	if !ok {
		return fmt.Errorf("expected a Database object but got %T", obj)
	}
	databaselog.Info("Defaulting for Database", "name", database.GetName())

	// Set defaults based on namespace
	if database.Namespace == "production" {
		// Production defaults
		if database.Spec.Image == "" {
			database.Spec.Image = "postgres:14" // Stable version
		}
		if database.Spec.Replicas == nil {
			replicas := int32(3) // More replicas
			database.Spec.Replicas = &replicas
		}
	} else {
		// Development defaults
		if database.Spec.Image == "" {
			database.Spec.Image = "postgres:latest"
		}
		if database.Spec.Replicas == nil {
			replicas := int32(1)
			database.Spec.Replicas = &replicas
		}
	}

	// Common defaults
	if database.Spec.Storage.StorageClass == "" {
		database.Spec.Storage.StorageClass = "standard"
	}

	// Add labels (idempotent)
	if database.Labels == nil {
		database.Labels = make(map[string]string)
	}
	if _, exists := database.Labels["managed-by"]; !exists {
		database.Labels["managed-by"] = "database-operator"
	}

	// Add annotations (idempotent)
	if database.Annotations == nil {
		database.Annotations = make(map[string]string)
	}
	if _, exists := database.Annotations["database.example.com/version"]; !exists {
		database.Annotations["database.example.com/version"] = "v1"
	}

	return nil
}

// Note: To integrate this with your existing webhook, update SetupDatabaseWebhookWithManager:
//
// func SetupDatabaseWebhookWithManager(mgr ctrl.Manager) error {
//     return ctrl.NewWebhookManagedBy(mgr).For(&databasev1.Database{}).
//         WithValidator(&DatabaseCustomValidator{}).
//         WithDefaulter(&DatabaseCustomDefaulter{}).
//         Complete()
// }
