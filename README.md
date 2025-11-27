# Building Kubernetes Operators Course

A comprehensive, hands-on course for building production-ready Kubernetes operators using Kubebuilder.

## Course Overview

This course teaches you how to build Kubernetes operators from the ground up. You'll learn the fundamentals of Kubernetes architecture, the controller pattern, and how to use Kubebuilder to create custom operators that manage complex applications.

**Duration:** 8 weeks (40-50 hours total)  
**Level:** Intermediate to Advanced  
**Prerequisites:** Basic Kubernetes knowledge, Go programming fundamentals, understanding of containerization

## Course Structure

The course is divided into 8 modules, each building on the previous:

1. **[Module 1: Kubernetes Architecture Deep Dive](module-01/README.md)** ✅
   - [Control plane components](module-01/lessons/01-control-plane.md)
   - [API machinery](module-01/lessons/02-api-machinery.md)
   - [Controller pattern](module-01/lessons/03-controller-pattern.md)
   - [Custom Resources](module-01/lessons/04-custom-resources.md)

2. **Module 2: Introduction to Operators** (Coming Soon)
   - Operator pattern
   - Kubebuilder setup
   - Your first operator

3. **Module 3: Building Custom Controllers** (Coming Soon)
   - Controller runtime
   - API design
   - Reconciliation logic

4. **Module 4: Advanced Reconciliation Patterns** (Coming Soon)
   - Status management
   - Finalizers
   - Watching and indexing

5. **Module 5: Webhooks and Admission Control** (Coming Soon)
   - Validating webhooks
   - Mutating webhooks
   - Certificate management

6. **Module 6: Testing and Debugging** (Coming Soon)
   - Unit testing
   - Integration testing
   - Observability

7. **Module 7: Production Considerations** (Coming Soon)
   - Packaging and distribution
   - RBAC and security
   - High availability

8. **Module 8: Advanced Topics** (Coming Soon)
   - Multi-tenancy
   - Operator composition
   - Stateful applications

## Quick Navigation

### Module 1: Kubernetes Architecture Deep Dive ✅

- [Module Overview](module-01/README.md)
- [Lesson 1.1: Control Plane](module-01/lessons/01-control-plane.md) | [Lab 1.1](module-01/labs/lab-01-control-plane.md)
- [Lesson 1.2: API Machinery](module-01/lessons/02-api-machinery.md) | [Lab 1.2](module-01/labs/lab-02-api-machinery.md)
- [Lesson 1.3: Controller Pattern](module-01/lessons/03-controller-pattern.md) | [Lab 1.3](module-01/labs/lab-03-controller-pattern.md)
- [Lesson 1.4: Custom Resources](module-01/lessons/04-custom-resources.md) | [Lab 1.4](module-01/labs/lab-04-custom-resources.md)

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

## Course Materials

- **Syllabus:** [k8s-operators-course-syllabus.md](k8s-operators-course-syllabus.md)
- **Build Plan:** [COURSE_BUILD_PLAN.md](COURSE_BUILD_PLAN.md)

## Module Status

- ✅ Module 1: Complete
- 🚧 Module 2-8: In Progress

## Learning Approach

This course emphasizes:

- **Practical Learning:** Every concept is demonstrated through hands-on exercises
- **Visual Learning:** Extensive use of Mermaid diagrams for architecture and flows
- **Progressive Complexity:** Start simple, build to production-ready operators
- **Real-world Examples:** Build actual operators you can use

## Contributing

This is a course repository. If you find issues or have suggestions, please open an issue.

## License

See [LICENSE](LICENSE) file for details.

## Resources

- [Kubebuilder Documentation](https://book.kubebuilder.io/)
- [Kubernetes API Documentation](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

## Support

For questions and discussions, please open an issue in this repository.

