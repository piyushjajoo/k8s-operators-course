# Module 6: Testing and Debugging

## Overview

Now that you can build sophisticated operators with webhooks ([Module 5](../module-05/README.md)), it's time to learn how to test and debug them effectively. This module covers unit testing with envtest, integration testing, debugging techniques, and observability patterns that make operators production-ready.

**Duration:** 6-7 hours  
**Prerequisites:** 
- Completion of [Module 1: Kubernetes Architecture Deep Dive](../module-01/README.md)
- Completion of [Module 2: Introduction to Operators](../module-02/README.md)
- Completion of [Module 3: Building Custom Controllers](../module-03/README.md)
- Completion of [Module 4: Advanced Reconciliation Patterns](../module-04/README.md)
- Completion of [Module 5: Webhooks and Admission Control](../module-05/README.md)
- Understanding of Go testing fundamentals

## Learning Objectives

By the end of this module, you will:

- Write comprehensive unit tests using envtest
- Create integration test suites with Ginkgo/Gomega
- Debug operators effectively using Delve and logs
- Add observability with metrics, logging, and events
- Understand testing best practices for operators

## Module Structure

1. **[Lesson 6.1: Testing Fundamentals](lessons/01-testing-fundamentals.md)**
   - [Lab 6.1: Setting Up Testing Environment](labs/lab-01-testing-fundamentals.md)

2. **[Lesson 6.2: Unit Testing with envtest](lessons/02-unit-testing-envtest.md)**
   - [Lab 6.2: Writing Unit Tests](labs/lab-02-unit-testing-envtest.md)

3. **[Lesson 6.3: Integration Testing](lessons/03-integration-testing.md)**
   - [Lab 6.3: Creating Integration Tests](labs/lab-03-integration-testing.md)

4. **[Lesson 6.4: Debugging and Observability](lessons/04-debugging-observability.md)**
   - [Lab 6.4: Adding Observability](labs/lab-04-debugging-observability.md)

## Prerequisites Check

Before starting, ensure you've completed:

- ✅ [Module 5](../module-05/README.md): Operator with webhooks
- ✅ Understand Go testing from [Module 2](../module-02/README.md)
- ✅ Have a working operator from Module 3/4/5
- ✅ Basic understanding of Go testing (`go test`)

If you haven't completed Module 5, start with [Module 5: Webhooks and Admission Control](../module-05/README.md).

## What You'll Build

Throughout this module, you'll add testing and observability to your Database operator:

- Unit tests for reconciliation logic
- Integration tests for end-to-end scenarios
- Debugging setup for local development
- Metrics and logging for observability

## Setup

Before starting this module:

1. **Have your Database operator from Module 3/4/5:**
   - Should have a working operator
   - Webhooks implemented (from Module 5)
   - Ready to add tests

2. **Ensure development environment is ready:**
   ```bash
   ./scripts/setup-dev-environment.sh
   ```

3. **Install testing tools:**
   ```bash
   # Install Ginkgo and Gomega
   go install github.com/onsi/ginkgo/v2/ginkgo@latest
   
   # Install Delve debugger
   go install github.com/go-delve/delve/cmd/dlv@latest
   ```

## Hands-on Labs

Each lesson includes hands-on exercises that add testing and observability to your operator.

- [Lab 6.1: Setting Up Testing Environment](labs/lab-01-testing-fundamentals.md)
- [Lab 6.2: Writing Unit Tests](labs/lab-02-unit-testing-envtest.md)
- [Lab 6.3: Creating Integration Tests](labs/lab-03-integration-testing.md)
- [Lab 6.4: Adding Observability](labs/lab-04-debugging-observability.md)

## Solutions

Complete working solutions for all labs are available in the [solutions directory](solutions/):
- [Lab 6.1 Solutions](solutions/suite_test.go) - Test suite setup
- [Lab 6.2 Solutions](solutions/database_controller_test.go) - Unit test examples
- [Lab 6.3 Solutions](solutions/integration_test.go) - Integration test examples
- [Lab 6.4 Solutions](solutions/metrics.go, solutions/observability.go) - Observability examples

## Additional Resources

- [Module Summary](SUMMARY.md)
- [Testing Guide](TESTING.md)
- [Course Build Plan](../COURSE_BUILD_PLAN.md)

## Navigation

- [← Previous: Module 5 - Webhooks and Admission Control](../module-05/README.md)
- [Course Overview](../README.md)
- [Next: Module 7 - Production Deployment →](../module-07/README.md) (Coming Soon)

