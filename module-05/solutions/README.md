# Module 5 Solutions

This directory contains complete, working solutions for Module 5 labs.

## Files

- **validating-webhook.go**: Complete validating webhook implementation
- **mutating-webhook.go**: Complete mutating webhook implementation

## Usage

These solutions can be used as:
- Reference when implementing your own webhooks
- Starting point if you get stuck
- Examples of best practices

## Integration

To use these solutions in your operator:

1. Copy the webhook code to `api/v1/database_webhook.go`
2. Ensure your API types match the structure
3. Run `make generate` and `make manifests`
4. Set up certificates: `make certs && make install-cert`
5. Run your operator: `make run`

## Notes

- These are complete, working examples
- They follow best practices from the lessons
- Error messages are clear and actionable
- Mutations are idempotent
- Validation covers common scenarios

