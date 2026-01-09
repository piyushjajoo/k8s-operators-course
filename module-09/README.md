---
layout: default
title: "Module 9: API Evolution and Versioning"
nav_order: 9
parent: Modules
has_children: true
has_toc: false
permalink: /module-09/
mermaid: true
---

# Module 9: API Evolution and Versioning

## Overview

You now have a production-ready operator. This advanced module focuses on evolving your API safely with conversion webhooks, so you can introduce new versions without breaking existing users.

**Duration:** 3-4 hours  
**Prerequisites:**
- Completion of [Module 8](../module-08/README.md)
- Operator with admission webhooks deployed (from Module 5)
- Understanding of CRD versioning concepts

## Learning Objectives

By the end of this module, you will:

- Create a v2 API version for your Database resource
- Implement conversion functions between v1 and v2
- Configure conversion webhooks in the CRD
- Validate bidirectional conversion and round-trip integrity

## Module Structure

1. **[Lesson 9.1: Conversion Webhooks and API Versioning](lessons/01-conversion-webhooks.md)**
   - [Lab 9.1: Conversion Webhooks](labs/lab-01-conversion-webhooks.md)

## Prerequisites Check

Before starting, ensure you've completed:

- ✅ [Module 8](../module-08/README.md): Advanced topics and real-world patterns
- ✅ Webhook deployment from [Module 5](../module-05/README.md)
- ✅ A working operator you can deploy and test

If you haven't completed Module 8, start with [Module 8: Advanced Topics](../module-08/README.md).

## What You'll Build

Throughout this module, you'll evolve your Database API safely by:

- Adding a v2 API with new fields
- Converting resources between v1 and v2
- Validating conversion behavior with end-to-end tests

## Setup

Before starting this module:

1. **Have your operator ready from Module 8:**
   - Deployed to a cluster
   - Admission webhooks working

2. **Ensure development environment is ready:**
   ```bash
   ./scripts/setup-dev-environment.sh
   ```

3. **Have a kind cluster running:**
   ```bash
   ./scripts/setup-kind-cluster.sh
   ```

## Solutions

Complete working solutions are available in the [solutions directory](solutions/):
- [Lab 9.1 Solutions](solutions/conversion-webhook.go) - Conversion webhook implementation

## Navigation

- [← Previous: Module 8 - Advanced Topics](../module-08/README.md)
- [Course Overview](../README.md)
