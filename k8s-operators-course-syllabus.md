# Building Kubernetes Operators: Complete Course Syllabus

## Course Overview

**Duration:** 8 weeks (40-50 hours total)  
**Level:** Intermediate to Advanced  
**Prerequisites:** Basic Kubernetes knowledge, Go programming fundamentals, understanding of containerization

### Course Description

This comprehensive course teaches you how to build production-ready Kubernetes operators using the Operator SDK and Kubebuilder framework. You’ll learn to extend Kubernetes functionality by creating custom controllers that automate complex application management tasks.

### Learning Objectives

By the end of this course, you will be able to:

- Understand the Kubernetes operator pattern and its use cases
- Build custom controllers using Operator SDK and Kubebuilder
- Design and implement Custom Resource Definitions (CRDs)
- Manage complex stateful applications on Kubernetes
- Implement reconciliation loops and error handling
- Test, debug, and deploy operators to production
- Apply best practices for operator development

-----

## Module 1: Kubernetes Architecture Deep Dive

**Duration:** Week 1 (5-6 hours)

### Topics Covered

**1.1 Kubernetes Control Plane Review**

- API Server architecture and request flow
- Controller Manager and its built-in controllers
- Scheduler fundamentals
- etcd as the source of truth

**1.2 Kubernetes API Machinery**

- RESTful API design in Kubernetes
- Resource types: objects, lists, and subresources
- API versioning and groups
- Understanding API server extensions

**1.3 The Controller Pattern**

- Control loops and reconciliation
- Declarative vs imperative management
- Watch mechanisms and informers
- Leader election concepts

**1.4 Custom Resources**

- What are Custom Resource Definitions (CRDs)
- When to use CRDs vs ConfigMaps
- CRD structure and validation
- Schema and OpenAPI specifications

### Hands-on Lab

- Explore existing controllers using kubectl and API calls
- Analyze built-in controller behavior
- Create a simple CRD manually

### Assessment

- Quiz on Kubernetes architecture concepts
- Design exercise: Identify use cases for custom controllers

-----

## Module 2: Introduction to Operators

**Duration:** Week 1-2 (5-6 hours)

### Topics Covered

**2.1 The Operator Pattern**

- Definition and philosophy
- Operator capability levels (Basic to Auto Pilot)
- Comparing operators to Helm charts and other tools
- Real-world operator examples

**2.2 Operator Framework Overview**

- Operator SDK introduction
- Kubebuilder fundamentals
- KUDO and other frameworks
- Choosing the right framework

**2.3 Development Environment Setup**

- Installing Go and dependencies
- Setting up Operator SDK
- Configuring kubectl and cluster access
- Development cluster options (kind, minikube, k3d)

**2.4 Your First Operator**

- Scaffolding an operator project
- Project structure walkthrough
- Understanding generated code
- Running the operator locally

### Hands-on Lab

- Set up complete development environment
- Create and run a “Hello World” operator
- Deploy a CRD to your cluster
- Observe controller logs and behavior

### Assessment

- Environment verification checklist
- Simple operator modification exercise

-----

## Module 3: Building Custom Controllers

**Duration:** Week 2-3 (6-7 hours)

### Topics Covered

**3.1 Controller Runtime Deep Dive**

- Manager, reconciler, and client architecture
- Understanding the Reconcile function
- Result and error handling
- Requeue strategies

**3.2 Designing Your API**

- Spec and Status conventions
- Naming and versioning APIs
- Defining meaningful fields
- Validation and defaults using markers

**3.3 Implementing Reconciliation Logic**

- The reconciliation loop lifecycle
- Reading cluster state
- Creating and updating resources
- Owner references and garbage collection

**3.4 Working with Client-Go**

- Typed vs dynamic clients
- Listing and watching resources
- Patch strategies
- Optimistic concurrency control

### Hands-on Lab

- Build a database operator that deploys PostgreSQL
- Implement basic reconciliation logic
- Handle resource creation and updates
- Test with various scenarios

### Assessment

- Code review of reconciliation implementation
- Debugging exercise with intentional bugs

-----

## Module 4: Advanced Reconciliation Patterns

**Duration:** Week 3-4 (6-7 hours)

### Topics Covered

**4.1 Conditions and Status Management**

- Implementing status subresources
- Using conditions effectively
- Reporting progress and errors
- Status update strategies

**4.2 Finalizers and Cleanup**

- Understanding finalizers
- Implementing pre-delete hooks
- Resource cleanup patterns
- Avoiding finalizer deadlocks

**4.3 Watching and Indexing**

- Setting up watches for dependent resources
- Using indexes for efficient lookups
- Handling watch events
- Cross-namespace watching

**4.4 Advanced Patterns**

- Multi-phase reconciliation
- State machines in controllers
- Handling external dependencies
- Idempotency and stability

### Hands-on Lab

- Add status conditions to your operator
- Implement proper cleanup with finalizers
- Create watches for child resources
- Handle complex multi-step deployments

### Assessment

- Implement a specific reconciliation pattern
- Troubleshooting exercise

-----

## Module 5: Webhooks and Admission Control

