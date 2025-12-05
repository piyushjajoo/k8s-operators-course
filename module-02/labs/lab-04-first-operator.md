# Lab 2.4: Building Hello World Operator

**Related Lesson:** [Lesson 2.4: Your First Operator](../lessons/04-first-operator.md)  
**Navigation:** [← Previous Lab: Dev Environment](lab-03-dev-environment.md) | [Module Overview](../README.md)

## Objectives

- Build your first complete operator
- Understand operator project structure
- Run operator locally
- Create and manage Custom Resources
- Observe reconciliation in action

## Prerequisites

- Complete development environment from [Lab 2.3](lab-03-dev-environment.md)
- Kind cluster running
- Understanding of CRDs from [Module 1](../../module-01/README.md)

## Exercise 1: Initialize Project

### Task 1.1: Create Project Directory

```bash
# Create project directory
mkdir -p ~/hello-world-operator
cd ~/hello-world-operator

# Initialize git (optional but recommended)
git init
```

### Task 1.2: Initialize Kubebuilder Project

```bash
# Initialize kubebuilder project
kubebuilder init --domain example.com --repo github.com/example/hello-world-operator
```

**Observe:**
- What files were created?
- What's the project structure?

### Task 1.3: Verify Project Structure

```bash
# List files
ls -la

# Check main.go
head -20 main.go

# Check Makefile
head -30 Makefile
```

## Exercise 2: Create API

### Task 2.1: Create HelloWorld API

```bash
# Create API
kubebuilder create api --group hello --version v1 --kind HelloWorld
```

When prompted:
- Create Resource [y/n]: **y**
- Create Controller [y/n]: **y**

### Task 2.2: Examine Generated Files

```bash
# Check API types
cat api/v1/helloworld_types.go

# Check controller
cat internal/controller/helloworld_controller.go
```

## Exercise 3: Define API Types

### Task 3.1: Edit API Types

Edit `api/v1/helloworld_types.go`:

```go
package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelloWorldSpec defines the desired state of HelloWorld
type HelloWorldSpec struct {
    // Message is the message to display
    // +kubebuilder:validation:Required
    Message string `json:"message"`
    
    // Count is the number of times to display the message
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=100
    Count int32 `json:"count,omitempty"`
}

// HelloWorldStatus defines the observed state of HelloWorld
type HelloWorldStatus struct {
    // Phase represents the current phase
    Phase string `json:"phase,omitempty"`
    
    // ConfigMapCreated indicates if the ConfigMap was created
    ConfigMapCreated bool `json:"configMapCreated,omitempty"`
    
    // LastUpdated is when the status was last updated
    LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".spec.message"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// HelloWorld is the Schema for the helloworlds API
type HelloWorld struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   HelloWorldSpec   `json:"spec,omitempty"`
    Status HelloWorldStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HelloWorldList contains a list of HelloWorld
type HelloWorldList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []HelloWorld `json:"items"`
}

func init() {
    SchemeBuilder.Register(&HelloWorld{}, &HelloWorldList{})
}
```

### Task 3.2: Generate Code

```bash
# Generate code
make generate

# Generate manifests
make manifests
```

### Task 3.3: Verify CRD

```bash
# Check CRD was generated
ls -la config/crd/bases/

# Examine CRD
cat config/crd/bases/hello.example.com_helloworlds.yaml | head -50
```

## Exercise 4: Implement Controller

### Task 4.1: Edit Controller

Edit `internal/controller/helloworld_controller.go`:

```go
package controller

import (
    "context"
    "fmt"
    "time"
    
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"
    
    hellov1 "github.com/example/hello-world-operator/api/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hello.example.com,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hello.example.com,resources=helloworlds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hello.example.com,resources=helloworlds/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *HelloWorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    
    logger.Info("Reconciling HelloWorld", "name", req.NamespacedName)
    
    // Fetch the HelloWorld instance
    helloWorld := &hellov1.HelloWorld{}
    if err := r.Get(ctx, req.NamespacedName, helloWorld); err != nil {
        if errors.IsNotFound(err) {
            // Object not found, return
            logger.Info("HelloWorld not found, ignoring", "name", req.NamespacedName)
            return ctrl.Result{}, nil
        }
        // Error reading the object
        logger.Error(err, "Failed to get HelloWorld")
        return ctrl.Result{}, err
    }
    
    // Define the ConfigMap
    configMapName := helloWorld.Name + "-config"
    configMap := &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{
            Name:      configMapName,
            Namespace: helloWorld.Namespace,
        },
        Data: map[string]string{
            "message": helloWorld.Spec.Message,
            "count":   fmt.Sprintf("%d", helloWorld.Spec.Count),
        },
    }
    
    // Set owner reference
    if err := ctrl.SetControllerReference(helloWorld, configMap, r.Scheme); err != nil {
        logger.Error(err, "Failed to set controller reference")
        return ctrl.Result{}, err
    }
    
    // Check if ConfigMap already exists
    existingConfigMap := &corev1.ConfigMap{}
    err := r.Get(ctx, client.ObjectKey{
        Name:      configMap.Name,
        Namespace: configMap.Namespace,
    }, existingConfigMap)
    
    if err != nil && errors.IsNotFound(err) {
        // ConfigMap doesn't exist, create it
        logger.Info("Creating ConfigMap", "name", configMap.Name)
        if err := r.Create(ctx, configMap); err != nil {
            logger.Error(err, "Failed to create ConfigMap")
            return ctrl.Result{}, err
        }
    } else if err != nil {
        logger.Error(err, "Failed to get ConfigMap")
        return ctrl.Result{}, err
    } else {
        // ConfigMap exists, update it if needed
        if existingConfigMap.Data["message"] != configMap.Data["message"] ||
           existingConfigMap.Data["count"] != configMap.Data["count"] {
            logger.Info("Updating ConfigMap", "name", configMap.Name)
            existingConfigMap.Data = configMap.Data
            if err := r.Update(ctx, existingConfigMap); err != nil {
                logger.Error(err, "Failed to update ConfigMap")
                return ctrl.Result{}, err
            }
        }
    }
    
    // Update status
    now := metav1.Now()
    helloWorld.Status.Phase = "Ready"
    helloWorld.Status.ConfigMapCreated = true
    helloWorld.Status.LastUpdated = &now
    
    if err := r.Status().Update(ctx, helloWorld); err != nil {
        logger.Error(err, "Failed to update status")
        return ctrl.Result{}, err
    }
    
    logger.Info("Successfully reconciled HelloWorld", "name", req.NamespacedName)
    return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&hellov1.HelloWorld{}).
        Complete(r)
}
```

