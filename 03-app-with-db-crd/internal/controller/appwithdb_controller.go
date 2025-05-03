package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mindscov1 "minds.co/repo/api/v1"
)

type AppWithDbReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=minds.co.minds.co,resources=appwithdbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=minds.co.minds.co,resources=appwithdbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=minds.co.minds.co,resources=appwithdbs/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *AppWithDbReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var crd mindscov1.AppWithDb
	if err := r.Get(ctx, req.NamespacedName, &crd); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 1. Reconcile postgres deployment

	// 1.1. Define desired state
	postgresDeployment, err := r.definePostgresDeployment(&crd)
	if err != nil {
		logger.Error(err, "Failed to define postgres deployment")
		return ctrl.Result{}, err
	}

	// 1.2. Try fetching existing deployment
	existingPostgresDeployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: postgresDeployment.Namespace, Name: postgresDeployment.Name}, existingPostgresDeployment); err != nil {
		// 1.2.1. If doesn't exist -> create it
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, postgresDeployment); err != nil {
				logger.Error(err, "Failed to create postgres deployment")
				return ctrl.Result{}, err
			}

			// 1.2.1.1. On successful create, we restart (requeue) reconciliation
			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Failed to fetch existing postgres deployment")
		return ctrl.Result{}, err
	}

	// 2. Reconcile postgres service

	postgresService, err := r.definePostgresService(&crd)
	if err != nil {
		logger.Error(err, "Failed to define postgres service")
		return ctrl.Result{}, err
	}

	existingPostgresService := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: postgresService.Namespace, Name: postgresService.Name}, existingPostgresService); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, postgresService); err != nil {
				logger.Error(err, "Failed to create postgres service")
				return ctrl.Result{}, err
			}

			return ctrl.Result{Requeue: true}, nil
		}
		logger.Error(err, "Failed to fetch existing postgres service")
		return ctrl.Result{}, err
	}

	// 3. Reconcile app deployment
	appDeployment, err := r.defineAppDeployment(&crd)
	if err != nil {
		logger.Error(err, "Failed to define app deployment")
		return ctrl.Result{}, err
	}

	existingAppDeployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: appDeployment.Namespace, Name: appDeployment.Name}, existingAppDeployment); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, appDeployment); err != nil {
				logger.Error(err, "Failed to create app deployment")
				return ctrl.Result{}, err
			}

			return ctrl.Result{Requeue: true}, nil
		}

		logger.Error(err, "Failed to fetch existing app deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppWithDbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mindscov1.AppWithDb{}).
		Named("appwithdb").
		Complete(r)
}

func (r *AppWithDbReconciler) definePostgresDeployment(app *mindscov1.AppWithDb) (*appsv1.Deployment, error) {
	postgresDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name + "-postgres",
			Namespace: app.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": app.Name, "component": "db"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": app.Name, "component": "db"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "postgres",
						Image:           "postgres:15.1",
						ImagePullPolicy: "IfNotPresent",
						Env: []corev1.EnvVar{
							{Name: "POSTGRES_DB", Value: "db"},
							{Name: "POSTGRES_USER", Value: "user"},
							{Name: "POSTGRES_PASSWORD", Value: "password"},
						},
						Ports: []corev1.ContainerPort{{ContainerPort: 5432}},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("250m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1000m"),
								corev1.ResourceMemory: resource.MustParse("1024Mi"),
							},
						},
					}},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(app, postgresDep, r.Scheme); err != nil {
		return nil, err
	}

	return postgresDep, nil
}

func (r *AppWithDbReconciler) definePostgresService(app *mindscov1.AppWithDb) (*corev1.Service, error) {
	postgresSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name + "-postgres",
			Namespace: app.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": app.Name, "component": "db"},
			Ports: []corev1.ServicePort{{
				Port:     5432,
				Protocol: corev1.ProtocolTCP,
			}},
		},
	}

	if err := ctrl.SetControllerReference(app, postgresSvc, r.Scheme); err != nil {
		return nil, err
	}

	return postgresSvc, nil
}

func (r *AppWithDbReconciler) defineAppDeployment(app *mindscov1.AppWithDb) (*appsv1.Deployment, error) {
	connectionString := fmt.Sprintf("postgres://user:password@%s-postgres:5432/db?sslmode=disable", app.Name)

	appDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name + "-app",
			Namespace: app.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": app.Name, "component": "app"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": app.Name, "component": "app"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "app",
						Image:           app.Spec.Image,
						ImagePullPolicy: "IfNotPresent",
						Env: []corev1.EnvVar{
							{Name: "CONNECTION_STRING", Value: connectionString},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("250m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1000m"),
								corev1.ResourceMemory: resource.MustParse("1024Mi"),
							},
						},
					}},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(app, appDep, r.Scheme); err != nil {
		return nil, err
	}

	return appDep, nil
}
