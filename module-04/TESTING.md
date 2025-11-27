# Module 4 Testing Guide

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

This document provides instructions for testing Module 4 content.

## Prerequisites

Before testing, ensure you have:
- Completed [Module 3](../module-03/README.md)
- PostgreSQL operator from Module 3
- Go 1.21+ installed
- kubebuilder installed
- kubectl configured
- Docker or Podman running
- kind installed and cluster running

## Quick Test

### 1. Test Conditions

Follow [Lab 4.1](labs/lab-01-conditions-status.md) to add conditions to your operator.

**Verify:**
- Conditions are added to status
- Conditions update based on resource state
- Condition transitions work correctly

### 2. Test Finalizers

Follow [Lab 4.2](labs/lab-02-finalizers-cleanup.md) to add finalizers.

**Verify:**
- Finalizer is added on creation
- Resource is not deleted until finalizer removed
- Cleanup is performed before deletion

### 3. Test Watches

Follow [Lab 4.3](labs/lab-03-watching-indexing.md) to set up watches.

**Verify:**
- Owned resources trigger reconciliation
- Non-owned resources trigger reconciliation
- Indexes work correctly

### 4. Test Advanced Patterns

Follow [Lab 4.4](labs/lab-04-advanced-patterns.md) to implement state machine.

**Verify:**
- State machine transitions work
- Multi-phase reconciliation completes
- External dependencies are handled

## Verification Checklist

- [ ] Conditions are implemented and update correctly
- [ ] Finalizers prevent deletion until cleanup
- [ ] Cleanup is idempotent
- [ ] Watches trigger reconciliation
- [ ] Indexes enable efficient queries
- [ ] State machine transitions correctly
- [ ] Multi-phase reconciliation works
- [ ] All operations are idempotent

## Common Issues

### Issue: Conditions not updating
**Solution**: 
- Ensure Status().Update() is called
- Check RBAC for status updates
- Verify observed generation is set

### Issue: Finalizer not removed
**Solution**:
- Check cleanup completes successfully
- Verify Update() is called after removing finalizer
- Check for errors in cleanup

### Issue: Watches not triggering
**Solution**:
- Verify SetupWithManager is correct
- Check event predicates
- Ensure resources match watch criteria

## Testing the Complete Enhanced Operator

### Full Test Flow

1. **Add conditions** (Lab 4.1)
2. **Add finalizers** (Lab 4.2)
3. **Set up watches** (Lab 4.3)
4. **Implement state machine** (Lab 4.4)

5. **Test complete flow:**
   ```bash
   make install
   make run
   kubectl apply -f database.yaml
   
   # Watch conditions
   kubectl get database test-db -o jsonpath='{.status.conditions}'
   
   # Test deletion
   kubectl delete database test-db
   # Should wait for cleanup
   ```

## Cleanup

After testing:

```bash
# Delete Database resources
kubectl delete databases --all

# Uninstall CRD
make uninstall

# Stop operator (Ctrl+C)
```

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

