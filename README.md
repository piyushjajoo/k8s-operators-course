# Building Kubernetes Operators Course

A comprehensive, hands-on course for building production-ready Kubernetes operators using Kubebuilder.

## Course Overview

This course teaches you how to build Kubernetes operators from the ground up. You'll learn the fundamentals of Kubernetes architecture, the controller pattern, and how to use Kubebuilder to create custom operators that manage complex applications.

**Duration:** 8 weeks (40-50 hours total)  
**Level:** Intermediate to Advanced  
**Prerequisites:** Basic Kubernetes knowledge, Go programming fundamentals, understanding of containerization  
**License:** Free and open-source - Licensed under [MIT License](LICENSE)

## Course Structure

The course is divided into 8 modules, each building on the previous:

1. **[Module 1: Kubernetes Architecture Deep Dive](module-01/README.md)** ✅
   - [Control plane components](module-01/lessons/01-control-plane.md)
   - [API machinery](module-01/lessons/02-api-machinery.md)
   - [Controller pattern](module-01/lessons/03-controller-pattern.md)
   - [Custom Resources](module-01/lessons/04-custom-resources.md)

2. **[Module 2: Introduction to Operators](module-02/README.md)** ✅
   - [Operator pattern](module-02/lessons/01-operator-pattern.md)
   - [Kubebuilder fundamentals](module-02/lessons/02-kubebuilder-fundamentals.md)
   - [Development environment](module-02/lessons/03-dev-environment.md)
   - [Your first operator](module-02/lessons/04-first-operator.md)

3. **[Module 3: Building Custom Controllers](module-03/README.md)** ✅
   - [Controller runtime](module-03/lessons/01-controller-runtime.md)
   - [API design](module-03/lessons/02-designing-api.md)
   - [Reconciliation logic](module-03/lessons/03-reconciliation-logic.md)
   - [Client operations](module-03/lessons/04-client-go.md)

4. **[Module 4: Advanced Reconciliation Patterns](module-04/README.md)** ✅
   - [Conditions and status](module-04/lessons/01-conditions-status.md)
   - [Finalizers and cleanup](module-04/lessons/02-finalizers-cleanup.md)
   - [Watching and indexing](module-04/lessons/03-watching-indexing.md)
   - [Advanced patterns](module-04/lessons/04-advanced-patterns.md)

5. **[Module 5: Webhooks and Admission Control](module-05/README.md)** ✅
   - [Admission control](module-05/lessons/01-admission-control.md)
   - [Validating webhooks](module-05/lessons/02-validating-webhooks.md)
   - [Mutating webhooks](module-05/lessons/03-mutating-webhooks.md)
   - [Webhook deployment](module-05/lessons/04-webhook-deployment.md)

6. **[Module 6: Testing and Debugging](module-06/README.md)** ✅
   - [Testing fundamentals](module-06/lessons/01-testing-fundamentals.md)
   - [Unit testing](module-06/lessons/02-unit-testing-envtest.md)
   - [Integration testing](module-06/lessons/03-integration-testing.md)
   - [Debugging and observability](module-06/lessons/04-debugging-observability.md)

7. **[Module 7: Production Considerations](module-07/README.md)** ✅
   - [Packaging and distribution](module-07/lessons/01-packaging-distribution.md)
   - [RBAC and security](module-07/lessons/02-rbac-security.md)
   - [High availability](module-07/lessons/03-high-availability.md)
   - [Performance and scalability](module-07/lessons/04-performance-scalability.md)

8. **[Module 8: Advanced Topics and Real-World Patterns](module-08/README.md)** ✅
   - [Multi-tenancy and namespace isolation](module-08/lessons/01-multi-tenancy.md)
   - [Operator composition](module-08/lessons/02-operator-composition.md)
   - [Stateful application management](module-08/lessons/03-stateful-applications.md)
   - [Real-world patterns and best practices](module-08/lessons/04-real-world-patterns.md)

## Quick Navigation

### Module 1: Kubernetes Architecture Deep Dive ✅

- [Module Overview](module-01/README.md)
- [Lesson 1.1: Control Plane](module-01/lessons/01-control-plane.md) | [Lab 1.1](module-01/labs/lab-01-control-plane.md)
- [Lesson 1.2: API Machinery](module-01/lessons/02-api-machinery.md) | [Lab 1.2](module-01/labs/lab-02-api-machinery.md)
- [Lesson 1.3: Controller Pattern](module-01/lessons/03-controller-pattern.md) | [Lab 1.3](module-01/labs/lab-03-controller-pattern.md)
- [Lesson 1.4: Custom Resources](module-01/lessons/04-custom-resources.md) | [Lab 1.4](module-01/labs/lab-04-custom-resources.md)

