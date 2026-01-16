---
layout: default
title: "Module 5: Webhooks & Admission Control"
nav_order: 5
parent: Modules
has_children: true
has_toc: false
permalink: /module-05/
mermaid: true
---

# Module 5: Webhooks and Admission Control

## Overview

Now that you can build sophisticated operators ([Module 4](../module-04/README.md)), it's time to add webhooks for validation and mutation. Webhooks allow you to validate and modify resources before they're stored in etcd, providing powerful control over your Custom Resources.

**Duration:** 6-7 hours  
**Prerequisites:** 
- Completion of [Module 1: Kubernetes Architecture Deep Dive](../module-01/README.md)
- Completion of [Module 2: Introduction to Operators](../module-02/README.md)
- Completion of [Module 3: Building Custom Controllers](../module-03/README.md)
- Completion of [Module 4: Advanced Reconciliation Patterns](../module-04/README.md)
- Understanding of API design and validation

## Learning Objectives

By the end of this module, you will:

- Understand Kubernetes admission control and webhooks
- Implement validating webhooks for custom validation
- Implement mutating webhooks for defaulting and mutation
- Manage webhook certificates and deployment
- Test webhooks locally and in production

## Module Structure

1. **[Lesson 5.1: Kubernetes Admission Control](lessons/01-admission-control.md)**
   - [Lab 5.1: Exploring Admission Control](labs/lab-01-admission-control.md)

2. **[Lesson 5.2: Implementing Validating Webhooks](lessons/02-validating-webhooks.md)**
   - [Lab 5.2: Building Validating Webhook](labs/lab-02-validating-webhooks.md)

3. **[Lesson 5.3: Implementing Mutating Webhooks](lessons/03-mutating-webhooks.md)**
   - [Lab 5.3: Building Mutating Webhook](labs/lab-03-mutating-webhooks.md)

4. **[Lesson 5.4: Webhook Deployment and Certificates](lessons/04-webhook-deployment.md)**
   - [Lab 5.4: Certificate Management](labs/lab-04-webhook-deployment.md)

## Prerequisites Check

Before starting, ensure you've completed:

- ✅ [Module 4](../module-04/README.md): Enhanced operator with conditions and finalizers
- ✅ Understand API design from [Lesson 3.2](../module-03/lessons/02-designing-api.md)
- ✅ Have a working operator from Module 3/4
- ✅ Understand CRD validation from [Lesson 1.4](../module-01/lessons/04-custom-resources.md)

If you haven't completed Module 4, start with [Module 4: Advanced Reconciliation Patterns](../module-04/README.md).

## What You'll Build

Throughout this module, you'll add webhooks to your Database operator:

- Validating webhook for custom validation rules
- Mutating webhook for defaulting values
- Certificate management for webhook security
- Local testing setup for webhook development

## Setup

Before starting this module:

1. **Have your Database operator from Module 3/4:**
   - Should have a working operator
   - API should be well-defined
   - Basic validation in CRD schema

2. **Ensure development environment is ready:**
   ```bash
   ./scripts/setup-dev-environment.sh
   ```

3. **Have a kind cluster running:**
   ```bash
   ./scripts/setup-kind-cluster.sh
   ```

## Hands-on Labs

Each lesson includes hands-on exercises that add webhooks to your operator.

- [Lab 5.1: Exploring Admission Control](labs/lab-01-admission-control.md)
- [Lab 5.2: Building Validating Webhook](labs/lab-02-validating-webhooks.md)
- [Lab 5.3: Building Mutating Webhook](labs/lab-03-mutating-webhooks.md)
- [Lab 5.4: Certificate Management](labs/lab-04-webhook-deployment.md)

## Solutions

Complete working solutions for all labs are available in the [solutions directory](solutions/):
- [Lab 5.2 Solutions](solutions/validating-webhook.go) - Complete validating webhook
- [Lab 5.3 Solutions](solutions/mutating-webhook.go) - Complete mutating webhook


## Navigation

- [← Previous: Module 4 - Advanced Reconciliation Patterns](../module-04/README.md)
- [Course Overview](../README.md)
- [Next: Module 6 - Testing and Debugging →](../module-06/README.md)
