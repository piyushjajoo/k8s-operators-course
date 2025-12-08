# Module 4 Solutions

This directory contains complete, working solutions for Module 4 labs.

## Files

- **conditions-helpers.go**: Helper functions for managing conditions
- **finalizer-handler.go**: Complete finalizer implementation
- **watch-setup.go**: Watch setup examples
- **state-machine-controller.go**: Complete multi-phase reconciliation with state machine

## Usage

These solutions can be used as:
- Reference when adding conditions to your operator
- Starting point for finalizer implementation
- Examples of watch patterns
- Template for implementing state machine reconciliation

## Integration

To use these solutions:

1. Add condition helpers to your controller
2. Integrate finalizer handler into Reconcile function
3. Update SetupWithManager with watch configuration
4. Update your Database status type to include Conditions
5. **For state machine**: Replace your main Reconcile function to call `reconcileWithStateMachine`

## State Machine (Lab 4)

The state machine implementation provides multi-phase reconciliation with the following state flow:

```
Pending → Provisioning → Configuring → Deploying → Verifying → Ready
                                                              ↓
                                                           Failed (on error)
```

### Prerequisites for State Machine

1. **Update API Types** - Edit `api/v1/database_types.go` and update the Phase field enum:
   ```go
   // +kubebuilder:validation:Enum=Pending;Provisioning;Configuring;Deploying;Verifying;Ready;Failed
   Phase string `json:"phase,omitempty"`
   ```

2. **Regenerate and reinstall CRD**:
   ```bash
   make manifests
   make install
   ```

3. **Update Reconcile function** - The main `Reconcile` function must call `reconcileWithStateMachine(ctx, db)` 
   instead of directly calling resource reconciliation functions.

> **Note:** If you skip step 1-2, you'll see validation errors like:
> `phase: Unsupported value: "Provisioning": supported values: "Pending", "Creating", "Ready", "Failed"`
>
> If you skip step 3, you'll only see `Pending → Creating → Ready` transitions.

## Notes

- These are complete, working examples
- Conditions follow Kubernetes standards
- Finalizers handle cleanup gracefully
- Watches are properly configured
- State machine handles all phases including error recovery
- Ready for Module 5 webhooks