### Module 2: Introduction to Operators ✅

- [Module Overview](module-02/README.md)
- [Lesson 2.1: Operator Pattern](module-02/lessons/01-operator-pattern.md) | [Lab 2.1](module-02/labs/lab-01-operator-pattern.md)
- [Lesson 2.2: Kubebuilder Fundamentals](module-02/lessons/02-kubebuilder-fundamentals.md) | [Lab 2.2](module-02/labs/lab-02-kubebuilder-fundamentals.md)
- [Lesson 2.3: Dev Environment](module-02/lessons/03-dev-environment.md) | [Lab 2.3](module-02/labs/lab-03-dev-environment.md)
- [Lesson 2.4: First Operator](module-02/lessons/04-first-operator.md) | [Lab 2.4](module-02/labs/lab-04-first-operator.md)

### Module 3: Building Custom Controllers ✅

- [Module Overview](module-03/README.md)
- [Lesson 3.1: Controller Runtime](module-03/lessons/01-controller-runtime.md) | [Lab 3.1](module-03/labs/lab-01-controller-runtime.md)
- [Lesson 3.2: API Design](module-03/lessons/02-designing-api.md) | [Lab 3.2](module-03/labs/lab-02-designing-api.md)
- [Lesson 3.3: Reconciliation Logic](module-03/lessons/03-reconciliation-logic.md) | [Lab 3.3](module-03/labs/lab-03-reconciliation-logic.md)
- [Lesson 3.4: Client Operations](module-03/lessons/04-client-go.md) | [Lab 3.4](module-03/labs/lab-04-client-go.md)

### Module 4: Advanced Reconciliation Patterns ✅

- [Module Overview](module-04/README.md)
- [Lesson 4.1: Conditions and Status](module-04/lessons/01-conditions-status.md) | [Lab 4.1](module-04/labs/lab-01-conditions-status.md)
- [Lesson 4.2: Finalizers and Cleanup](module-04/lessons/02-finalizers-cleanup.md) | [Lab 4.2](module-04/labs/lab-02-finalizers-cleanup.md)
- [Lesson 4.3: Watching and Indexing](module-04/lessons/03-watching-indexing.md) | [Lab 4.3](module-04/labs/lab-03-watching-indexing.md)
- [Lesson 4.4: Advanced Patterns](module-04/lessons/04-advanced-patterns.md) | [Lab 4.4](module-04/labs/lab-04-advanced-patterns.md)

### Module 5: Webhooks and Admission Control ✅

- [Module Overview](module-05/README.md)
- [Lesson 5.1: Admission Control](module-05/lessons/01-admission-control.md) | [Lab 5.1](module-05/labs/lab-01-admission-control.md)
- [Lesson 5.2: Validating Webhooks](module-05/lessons/02-validating-webhooks.md) | [Lab 5.2](module-05/labs/lab-02-validating-webhooks.md)
- [Lesson 5.3: Mutating Webhooks](module-05/lessons/03-mutating-webhooks.md) | [Lab 5.3](module-05/labs/lab-03-mutating-webhooks.md)
- [Lesson 5.4: Webhook Deployment](module-05/lessons/04-webhook-deployment.md) | [Lab 5.4](module-05/labs/lab-04-webhook-deployment.md)

### Module 6: Testing and Debugging ✅

- [Module Overview](module-06/README.md)
- [Lesson 6.1: Testing Fundamentals](module-06/lessons/01-testing-fundamentals.md) | [Lab 6.1](module-06/labs/lab-01-testing-fundamentals.md)
- [Lesson 6.2: Unit Testing](module-06/lessons/02-unit-testing-envtest.md) | [Lab 6.2](module-06/labs/lab-02-unit-testing-envtest.md)
- [Lesson 6.3: Integration Testing](module-06/lessons/03-integration-testing.md) | [Lab 6.3](module-06/labs/lab-03-integration-testing.md)
- [Lesson 6.4: Debugging and Observability](module-06/lessons/04-debugging-observability.md) | [Lab 6.4](module-06/labs/lab-04-debugging-observability.md)

### Module 7: Production Considerations ✅

