# Kubernetes Operators Course - Build Plan

## Course Philosophy

- **Practical First**: Every concept explained through examples and hands-on exercises
- **Visual Learning**: Extensive use of Mermaid diagrams for architecture, flows, and concepts
- **Kubebuilder Focus**: All examples and labs use Kubebuilder framework
- **Kind Cluster**: All hands-on labs use kind cluster (Docker/Podman)
- **Progressive Complexity**: Start simple, build up to production-ready operators

---

## Prerequisites Setup (Before Module 1)

### Development Environment Setup Guide

**Tools Required:**
- Go 1.21+
- Docker or Podman
- kubectl
- kind
- kubebuilder CLI

**Kind Cluster Setup Script:**
- Create `scripts/setup-kind-cluster.sh` with:
  - Kind cluster creation
  - Load balancer setup (MetalLB or similar)
  - Ingress controller installation
  - Verification steps

**Mermaid Diagrams Needed:**
- Development environment architecture
- Kind cluster networking diagram

---

## Module 1: Kubernetes Architecture Deep Dive

### Structure

**1.1 Kubernetes Control Plane Review**
- **Mermaid Diagrams:**
  - Control plane component architecture
  - API Server request flow
  - Controller Manager architecture
  - etcd interaction flow
- **Hands-on:**
  - Explore API Server with `kubectl proxy`
  - Inspect etcd data (read-only)
  - Watch controller manager logs
  - Trace a Pod creation request flow

**1.2 Kubernetes API Machinery**
- **Mermaid Diagrams:**
  - RESTful API structure
  - API versioning flow
  - API groups organization
  - Resource lifecycle diagram
- **Hands-on:**
  - Direct API calls with curl
  - Explore API discovery
  - Create/update/delete resources via API
  - Understand API versions and groups

**1.3 The Controller Pattern**
- **Mermaid Diagrams:**
  - Control loop diagram
  - Reconciliation flow
  - Watch/Informer pattern
  - Leader election flow
- **Hands-on:**
  - Observe Deployment controller behavior
  - Watch resource changes in real-time
  - Simulate reconciliation scenarios
  - Understand declarative vs imperative

**1.4 Custom Resources**
- **Mermaid Diagrams:**
  - CRD registration flow
  - CRD vs ConfigMap decision tree
  - CRD schema structure
  - Custom resource lifecycle
- **Hands-on:**
  - Create a simple CRD manually (YAML)
  - Deploy and use the CRD
  - Add validation schema
  - Compare CRD vs ConfigMap use cases

### Module 1 Deliverables
- `module-01/` directory with:
  - Lesson content (markdown)
  - Mermaid diagrams (`.mmd` files)
  - Hands-on lab scripts
  - Solutions directory

---

## Module 2: Introduction to Operators

### Structure

**2.1 The Operator Pattern**
- **Mermaid Diagrams:**
  - Operator pattern overview
  - Operator capability levels (1-5)
  - Operator vs Helm comparison
  - Operator decision flowchart
- **Hands-on:**
  - Analyze existing operators (Prometheus, etc.)
  - Compare operator vs Helm deployment
  - Identify operator use cases

**2.2 Kubebuilder Fundamentals**
- **Mermaid Diagrams:**
  - Kubebuilder architecture
  - Project structure diagram
  - Code generation flow
  - Kubebuilder vs Operator SDK comparison
- **Hands-on:**
  - Install kubebuilder
  - Understand kubebuilder CLI commands
  - Explore project structure

**2.3 Development Environment Setup**
- **Mermaid Diagrams:**
  - Local development setup
  - Kind cluster architecture
  - Development workflow
- **Hands-on:**
  - Complete environment setup
  - Create kind cluster
  - Verify all tools
  - Configure kubectl context

**2.4 Your First Operator**
- **Mermaid Diagrams:**
  - Operator scaffolding process
  - Generated code structure
  - Operator runtime flow
  - First reconciliation flow
- **Hands-on:**
  - Scaffold "Hello World" operator with kubebuilder
  - Understand generated files
  - Run operator locally
  - Create and observe CustomResource
  - Modify reconciliation logic

### Module 2 Deliverables
- `module-02/` directory with:
  - Lesson content
  - Mermaid diagrams
  - Complete "Hello World" operator example
  - Setup scripts
  - Lab exercises

---

## Module 3: Building Custom Controllers

### Structure

