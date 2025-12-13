---
layout: default
title: "Lab 08.4: Final Project"
nav_order: 14
parent: "Module 8: Advanced Topics"
grand_parent: Modules
mermaid: true
---

# Lab 8.4: Final Project

**Related Lesson:** [Lesson 8.4: Real-World Patterns and Best Practices](../lessons/04-real-world-patterns.md)  
**Navigation:** [← Previous Lab: Stateful Applications](lab-03-stateful-applications.md) | [Module Overview](../README.md) | [Course Overview](../../README.md)

## Objectives

Build a complete, production-ready operator for a stateful application that demonstrates all concepts learned throughout the course. This final project integrates everything you've learned from Modules 1-8 into a single, comprehensive operator.

## Prerequisites

- Completion of all previous modules (Modules 1-7)
- Completion of [Lab 8.1](lab-01-multi-tenancy.md), [Lab 8.2](lab-02-operator-composition.md), and [Lab 8.3](lab-03-stateful-applications.md)
- Understanding of all operator concepts:
  - Kubernetes architecture and API machinery ([Module 1](../../module-01/README.md))
  - Operator pattern and Kubebuilder ([Module 2](../../module-02/README.md))
  - Controller runtime and reconciliation ([Module 3](../../module-03/README.md))
  - Advanced reconciliation patterns ([Module 4](../../module-04/README.md))
  - Webhooks and admission control ([Module 5](../../module-05/README.md))
  - Testing and debugging ([Module 6](../../module-06/README.md))
  - Production considerations ([Module 7](../../module-07/README.md))
- A working development environment with kubebuilder, Go, Docker/Podman, and kind

## Project Requirements

Your final operator must include:

1. **Full CRUD Operations**
   - Create, Read, Update, Delete
   - Proper error handling
   - Idempotent operations

2. **Status Reporting**
   - Status subresource
   - Conditions
   - Progress tracking
   - Observed generation

3. **Webhooks**
   - Validating webhooks
   - Mutating webhooks
   - Default values
   - Validation rules

4. **Testing**
   - Unit tests
   - Integration tests
   - Test coverage > 80%

5. **Production Features**
   - RBAC configuration
   - Security hardening
   - High availability
   - Performance optimization

6. **Advanced Features**
   - Multi-tenancy support
   - Backup/restore
   - Rolling updates
   - Documentation

## Project Structure

```text
final-operator/
├── api/
│   └── v1/
│       ├── groupversion_info.go
│       └── <resource>_types.go
├── internal/controller/
│   └── <resource>_controller.go
├── config/
│   ├── crd/
│   ├── rbac/
│   ├── manager/
│   └── webhook/
├── internal/
│   ├── backup/
│   └── restore/
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Exercise 1: Choose Your Application

### Options

1. **Database Operator** (PostgreSQL, MySQL, MongoDB)
2. **Message Queue Operator** (RabbitMQ, Kafka)
3. **Cache Operator** (Redis, Memcached)
4. **Search Engine Operator** (Elasticsearch)
5. **Your Choice** (Any stateful application)

### Task 1.1: Scaffold Your Operator Project

Start by creating a new operator project using kubebuilder:

```bash
# Create a new directory for your final project
mkdir -p ~/final-operator
cd ~/final-operator

# Initialize kubebuilder project
kubebuilder init --domain example.com --project-name final-operator

# Create your API (replace <resource> with your choice, e.g., Database, Cache, Queue)
kubebuilder create api \
  --group apps \
  --version v1 \
  --kind <Resource> \
  --resource --controller

# When prompted:
# Create Resource [y/n]: y
# Create Controller [y/n]: y
```

### Task 1.2: Define Comprehensive API Types

Edit `api/v1/<resource>_types.go` to create a production-ready API. Here's a complete example for a Database operator:

```go
package v1

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
    // Image is the database image to use
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

    // PasswordSecretRef references a Secret containing the password
    // +optional
    PasswordSecretRef *corev1.SecretKeySelector `json:"passwordSecretRef,omitempty"`

    // Backup configuration
    // +optional
    Backup *BackupConfig `json:"backup,omitempty"`

    // Monitoring configuration
    // +optional
    Monitoring *MonitoringConfig `json:"monitoring,omitempty"`
}

