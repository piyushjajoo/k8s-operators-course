# Module 1 Solutions

This directory contains complete, working solutions for Module 1 labs.

## Files

- [**website-crd.yaml**](https://github.com/piyushjajoo/k8s-operators-course/blob/main/module-01/solutions/website-crd.yaml): Complete Website CRD definition
- [**example-website.yaml**](https://github.com/piyushjajoo/k8s-operators-course/blob/main/module-01/solutions/example-website.yaml): Example Website Custom Resource

## Usage

These solutions can be used as:
- Reference when creating your own CRDs
- Starting point if you get stuck
- Examples of CRD best practices

## Integration

To use these solutions:

1. Apply the CRD: `kubectl apply -f website-crd.yaml`
2. Wait for CRD to be established: `kubectl wait --for condition=established crd/websites.example.com`
3. Create a Website: `kubectl apply -f example-website.yaml`
4. Verify: `kubectl get websites`

## Notes

- These are complete, working examples
- They follow Kubernetes best practices
- Validation rules are included
- Status subresource is properly configured

