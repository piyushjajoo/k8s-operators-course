---
layout: default
title: Home
nav_order: 1
description: "A comprehensive, hands-on course for building production-ready Kubernetes operators using Kubebuilder"
permalink: /
mermaid: true
---

# Building Kubernetes Operators

A comprehensive, hands-on course for building production-ready Kubernetes operators using Kubebuilder.
{: .fs-6 .fw-300 }

[Get Started]({{ site.baseurl }}/modules){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View on GitHub](https://github.com/piyushjajoo/k8s-operators-course){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## Course Overview

This course teaches you how to build Kubernetes operators from the ground up. You'll learn the fundamentals of Kubernetes architecture, the controller pattern, and how to use Kubebuilder to create custom operators that manage complex applications.

| **Duration** | 8 weeks (40-50 hours total) |
| **Level** | Intermediate to Advanced |
| **Prerequisites** | Basic Kubernetes knowledge, Go programming fundamentals, understanding of containerization |
| **License** | Free and open-source - [MIT License](https://github.com/piyushjajoo/k8s-operators-course/blob/main/LICENSE) |

---

## What You'll Learn

| Module | Description |
|:-------|:------------|
| **Module 1: Kubernetes Architecture Deep Dive** | Learn how the control plane works, API machinery operates, and understand the controller pattern. |
| **Module 2: Introduction to Operators** | Understand the operator pattern and build your first operator with Kubebuilder. |
| **Module 3: Building Custom Controllers** | Master controller-runtime, API design, and reconciliation logic. |
| **Module 4: Advanced Reconciliation Patterns** | Handle conditions, finalizers, watching, and advanced patterns. |
| **Module 5: Webhooks and Admission Control** | Implement validating and mutating webhooks. |
| **Module 6: Testing and Debugging** | Unit testing, integration testing, and observability. |
| **Module 7: Production Considerations** | Packaging, RBAC, high availability, and performance. |
| **Module 8: Advanced Topics** | Multi-tenancy, operator composition, and real-world patterns. |

---

## Learning Approach

This course emphasizes:

- **Practical Learning** — Every concept is demonstrated through hands-on exercises
- **Visual Learning** — Extensive use of Mermaid diagrams for architecture and flows
- **Progressive Complexity** — Start simple, build to production-ready operators
- **Real-world Examples** — Build actual operators you can use

---

## Prerequisites

Before starting, ensure you have:

```bash
# Required tools
Go 1.24+
kubectl
Docker or Podman
kind v0.29+
Kubebuilder 4.7+
```

---

## Quick Start

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

4. **Start with Module 1:**
   Navigate to the [Modules]({{ site.baseurl }}/modules) page to begin!

---

## Resources

- [Kubebuilder Documentation](https://book.kubebuilder.io/)
- [Kubernetes API Documentation](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [Hello World Operator Code](https://github.com/piyushjajoo/hello-world-operator) — Built following this course
- [Postgres Operator Code](https://github.com/piyushjajoo/postgres-operator) — Built following this course

---

## Share Your Project

If you've completed the course and built an operator, share your project on LinkedIn and tag [Piyush Jajoo](https://www.linkedin.com/in/pjajoo) for feedback! Please consider ⭐ing the project if you found it useful.

---

## Contributing

We welcome contributions! Open an issue for bugs, typos, or feature requests.

{: .note }
This course is **free and open-source** under the MIT License.

