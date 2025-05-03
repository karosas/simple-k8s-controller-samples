/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=namespaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=namespaces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// If namespace doesn't have dev prefix or suffix - end reconciliation
	if !strings.HasPrefix(req.Name, "dev-") && !strings.HasSuffix(req.Name, "-dev") {
		return ctrl.Result{}, nil
	}

	// fetch resource from cache
	var resource corev1.Namespace
	if err := r.Get(ctx, req.NamespacedName, &resource); err != nil {
		if errors.IsNotFound(err) {
			// namespace not found, most likely deleted
			return ctrl.Result{}, nil
		}
		logger.Error(err, fmt.Sprintf("Couldn't fetch namespace %s", req.Name))
		return ctrl.Result{}, err
	}

	// If resource already contains correct label - end reconciliation
	if resource.Labels["env"] == "dev" {
		return ctrl.Result{}, nil
	}

	resource.Labels["env"] = "dev"
	if err := r.Update(ctx, &resource); err != nil {
		logger.Error(err, fmt.Sprintf("Couldn't update namespace %s", req.Name))
		return ctrl.Result{}, err
	}

	logger.Info(fmt.Sprintf("Successfully labelled namespace %s", req.Name))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Named("namespace").
		Complete(r)
}
