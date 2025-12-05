# Lab 3.1: Exploring Controller Runtime

**Related Lesson:** [Lesson 3.1: Controller Runtime Deep Dive](../lessons/01-controller-runtime.md)  
**Navigation:** [Module Overview](../README.md) | [Next Lab: Designing API →](lab-02-designing-api.md)

## Objectives

- Explore controller-runtime architecture
- Understand Manager setup
- Implement different requeue scenarios
- Trace reconciliation calls

## Prerequisites

- Completion of [Module 2](../module-02/README.md)
- Kind cluster running
- Understanding of basic operator structure

## Exercise 1: Examine Manager Setup

### Task 1.1: Review Your Hello World Operator

```bash
# Navigate to your hello-world-operator from Module 2
cd ~/hello-world-operator

# Examine main.go
cat main.go
```

**Questions:**
1. How is the Manager created?
2. What options are configured?
3. How is the reconciler set up?

### Task 1.2: Understand Manager Options

```go
// In main.go, examine the Manager options:
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    Scheme:                 scheme,
    MetricsBindAddress:     metricsAddr,
    Port:                   9443,
    HealthProbeBindAddress: probeAddr,
    LeaderElection:          enableLeaderElection,
})
```

**Questions:**
1. What does each option do?
2. Why is leader election important?
3. What's the purpose of metrics and health probes?

## Exercise 2: Explore Reconcile Function

### Task 2.1: Examine Current Reconcile Function

```bash
# Look at your controller
cat internal/controller/helloworld_controller.go
```

**Observe:**
- Function signature
- How it reads resources
- How it returns results
- Error handling

### Task 2.2: Add Logging

Add detailed logging to understand the flow:

```go
func (r *HelloWorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    
    log.Info("Reconcile called", "request", req)
    
    // ... existing code ...
    
    log.Info("Reconcile completed", "result", result)
    return result, err
}
```

## Exercise 3: Implement Different Requeue Scenarios

### Task 3.1: Immediate Requeue

Modify your controller to requeue immediately on certain conditions:

```go
// If ConfigMap is being created, requeue to check status
if !configMapCreated {
    log.Info("ConfigMap not ready, requeuing")
    return ctrl.Result{Requeue: true}, nil
}
```

### Task 3.2: Delayed Requeue

Add a delayed requeue for rate limiting:

```go
// If external dependency is not ready, check again in 10 seconds
if !dependencyReady {
    log.Info("Dependency not ready, requeuing in 10s")
    return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}
```

### Task 3.3: No Requeue

Ensure success cases don't requeue:

```go
// Everything is in desired state
log.Info("Reconciliation successful")
return ctrl.Result{}, nil
```

## Exercise 4: Trace Reconciliation Calls

### Task 4.1: Run Operator with Verbose Logging

```bash
# Run operator
make run

# In another terminal, create a resource
kubectl apply -f - <<EOF
apiVersion: hello.example.com/v1
kind: HelloWorld
metadata:
  name: trace-test
spec:
  message: "Trace me"
  count: 3
EOF
```

**Observe:**
- When Reconcile is called
- What request is passed
- What result is returned
- How often it's called

### Task 4.2: Modify Resource and Observe

```bash
# Update the resource
kubectl patch helloworld trace-test --type merge -p '{"spec":{"count":5}}'

# Watch logs - see reconciliation triggered
```

## Exercise 5: Understand Client Usage

### Task 5.1: Examine Client Operations

In your controller, identify:
- `r.Get()` calls
- `r.Create()` calls
- `r.Update()` calls
- `r.Status().Update()` calls

### Task 5.2: Add Client Error Handling

Improve error handling:

```go
// Get resource with proper error handling
if err := r.Get(ctx, req.NamespacedName, helloWorld); err != nil {
    if errors.IsNotFound(err) {
        log.Info("Resource not found, may have been deleted")
        return ctrl.Result{}, nil
    }
    log.Error(err, "Failed to get resource")
    return ctrl.Result{}, err
}
```

## Exercise 6: Test Different Scenarios

### Task 6.1: Test Resource Deletion

```bash
# Delete resource
kubectl delete helloworld trace-test

# Observe logs - see reconciliation on deletion
```

### Task 6.2: Test Concurrent Updates

```bash
# Create resource
kubectl apply -f resource.yaml

# Quickly update multiple times
kubectl patch helloworld test --type merge -p '{"spec":{"count":1}}'
kubectl patch helloworld test --type merge -p '{"spec":{"count":2}}'
kubectl patch helloworld test --type merge -p '{"spec":{"count":3}}'

# Observe how reconciliation handles rapid changes
```

## Cleanup

```bash
# Delete test resources
kubectl delete helloworld trace-test test 2>/dev/null || true
```

## Lab Summary

In this lab, you:
- Explored Manager setup and configuration
- Understood Reconcile function flow
- Implemented different requeue strategies
- Traced reconciliation calls
- Improved error handling

## Key Learnings

1. Manager coordinates all controller components
2. Reconcile function is called for each resource change
3. Different requeue strategies for different scenarios
4. Client provides type-safe access to resources
5. Proper error handling is crucial
6. Logging helps understand reconciliation flow

## Next Steps

Now that you understand controller-runtime, let's design a proper API for a database operator!

**Navigation:** [← Module Overview](../README.md) | [Related Lesson](../lessons/01-controller-runtime.md) | [Next Lab: Designing API →](lab-02-designing-api.md)

