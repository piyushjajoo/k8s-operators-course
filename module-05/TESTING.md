# Module 5 Testing Guide

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

This document provides instructions for testing Module 5 content.

## Prerequisites

Before testing, ensure you have:
- Completed [Module 4](../module-04/README.md)
- Database operator from Module 3/4
- Go 1.21+ installed
- kubebuilder installed
- kubectl configured
- Docker or Podman running
- kind installed and cluster running

## Quick Test

### 1. Test Validating Webhook

Follow [Lab 5.2](labs/lab-02-validating-webhooks.md) to add validating webhook.

**Verify:**
- Webhook is registered
- Valid resources are accepted
- Invalid resources are rejected
- Error messages are clear

### 2. Test Mutating Webhook

Follow [Lab 5.3](labs/lab-03-mutating-webhooks.md) to add mutating webhook.

**Verify:**
- Defaults are applied
- Mutations are idempotent
- Context-aware defaults work

### 3. Test Certificate Management

Follow [Lab 5.4](labs/lab-04-webhook-deployment.md) to set up certificates.

**Verify:**
- Certificates are generated
- Webhook service works
- TLS connection is established

## Verification Checklist

- [ ] Validating webhook is created
- [ ] Mutating webhook is created
- [ ] Certificates are generated
- [ ] Webhook service is configured
- [ ] Valid resources pass validation
- [ ] Invalid resources are rejected
- [ ] Defaults are applied correctly
- [ ] Webhooks work in cluster
- [ ] Certificate rotation works (if using cert-manager)

## Common Issues

### Issue: Webhook not called
**Solution**: 
- Check webhook configuration exists
- Verify service is accessible
- Check certificate and CA bundle match
- Ensure webhook pod is running

### Issue: Certificate errors
**Solution**:
- Regenerate certificates: `make certs`
- Reinstall certificates: `make install-cert`
- Check certificate secret exists
- Verify CA bundle in webhook config

### Issue: Connection refused
**Solution**:
- Check webhook service endpoints
- Verify operator pod is running
- Check service selector matches pod labels
- Verify port configuration

## Testing the Complete Webhook Setup

### Full Test Flow

1. **Scaffold webhooks:**
   ```bash
   kubebuilder create webhook --group database --version v1 --kind Database --programmatic-validation --defaulting
   ```

2. **Implement validation and defaulting** (Labs 5.2 and 5.3)

3. **Set up certificates:**
   ```bash
   make certs
   make install-cert
   ```

4. **Test:**
   ```bash
   make run
   kubectl apply -f database.yaml
   ```

5. **Verify:**
   - Defaults applied (mutating webhook)
   - Validation works (validating webhook)
   - Invalid resources rejected

## Cleanup

After testing:

```bash
# Delete Database resources
kubectl delete databases --all

# Uninstall operator
make undeploy

# Remove certificates
make uninstall-cert
```

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