// StorageSpec defines storage configuration
type StorageSpec struct {
    // Size is the storage size
    // +kubebuilder:validation:Required
    Size string `json:"size"`

    // StorageClassName is the storage class to use
    // +optional
    StorageClassName *string `json:"storageClassName,omitempty"`
}

// BackupConfig defines backup configuration
type BackupConfig struct {
    // Enabled enables automatic backups
    Enabled bool `json:"enabled"`

    // Schedule is the cron schedule for backups
    // +optional
    Schedule string `json:"schedule,omitempty"`

    // Retention is the number of backups to retain
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:default=7
    Retention int `json:"retention,omitempty"`
}

// MonitoringConfig defines monitoring configuration
type MonitoringConfig struct {
    // Enabled enables monitoring
    Enabled bool `json:"enabled"`

    // ServiceMonitor creates a ServiceMonitor for Prometheus
    // +optional
    ServiceMonitor bool `json:"serviceMonitor,omitempty"`
}

// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
    // Conditions represent the latest observations
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // Phase is the current phase
    // +kubebuilder:validation:Enum=Pending;Creating;Ready;Failed;Updating
    Phase string `json:"phase,omitempty"`

    // Ready indicates if the database is ready
    Ready bool `json:"ready,omitempty"`

    // ReadyReplicas is the number of ready replicas
    ReadyReplicas int32 `json:"readyReplicas,omitempty"`

    // Endpoint is the database endpoint
    Endpoint string `json:"endpoint,omitempty"`

    // ObservedGeneration is the generation observed by the controller
    ObservedGeneration int64 `json:"observedGeneration,omitempty"`

    // LastBackupTime is when the last backup was completed
    // +optional
    LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".status.readyReplicas"
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
```

### Task 1.3: Generate CRDs and Verify

```bash
# Generate code and CRD manifests
make generate
make manifests

# Review the generated CRD
cat config/crd/bases/apps.example.com_databases.yaml

# Install CRDs
make install

# Verify CRD is installed
kubectl get crd databases.apps.example.com
```

## Exercise 2: Implement Core Functionality

### Task 2.1: Implement Complete Reconciliation Logic

Edit `internal/controller/<resource>_controller.go` to implement full reconciliation. Reference implementations from [Module 3 solutions](../../module-03/solutions/database-controller.go) and [Module 4 solutions](../../module-04/solutions/state-machine-controller.go) for complete examples.

Key components to implement:

- `Reconcile()` - Main reconciliation loop with phase handling
- `handlePending()` - Initial state handling
- `handleCreating()` - Resource creation logic
- `handleReady()` - Ready state monitoring
- `handleUpdating()` - Update handling
- `handleDeletion()` - Cleanup with finalizers
- `reconcileSecret()` - Secret management
- `reconcileStatefulSet()` - StatefulSet creation/updates
- `reconcileService()` - Service management
- `updateStatus()` - Status and condition updates

### Task 2.2: Implement Status Management Helpers

Add comprehensive status management functions. Reference implementations from [Module 4 solutions](../../module-04/solutions/conditions-helpers.go) for complete condition management.

## Exercise 3: Add Webhooks

### Task 3.1: Scaffold Webhooks

```bash
# Create validating and mutating webhooks
kubebuilder create webhook \
  --group apps \
  --version v1 \
  --kind Database \
  --programmatic-validation \
  --defaulting
```

### Task 3.2: Implement Validating Webhook

Edit `internal/webhook/v1/database_webhook.go` to add comprehensive validation. Reference [Module 5 Lab 2](../../module-05/labs/lab-02-validating-webhooks.md) for detailed examples.

Key validations to implement:

- Validate database name format
- Validate username format
- Validate storage size format
- Validate replicas range
- Prevent changing immutable fields on update

### Task 3.3: Implement Mutating Webhook

Add defaulting logic for:

- Default image if not specified
- Default replicas if not specified
- Default backup retention
- Default resource limits

### Task 3.4: Generate Webhook Manifests

```bash
# Generate webhook manifests
make manifests

# Review webhook configuration
cat config/webhook/manifests.yaml