**3.1 Controller Runtime Deep Dive**
- **Mermaid Diagrams:**
  - Controller-runtime architecture
  - Manager, Reconciler, Client relationship
  - Reconcile function flow
  - Requeue strategies diagram
- **Hands-on:**
  - Explore controller-runtime code
  - Understand Manager setup
  - Implement different requeue scenarios
  - Trace reconciliation calls

**3.2 Designing Your API**
- **Mermaid Diagrams:**
  - API design process
  - Spec vs Status separation
  - API versioning strategy
  - Validation flow
- **Hands-on:**
  - Design API for database operator
  - Use kubebuilder markers for validation
  - Generate CRD with proper schema
  - Test API validation

**3.3 Implementing Reconciliation Logic**
- **Mermaid Diagrams:**
  - Reconciliation loop lifecycle
  - Resource creation flow
  - Update detection flow
  - Owner reference chain
- **Hands-on:**
  - Build PostgreSQL operator (basic)
  - Implement create/update logic
  - Handle owner references
  - Test idempotency

**3.4 Working with Client-Go**
- **Mermaid Diagrams:**
  - Client types comparison
  - Watch mechanism
  - Patch operations flow
  - Optimistic concurrency
- **Hands-on:**
  - Use typed client
  - Implement watch for dependent resources
  - Use strategic merge patch
  - Handle conflicts

### Module 3 Deliverables
- `module-03/` directory with:
  - Lesson content
  - Mermaid diagrams
  - Complete PostgreSQL operator (basic version)
  - Lab exercises
  - Code examples

---

## Module 4: Advanced Reconciliation Patterns

### Structure

**4.1 Conditions and Status Management**
- **Mermaid Diagrams:**
  - Status subresource flow
  - Condition lifecycle
  - Status update strategy
  - Condition state machine
- **Hands-on:**
  - Add status subresource to operator
  - Implement conditions (Ready, Progressing, etc.)
  - Update status based on resource state
  - Observe status changes

**4.2 Finalizers and Cleanup**
- **Mermaid Diagrams:**
  - Finalizer flow
  - Deletion flow with finalizers
  - Cleanup process
  - Finalizer deadlock prevention
- **Hands-on:**
  - Add finalizer to operator
  - Implement cleanup logic
  - Handle graceful deletion
  - Test cleanup scenarios

**4.3 Watching and Indexing**
- **Mermaid Diagrams:**
  - Watch setup flow
  - Index structure
  - Event handling flow
  - Cross-namespace watching
- **Hands-on:**
  - Set up watches for child resources
  - Create indexes for efficient lookups
  - Handle watch events
  - Implement cross-namespace watching

**4.4 Advanced Patterns**
- **Mermaid Diagrams:**
  - Multi-phase reconciliation
  - State machine pattern
  - External dependency handling
  - Idempotency guarantees
- **Hands-on:**
  - Implement multi-phase deployment
  - Create state machine for operator
  - Handle external API dependencies
  - Ensure idempotent operations

### Module 4 Deliverables
- `module-04/` directory with:
  - Lesson content
  - Mermaid diagrams
  - Enhanced PostgreSQL operator
  - Advanced pattern examples
  - Lab exercises

---

## Module 5: Webhooks and Admission Control

### Structure

**5.1 Kubernetes Admission Control**
- **Mermaid Diagrams:**
  - Admission control flow
  - Webhook request/response flow
  - Mutating vs Validating comparison
  - Webhook registration process
- **Hands-on:**
  - Explore existing admission controllers
  - Understand webhook configuration
  - Test webhook endpoints

**5.2 Implementing Validating Webhooks**
- **Mermaid Diagrams:**
  - Validation webhook flow
  - Request validation process
  - Error response format
  - Validation decision tree
- **Hands-on:**
  - Scaffold validating webhook with kubebuilder
  - Implement custom validation logic
  - Test with valid/invalid resources
  - Provide meaningful error messages

**5.3 Implementing Mutating Webhooks**
- **Mermaid Diagrams:**
  - Mutation webhook flow
  - Defaulting process
  - JSON patch operations
  - Mutation strategies
- **Hands-on:**
  - Scaffold mutating webhook
  - Implement defaulting logic
  - Use JSON patch for mutations
  - Test mutation scenarios

**5.4 Webhook Deployment and Certificates**
- **Mermaid Diagrams:**
  - Certificate management flow
  - Webhook service architecture
  - cert-manager integration
  - Local development setup
