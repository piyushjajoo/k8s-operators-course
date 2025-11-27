# Lab 1.4: Creating Your First CRD

**Related Lesson:** [Lesson 1.4: Custom Resources](../lessons/04-custom-resources.md)  
**Navigation:** [← Previous Lab: Controller Pattern](lab-03-controller-pattern.md) | [Module Overview](../README.md)

## Objectives

- Create a Custom Resource Definition (CRD)
- Create and manage Custom Resources
- Understand CRD validation
- Work with status subresources
- Understand when to use CRDs

## Prerequisites

- Kind cluster running
- kubectl configured

## Exercise 1: Create a Simple CRD

### Task 1.1: Define a Website CRD

```bash
# Create a CRD for managing websites
cat <<EOF | kubectl apply -f -
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: websites.example.com
spec:
  group: example.com
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              url:
                type: string
                pattern: '^https?://'
                description: The website URL
              replicas:
                type: integer
                minimum: 1
                maximum: 10
                description: Number of replicas
              environment:
                type: string
                enum: [development, staging, production]
                default: development
            required:
            - url
            - replicas
          status:
            type: object
            properties:
              phase:
                type: string
                enum: [Pending, Running, Failed]
              readyReplicas:
                type: integer
              lastUpdated:
                type: string
                format: date-time
  scope: Namespaced
  names:
    plural: websites
    singular: website
    kind: Website
    shortNames:
    - ws
EOF

# Verify CRD was created
kubectl get crd websites.example.com

# Check API discovery
kubectl api-resources | grep websites
```

### Task 1.2: Verify API Endpoint

```bash
# Check the API endpoint is available
kubectl get --raw /apis/example.com/v1

# Get the CRD definition
kubectl get crd websites.example.com -o yaml | head -50
```

## Exercise 2: Create Custom Resources

### Task 2.1: Create a Valid Website Resource

```bash
# Create a website resource
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: my-blog
spec:
  url: https://example.com/blog
  replicas: 3
  environment: production
EOF

# Verify it was created
kubectl get websites
kubectl get website my-blog
kubectl get ws my-blog  # Using short name

# View the full resource
kubectl get website my-blog -o yaml
```

### Task 2.2: Create Multiple Websites

```bash
# Create more websites
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: my-shop
spec:
  url: https://example.com/shop
  replicas: 5
  environment: production
---
apiVersion: example.com/v1
kind: Website
metadata:
  name: dev-site
spec:
  url: http://dev.example.com
  replicas: 1
  environment: development
EOF

# List all websites
kubectl get websites

# Get specific website
kubectl get website my-shop -o yaml
```

## Exercise 3: Test Validation

### Task 3.1: Test Required Fields

```bash
# Try to create website without required field
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: invalid-website
spec:
  url: https://example.com
  # Missing replicas field
EOF
```

**Expected Result:** Validation error about missing required field.

### Task 3.2: Test URL Pattern Validation

```bash
# Try invalid URL (doesn't match pattern)
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: invalid-url
spec:
  url: not-a-valid-url
  replicas: 2
EOF
```

**Expected Result:** Validation error about URL pattern.

### Task 3.3: Test Replica Range Validation

```bash
# Try replicas below minimum
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: too-few-replicas
spec:
  url: https://example.com
  replicas: 0
EOF
```

**Expected Result:** Validation error about minimum value.

```bash
# Try replicas above maximum
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: too-many-replicas
spec:
  url: https://example.com
  replicas: 20
EOF
```

**Expected Result:** Validation error about maximum value.

### Task 3.4: Test Enum Validation

```bash
# Try invalid environment value
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: invalid-env
spec:
  url: https://example.com
  replicas: 2
  environment: invalid-env
EOF
```

**Expected Result:** Validation error about enum value.

### Task 3.5: Test Default Values

```bash
# Create website without environment (should use default)
cat <<EOF | kubectl apply -f -
apiVersion: example.com/v1
kind: Website
metadata:
  name: default-env
spec:
  url: https://example.com
  replicas: 2
EOF

# Check the default was applied
kubectl get website default-env -o jsonpath='{.spec.environment}'
echo
```

**Expected Result:** Environment should be "development" (the default).

## Exercise 4: Update Custom Resources

### Task 4.1: Update Spec

```bash
# Update the website
kubectl patch website my-blog --type merge -p '{"spec":{"replicas":5}}'

# Verify the update
kubectl get website my-blog -o jsonpath='{.spec.replicas}'
echo

# Update URL
kubectl patch website my-blog --type merge -p '{"spec":{"url":"https://newurl.com"}}'

# Verify
kubectl get website my-blog -o jsonpath='{.spec.url}'
echo
```

### Task 4.2: Update via YAML

