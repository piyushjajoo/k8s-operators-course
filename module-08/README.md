# Module 8: Advanced Topics and Real-World Patterns

## Overview

Congratulations on reaching the final module! Now that you can build production-ready operators ([Module 7](../module-07/README.md)), it's time to explore advanced topics and real-world patterns. This module covers multi-tenancy, operator composition, stateful application management, and best practices from popular operators.

**Duration:** 6-7 hours  
**Prerequisites:** 
- Completion of all previous modules (Modules 1-7)
- Production-ready operator from Module 7
- Understanding of advanced Kubernetes concepts

## Learning Objectives

By the end of this module, you will:

- Build cluster-scoped and multi-tenant operators
- Compose multiple operators together
- Manage stateful applications with backups and migrations
- Understand real-world operator patterns
- Apply best practices from popular operators

## Module Structure

1. **[Lesson 8.1: Multi-Tenancy and Namespace Isolation](lessons/01-multi-tenancy.md)**
   - [Lab 8.1: Building Multi-Tenant Operator](labs/lab-01-multi-tenancy.md)

2. **[Lesson 8.2: Operator Composition](lessons/02-operator-composition.md)**
   - [Lab 8.2: Composing Operators](labs/lab-02-operator-composition.md)

3. **[Lesson 8.3: Stateful Application Management](lessons/03-stateful-applications.md)**
   - [Lab 8.3: Managing Stateful Applications](labs/lab-03-stateful-applications.md)

4. **[Lesson 8.4: Real-World Patterns and Best Practices](lessons/04-real-world-patterns.md)**
   - [Lab 8.4: Final Project](labs/lab-04-final-project.md)

## Prerequisites Check

Before starting, ensure you've completed:

- ✅ [Module 7](../module-07/README.md): Production-ready operator
- ✅ Have a working operator with all features
- ✅ Understand production deployment
- ✅ Ready for advanced topics

If you haven't completed Module 7, start with [Module 7: Production Considerations](../module-07/README.md).

## What You'll Build

Throughout this module, you'll extend your Database operator with:

- Multi-tenant support
- Backup and restore functionality
- Advanced stateful application patterns
- Real-world best practices

## Setup

Before starting this module:

1. **Have your production-ready operator from Module 7:**
   - Should be packaged and deployed
   - Should have HA and security configured
   - Ready for advanced features

2. **Ensure development environment is ready:**
   ```bash
   ./scripts/setup-dev-environment.sh
   ```

3. **Have a kind cluster running:**
   ```bash
   ./scripts/setup-kind-cluster.sh
   ```

## Hands-on Labs

Each lesson includes hands-on exercises that add advanced features to your operator.

- [Lab 8.1: Building Multi-Tenant Operator](labs/lab-01-multi-tenancy.md)
- [Lab 8.2: Composing Operators](labs/lab-02-operator-composition.md)
- [Lab 8.3: Managing Stateful Applications](labs/lab-03-stateful-applications.md)
- [Lab 8.4: Final Project](labs/lab-04-final-project.md)

## Solutions

Complete working solutions for all labs are available in the [solutions directory](solutions/):
- [Lab 8.1 Solutions](solutions/) - Multi-tenant operator examples
  - [cluster-scoped-crd.yaml](solutions/cluster-scoped-crd.yaml) - Cluster-scoped CRD example
  - [multi-tenant-controller.go](solutions/multi-tenant-controller.go) - Multi-tenant controller implementation
- [Lab 8.2 Solutions](solutions/) - Operator composition examples
  - [backup-operator.go](solutions/backup-operator.go) - Complete backup operator
  - [operator-coordination.go](solutions/operator-coordination.go) - Operator coordination examples
- [Lab 8.3 Solutions](solutions/) - Stateful application management
  - [backup.go](solutions/backup.go) - Backup functionality implementation
  - [restore.go](solutions/restore.go) - Restore functionality implementation
  - [rolling-update.go](solutions/rolling-update.go) - Rolling update handling
- [Lab 8.4 Solutions](solutions/) - Final project examples
  - See [solutions README](solutions/README.md) for complete examples


## Navigation

- [← Previous: Module 7 - Production Considerations](../module-07/README.md)
- [Course Overview](../README.md)