- **Hands-on:**
  - Set up certificate management
  - Configure webhook service
  - Use cert-manager for production
  - Test webhooks locally with kind

### Module 5 Deliverables
- `module-05/` directory with:
  - Lesson content
  - Mermaid diagrams
  - Webhook examples
  - Certificate setup scripts
  - Lab exercises

---

## Module 6: Testing and Debugging

### Structure

**6.1 Unit Testing Controllers**
- **Mermaid Diagrams:**
  - envtest architecture
  - Test setup flow
  - Test execution flow
  - Mocking strategy
- **Hands-on:**
  - Set up envtest environment
  - Write unit tests for reconciler
  - Test reconciliation logic
  - Use table-driven tests
  - Achieve good test coverage

**6.2 Integration Testing**
- **Mermaid Diagrams:**
  - Integration test flow
  - Test cluster setup
  - Ginkgo test structure
  - CI/CD integration
- **Hands-on:**
  - Set up test cluster with kind
  - Write Ginkgo/Gomega tests
  - Create end-to-end test suite
  - Integrate with CI/CD

**6.3 Debugging Techniques**
- **Mermaid Diagrams:**
  - Debugging workflow
  - Delve debugger setup
  - Log analysis flow
  - Common issues flowchart
- **Hands-on:**
  - Set up Delve debugger
  - Debug operator locally
  - Analyze controller logs
  - Trace reconciliation loops
  - Fix common bugs

**6.4 Observability**
- **Mermaid Diagrams:**
  - Observability stack
  - Metrics collection flow
  - Event flow
  - Tracing architecture
- **Hands-on:**
  - Add structured logging
  - Expose Prometheus metrics
  - Emit Kubernetes events
  - Set up basic tracing
  - Create dashboards

### Module 6 Deliverables
- `module-06/` directory with:
  - Lesson content
  - Mermaid diagrams
  - Complete test suite examples
  - Debugging guides
  - Observability examples
  - Lab exercises

---

## Module 7: Production Considerations

### Structure

**7.1 Packaging and Distribution**
- **Mermaid Diagrams:**
  - Operator packaging flow
  - Image build process
  - OLM bundle structure
  - Distribution strategy
- **Hands-on:**
  - Build operator container image
  - Create Helm chart for operator
  - Package as OLM bundle
  - Version and tag properly
  - Push to container registry

**7.2 RBAC and Security**
- **Mermaid Diagrams:**
  - RBAC architecture
  - Permission model
  - Security scanning flow
  - Least privilege principle
- **Hands-on:**
  - Define RBAC requirements
  - Generate RBAC manifests with kubebuilder
  - Review and minimize permissions
  - Run security scans
  - Harden operator image

**7.3 High Availability**
- **Mermaid Diagrams:**
  - Leader election flow
  - HA architecture
  - Failover process
  - Replica coordination
- **Hands-on:**
  - Enable leader election
  - Deploy multiple replicas
  - Test failover scenarios
  - Set resource limits
  - Monitor leader status

**7.4 Performance and Scalability**
- **Mermaid Diagrams:**
  - Performance optimization flow
  - Rate limiting strategy
  - Caching architecture
  - Scalability patterns
- **Hands-on:**
  - Implement rate limiting
  - Add caching layer
  - Optimize reconciliation
  - Load test operator
  - Profile and optimize

### Module 7 Deliverables
- `module-07/` directory with:
  - Lesson content
  - Mermaid diagrams
  - Production-ready operator example
  - Packaging scripts
  - Security checklists
  - Performance benchmarks
  - Lab exercises

---

## Module 8: Advanced Topics and Real-World Patterns

### Structure

**8.1 Multi-Tenancy and Namespace Isolation**
- **Mermaid Diagrams:**
  - Multi-tenancy architecture
  - Namespace isolation model
  - Resource quota flow
  - Cluster-scoped vs namespaced
- **Hands-on:**
  - Build cluster-scoped operator
  - Implement namespace isolation
  - Handle resource quotas
  - Test multi-tenant scenarios

**8.2 Operator Composition**
- **Mermaid Diagrams:**
  - Operator composition pattern
  - Dependency management
  - Coordination strategies
  - Composite operator flow
- **Hands-on:**
  - Use multiple operators together
  - Manage operator dependencies
  - Implement coordination
  - Test composite scenarios