```bash
# Get current resource
kubectl get website my-shop -o yaml > /tmp/website.yaml

# Edit the file (or use sed)
sed -i '' 's/replicas: 5/replicas: 7/' /tmp/website.yaml

# Apply the update
kubectl apply -f /tmp/website.yaml

# Verify
kubectl get website my-shop -o jsonpath='{.spec.replicas}'
echo
```

## Exercise 5: Status Subresource

### Task 5.1: Examine Status Field

```bash
# Get website with status
kubectl get website my-blog -o yaml | grep -A 10 status

# The status field exists but is empty (no controller to update it)
# In a real operator, the controller would update this
```

### Task 5.2: Understand Spec vs Status

```bash
# Compare spec and status
echo "=== SPEC (Desired State) ==="
kubectl get website my-blog -o jsonpath='{.spec}' | jq '.'

echo -e "\n=== STATUS (Actual State) ==="
kubectl get website my-blog -o jsonpath='{.status}' | jq '.'

# Spec is what the user wants
# Status is what actually exists (updated by controller)
```

## Exercise 6: CRD vs ConfigMap Comparison

### Task 6.1: Create Equivalent ConfigMap

```bash
# Create a ConfigMap with similar data
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: website-config
data:
  url: "https://example.com"
  replicas: "3"
  environment: "production"
EOF

# Compare the two approaches
echo "=== Custom Resource ==="
kubectl get website my-blog -o yaml | head -20

echo -e "\n=== ConfigMap ==="
kubectl get configmap website-config -o yaml
```

**Key Differences:**
1. CRD has structured schema and validation
2. ConfigMap is just key-value pairs
3. CRD has API semantics (can watch, has resourceVersion)
4. CRD can have status subresource

### Task 6.2: Try Invalid ConfigMap Data

```bash
# ConfigMap accepts any data (no validation)
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: invalid-config
data:
  url: "not-a-url"
  replicas: "not-a-number"
  environment: "invalid-env"
EOF

# This works! No validation
kubectl get configmap invalid-config

# But our CRD would reject this
```

## Exercise 7: Explore CRD Details

### Task 7.1: Examine CRD Schema

```bash
# Get the full CRD definition
kubectl get crd websites.example.com -o yaml > /tmp/crd.yaml

# View the schema
kubectl get crd websites.example.com -o jsonpath='{.spec.versions[0].schema}' | jq '.'

# View validation rules
kubectl get crd websites.example.com -o jsonpath='{.spec.versions[0].schema.openAPIV3Schema.properties.spec.properties}' | jq '.'
```

### Task 7.2: API Discovery

```bash
# Discover the API
kubectl get --raw /apis/example.com/v1 | jq '.'

# See available resources
kubectl get --raw /apis/example.com/v1 | jq '.resources[].name'

# Get a specific website via API
kubectl get --raw /apis/example.com/v1/namespaces/default/websites/my-blog | jq '.'
```

## Exercise 8: Delete and Cleanup

### Task 8.1: Delete Custom Resources

```bash
# Delete individual websites
kubectl delete website my-blog
kubectl delete website my-shop

# Delete all websites
kubectl delete websites --all

# Verify they're gone
kubectl get websites
```

### Task 8.2: Delete CRD

```bash
# Delete the CRD
kubectl delete crd websites.example.com

# Verify it's gone
kubectl get crd websites.example.com

# Try to create a website (should fail)
kubectl create website test --url=https://test.com --replicas=2
```

**Note:** Deleting a CRD also deletes all Custom Resources of that type!

## Cleanup

```bash
# Clean up any remaining resources
kubectl delete websites --all 2>/dev/null
kubectl delete crd websites.example.com 2>/dev/null
kubectl delete configmap website-config invalid-config 2>/dev/null
rm -f /tmp/website.yaml /tmp/crd.yaml
```

## Lab Summary

In this lab, you:
- Created a Custom Resource Definition (CRD)
- Created and managed Custom Resources
- Tested CRD validation rules
- Compared CRDs with ConfigMaps
- Understood spec vs status separation
- Explored CRD schema and API discovery

## Key Learnings

1. CRDs extend Kubernetes with domain-specific resources
2. CRDs provide schema validation (unlike ConfigMaps)
3. CRDs have API semantics (watch, resourceVersion, etc.)
4. Spec describes desired state, status describes actual state
5. Validation happens at the API level before storage
6. CRDs are the foundation for building operators

## When to Use CRDs

**Use CRDs when:**
- You need structured data with validation
- You want API semantics (watch, etc.)
- You're building an operator
- You need status subresource

**Use ConfigMaps when:**
- Simple key-value configuration
- No validation needed
- No API semantics required

**Navigation:** [← Previous Lab: Controller Pattern](lab-03-controller-pattern.md) | [Related Lesson](../lessons/04-custom-resources.md) | [Module Overview](../README.md)