**Duration:** Week 4-5 (5-6 hours)

### Topics Covered

**5.1 Kubernetes Admission Control**

- Admission controller overview
- Mutating vs validating webhooks
- Webhook configuration and registration

**5.2 Implementing Validating Webhooks**

- Validation logic implementation
- Handling webhook requests
- Error responses and user feedback
- Schema validation vs webhook validation

**5.3 Implementing Mutating Webhooks**

- Defaulting values
- Mutating requests
- JSON patch operations
- Common mutation patterns

**5.4 Webhook Deployment and Certificates**

- Certificate management strategies
- Using cert-manager
- Webhook service setup
- Testing webhooks locally

### Hands-on Lab

- Add validation webhook to your operator
- Implement defaulting with mutation webhook
- Configure certificate management
- Test webhook behavior with valid and invalid resources

### Assessment

- Implement custom validation rules
- Webhook debugging challenge

-----

## Module 6: Testing and Debugging

**Duration:** Week 5-6 (6-7 hours)

### Topics Covered

**6.1 Unit Testing Controllers**

- Using envtest for controller testing
- Mocking Kubernetes clients
- Testing reconciliation logic
- Table-driven tests

**6.2 Integration Testing**

- Setting up test clusters
- End-to-end test patterns
- Using Ginkgo and Gomega
- CI/CD integration

**6.3 Debugging Techniques**

- Using delve debugger with operators
- Reading controller logs effectively
- Tracing reconciliation loops
- Common pitfalls and solutions

**6.4 Observability**

- Adding structured logging
- Exposing Prometheus metrics
- Using events effectively
- Tracing with OpenTelemetry

### Hands-on Lab

- Write comprehensive unit tests for your operator
- Create integration test suite
- Debug a failing reconciliation
- Add metrics and monitoring

### Assessment

- Test coverage exercise
- Debug a broken operator scenario

-----

## Module 7: Production Considerations

**Duration:** Week 6-7 (6-7 hours)

### Topics Covered

**7.1 Packaging and Distribution**

- Building operator images
- Creating Helm charts for operators
- OLM (Operator Lifecycle Manager) bundles
- Versioning and upgrades

**7.2 RBAC and Security**

- Principle of least privilege
- Creating appropriate roles and bindings
- Service account configuration
- Security scanning and hardening

**7.3 High Availability**

- Leader election implementation
- Multiple controller replicas
- Handling failover
- Resource limits and requests

**7.4 Performance and Scalability**

- Rate limiting and backoff
- Batch reconciliation
- Caching strategies
- Managing large clusters

### Hands-on Lab

- Package your operator for distribution
- Configure proper RBAC
- Implement leader election
- Load test your operator

### Assessment

- Security audit exercise
- Performance optimization challenge

-----

## Module 8: Advanced Topics and Real-World Patterns

**Duration:** Week 7-8 (6-7 hours)

### Topics Covered

**8.1 Multi-Tenancy and Namespace Isolation**

- Cluster-scoped vs namespaced operators
- Handling multi-tenant scenarios
- Resource quotas and limits
- Namespace lifecycle management

**8.2 Operator Composition**

- Using multiple operators together
- Managing dependencies between operators
- Composite operators pattern
- Coordination strategies

**8.3 Stateful Application Management**

- Managing StatefulSets
- Backup and restore patterns
- Rolling updates and migrations
- Data consistency guarantees

**8.4 Case Studies and Best Practices**

- Analysis of popular operators (Prometheus, Elasticsearch, etc.)
- Common anti-patterns to avoid
- Documentation and user experience
- Community and contribution guidelines

### Hands-on Lab

- Extend your operator with advanced features
- Implement backup/restore functionality
- Create comprehensive documentation
- Deploy to a production-like environment

### Final Project

Build a complete operator for a stateful application of your choice that includes:

- Full CRUD operations
- Status reporting
- Webhooks for validation
- Backup/restore capabilities
- Comprehensive tests
- Production-ready packaging

### Assessment

- Final project presentation and code review
- Written reflection on lessons learned

-----

## Course Materials

### Required Tools

- Kubernetes cluster (v1.24+)
- Go 1.21+
- Operator SDK
- kubectl
- Docker or Podman
- IDE with Go support (VS Code, GoLand)

### Recommended Reading

- “Kubernetes Operators” by Jason Dobies and Joshua Wood
- “Programming Kubernetes” by Michael Hausenblas and Stefan Schimanski
- Kubernetes documentation on Custom Resources
- Operator SDK documentation

### Additional Resources

- OperatorHub.io for operator examples
- Kubernetes Slack channels
- CNCF operator working group materials
- Sample operator repositories

-----

## Grading and Certification

**Grading Breakdown:**

- Weekly labs and exercises: 40%
- Module assessments: 30%
- Final project: 30%

**Certification Requirements:**

- Complete all 8 modules
- Achieve 80% or higher on assessments
- Submit working final project
- Participate in code reviews

-----

## Support and Community

- Weekly office hours for Q&A
- Private Slack channel for course participants
- Peer code review sessions
- Access to instructor for 3 months post-course
