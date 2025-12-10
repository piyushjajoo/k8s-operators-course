// Solution: Backup Implementation from Module 8
// This demonstrates backup functionality for stateful applications
//
// IMPORTANT: This implementation uses pg_dump which requires PostgreSQL client tools
// to be installed in your operator container image. Update your Dockerfile to include
// postgresql-client package. See the Dockerfile example in this solutions directory.

package backup

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	databasev1 "github.com/example/postgres-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func PerformBackup(ctx context.Context, k8sClient client.Client, db *databasev1.Database) (string, error) {
	// Connect to database
	endpoint := db.Status.Endpoint
	if endpoint == "" {
		return "", fmt.Errorf("database endpoint not available")
	}

	// Get password from Secret
	secretName := db.Status.SecretName
	if secretName == "" {
		// Fallback to default secret name pattern if SecretName not set in status
		secretName = fmt.Sprintf("%s-credentials", db.Name)
	}

	secret := &corev1.Secret{}
	err := k8sClient.Get(ctx, client.ObjectKey{
		Name:      secretName,
		Namespace: db.Namespace,
	}, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", fmt.Errorf("secret %s not found", secretName)
		}
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	// Extract password from Secret
	passwordBytes, exists := secret.Data["password"]
	if !exists {
		return "", fmt.Errorf("password key not found in secret %s", secretName)
	}
	password := string(passwordBytes)

	// Create backup filename
	backupFile := fmt.Sprintf("/backups/%s-%s.sql",
		db.Name,
		time.Now().Format("20060102-150405"))

	// Perform pg_dump with password from Secret
	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", endpoint,
		"-U", db.Spec.Username,
		"-d", db.Spec.DatabaseName,
		"-f", backupFile)

	// Set password as environment variable for pg_dump
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("backup failed: %v, output: %s", err, string(output))
	}

	// Save to storage (S3, PVC, etc.)
	backupLocation, err := saveToStorage(backupFile)
	if err != nil {
		return "", fmt.Errorf("failed to save backup: %v", err)
	}

	return backupLocation, nil
}

func saveToStorage(backupFile string) (string, error) {
	// In real implementation, this would:
	// 1. Upload to S3
	// 2. Copy to PVC
	// 3. Store in object storage

	// For example, upload to S3:
	// s3Location := fmt.Sprintf("s3://backups/%s", backupFile)
	// aws s3 cp backupFile s3Location

	// Simplified for example
	return fmt.Sprintf("s3://backups/%s", backupFile), nil
}

func PerformScheduledBackup(ctx context.Context, db *databasev1.Database, schedule string) error {
	// Parse cron schedule
	// Calculate next backup time
	// Schedule backup job

	// For Kubernetes, you might use CronJob
	// For operator, you can use RequeueAfter

	return nil
}

