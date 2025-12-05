# Lab 2.2: Kubebuilder CLI and Project Structure

**Related Lesson:** [Lesson 2.2: Kubebuilder Fundamentals](../lessons/02-kubebuilder-fundamentals.md)  
**Navigation:** [← Previous Lab: Operator Pattern](lab-01-operator-pattern.md) | [Module Overview](../README.md) | [Next Lab: Dev Environment →](lab-03-dev-environment.md)

## Objectives

- Install and verify kubebuilder
- Understand kubebuilder CLI commands
- Explore kubebuilder project structure
- Understand code generation

## Prerequisites

- Go 1.21+ installed
- Understanding of operators from [Lesson 2.1](../lessons/01-operator-pattern.md)

## Exercise 1: Install Kubebuilder

### Task 1.1: Check if Kubebuilder is Installed

```bash
# Check kubebuilder version
kubebuilder version
```

If not installed, proceed to installation.

### Task 1.2: Install Kubebuilder

```bash
# Download kubebuilder
curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
chmod +x kubebuilder
sudo mv kubebuilder /usr/local/bin/

# Verify installation
kubebuilder version
```

### Task 1.3: Verify Installation

```bash
# Check kubebuilder is in PATH
which kubebuilder

# Check version
kubebuilder version

# List available commands
kubebuilder --help
```

## Exercise 2: Explore Kubebuilder Commands

### Task 2.1: Initialize a Test Project

```bash
# Create a test directory
mkdir -p /tmp/kubebuilder-test
cd /tmp/kubebuilder-test

# Initialize project
kubebuilder init --domain example.com --repo github.com/example/test-operator
```

**Observe:**
- What files were created?
- What directories were created?
- What's in the Makefile?

### Task 2.2: Examine Project Structure

```bash
# List all files
find . -type f | head -20

# Examine main.go
cat ./cmd/main.go

# Examine Makefile
cat Makefile | head -30

# Check go.mod
cat go.mod
```

**Questions:**
1. What's the purpose of `main.go`?
2. What Makefile targets are available?
3. What dependencies are in `go.mod`?

### Task 2.3: Create an API

```bash
# Create an API
kubebuilder create api --group test --version v1 --kind TestResource
```

When prompted:
- Create Resource [y/n]: **y**
- Create Controller [y/n]: **y**

**Observe:**
- What new files were created?
- What directories were added?
- What's in the API directory?

## Exercise 3: Understand Generated Code

### Task 3.1: Examine API Types

```bash
# Look at generated types
cat api/v1/testresource_types.go
```

**Key Observations:**
- Spec structure
- Status structure
- Kubebuilder markers (comments starting with `//+kubebuilder:`)

### Task 3.2: Examine Controller

```bash
# Look at generated controller
cat controllers/testresource_controller.go
```

**Key Observations:**
- Reconcile function skeleton
- RBAC markers
- SetupWithManager function

### Task 3.3: Generate Code

```bash
# Generate code
make generate

# Generate manifests
make manifests
```

**Observe:**
- What files were generated?
- Check `config/crd/bases/` directory
- Check `config/rbac/` directory

## Exercise 4: Explore Generated Manifests

### Task 4.1: Examine CRD

```bash
# List generated CRDs
ls -la config/crd/bases/

# Examine CRD YAML
cat config/crd/bases/test.example.com_testresources.yaml | head -50
```

**Questions:**
1. What API group is used?
2. What's the resource name?
3. What validation is included?

### Task 4.2: Examine RBAC

```bash
# List RBAC files
ls -la config/rbac/

# Examine role
cat config/rbac/role.yaml
```

**Questions:**
1. What permissions are granted?
2. How are permissions determined?

## Exercise 5: Understand Code Generation Flow

### Task 5.1: Modify Types

Edit `api/v1/testresource_types.go` and add a field:

```go
// TestResourceSpec defines the desired state of TestResource
type TestResourceSpec struct {
    // Message is a test message
    Message string `json:"message,omitempty"`
}
```

### Task 5.2: Regenerate

```bash
# Regenerate code
make generate
make manifests

# Check CRD was updated
cat config/crd/bases/test.example.com_testresources.yaml | grep -A 5 message
```

**Observation:** The CRD schema was updated automatically!

## Exercise 6: Explore Makefile Targets

### Task 6.1: List Available Targets

```bash
# List all Makefile targets
make help
```

### Task 6.2: Understand Key Targets

**Important targets:**
- `make generate` - Generates code
- `make manifests` - Generates manifests
- `make install` - Installs CRDs
- `make run` - Runs operator locally
- `make docker-build` - Builds container image

### Task 6.3: Try Some Targets

```bash
# Generate everything
make generate manifests

# Check what was created
ls -la config/crd/bases/
ls -la config/rbac/
```

## Exercise 7: Project Structure Deep Dive

### Task 7.1: Map the Structure

Create a mental map of the project:

```
project-root/
├── api/              # API type definitions
│   └── v1/           # API version
├── controllers/      # Controller implementations
├── config/           # Generated manifests
│   ├── crd/          # CRD definitions
│   ├── rbac/         # RBAC rules
│   └── manager/      # Manager deployment
├── main.go           # Entry point
├── Makefile          # Build targets
└── go.mod            # Go dependencies
```

### Task 7.2: Understand Each Component

For each directory, understand its purpose:

- **api/**: Your Custom Resource type definitions
- **controllers/**: Your reconciliation logic
- **config/crd/**: Generated CRD YAML files
- **config/rbac/**: Generated RBAC manifests
- **main.go**: Sets up and starts the manager

## Cleanup

```bash
# Remove test project
cd ~
rm -rf /tmp/kubebuilder-test
```

## Lab Summary

In this lab, you:
- Installed and verified kubebuilder
- Explored kubebuilder CLI commands
- Created a test project
- Examined generated code structure
- Understood code generation flow
- Explored project structure

## Key Learnings

1. Kubebuilder provides scaffolding for operator projects
2. `kubebuilder init` creates project structure
3. `kubebuilder create api` generates API types and controller
4. `make generate` and `make manifests` generate code and YAML
5. Project structure is standardized and organized
6. Kubebuilder markers control code generation

## Next Steps

Now that you understand kubebuilder, let's set up your complete development environment!

**Navigation:** [← Previous Lab: Operator Pattern](lab-01-operator-pattern.md) | [Related Lesson](../lessons/02-kubebuilder-fundamentals.md) | [Next Lab: Dev Environment →](lab-03-dev-environment.md)

