---
layout: default
title: "Lab 03.2: Designing Api"
nav_order: 12
parent: "Module 3: Building Custom Controllers"
grand_parent: Modules
mermaid: true
---

# Lab 3.2: API Design for Database Operator

**Related Lesson:** [Lesson 3.2: Designing Your API](../lessons/02-designing-api.md)  
**Navigation:** [← Previous Lab: Controller Runtime](lab-01-controller-runtime.md) | [Module Overview](../README.md) | [Next Lab: Reconciliation Logic →](lab-03-reconciliation-logic.md)

## Objectives

- Design API for a PostgreSQL database operator
- Use kubebuilder markers for validation
- Generate CRD with proper schema
- Test API validation

## Prerequisites

- Completion of [Module 2](../../module-02/README.md)
- Understanding of API design principles
- kubebuilder installed

## Exercise 1: Initialize Database Operator Project

### Task 1.1: Create Project

```bash
# Create new project
mkdir -p ~/postgres-operator
cd ~/postgres-operator

# Initialize kubebuilder project
kubebuilder init --domain example.com --repo github.com/example/postgres-operator
```

### Task 1.2: Create Database API

```bash
# Create Database API
kubebuilder create api --group database --version v1 --kind Database

# When prompted:
# Create Resource [y/n]: y
# Create Controller [y/n]: y
```

## Exercise 2: Design Database Spec

### Task 2.1: Define DatabaseSpec

Edit `api/v1/database_types.go`:

```go
package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
)

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
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
```

### Task 2.2: Define DatabaseStatus

```go
// DatabaseStatus defines the observed state of Database
type DatabaseStatus struct {
    // Phase is the current phase
    // +kubebuilder:validation:Enum=Pending;Creating;Ready;Failed
    Phase string `json:"phase,omitempty"`
    
    // Ready indicates if the database is ready
    Ready bool `json:"ready,omitempty"`
    
    // Endpoint is the database endpoint
    Endpoint string `json:"endpoint,omitempty"`
    
    // SecretName is the name of the Secret containing database credentials
    SecretName string `json:"secretName,omitempty"`
}
```

### Task 2.3: Complete Database Type

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
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

## Exercise 3: Generate and Verify CRD

### Task 3.1: Generate Code

```bash
# Generate code
make generate

# Generate manifests
make manifests
```

### Task 3.2: Examine Generated CRD

```bash
# Check CRD was generated and verify validation rules
cat config/crd/bases/database.example.com_databases.yaml | head -100
```

**Questions:**
1. Are validation rules present?
2. Are default values set?
3. Are print columns defined?

## Exercise 4: Test API Validation

### Task 4.1: Install CRD

```bash
# Install CRD
make install

# Verify
kubectl get crd databases.database.example.com
```

### Task 4.2: Test Valid Resource

```bash
# Create valid Database resource
cat <<EOF | kubectl apply -f -
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: test-db
spec:
  image: postgres:14
  replicas: 1
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Verify it was created
kubectl get database test-db
```

### Task 4.3: Test Invalid Resources

```bash
# Test missing required field
cat <<EOF | kubectl apply -f -
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: invalid-db
spec:
  image: postgres:14
  # Missing databaseName
EOF

# Should fail validation

# Test invalid replica count
cat <<EOF | kubectl apply -f -
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: invalid-replicas
spec:
  image: postgres:14
  replicas: 20  # Exceeds maximum
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
EOF

# Should fail validation

# Test invalid storage size
cat <<EOF | kubectl apply -f -
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: invalid-storage
spec:
  image: postgres:14
  databaseName: mydb
  username: admin
  storage:
    size: invalid  # Doesn't match pattern
EOF

# Should fail validation
```

## Exercise 5: Test Print Columns

### Task 5.1: Create Multiple Databases

```bash
# Create a few databases
kubectl apply -f - <<EOF
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: db1
spec:
  image: postgres:14
  replicas: 1
  databaseName: db1
  username: user1
  storage:
    size: 10Gi
---
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: db2
spec:
  image: postgres:13
  replicas: 2
  databaseName: db2
  username: user2
  storage:
    size: 20Gi
EOF
```

### Task 5.2: Verify Print Columns

```bash
# List databases - should show print columns
kubectl get databases

# Should show: NAME, PHASE, REPLICAS, READY, AGE
```

## Exercise 6: Test Default Values

### Task 6.1: Create Resource with Minimal Spec

```bash
# Create with only required fields
cat <<EOF | kubectl apply -f -
apiVersion: database.example.com/v1
kind: Database
metadata:
  name: minimal-db
spec:
  databaseName: mydb
  username: admin
  storage:
    size: 10Gi
  # image and replicas should use defaults
EOF
```

### Task 6.2: Verify Defaults

```bash
# Check if defaults were applied
kubectl get database minimal-db -o jsonpath='{.spec.image}'
kubectl get database minimal-db -o jsonpath='{.spec.replicas}'
```

## Cleanup

```bash
# Delete test resources
kubectl delete databases --all

# Uninstall CRD
make uninstall
```

## Lab Summary

In this lab, you:
- Designed a complete API for a database operator
- Used kubebuilder markers for validation
- Generated CRD with proper schema
- Tested API validation rules
- Verified print columns work
- Tested default values

## Key Learnings

1. API design follows Kubernetes conventions
2. Spec contains desired state, Status contains actual state
3. Validation markers enforce constraints
4. Print columns improve user experience
5. Default values make APIs easier to use
6. Proper versioning is important

## Solutions

The API design from this lab is used in the complete Database operator solution:
- [Database Types](../solutions/database-types.go) - Complete API type definitions with validation markers

## Next Steps

Now that you have a well-designed API, let's implement the reconciliation logic to make it work!

**Navigation:** [← Previous Lab: Controller Runtime](lab-01-controller-runtime.md) | [Related Lesson](../lessons/02-designing-api.md) | [Next Lab: Reconciliation Logic →](lab-03-reconciliation-logic.md)
