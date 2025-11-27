# Module 8 Testing Guide

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

This document provides instructions for testing Module 8 content.

## Prerequisites

Before testing, ensure you have:
- Completed all previous modules (Modules 1-7)
- Production-ready operator from Module 7
- Go 1.21+ installed
- kubebuilder installed
- kubectl configured
- Docker or Podman running
- kind installed and cluster running

## Quick Test

### 1. Test Multi-Tenancy

Follow [Lab 8.1](labs/lab-01-multi-tenancy.md) to implement multi-tenancy.

**Verify:**
- Cluster-scoped CRD works
- Namespace isolation works
- Resource quotas enforced
- Multi-tenant scenarios work

### 2. Test Operator Composition

Follow [Lab 8.2](labs/lab-02-operator-composition.md) to compose operators.

**Verify:**
- Backup operator works
- Operators coordinate
- Resource references work
- Status conditions coordinate

### 3. Test Stateful Applications

Follow [Lab 8.3](labs/lab-03-stateful-applications.md) to manage stateful apps.

**Verify:**
- Backup functionality works
- Restore operations work
- Rolling updates work
- Data consistency maintained

### 4. Test Final Project

Follow [Lab 8.4](labs/lab-04-final-project.md) to build final project.

**Verify:**
- All requirements met
- Tests pass
- Documentation complete
- Production ready

## Verification Checklist

- [ ] Multi-tenancy implemented
- [ ] Operator composition works
- [ ] Backup/restore functional
- [ ] Rolling updates work
- [ ] Data consistency ensured
- [ ] Final project complete
- [ ] All tests pass
- [ ] Documentation complete
- [ ] Production ready

## Common Issues

### Issue: Cluster-scoped CRD not working
**Solution**: 
- Verify scope is set to Cluster
- Check RBAC permissions
- Verify namespace handling

### Issue: Operators not coordinating
**Solution**:
- Check resource references
- Verify status conditions
- Check event emission

### Issue: Backup/restore failing
**Solution**:
- Verify database connectivity
- Check storage access
- Verify permissions

## Testing the Complete Course

### Full Test Flow

1. **Test all modules** (Modules 1-8)
2. **Build final project** (Lab 8.4)
3. **Verify all features** work
4. **Run all tests**
5. **Check documentation**

## Cleanup

After testing:

```bash
# Delete test resources
kubectl delete databases --all
kubectl delete backups --all
kubectl delete restores --all
kubectl delete namespaces tenant-1 tenant-2
```

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