# For local development, use cert-manager or manual certificates
# See Module 5 Lab 4 for webhook deployment details
```

## Exercise 4: Add Testing

### Task 4.1: Write Comprehensive Unit Tests

Edit `internal/controller/<resource>_controller_test.go` to add comprehensive tests. Reference [Module 6 solutions](../../module-06/solutions/database_controller_test.go) for complete examples.

Test scenarios to cover:

- Resource creation
- Resource updates
- Resource deletion
- Status updates
- Error handling
- Finalizer handling

### Task 4.2: Write Integration Tests

Create `internal/controller/integration_test.go` for end-to-end tests. Reference [Module 6 Lab 3](../../module-06/labs/lab-03-integration-testing.md).

### Task 4.3: Run Tests and Check Coverage

```bash
# Setup envtest binaries
make setup-envtest

# Run all tests
make test

# Run with coverage
go test -coverprofile=coverage.out ./internal/controller/...
go tool cover -html=coverage.out

# Verify coverage is > 80%
go test -cover ./internal/controller/...
```

### Task 4.4: Add Test Examples for Advanced Features

Add tests for:

- Backup/restore functionality
- Multi-tenancy scenarios
- Error handling and retries
- Webhook validation
- Status condition updates

Reference [Module 6 solutions](../../module-06/solutions/) for complete test examples.

## Exercise 5: Production Features

### Task 5.1: Configure RBAC

```bash
# Generate RBAC manifests
make manifests

# Review RBAC configuration
cat config/rbac/role.yaml

# Optimize RBAC - remove unnecessary permissions
# Only grant permissions your operator actually needs
```

Review and optimize the generated RBAC. Reference [Module 7 Lab 2](../../module-07/labs/lab-02-rbac-security.md) for best practices.

### Task 5.2: Security Hardening

#### Update Dockerfile for Security

Ensure your `Dockerfile` uses distroless images and runs as non-root. Reference [Module 7 solutions](../../module-07/solutions/Dockerfile) for complete example.

#### Add Security Contexts

Update `config/manager/manager.yaml` to include security contexts. Reference [Module 7 solutions](../../module-07/solutions/security.yaml) for complete security configuration.

### Task 5.3: Enable High Availability

#### Enable Leader Election

Update `cmd/main.go` to enable leader election. Reference [Module 7 Lab 3](../../module-07/labs/lab-03-high-availability.md) for complete HA setup.

#### Add Pod Disruption Budget

Create `config/manager/pdb.yaml` for Pod Disruption Budget configuration.

### Task 5.4: Performance Optimization

#### Add Rate Limiting

Update controller setup to use rate limiting. Reference [Module 7 solutions](../../module-07/solutions/ratelimiter.go) for advanced rate limiting.

### Task 5.5: Add Observability

Add metrics and logging. Reference [Module 6 Lab 4](../../module-06/labs/lab-04-debugging-observability.md) for complete observability setup.

## Exercise 6: Documentation

### Task 6.1: Create Comprehensive README

Create a `README.md` in your project root with:

- Quick start guide
- Architecture overview
- API documentation
- Examples
- Troubleshooting guide

### Task 6.2: Create Example Resources

Create `config/samples/apps_v1_database.yaml` and additional examples in an `examples/` directory:

- Basic usage example
- Advanced scenarios
- Multi-tenant setup
- Backup/restore examples

### Task 6.3: Add API Documentation

Document all API fields with clear descriptions and examples. Use kubebuilder markers for automatic documentation generation.

## Exercise 7: Build and Deploy

### Task 7.1: Build Container Image

```bash
# Build the image
make docker-build IMG=final-operator:v1.0.0

# For kind, load image
kind load docker-image final-operator:v1.0.0 --name k8s-operators-course

# Or push to registry
docker push final-operator:v1.0.0
```

### Task 7.2: Deploy Operator

```bash
# Update image in config/manager/manager.yaml
# Then deploy
make deploy IMG=final-operator:v1.0.0

# Verify deployment
kubectl get pods -n final-operator-system

# Check logs
kubectl logs -n final-operator-system -l control-plane=controller-manager
```

### Task 7.3: Test Your Operator

```bash
# Create a test resource
kubectl apply -f config/samples/apps_v1_database.yaml

