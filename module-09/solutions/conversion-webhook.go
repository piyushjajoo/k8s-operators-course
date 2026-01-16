package v1

import (
    "fmt"
    "sigs.k8s.io/controller-runtime/pkg/conversion"
    databasev2 "github.com/example/postgres-operator/api/v2"
)

// ConvertTo converts this Database to the Hub version (v2)
// This function is called when Kubernetes needs to convert a v1 Database to v2
// Note: v1 remains the storage version; this conversion is for serving v2 API requests
func (src *Database) ConvertTo(dstRaw conversion.Hub) error {
    dst, ok := dstRaw.(*databasev2.Database)
    if !ok {
        return fmt.Errorf("expected *v2.Database, got %T", dstRaw)
    }

    // Convert metadata - preserve all metadata
    dst.ObjectMeta = src.ObjectMeta

    // Convert spec: v1 → v2
    dst.Spec.Image = src.Spec.Image
    dst.Spec.DatabaseName = src.Spec.DatabaseName
    dst.Spec.Username = src.Spec.Username
    dst.Spec.Storage = src.Spec.Storage // Same structure in both versions
    dst.Spec.Resources = src.Spec.Resources
    
    // Convert replication: v1 has Replicas at top level, v2 has ReplicationConfig
    if src.Spec.Replicas != nil {
        if dst.Spec.Replication == nil {
            dst.Spec.Replication = &databasev2.ReplicationConfig{}
        }
        dst.Spec.Replication.Replicas = src.Spec.Replicas
        // Set default mode if not specified
        if dst.Spec.Replication.Mode == "" {
            dst.Spec.Replication.Mode = "async"
        }
    }
    
    // Backup config is new in v2, leave nil (no v1 equivalent)

    // Convert status
    dst.Status.Phase = src.Status.Phase
    dst.Status.Ready = src.Status.Ready
    dst.Status.Endpoint = src.Status.Endpoint
    dst.Status.SecretName = src.Status.SecretName
    dst.Status.Conditions = src.Status.Conditions

    return nil
}

// ConvertFrom converts from the Hub version (v2) to this version (v1)
// This function is called when Kubernetes needs to convert a v2 Database to v1
// Note: v1 is the storage version, so this conversion happens when storing v2 resources
func (dst *Database) ConvertFrom(srcRaw conversion.Hub) error {
    src, ok := srcRaw.(*databasev2.Database)
    if !ok {
        return fmt.Errorf("expected *v2.Database, got %T", srcRaw)
    }

    // Convert metadata - preserve all metadata
    dst.ObjectMeta = src.ObjectMeta

    // Convert spec: v2 → v1
    dst.Spec.Image = src.Spec.Image
    dst.Spec.DatabaseName = src.Spec.DatabaseName
    dst.Spec.Username = src.Spec.Username
    dst.Spec.Storage = src.Spec.Storage // Same structure in both versions
    dst.Spec.Resources = src.Spec.Resources
    
    // Convert replication: v2 has ReplicationConfig, v1 has Replicas at top level
    if src.Spec.Replication != nil {
        dst.Spec.Replicas = src.Spec.Replication.Replicas
    }
    // Note: Replication.Mode is lost in v1 conversion (acceptable - v1 doesn't support it)
    
    // Note: Backup config is lost in v1 conversion (v1 doesn't support backups)

    // Convert status
    dst.Status.Phase = src.Status.Phase
    dst.Status.Ready = src.Status.Ready
    dst.Status.Endpoint = src.Status.Endpoint
    dst.Status.SecretName = src.Status.SecretName
    dst.Status.Conditions = src.Status.Conditions

    return nil
}

