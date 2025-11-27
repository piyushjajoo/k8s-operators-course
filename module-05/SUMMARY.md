# Module 5: Summary

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

## What We Built

Module 5 teaches webhooks and admission control, adding powerful validation and mutation capabilities to operators. This module builds on [Module 4](../module-04/README.md) by adding webhook-based validation and defaulting that goes beyond CRD schema validation.

### Content Structure

1. **4 Complete Lessons** with Mermaid diagrams:
   - Lesson 5.1: Kubernetes Admission Control
   - Lesson 5.2: Implementing Validating Webhooks
   - Lesson 5.3: Implementing Mutating Webhooks
   - Lesson 5.4: Webhook Deployment and Certificates

2. **4 Hands-on Labs**:
   - Lab 5.1: Exploring Admission Control
   - Lab 5.2: Building Validating Webhook
   - Lab 5.3: Building Mutating Webhook
   - Lab 5.4: Certificate Management

3. **Mermaid Diagrams**:
   - Admission control flow
   - Admission sequence diagram
   - Webhook request/response flow
   - Certificate management flow

## Key Concepts Covered

### Admission Control
- Admission control overview
- Mutating vs validating webhooks
- Webhook configuration and registration
- Admission request/response structures

### Validating Webhooks
- Validation logic implementation
- Handling webhook requests
- Error responses and user feedback
- Schema validation vs webhook validation
- Cross-field validation

### Mutating Webhooks
- Defaulting values
- Mutating requests
- JSON patch operations
- Common mutation patterns
- Idempotent mutations

### Webhook Deployment
- Certificate management strategies
- Using cert-manager
- Webhook service setup
- Testing webhooks locally
- Certificate rotation

## Learning Outcomes

After completing Module 5, students will:
- ✅ Understand Kubernetes admission control
- ✅ Implement validating webhooks for custom validation
- ✅ Implement mutating webhooks for defaulting
- ✅ Manage webhook certificates and deployment
- ✅ Test webhooks locally and in production
- ✅ Troubleshoot webhook issues

## Connection to Previous Modules

Module 5 builds on:

- **Module 1** ([Lesson 1.4](../module-01/lessons/04-custom-resources.md)): CRD schema validation knowledge
- **Module 3** ([Lesson 3.2](../module-03/lessons/02-designing-api.md)): API design and validation
- **Module 4**: Enhanced operator patterns

## What Students Build

By the end of Module 5, students have:
- Validating webhook for custom validation rules
- Mutating webhook for defaulting values
- Certificate management setup
- Complete webhook deployment
- Understanding of admission control

## Files Created

```
module-05/
├── README.md
├── SUMMARY.md
├── TESTING.md
├── diagrams/
│   ├── 01-admission-control-flow.mmd
│   └── 01-admission-sequence.mmd
├── labs/
│   ├── lab-01-admission-control.md
│   ├── lab-02-validating-webhooks.md
│   ├── lab-03-mutating-webhooks.md
│   └── lab-04-webhook-deployment.md
└── lessons/
    ├── 01-admission-control.md
    ├── 02-validating-webhooks.md
    ├── 03-mutating-webhooks.md
    └── 04-webhook-deployment.md
```

## Notes

- All examples add webhooks to the Database operator
- Practical, hands-on approach throughout
- Mermaid diagrams for visual learning
- Content builds on Module 4 concepts
- Students add production-ready webhooks
- Ready for students to use immediately

## Next Steps

Module 5 adds webhook capabilities. In Module 6, students will:
- Learn about testing operators
- Use envtest for unit testing
- Create integration tests
- Debug operators effectively
- Add observability

**Navigation:** [Module Overview](README.md) | [Course Overview](../README.md)

