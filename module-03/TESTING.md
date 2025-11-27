# Module 3 Testing Guide

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

This document provides instructions for testing Module 3 content.

## Prerequisites

Before testing, ensure you have:
- Completed [Module 1](../module-01/README.md) and [Module 2](../module-02/README.md)
- Go 1.21+ installed
- kubebuilder installed
- kubectl configured
- Docker or Podman running
- kind installed and cluster running

## Quick Test

### 1. Test PostgreSQL Operator

Follow the complete labs to build the PostgreSQL operator:

1. **Lab 3.2**: Design and create the Database API
2. **Lab 3.3**: Implement reconciliation logic
3. **Lab 3.4**: Add advanced client operations

### 2. Verify Operator Works

```bash
# Install CRD
make install

# Run operator
make run

# Create Database resource
kubectl apply -f - <<EOF
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

# Verify resources created
kubectl get database test-db
kubectl get statefulset test-db
kubectl get service test-db

# Check status
kubectl get database test-db -o jsonpath='{.status}'
```

## Verification Checklist

- [ ] Can create Database API with kubebuilder
- [ ] CRD generated with proper validation
- [ ] API validation works correctly
- [ ] Can implement reconciliation logic
- [ ] StatefulSet is created correctly
- [ ] Service is created correctly
- [ ] Owner references work (cascade delete)
- [ ] Status updates correctly
- [ ] Idempotency works (multiple applies)
- [ ] Updates work correctly
- [ ] Advanced client operations work

## Common Issues

### Issue: StatefulSet not creating
**Solution**: 
- Check RBAC permissions
- Verify owner reference is set
- Check operator logs

### Issue: Status not updating
**Solution**:
- Ensure status subresource is enabled
- Check RBAC for status updates
- Verify Status().Update() is used

### Issue: Validation not working
**Solution**:
- Regenerate manifests: `make manifests`
- Reinstall CRD: `make uninstall && make install`
- Check kubebuilder markers are correct

## Testing the Complete Operator

### Full Test Flow

1. **Initialize project:**
   ```bash
   kubebuilder init --domain database.example.com --repo github.com/example/postgres-operator
   kubebuilder create api --group database --version v1 --kind Database
   ```

2. **Implement API** (Lab 3.2)

3. **Implement controller** (Lab 3.3)

4. **Add advanced features** (Lab 3.4)

5. **Test:**
   ```bash
   make install
   make run
   kubectl apply -f database.yaml
   ```

6. **Verify:**
   - Database resource created
   - StatefulSet created
   - Service created
   - Status updated
   - Cascade delete works

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

