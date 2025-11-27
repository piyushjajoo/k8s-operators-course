# Lab 8.4: Final Project

**Related Lesson:** [Lesson 8.4: Real-World Patterns and Best Practices](../lessons/04-real-world-patterns.md)  
**Navigation:** [← Previous Lab: Stateful Applications](lab-03-stateful-applications.md) | [Module Overview](../README.md)

## Objectives

Build a complete, production-ready operator for a stateful application that demonstrates all concepts learned throughout the course.

## Project Requirements

Your final operator must include:

1. **Full CRUD Operations**
   - Create, Read, Update, Delete
   - Proper error handling
   - Idempotent operations

2. **Status Reporting**
   - Status subresource
   - Conditions
   - Progress tracking
   - Observed generation

3. **Webhooks**
   - Validating webhooks
   - Mutating webhooks
   - Default values
   - Validation rules

4. **Testing**
   - Unit tests
   - Integration tests
   - Test coverage > 80%

5. **Production Features**
   - RBAC configuration
   - Security hardening
   - High availability
   - Performance optimization

6. **Advanced Features**
   - Multi-tenancy support
   - Backup/restore
   - Rolling updates
   - Documentation

## Project Structure

```
final-operator/
├── api/
│   └── v1/
│       ├── groupversion_info.go
│       └── <resource>_types.go
├── controllers/
│   └── <resource>_controller.go
├── config/
│   ├── crd/
│   ├── rbac/
│   ├── manager/
│   └── webhook/
├── internal/
│   ├── backup/
│   └── restore/
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Exercise 1: Choose Your Application

### Options

1. **Database Operator** (PostgreSQL, MySQL, MongoDB)
2. **Message Queue Operator** (RabbitMQ, Kafka)
3. **Cache Operator** (Redis, Memcached)
4. **Search Engine Operator** (Elasticsearch)
5. **Your Choice** (Any stateful application)

### Task 1.1: Define Your CRD

Create comprehensive API types:

```go
type <Resource>Spec struct {
    Image     string `json:"image"`
    Replicas  int32  `json:"replicas"`
    Storage   StorageSpec `json:"storage"`
    Resources corev1.ResourceRequirements `json:"resources,omitempty"`
    // Add your application-specific fields
}

type <Resource>Status struct {
    Conditions []metav1.Condition `json:"conditions,omitempty"`
    Phase      string            `json:"phase,omitempty"`
    ReadyReplicas int32          `json:"readyReplicas,omitempty"`
    Endpoint   string            `json:"endpoint,omitempty"`
    ObservedGeneration int64    `json:"observedGeneration,omitempty"`
}
```

## Exercise 2: Implement Core Functionality

### Task 2.1: Create Controller

Implement full reconciliation logic:

```go
func (r *<Resource>Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Get resource
    // Check deletion
    // Reconcile state
    // Update status
    // Handle errors
}
```

### Task 2.2: Implement Status Management

```go
func (r *<Resource>Reconciler) updateStatus(ctx context.Context, resource *<resource>v1.<Resource>) error {
    // Update conditions
    // Update phase
    // Update observed generation
    // Update status fields
}
```

## Exercise 3: Add Webhooks

### Task 3.1: Validating Webhook

```go
func (r *<Resource>) ValidateCreate() error {
    // Validate on create
}

func (r *<Resource>) ValidateUpdate(old runtime.Object) error {
    // Validate on update
}

func (r *<Resource>) ValidateDelete() error {
    // Validate on delete
}
```

### Task 3.2: Mutating Webhook

```go
func (r *<Resource>) Default() {
    // Set defaults
}
```

## Exercise 4: Add Testing

### Task 4.1: Unit Tests

```go
func Test<Resource>Reconciler(t *testing.T) {
    // Test reconciliation
    // Test error handling
    // Test status updates
}
```

### Task 4.2: Integration Tests

```go
func Test<Resource>Integration(t *testing.T) {
    // Test full lifecycle
    // Test backup/restore
    // Test scaling
}
```

## Exercise 5: Production Features

### Task 5.1: RBAC

```bash
# Generate RBAC
make manifests

# Review and optimize
cat config/rbac/role.yaml
```

### Task 5.2: Security

- Use distroless images
- Run as non-root
- Apply security contexts
- Add network policies

### Task 5.3: High Availability

- Enable leader election
- Deploy multiple replicas
- Add Pod Disruption Budget

## Exercise 6: Documentation

### Task 6.1: README

Create comprehensive README with:
- Quick start
- Architecture
- API documentation
- Examples
- Troubleshooting

### Task 6.2: Examples

Create example resources:
- Basic usage
- Advanced scenarios
- Multi-tenant setup
- Backup/restore

## Submission Checklist

- [ ] Full CRUD operations implemented
- [ ] Status reporting complete
- [ ] Webhooks configured
- [ ] Tests written (>80% coverage)
- [ ] RBAC optimized
- [ ] Security hardened
- [ ] HA configured
- [ ] Performance optimized
- [ ] Documentation complete
- [ ] Examples provided
- [ ] Operator packaged
- [ ] Ready for production

## Evaluation Criteria

1. **Functionality** (30%)
   - All features work correctly
   - Handles edge cases
   - Error handling

2. **Code Quality** (20%)
   - Clean, readable code
   - Follows best practices
   - Well-structured

3. **Testing** (20%)
   - Comprehensive tests
   - Good coverage
   - Integration tests

4. **Production Ready** (20%)
   - Security configured
   - HA enabled
   - Performance optimized

5. **Documentation** (10%)
   - Clear documentation
   - Good examples
   - Troubleshooting guide

## Solutions

Complete example solutions are available in the [solutions directory](../solutions/):
- [Final Project Template](../solutions/final-project-template/) - Complete project structure
- [Best Practices](../solutions/best-practices.md) - Best practices guide

## Congratulations!

You've completed the entire Kubernetes Operators course! You now have the knowledge and skills to build production-ready operators.

**Navigation:** [← Previous Lab: Stateful Applications](lab-03-stateful-applications.md) | [Related Lesson](../lessons/04-real-world-patterns.md) | [Module Overview](../README.md) | [Course Overview](../../README.md)

