# Module 6 Testing Guide

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

This document provides instructions for testing Module 6 content.

## Prerequisites

Before testing, ensure you have:
- Completed [Module 5](../module-05/README.md)
- Database operator from Module 3/4/5
- Go 1.21+ installed
- kubebuilder installed
- kubectl configured
- Docker or Podman running
- kind installed and cluster running

## Quick Test

### 1. Test Unit Testing Setup

Follow [Lab 6.1](labs/lab-01-testing-fundamentals.md) to set up testing environment.

**Verify:**
- Ginkgo and Gomega installed
- envtest tools installed
- Delve debugger installed
- Test suite created

### 2. Test Unit Tests

Follow [Lab 6.2](labs/lab-02-unit-testing-envtest.md) to write unit tests.

**Verify:**
- Unit tests run successfully
- Tests cover reconciliation logic
- Test coverage is good (80%+)
- Error cases are tested

### 3. Test Integration Tests

Follow [Lab 6.3](labs/lab-03-integration-testing.md) to create integration tests.

**Verify:**
- Integration tests run successfully
- End-to-end workflows work
- Webhooks are tested
- CI/CD integration works

### 4. Test Observability

Follow [Lab 6.4](labs/lab-04-debugging-observability.md) to add observability.

**Verify:**
- Structured logging works
- Metrics are exposed
- Events are emitted
- Debugging setup works

## Verification Checklist

- [ ] Testing tools installed (Ginkgo, Gomega, envtest, Delve)
- [ ] Unit test suite created
- [ ] Unit tests pass
- [ ] Integration test suite created
- [ ] Integration tests pass
- [ ] Structured logging added
- [ ] Metrics exposed
- [ ] Events emitted
- [ ] Debugging setup works
- [ ] CI/CD integration works

## Common Issues

### Issue: envtest not starting
**Solution**: 
- Check CRD paths are correct
- Verify envtest binaries are downloaded
- Check for port conflicts

### Issue: Tests timing out
**Solution**:
- Increase timeout values
- Check cluster is accessible
- Verify operator is running

### Issue: Metrics not exposed
**Solution**:
- Check metrics registry
- Verify metrics endpoint
- Check port configuration

## Testing the Complete Setup

### Full Test Flow

1. **Set up testing environment** (Lab 6.1)
2. **Write unit tests** (Lab 6.2)
3. **Create integration tests** (Lab 6.3)
4. **Add observability** (Lab 6.4)

5. **Run all tests:**
   ```bash
   # Unit tests
   ginkgo -v ./controllers
   
   # Integration tests
   ginkgo -v ./test/integration
   
   # Check coverage
   go test -cover ./controllers/...
   ```

## Cleanup

After testing:

```bash
# Clean up test resources
kubectl delete databases --all

# Clean up test cluster (if created)
kind delete cluster
```

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

