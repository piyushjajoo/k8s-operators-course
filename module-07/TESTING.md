# Module 7 Testing Guide

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

This document provides instructions for testing Module 7 content.

## Prerequisites

Before testing, ensure you have:
- Completed [Module 6](../module-06/README.md)
- Database operator from Module 3/4/5/6
- Go 1.21+ installed
- kubebuilder installed
- kubectl configured
- Docker or Podman running
- kind installed and cluster running
- Helm installed (optional)

## Quick Test

### 1. Test Packaging

Follow [Lab 7.1](labs/lab-01-packaging-distribution.md) to package your operator.

**Verify:**
- Dockerfile builds successfully
- Image is created
- Helm chart packages correctly
- Operator deploys with Helm

### 2. Test RBAC

Follow [Lab 7.2](labs/lab-02-rbac-security.md) to configure RBAC.

**Verify:**
- RBAC is optimized
- Service account works
- Security contexts applied
- Network policies work

### 3. Test High Availability

Follow [Lab 7.3](labs/lab-03-high-availability.md) to implement HA.

**Verify:**
- Leader election works
- Multiple replicas run
- Failover works
- PDB protects pods

### 4. Test Performance

Follow [Lab 7.4](labs/lab-04-performance-scalability.md) to optimize performance.

**Verify:**
- Rate limiting works
- Caching improves performance
- Metrics are exposed
- Load testing passes

## Verification Checklist

- [ ] Container image builds successfully
- [ ] Helm chart packages correctly
- [ ] RBAC is optimized (least privilege)
- [ ] Security contexts applied
- [ ] Leader election works
- [ ] Multiple replicas run correctly
- [ ] Failover works
- [ ] Rate limiting implemented
- [ ] Performance metrics exposed
- [ ] Operator ready for production

## Common Issues

### Issue: Image build fails
**Solution**: 
- Check Dockerfile syntax
- Verify Go version matches
- Check build context

### Issue: Leader election not working
**Solution**:
- Verify leader election is enabled
- Check RBAC for lease permissions
- Verify namespace is correct

### Issue: Performance issues
**Solution**:
- Check rate limiting
- Verify caching is working
- Monitor metrics
- Profile reconciliation

## Testing the Complete Production Setup

### Full Test Flow

1. **Package operator** (Lab 7.1)
2. **Configure RBAC** (Lab 7.2)
3. **Enable HA** (Lab 7.3)
4. **Optimize performance** (Lab 7.4)

5. **Deploy to production:**
   ```bash
   # Build and push image
   docker build -t database-operator:v0.1.0 .
   docker push database-operator:v0.1.0
   
   # Install with Helm
   helm install database-operator ./helm-chart
   
   # Verify deployment
   kubectl get deployment database-operator
   kubectl get pods -l app=database-operator
   ```

## Cleanup

After testing:

```bash
# Uninstall Helm release
helm uninstall database-operator

# Delete resources
kubectl delete databases --all
```

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