# Watch the resource
kubectl get database database-sample -w

# Verify resources were created
kubectl get statefulset,service,secret -l app=database-sample

# Test updates
kubectl patch database database-sample --type=merge -p '{"spec":{"replicas":3}}'

# Test deletion
kubectl delete database database-sample
```

## Submission Checklist

Use this checklist to ensure your operator is complete:

### Core Functionality

- [ ] Full CRUD operations implemented (Create, Read, Update, Delete)
- [ ] Proper error handling throughout
- [ ] Idempotent reconciliation logic
- [ ] Finalizers implemented for cleanup

### Status Management

- [ ] Status subresource configured
- [ ] Conditions implemented and updated correctly
- [ ] Phase tracking (Pending, Creating, Ready, Failed, etc.)
- [ ] Observed generation tracking
- [ ] Progress tracking for long operations

### Webhooks

- [ ] Validating webhook implemented
- [ ] Mutating webhook implemented
- [ ] Default values set correctly
- [ ] Validation rules comprehensive
- [ ] Webhook certificates configured

### Testing

- [ ] Unit tests written (>80% coverage)
- [ ] Integration tests implemented
- [ ] Test suite runs successfully (`make test`)
- [ ] Edge cases covered
- [ ] Error scenarios tested

### Production Features

- [ ] RBAC configured and optimized
- [ ] Security hardened (distroless, non-root, security contexts)
- [ ] High availability enabled (leader election)
- [ ] Performance optimized (rate limiting, concurrency)
- [ ] Observability added (metrics, logging)

### Advanced Features

- [ ] Multi-tenancy support (if applicable)
- [ ] Backup/restore functionality (if applicable)
- [ ] Rolling updates handled correctly
- [ ] Resource quotas respected

### Documentation

- [ ] README.md complete with:
  - Quick start guide
  - Architecture overview
  - API documentation
  - Examples
  - Troubleshooting guide
- [ ] Example resources provided
- [ ] Code comments comprehensive
- [ ] API fields documented

### Packaging

- [ ] Container image builds successfully
- [ ] Image pushed to registry (or loaded into kind)
- [ ] Operator deploys successfully
- [ ] All resources created correctly
- [ ] Operator works end-to-end

## Cleanup

After completing your project:

```bash
# Delete test resources
kubectl delete database --all --all-namespaces

# Undeploy operator
make undeploy

# Uninstall CRDs
make uninstall