### Task 4.2: Regenerate Manifests

```bash
# Regenerate RBAC (controller has new permissions)
make manifests
```

## Exercise 5: Install and Run Operator

### Task 5.1: Install CRD

```bash
# Install CRD to cluster
make install

# Verify CRD was created
kubectl get crd helloworlds.hello.example.com

# Examine CRD
kubectl get crd helloworlds.hello.example.com -o yaml | head -30
```

### Task 5.2: Run Operator Locally

In one terminal, run the operator:

```bash
# Run operator (connects to kind cluster)
make run
```

**Observe:**
- Operator starts up
- Logs show it's ready
- It's watching for HelloWorld resources

### Task 5.3: Create HelloWorld Resource

In another terminal, create a HelloWorld:

```bash
# Create HelloWorld resource
cat <<EOF | kubectl apply -f -
apiVersion: hello.example.com/v1
kind: HelloWorld
metadata:
  name: hello-example
spec:
  message: "Hello from my first operator!"
  count: 5
EOF
```

### Task 5.4: Observe Reconciliation

Watch what happens:

```bash
# Check HelloWorld resource
kubectl get helloworld hello-example

# Get detailed view
kubectl get helloworld hello-example -o yaml

# Check ConfigMap was created
kubectl get configmap hello-example-config

# View ConfigMap data
kubectl get configmap hello-example-config -o jsonpath='{.data}'

# Watch operator logs (in the terminal running make run)
# You should see reconciliation logs
```

## Exercise 6: Test Updates

### Task 6.1: Update HelloWorld

```bash
# Update the message
kubectl patch helloworld hello-example --type merge -p '{"spec":{"message":"Updated message!"}}'

# Watch operator logs
# Check ConfigMap was updated
kubectl get configmap hello-example-config -o jsonpath='{.data.message}'
```

### Task 6.2: Verify Status Updates

```bash
# Check status was updated
kubectl get helloworld hello-example -o jsonpath='{.status}'
```

## Exercise 7: Test Deletion

### Task 7.1: Delete HelloWorld

```bash
# Delete HelloWorld
kubectl delete helloworld hello-example

# Check ConfigMap (should be deleted due to owner reference)
kubectl get configmap hello-example-config
```

**Expected:** ConfigMap should be automatically deleted (owner reference from [Module 1](../../module-01/lessons/03-controller-pattern.md))

## Exercise 8: Create Multiple Resources

### Task 8.1: Create Multiple HelloWorlds

```bash
# Create multiple HelloWorld resources
cat <<EOF | kubectl apply -f -
apiVersion: hello.example.com/v1
kind: HelloWorld
metadata:
  name: hello-1
spec:
  message: "First hello"
  count: 3
---
apiVersion: hello.example.com/v1
kind: HelloWorld
metadata:
  name: hello-2
spec:
  message: "Second hello"
  count: 7
EOF
```

### Task 8.2: Verify All Resources

```bash
# List all HelloWorlds
kubectl get helloworlds

# Check all ConfigMaps
kubectl get configmaps | grep hello
```

## Cleanup

```bash
# Delete all HelloWorld resources
kubectl delete helloworlds --all

# Uninstall CRD
make uninstall

# Stop operator (Ctrl+C in the terminal running make run)
```

## Lab Summary

In this lab, you:
- Created a complete operator project
- Defined Custom Resource types
- Implemented reconciliation logic
- Ran operator locally
- Created and managed Custom Resources
- Observed reconciliation in action
- Tested updates and deletions

## Key Learnings

1. Kubebuilder scaffolds complete operator projects
2. You define API types (spec and status)
3. You implement the Reconcile function
4. Operator follows reconciliation pattern from Module 1
5. Owner references manage resource lifecycle
6. Status updates reflect actual state
7. Operators run locally but connect to cluster

## Congratulations!

You've built your first operator! This demonstrates:
- ✅ CRD creation and management
- ✅ Controller implementation
- ✅ Reconciliation pattern
- ✅ Resource creation and updates
- ✅ Status management
- ✅ Owner references

## Solutions

Complete working solutions for this lab are available in the [solutions directory](../solutions/):
- [main.go](../solutions/hello-world-operator-main.go) - Complete operator entry point
- [Controller](../solutions/hello-world-controller.go) - Complete controller implementation
- [API Types](../solutions/hello-world-types.go) - Complete API type definitions

## Next Steps

In Module 3, you'll learn to build more sophisticated controllers with advanced patterns!

**Navigation:** [← Previous Lab: Dev Environment](lab-03-dev-environment.md) | [Related Lesson](../lessons/04-first-operator.md) | [Module Overview](../README.md)

