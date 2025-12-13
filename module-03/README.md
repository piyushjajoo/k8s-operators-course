---
layout: default
title: "Module 3: Building Custom Controllers"
nav_order: 3
parent: Modules
has_children: true
has_toc: false
permalink: /module-03/
mermaid: true
---

# Module 3: Building Custom Controllers

## Overview

Now that you've built your first operator in [Module 2](../module-02/README.md), it's time to dive deeper into building sophisticated controllers. This module teaches you the advanced patterns and techniques needed to build production-ready operators that manage complex applications.

**Duration:** 6-7 hours  
**Prerequisites:** 
- Completion of [Module 1: Kubernetes Architecture Deep Dive](../module-01/README.md)
- Completion of [Module 2: Introduction to Operators](../module-02/README.md)
- Understanding of the basic operator pattern

## Learning Objectives

By the end of this module, you will:

- Understand controller-runtime architecture in depth
- Design well-structured APIs for your operators
- Implement robust reconciliation logic
- Work effectively with the Kubernetes client
- Build a database operator that manages PostgreSQL

## Module Structure

1. **[Lesson 3.1: Controller Runtime Deep Dive](lessons/01-controller-runtime.md)**
   - [Lab 3.1: Exploring Controller Runtime](labs/lab-01-controller-runtime.md)

2. **[Lesson 3.2: Designing Your API](lessons/02-designing-api.md)**
   - [Lab 3.2: API Design for Database Operator](labs/lab-02-designing-api.md)

3. **[Lesson 3.3: Implementing Reconciliation Logic](lessons/03-reconciliation-logic.md)**
   - [Lab 3.3: Building PostgreSQL Operator](labs/lab-03-reconciliation-logic.md)

4. **[Lesson 3.4: Working with Client-Go](lessons/04-client-go.md)**
   - [Lab 3.4: Advanced Client Operations](labs/lab-04-client-go.md)

## Prerequisites Check

Before starting, ensure you've completed:

- ✅ [Module 1](../module-01/README.md): Understand CRDs, controllers, and reconciliation
- ✅ [Module 2](../module-02/README.md): Built your first "Hello World" operator
- ✅ Can scaffold kubebuilder projects
- ✅ Understand the Reconcile function basics

If you haven't completed Module 2, start with [Module 2: Introduction to Operators](../module-02/README.md).

## What You'll Build

Throughout this module, you'll build a **PostgreSQL operator** that:

- Manages PostgreSQL database instances
- Handles database creation and configuration
- Manages StatefulSets and Services
- Implements proper reconciliation logic
- Uses owner references for resource management

This builds on your "Hello World" operator from Module 2, adding complexity and real-world patterns.

## Setup

Before starting this module:

1. **Verify Module 2 completion:**
   - You should have built a "Hello World" operator
   - You understand kubebuilder project structure
   - You can run operators locally

2. **Ensure development environment is ready:**
   ```bash
   ./scripts/setup-dev-environment.sh
   ```

3. **Have a kind cluster running:**
   ```bash
   ./scripts/setup-kind-cluster.sh
   ```

## Hands-on Labs

Each lesson includes hands-on exercises building toward a complete PostgreSQL operator.

- [Lab 3.1: Exploring Controller Runtime](labs/lab-01-controller-runtime.md)
- [Lab 3.2: API Design for Database Operator](labs/lab-02-designing-api.md)
- [Lab 3.3: Building PostgreSQL Operator](labs/lab-03-reconciliation-logic.md)
- [Lab 3.4: Advanced Client Operations](labs/lab-04-client-go.md)

## Solutions

Complete working solutions for all labs are available in the [solutions directory](solutions/):
- [Lab 3.3 Solutions](solutions/) - Complete Database operator (types, controller with StatefulSet/Service)


## Navigation

- [← Previous: Module 2 - Introduction to Operators](../module-02/README.md)
- [Course Overview](../README.md)
- [Next: Module 4 - Advanced Reconciliation Patterns →](../module-04/README.md)