- [Module Overview](module-07/README.md)
- [Lesson 7.1: Packaging and Distribution](module-07/lessons/01-packaging-distribution.md) | [Lab 7.1](module-07/labs/lab-01-packaging-distribution.md)
- [Lesson 7.2: RBAC and Security](module-07/lessons/02-rbac-security.md) | [Lab 7.2](module-07/labs/lab-02-rbac-security.md)
- [Lesson 7.3: High Availability](module-07/lessons/03-high-availability.md) | [Lab 7.3](module-07/labs/lab-03-high-availability.md)
- [Lesson 7.4: Performance and Scalability](module-07/lessons/04-performance-scalability.md) | [Lab 7.4](module-07/labs/lab-04-performance-scalability.md)

### Module 8: Advanced Topics and Real-World Patterns ✅

- [Module Overview](module-08/README.md)
- [Lesson 8.1: Multi-Tenancy](module-08/lessons/01-multi-tenancy.md) | [Lab 8.1](module-08/labs/lab-01-multi-tenancy.md)
- [Lesson 8.2: Operator Composition](module-08/lessons/02-operator-composition.md) | [Lab 8.2](module-08/labs/lab-02-operator-composition.md)
- [Lesson 8.3: Stateful Applications](module-08/lessons/03-stateful-applications.md) | [Lab 8.3](module-08/labs/lab-03-stateful-applications.md)
- [Lesson 8.4: Real-World Patterns](module-08/lessons/04-real-world-patterns.md) | [Lab 8.4](module-08/labs/lab-04-final-project.md)

## Getting Started

### Prerequisites

- Go 1.21+
- kubectl
- Docker or Podman
- kind
- kubebuilder

### Setup

1. **Clone this repository:**
   ```bash
   git clone <repository-url>
   cd k8s-operators-course
   ```

2. **Set up your development environment:**
   ```bash
   ./scripts/setup-dev-environment.sh
   ```

3. **Create a kind cluster:**
   ```bash
   ./scripts/setup-kind-cluster.sh
   ```

4. **Start with [Module 1](module-01/README.md):**
   ```bash
   cd module-01
   cat README.md
   ```

## Module Status

- ✅ Module 1: Complete
- ✅ Module 2: Complete
- ✅ Module 3: Complete
- ✅ Module 4: Complete
- ✅ Module 5: Complete
- ✅ Module 6: Complete
- ✅ Module 7: Complete
- ✅ Module 8: Complete

## Learning Approach

This course emphasizes:

- **Practical Learning:** Every concept is demonstrated through hands-on exercises
- **Visual Learning:** Extensive use of Mermaid diagrams for architecture and flows
- **Progressive Complexity:** Start simple, build to production-ready operators
- **Real-world Examples:** Build actual operators you can use

## Contributing

This is a course repository. If you find issues or have suggestions, please open an issue.

## License

This course is **free and open-source**, licensed under the [MIT License](LICENSE). You are free to:

- Use, share, and modify the course materials
- Use for personal or commercial purposes
- Distribute and sublicense the materials

The only requirement is that you include the original copyright notice and license text. See the [LICENSE](LICENSE) file for full details.

## Solutions

Complete working solutions for all labs are available in each module's `solutions/` directory:

- **Module 1**: [CRD examples](module-01/solutions/) - Website CRD and example resources
- **Module 2**: [Hello World operator](module-02/solutions/) - Complete operator (main.go, controller, types)
- **Module 3**: [Database operator](module-03/solutions/) - Complete Database operator (types, controller)
- **Module 4**: [Advanced patterns](module-04/solutions/) - Condition helpers, finalizers, watch setup
- **Module 5**: [Webhooks](module-05/solutions/) - Validating and mutating webhooks
- **Module 6**: [Testing](module-06/solutions/) - Unit tests, integration tests, metrics, and observability examples
- **Module 7**: [Production](module-07/solutions/) - Dockerfile, Helm chart, RBAC, security, HA, and performance examples
- **Module 8**: [Advanced](module-08/solutions/) - Multi-tenancy, operator composition, backup/restore, and rolling update examples

Each solution includes:
- Complete, working code examples
- Best practices from the lessons
- README files with usage instructions
- Ready to use as reference or starting point

## Resources

- [Kubebuilder Documentation](https://book.kubebuilder.io/)
- [Kubernetes API Documentation](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

## Support

For questions and discussions, please open an issue in this repository.

