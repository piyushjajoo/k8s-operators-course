# Module 1: Kubernetes Architecture Deep Dive

## Overview

This module provides a deep understanding of Kubernetes architecture, focusing on the components and patterns that operators build upon. You'll learn how the control plane works, how the API machinery operates, the controller pattern, and how custom resources extend Kubernetes.

**Duration:** 5-6 hours  
**Prerequisites:** Basic Kubernetes knowledge, kubectl familiarity

## Learning Objectives

By the end of this module, you will:

- Understand Kubernetes control plane components and their interactions
- Know how the Kubernetes API machinery works
- Comprehend the controller pattern and reconciliation loops
- Understand Custom Resource Definitions (CRDs) and when to use them

## Module Structure

1. **Lesson 1.1**: Kubernetes Control Plane Review
2. **Lesson 1.2**: Kubernetes API Machinery
3. **Lesson 1.3**: The Controller Pattern
4. **Lesson 1.4**: Custom Resources

## Setup

Before starting, ensure you have:

1. Completed the development environment setup:
   ```bash
   ./scripts/setup-dev-environment.sh
   ```

2. Created a kind cluster:
   ```bash
   ./scripts/setup-kind-cluster.sh
   ```

3. Verified cluster access:
   ```bash
   kubectl cluster-info
   kubectl get nodes
   ```

## Hands-on Labs

Each lesson includes hands-on exercises. All labs use the kind cluster you set up.

- **Lab 1.1**: Exploring the Control Plane
- **Lab 1.2**: Working with the Kubernetes API
- **Lab 1.3**: Observing Controllers in Action
- **Lab 1.4**: Creating Your First CRD

## Next Steps

After completing this module, proceed to [Module 2: Introduction to Operators](../module-02/README.md).