**8.3 Stateful Application Management**
- **Mermaid Diagrams:**
  - StatefulSet management
  - Backup/restore flow
  - Rolling update process
  - Data consistency model
- **Hands-on:**
  - Manage StatefulSets in operator
  - Implement backup functionality
  - Handle rolling updates
  - Ensure data consistency
  - Test migration scenarios

**8.4 Case Studies and Best Practices**
- **Mermaid Diagrams:**
  - Popular operator architectures
  - Best practices flow
  - Anti-patterns to avoid
  - Documentation structure
- **Hands-on:**
  - Analyze Prometheus operator
  - Review Elasticsearch operator
  - Identify patterns
  - Document your operator
  - Create user guides

### Module 8 Deliverables
- `module-08/` directory with:
  - Lesson content
  - Mermaid diagrams
  - Advanced operator examples
  - Case study analyses
  - Best practices guide
  - Final project template

---

## Final Project

### Requirements

Build a complete operator for a stateful application that includes:

1. **Full CRUD Operations**
   - Create, Read, Update, Delete
   - Proper error handling

2. **Status Reporting**
   - Status subresource
   - Conditions
   - Progress tracking

3. **Webhooks**
   - Validation webhook
   - Mutating webhook (defaulting)

4. **Backup/Restore**
   - Backup functionality
   - Restore capability
   - Backup scheduling

5. **Comprehensive Tests**
   - Unit tests (80%+ coverage)
   - Integration tests
   - E2E tests

6. **Production-Ready Packaging**
   - Container image
   - Helm chart
   - RBAC configuration
   - Documentation

### Project Suggestions
- Redis operator
- MongoDB operator
- Elasticsearch operator (simplified)
- Custom application operator

---

## Course Structure

### Directory Layout

```
k8s-operators-course/
├── README.md
├── COURSE_BUILD_PLAN.md (this file)
├── k8s-operators-course-syllabus.md
├── scripts/
│   ├── setup-kind-cluster.sh
│   ├── setup-dev-environment.sh
│   └── cleanup.sh
├── module-01/
│   ├── README.md
│   ├── lessons/
│   │   ├── 01-control-plane.md
│   │   ├── 02-api-machinery.md
│   │   ├── 03-controller-pattern.md
│   │   └── 04-custom-resources.md
│   ├── diagrams/
│   │   └── *.mmd
│   ├── labs/
│   │   └── lab-*.md
│   └── solutions/
├── module-02/
│   └── [similar structure]
├── ...
└── examples/
    ├── hello-world-operator/
    ├── postgres-operator/
    └── final-project-template/
```

### Content Format

Each lesson should follow this structure:

1. **Brief Introduction** (1-2 paragraphs)
2. **Concept Explanation** (with Mermaid diagrams)
3. **Hands-on Exercise** (step-by-step)
4. **Key Takeaways** (bullet points)
5. **Further Reading** (optional)

### Mermaid Diagram Guidelines

- Use Mermaid for all architecture diagrams
- Include flowcharts for processes
- Use sequence diagrams for interactions
- Create state diagrams for state machines
- Keep diagrams simple and focused
- Include diagram files (`.mmd`) alongside content

### Hands-on Lab Guidelines

- Provide clear step-by-step instructions
- Include expected outputs
- Add troubleshooting sections
- Provide solutions (in solutions directory)
- Use kind cluster for all labs
- Include cleanup steps

---

## Build Order

1. **Module 1** - Foundation concepts
2. **Module 2** - First operator experience
3. **Module 3** - Core controller development
4. **Module 4** - Advanced patterns
5. **Module 5** - Webhooks
6. **Module 6** - Testing
7. **Module 7** - Production readiness
8. **Module 8** - Advanced topics

Each module builds on previous ones, so build sequentially.

---

## Quality Checklist

For each module, ensure:

- [ ] All Mermaid diagrams render correctly
- [ ] All code examples work with specified kubebuilder version
- [ ] All labs tested on kind cluster
- [ ] Solutions provided for all exercises
- [ ] Clear learning objectives met
- [ ] Practical examples over theory
- [ ] Proper code formatting and comments
- [ ] Links between related concepts
- [ ] Troubleshooting guides included

---

## Notes

- Focus on kubebuilder throughout (not Operator SDK)
- Use kind cluster for all hands-on labs
- Keep theory minimal, maximize practical examples
- Use Mermaid diagrams extensively
- Build one module at a time
- Test all examples before including
- Provide complete, working code examples