# Clean up kind cluster (if using)
kind delete cluster --name k8s-operators-course
```

## Self-Evaluation Checklist

Use this checklist to ensure your operator meets production-ready standards:

### Functionality

- **Core Operations**: All CRUD operations work correctly
- **Edge Cases**: Handles edge cases gracefully (deletion, updates, failures)
- **Error Handling**: Proper error handling and retry logic
- **Status Management**: Status accurately reflects resource state
- **Webhooks**: Validation and mutation work correctly

### Code Quality

- **Structure**: Well-organized, follows Go best practices
- **Readability**: Clear variable names, comments where needed
- **Idempotency**: Reconciliation is idempotent
- **Resource Management**: Proper use of finalizers, owner references
- **Error Messages**: Clear, actionable error messages

### Testing Standards

- **Coverage**: Test coverage > 80%
- **Unit Tests**: Comprehensive unit tests for controller logic
- **Integration Tests**: End-to-end integration tests
- **Edge Cases**: Tests cover error scenarios and edge cases
- **Test Quality**: Tests are maintainable and well-structured

### Production Readiness

- **Security**: Uses distroless images, non-root, security contexts
- **RBAC**: Minimal, least-privilege RBAC configuration
- **High Availability**: Leader election enabled
- **Performance**: Rate limiting, concurrency limits configured
- **Observability**: Metrics and logging implemented

### Documentation Standards

- **README**: Comprehensive README with quick start
- **API Documentation**: All API fields documented
- **Examples**: Multiple example resources provided
- **Troubleshooting**: Common issues and solutions documented
- **Code Comments**: Important logic explained

## Integration with Previous Modules

This final project integrates concepts from all previous modules:

- **Module 1**: Understanding Kubernetes architecture and API machinery
- **Module 2**: Using Kubebuilder to scaffold operators
- **Module 3**: Controller runtime and reconciliation patterns
- **Module 4**: Advanced patterns (conditions, finalizers, watching)
- **Module 5**: Webhooks and admission control
- **Module 6**: Testing and observability
- **Module 7**: Production considerations (packaging, security, HA)
- **Module 8**: Advanced topics (multi-tenancy, composition, stateful apps)

Reference solutions from previous modules:

- [Module 3 Solutions](../../module-03/solutions/) - Controller implementation
- [Module 4 Solutions](../../module-04/solutions/) - Advanced patterns
- [Module 5 Solutions](../../module-05/solutions/) - Webhooks
- [Module 6 Solutions](../../module-06/solutions/) - Testing
- [Module 7 Solutions](../../module-07/solutions/) - Production features
- [Module 8 Solutions](../solutions/) - Advanced features

## Solutions

Complete example solutions and reference implementations are available:

- **Previous Module Solutions**: Reference solutions from Modules 3-7 for implementation patterns
- **Module 8 Solutions**: See [solutions directory](../solutions/) for:
  - Multi-tenant operator patterns
  - Operator composition examples
  - Backup/restore implementations
  - Rolling update patterns

## Lab Summary

In this final project lab, you:

1. **Designed a complete API** - Created comprehensive CRD with spec and status
2. **Implemented full reconciliation** - Built complete controller with all CRUD operations
3. **Added webhooks** - Implemented validating and mutating webhooks
4. **Wrote comprehensive tests** - Created unit and integration tests with >80% coverage
5. **Configured production features** - Set up RBAC, security, HA, and performance optimizations
6. **Created documentation** - Wrote README, examples, and API documentation
7. **Built and deployed** - Packaged operator as container image and deployed to cluster

## Key Learnings

Through this final project, you've demonstrated mastery of:

1. **Operator Development Lifecycle** - From scaffolding to production deployment
2. **Kubernetes API Patterns** - CRDs, status subresources, conditions, finalizers
3. **Controller Patterns** - Reconciliation, idempotency, error handling
4. **Admission Control** - Validating and mutating webhooks
5. **Testing Strategies** - Unit tests with envtest, integration tests
6. **Production Readiness** - Security, HA, performance, observability
7. **Best Practices** - Code organization, documentation, examples

## Next Steps

After completing this course, you can:

1. **Build Real Operators** - Apply these patterns to build operators for your applications
2. **Contribute to Open Source** - Contribute to existing operators or create new ones
3. **Advanced Topics** - Explore:
   - Operator SDK (alternative to Kubebuilder)
   - Operator Lifecycle Manager (OLM)
   - Operator Framework
   - Multi-cluster operators
   - Operator metrics and dashboards

## Share Your Project

We'd love to see what you've built! If you've completed your final project and want to share it:

1. **Post on LinkedIn** - Share your operator project, what you learned, and what you built
2. **Tag the Course Creator** - Tag [Piyush Jajoo](https://www.linkedin.com/in/pjajoo) in your post
3. **Include Details** - Share:
   - What operator you built
   - Key features you implemented
   - What you learned from the course
   - Link to your GitHub repository (if public)

I'll make my best effort in my free time to review your code and provide feedback!

## Additional Resources

- [Kubebuilder Book](https://book.kubebuilder.io/) - Comprehensive Kubebuilder documentation
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md) - API design guidelines
- [Operator Best Practices](https://sdk.operatorframework.io/docs/best-practices/) - Operator SDK best practices
- [Example Operators](https://github.com/operator-framework/awesome-operators) - List of example operators

## Congratulations! 🎉

You've completed the entire **Building Kubernetes Operators Course**!

You now have:

- ✅ Deep understanding of Kubernetes architecture and operators
- ✅ Hands-on experience building production-ready operators
- ✅ Knowledge of best practices and patterns
- ✅ Skills to build operators for any application

**You are now ready to build production-ready Kubernetes operators!**

**Navigation:** [← Previous Lab: Stateful Applications](lab-03-stateful-applications.md) | [Related Lesson](../lessons/04-real-world-patterns.md) | [Module Overview](../README.md) | [Course Overview](../../README.md)
