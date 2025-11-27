# Lesson 6.1: Testing Fundamentals

**Navigation:** [Module Overview](../README.md) | [Next Lesson: Unit Testing with envtest →](02-unit-testing-envtest.md)

## Introduction

Testing operators is crucial for reliability and confidence in production. Operators manage critical infrastructure, so comprehensive testing is essential. This lesson covers testing fundamentals, testing strategies, and the tools you'll use to test Kubernetes operators.

## Why Test Operators?

Operators manage critical resources:

```mermaid
graph TB
    OPERATOR[Operator] --> CRITICAL[Critical Resources]
    
    CRITICAL --> DATABASES[Databases]
    CRITICAL --> STORAGE[Storage]
    CRITICAL --> NETWORKING[Networking]
    CRITICAL --> SECURITY[Security]
    
    TESTING[Testing] --> CONFIDENCE[Confidence]
    TESTING --> RELIABILITY[Reliability]
    TESTING --> SAFETY[Safety]
    
    style OPERATOR fill:#FFB6C1
    style TESTING fill:#90EE90
```

**Benefits:**
- Catch bugs before production
- Ensure correctness of reconciliation
- Validate edge cases
- Enable safe refactoring
- Document expected behavior

## Testing Pyramid for Operators

```mermaid
graph TB
    PYRAMID[Testing Pyramid]
    
    PYRAMID --> UNIT[Unit Tests<br/>Many, Fast]
    PYRAMID --> INTEGRATION[Integration Tests<br/>Some, Slower]
    PYRAMID --> E2E[End-to-End Tests<br/>Few, Slowest]
    
    UNIT --> LOGIC[Test Logic]
    UNIT --> FUNCTIONS[Test Functions]
    
    INTEGRATION --> CLUSTER[Test with Cluster]
    INTEGRATION --> RESOURCES[Test Resources]
    
    E2E --> SCENARIOS[Test Scenarios]
    E2E --> WORKFLOWS[Test Workflows]
    
    style UNIT fill:#90EE90
    style INTEGRATION fill:#FFE4B5
    style E2E fill:#FFB6C1
```

## Testing Strategies

### Strategy 1: Unit Testing

Test individual functions and logic:

```mermaid
flowchart LR
    UNIT[Unit Test] --> FUNCTION[Function]
    FUNCTION --> MOCK[Mock Dependencies]
    MOCK --> ASSERT[Assert Results]
    
    style UNIT fill:#90EE90
```

**Use for:**
- Reconciliation logic
- Helper functions
- Validation logic
- Transformation functions

### Strategy 2: Integration Testing

Test with real Kubernetes API:

```mermaid
flowchart LR
    INTEGRATION[Integration Test] --> CLUSTER[Test Cluster]
    CLUSTER --> API[Kubernetes API]
    API --> RESOURCES[Create Resources]
    RESOURCES --> VERIFY[Verify State]
    
    style INTEGRATION fill:#FFE4B5
```

**Use for:**
- End-to-end workflows
- Resource creation/updates
- Webhook behavior
- Controller interactions

### Strategy 3: End-to-End Testing

Test complete scenarios:

```mermaid
flowchart LR
    E2E[E2E Test] --> SCENARIO[Scenario]
    SCENARIO --> OPERATOR[Run Operator]
    OPERATOR --> CLUSTER[Real Cluster]
    CLUSTER --> VERIFY[Verify Results]
    
    style E2E fill:#FFB6C1
```

**Use for:**
- Complete user workflows
- Production-like scenarios
- Performance testing
- Regression testing

## Testing Tools

### envtest

**Purpose:** Lightweight Kubernetes API server for unit testing

```mermaid
graph TB
    ENVTEST[envtest]
    
    ENVTEST --> API[Kubernetes API Server]
    ENVTEST --> ETCD[etcd]
    
    API --> TEST[Your Tests]
    ETCD --> TEST
    
    style ENVTEST fill:#90EE90
```

**Features:**
- No full cluster needed
- Fast test execution
- Isolated test environment
- Real Kubernetes API

### Ginkgo and Gomega

**Purpose:** BDD-style testing framework

```mermaid
graph TB
    GINKGO[Ginkgo/Gomega]
    
    GINKGO --> BDD[BDD Style]
    GINKGO --> MATCHERS[Rich Matchers]
    GINKGO --> STRUCTURE[Test Structure]
    
    style GINKGO fill:#90EE90
```

**Features:**
- Descriptive test structure
- Rich assertion library
- Parallel test execution
- Test organization

### Delve Debugger

**Purpose:** Go debugger for operators

```mermaid
graph TB
    DELVE[Delve]
    
    DELVE --> BREAKPOINTS[Breakpoints]
    DELVE --> INSPECT[Inspect Variables]
    DELVE --> STEP[Step Through Code]
    
    style DELVE fill:#FFB6C1
```

**Features:**
- Set breakpoints
- Inspect variables
- Step through code
- Debug running operators

## Test Structure

### Basic Test Structure

```go
func TestReconcile(t *testing.T) {
    // Arrange: Set up test environment
    // Act: Execute the function
    // Assert: Verify results
}
```

### Table-Driven Tests

```go
func TestReconcile(t *testing.T) {
    tests := []struct {
        name    string
        input   *Database
        want    ctrl.Result
        wantErr bool
    }{
        {
            name: "successful reconciliation",
            input: &Database{...},
            want: ctrl.Result{},
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

## Test Coverage Goals

```mermaid
graph LR
    COVERAGE[Test Coverage]
    
    COVERAGE --> HIGH[High Coverage]
    COVERAGE --> CRITICAL[Critical Paths]
    COVERAGE --> EDGE[Edge Cases]
    
    HIGH --> 80[80%+]
    CRITICAL --> 100[100%]
    EDGE --> ALL[All Cases]
    
    style HIGH fill:#90EE90
    style CRITICAL fill:#FFB6C1
```

**Targets:**
- Overall: 80%+ coverage
- Critical paths: 100% coverage
- Edge cases: All covered
- Error paths: All tested

## Key Takeaways

- **Testing is essential** for operator reliability
- **Unit tests** are fast and test logic
- **Integration tests** test with real Kubernetes API
- **E2E tests** test complete scenarios
- **envtest** provides lightweight Kubernetes API
- **Ginkgo/Gomega** provide BDD-style testing
- **Delve** enables debugging operators
- **Table-driven tests** organize test cases
- **Aim for 80%+ coverage** with 100% on critical paths

## Understanding for Building Operators

When testing operators:
- Write unit tests for all logic
- Use integration tests for workflows
- Test error cases and edge cases
- Use table-driven tests for multiple scenarios
- Aim for high coverage
- Test in isolation when possible
- Use real Kubernetes API for integration tests

## Related Lab

- [Lab 6.1: Setting Up Testing Environment](../labs/lab-01-testing-fundamentals.md) - Hands-on exercises for this lesson

## Next Steps

Now that you understand testing fundamentals, let's set up envtest and write unit tests.

**Navigation:** [← Module Overview](../README.md) | [Next: Unit Testing with envtest →](02-unit-testing-envtest.md)

