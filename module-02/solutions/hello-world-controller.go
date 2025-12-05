// Solution: Complete HelloWorld Controller from Module 2
// This implements the reconciliation logic for the Hello World operator

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hellov1 "github.com/example/hello-world-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HelloWorldReconciler reconciles a HelloWorld object
type HelloWorldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hello.example.com,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hello.example.com,resources=helloworlds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hello.example.com,resources=helloworlds/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop
func (r *HelloWorldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the HelloWorld instance
	helloWorld := &hellov1.HelloWorld{}
	if err := r.Get(ctx, req.NamespacedName, helloWorld); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return
			return ctrl.Result{}, nil
		}
		// Error reading the object
		return ctrl.Result{}, err
	}

	// Define the ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helloWorld.Name + "-config",
			Namespace: helloWorld.Namespace,
		},
		Data: map[string]string{
			"message": helloWorld.Spec.Message,
			"count":   fmt.Sprintf("%d", helloWorld.Spec.Count),
		},
	}

	// Set owner reference
	if err := ctrl.SetControllerReference(helloWorld, configMap, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// Check if ConfigMap already exists
	existingConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{
		Name:      configMap.Name,
		Namespace: configMap.Namespace,
	}, existingConfigMap)

	if err != nil && errors.IsNotFound(err) {
		// ConfigMap doesn't exist, create it
		logger.Info("Creating ConfigMap", "name", configMap.Name)
		if err := r.Create(ctx, configMap); err != nil {
			return ctrl.Result{}, err
		}
	} else if err != nil {
		return ctrl.Result{}, err
	} else {
		// ConfigMap exists, update it if needed
		if existingConfigMap.Data["message"] != configMap.Data["message"] ||
			existingConfigMap.Data["count"] != configMap.Data["count"] {
			logger.Info("Updating ConfigMap", "name", configMap.Name)
			existingConfigMap.Data = configMap.Data
			if err := r.Update(ctx, existingConfigMap); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// Update status
	helloWorld.Status.Phase = "Ready"
	helloWorld.Status.ConfigMapCreated = true
	if err := r.Status().Update(ctx, helloWorld); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloWorldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hellov1.HelloWorld{}).
		Complete(r)
}
