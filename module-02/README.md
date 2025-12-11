# Module 2: Introduction to Operators

## Overview

Now that you understand Kubernetes architecture, the controller pattern, and Custom Resources (from [Module 1](../module-01/README.md)), it's time to build your first operator! This module introduces the operator pattern and teaches you how to use Kubebuilder to create operators that manage custom resources.

**Duration:** 5-6 hours  
**Prerequisites:** 
- Completion of [Module 1: Kubernetes Architecture Deep Dive](../module-01/README.md)
- Basic Kubernetes knowledge
- Go programming fundamentals

## Learning Objectives

By the end of this module, you will:

- Understand the operator pattern and when to use it
- Know how Kubebuilder works and its architecture
- Have a complete development environment set up
- Build and run your first "Hello World" operator
- Understand operator project structure

## Module Structure

1. **[Lesson 2.1: The Operator Pattern](lessons/01-operator-pattern.md)**
   - [Lab 2.1: Exploring Existing Operators](labs/lab-01-operator-pattern.md)

2. **[Lesson 2.2: Kubebuilder Fundamentals](lessons/02-kubebuilder-fundamentals.md)**
   - [Lab 2.2: Kubebuilder CLI and Project Structure](labs/lab-02-kubebuilder-fundamentals.md)

3. **[Lesson 2.3: Development Environment Setup](lessons/03-dev-environment.md)**
   - [Lab 2.3: Setting Up Your Environment](labs/lab-03-dev-environment.md)

4. **[Lesson 2.4: Your First Operator](lessons/04-first-operator.md)**
   - [Lab 2.4: Building Hello World Operator](labs/lab-04-first-operator.md)

## Prerequisites Check

Before starting, ensure you've completed Module 1:

- ✅ Understand Kubernetes control plane components
- ✅ Know how the API machinery works
- ✅ Understand the controller pattern and reconciliation
- ✅ Can create and use Custom Resources (CRDs)

If you haven't completed Module 1, start with [Module 1: Kubernetes Architecture Deep Dive](../module-01/README.md).

## Setup

Before starting this module:

1. **Verify Module 1 completion:**
   - You should understand CRDs from [Lesson 1.4](../module-01/lessons/04-custom-resources.md)
   - You should understand controllers from [Lesson 1.3](../module-01/lessons/03-controller-pattern.md)

2. **Ensure development environment is ready:**
   ```bash
   ./scripts/setup-dev-environment.sh
   ```

3. **Have a kind cluster running:**
   ```bash
   ./scripts/setup-kind-cluster.sh
   ```

4. **Verify kubebuilder is installed:**
   ```bash
   kubebuilder version
   ```

## Hands-on Labs

Each lesson includes hands-on exercises. All labs use the kind cluster and kubebuilder.

- [Lab 2.1: Exploring Existing Operators](labs/lab-01-operator-pattern.md)
- [Lab 2.2: Kubebuilder CLI and Project Structure](labs/lab-02-kubebuilder-fundamentals.md)
- [Lab 2.3: Setting Up Your Environment](labs/lab-03-dev-environment.md)
- [Lab 2.4: Building Hello World Operator](labs/lab-04-first-operator.md)

## What You'll Build

By the end of this module, you'll have:

- A complete "Hello World" operator that manages a custom resource
- Understanding of kubebuilder project structure
- Ability to run operators locally for development
- Foundation for building more complex operators in Module 3

## Solutions

Complete working solutions for all labs are available in the [solutions directory](solutions/):
- [Lab 2.4 Solutions](solutions/) - Complete Hello World operator (main.go, controller, types)


## Navigation

- [← Previous: Module 1 - Kubernetes Architecture](../module-01/README.md)
- [Course Overview](../README.md)
- [Next: Module 3 - Building Custom Controllers →](../module-03/README.md)

