---
layout: default
title: Course Overview
nav_order: 0
nav_exclude: true
---

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

1. **[Module 1: Kubernetes Architecture Deep Dive](module-01/README.md)**
   - [Control plane components](module-01/lessons/01-control-plane.md)
   - [API machinery](module-01/lessons/02-api-machinery.md)
   - [Controller pattern](module-01/lessons/03-controller-pattern.md)
   - [Custom Resources](module-01/lessons/04-custom-resources.md)

2. **[Module 2: Introduction to Operators](module-02/README.md)**
   - [Operator pattern](module-02/lessons/01-operator-pattern.md)
   - [Kubebuilder fundamentals](module-02/lessons/02-kubebuilder-fundamentals.md)
   - [Development environment](module-02/lessons/03-dev-environment.md)
   - [Your first operator](module-02/lessons/04-first-operator.md)

3. **[Module 3: Building Custom Controllers](module-03/README.md)**
   - [Controller runtime](module-03/lessons/01-controller-runtime.md)
   - [API design](module-03/lessons/02-designing-api.md)
   - [Reconciliation logic](module-03/lessons/03-reconciliation-logic.md)
   - [Client operations](module-03/lessons/04-client-go.md)

4. **[Module 4: Advanced Reconciliation Patterns](module-04/README.md)**
   - [Conditions and status](module-04/lessons/01-conditions-status.md)
   - [Finalizers and cleanup](module-04/lessons/02-finalizers-cleanup.md)
   - [Watching and indexing](module-04/lessons/03-watching-indexing.md)
   - [Advanced patterns](module-04/lessons/04-advanced-patterns.md)

5. **[Module 5: Webhooks and Admission Control](module-05/README.md)**
   - [Admission control](module-05/lessons/01-admission-control.md)
   - [Validating webhooks](module-05/lessons/02-validating-webhooks.md)
   - [Mutating webhooks](module-05/lessons/03-mutating-webhooks.md)
   - [Webhook deployment](module-05/lessons/04-webhook-deployment.md)
   - [Conversion webhooks](module-05/lessons/05-conversion-webhooks.md)

6. **[Module 6: Testing and Debugging](module-06/README.md)**
   - [Testing fundamentals](module-06/lessons/01-testing-fundamentals.md)
   - [Unit testing](module-06/lessons/02-unit-testing-envtest.md)
   - [Integration testing](module-06/lessons/03-integration-testing.md)
   - [Debugging and observability](module-06/lessons/04-debugging-observability.md)

7. **[Module 7: Production Considerations](module-07/README.md)**
   - [Packaging and distribution](module-07/lessons/01-packaging-distribution.md)
   - [RBAC and security](module-07/lessons/02-rbac-security.md)
   - [High availability](module-07/lessons/03-high-availability.md)
   - [Performance and scalability](module-07/lessons/04-performance-scalability.md)

8. **[Module 8: Advanced Topics and Real-World Patterns](module-08/README.md)**
   - [Multi-tenancy and namespace isolation](module-08/lessons/01-multi-tenancy.md)
   - [Operator composition](module-08/lessons/02-operator-composition.md)
   - [Stateful application management](module-08/lessons/03-stateful-applications.md)
   - [Real-world patterns and best practices](module-08/lessons/04-real-world-patterns.md)

## Getting Started

### Prerequisites

- Go 1.24+
- kubectl
- Docker or Podman
- kind v0.29+
- Kubebuilder 4.7+

### Setup

1. **Clone this repository:**

   ```bash
   git clone https://github.com/piyushjajoo/k8s-operators-course.git
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

## Learning Approach

This course emphasizes:

- **Practical Learning:** Every concept is demonstrated through hands-on exercises
- **Visual Learning:** Extensive use of Mermaid diagrams for architecture and flows
- **Progressive Complexity:** Start simple, build to production-ready operators
- **Real-world Examples:** Build actual operators you can use

## Resources

- [Kubebuilder Documentation](https://book.kubebuilder.io/)
- [Kubernetes API Documentation](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [Complete Hello World Operator Code](https://github.com/piyushjajoo/hello-world-operator) built following this course, you can refer.
- [Complete Postgres Operator Code](https://github.com/piyushjajoo/postgres-operator) built following this course, you can refer.

## Contributing

We welcome contributions and feedback! Here's how you can help improve this course:

### Reporting Issues

If you find bugs, typos, or errors in the course materials, please open an issue in this repository.

### Requesting New Concepts

Have an idea for a new concept, topic, or module you'd like to see added to the course? We'd love to hear from you!

**To request a new concept:**

1. **Open a new issue** in this repository with the label `enhancement` (if available) or use the title prefix `[Feature Request]`
2. **Include the following information:**
   - **Concept/Topic Name:** What concept would you like to see covered?
   - **Description:** A brief description of the concept and why it would be valuable
   - **Suggested Module:** Which module do you think this fits best in? (or suggest a new module)
   - **Use Case:** How would this help learners build better operators?
   - **Priority:** Is this a nice-to-have or a critical gap in the course?

3. **Example format:**
   ```
   [Feature Request] Operator SDK Comparison
   
   Description: Add a lesson comparing Kubebuilder with Operator SDK
   Suggested Module: Module 2 or new comparison module
   Use Case: Help learners understand when to choose which framework
   Priority: Nice-to-have
   ```

We review all requests and prioritize based on:
- Community interest and upvotes
- Alignment with course learning objectives
- Complexity and time required to develop
- Gaps in current course coverage

**Note:** While we can't guarantee every request will be implemented, we value your input and will consider all suggestions!

## License

This course is **free and open-source**, licensed under the [MIT License](LICENSE). You are free to:

- Use, share, and modify the course materials
- Use for personal or commercial purposes
- Distribute and sublicense the materials

The only requirement is that you include the original copyright notice and license text. See the [LICENSE](LICENSE) file for full details.

## Share Your Project

If you've completed the course and built an operator, we'd love to see it! Share your project on LinkedIn and tag [Piyush Jajoo](https://www.linkedin.com/in/pjajoo). I'll make my best effort in my free time to review your code and provide feedback. Please consider ⭐ing the project if you found it useful.

## Support

For questions and discussions, please open an issue in this repository.
