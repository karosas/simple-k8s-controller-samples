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

package v1

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// nolint:unused
// log is for logging in this package.
var podlog = logf.Log.WithName("pod-resource")

// SetupPodWebhookWithManager registers the webhook for Pod in the manager.
func SetupPodWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&corev1.Pod{}).
		WithValidator(&PodCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate--v1-pod,mutating=false,failurePolicy=fail,sideEffects=None,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod-v1.kb.io,admissionReviewVersions=v1

type PodCustomValidator struct {
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &PodCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("expected a Pod object but got %T", obj)
	}

	podlog.Info("Validation for Pod upon creation", "name", pod.GetName())

	return nil, validatePod(pod)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Pod.
func (v *PodCustomValidator) ValidateUpdate(_ context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	pod, ok := newObj.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("expected a Pod object for the newObj but got %T", newObj)
	}
	podlog.Info("Validation for Pod upon update", "name", pod.GetName())

	return nil, validatePod(pod)
}

func validatePod(pod *corev1.Pod) error {
	if b, err := json.Marshal(pod); err == nil {
		podlog.Info(string(b))
	}

	var allErrors field.ErrorList

	for i, container := range pod.Spec.Containers {
		if cpu := container.Resources.Requests[corev1.ResourceCPU]; cpu.IsZero() {
			podlog.Info(".spec.resources.requests.cpu is missing")
			allErrors = append(allErrors, field.Invalid(field.NewPath("spec").Child("containers").Index(i).Child("resources", "requests", "cpu"), cpu.String(), "cpu request must be specified"))
		}

		if memory := container.Resources.Requests[corev1.ResourceMemory]; memory.IsZero() {
			podlog.Info(".spec.resources.requests.memory is missing")
			allErrors = append(allErrors, field.Invalid(field.NewPath("spec").Child("containers").Index(i).Child("resources", "requests", "memory"), memory.String(), "memory request must be specified"))
		}

		if cpu := container.Resources.Limits[corev1.ResourceCPU]; cpu.IsZero() {
			podlog.Info(".spec.resources.limits.cpu is missing")
			allErrors = append(allErrors, field.Invalid(field.NewPath("spec").Child("containers").Index(i).Child("resources", "limits", "cpu"), cpu.String(), "cpu limit must be specified"))
		}

		if memory := container.Resources.Limits[corev1.ResourceMemory]; memory.IsZero() {
			podlog.Info(".spec.resources.limits.memory is missing")
			allErrors = append(allErrors, field.Invalid(field.NewPath("spec").Child("containers").Index(i).Child("resources", "limits", "memory"), memory.String(), "memory limit must be specified"))
		}
	}

	if len(allErrors) > 0 {
		return apierrors.NewInvalid(corev1.SchemeGroupVersion.WithKind("Pod").GroupKind(), pod.Name, allErrors)
	}

	return nil
}

func (v *PodCustomValidator) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	// we don't validate deletions
	return nil, nil
}
