/*
Copyright 2023.

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

package controllers

import (
	"context"

	infrav1alpha1 "github.com/cmwylie19/kubescrub-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ReaperReconciler reconciles a Reaper object
type ReaperReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	Name      = "kubescrub"
	Namespace = "kubescrub-operator-system"
)

//+kubebuilder:rbac:groups=infra.caseywylie.io,resources=reapers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infra.caseywylie.io,resources=reapers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infra.caseywylie.io,resources=reapers/finalizers,verbs=update

// +kubebuilder:rbac:groups=networking,resources=ingress,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Reaper object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ReaperReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling Reaper", "namespace", req.NamespacedName)

	// check on the Reaper resource
	kubescrub := &infrav1alpha1.Reaper{}
	err := r.Get(ctx, req.NamespacedName, kubescrub)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Reaper resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Reaper")
		return ctrl.Result{}, err
	}

	// Check if deployment already exists, if not create a new one
	deploy := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: Name, Namespace: Namespace}, deploy)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForKubescrub(kubescrub)
		logger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err

	}

	// Check on the service resources

	// check on the service account resource

	// check on the clusterrole resource

	// check on the clusterrolebinding resources

	return ctrl.Result{}, nil
}

func labelsForKubescrub(name string) map[string]string {
	return map[string]string{"app": "kubescrub", "kubescrub_cr": name}
}

func (r *ReaperReconciler) ingressForKubescrub(k *infrav1alpha1.Reaper) *networkingv1.Ingress {
	ls := labelsForKubescrub(k.Name)
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: Namespace,
			Labels:    ls,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &[]string{"nginx"}[0],
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/scrub",
									PathType: &[]networkingv1.PathType{networkingv1.PathTypePrefix}[0],
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: Name,
											Port: networkingv1.ServiceBackendPort{
												Number: 8080,
											},
										},
									},
								}, {
									Path:     "/",
									PathType: &[]networkingv1.PathType{networkingv1.PathTypePrefix}[0],
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: Name + "-web",
											Port: networkingv1.ServiceBackendPort{
												Number: 8080,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	ctrl.SetControllerReference(k, ingress, r.Scheme)
	return ingress
}

func (r *ReaperReconciler) clusterRoleBindingForKubescrub(k *infrav1alpha1.Reaper) *rbacv1.ClusterRoleBinding {
	ls := labelsForKubescrub(k.Name)
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: Namespace,
			Labels:    ls,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      Name,
				Namespace: Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     Name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	ctrl.SetControllerReference(k, crb, r.Scheme)
	return crb
}
func (r *ReaperReconciler) serviceForKubescrubWeb(k *infrav1alpha1.Reaper) *corev1.Service {
	ls := labelsForKubescrub(k.Name + "-web")

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name + "-web",
			Namespace: Namespace,
			Labels:    ls,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 8080,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
			},
			Selector: ls,
		},
	}
	ctrl.SetControllerReference(k, svc, r.Scheme)
	return svc
}
func (r *ReaperReconciler) serviceForKubescrub(k *infrav1alpha1.Reaper) *corev1.Service {
	ls := labelsForKubescrub(k.Name)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: Namespace,
			Labels:    ls,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port: 8080,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
			},
			Selector: ls,
		},
	}
	ctrl.SetControllerReference(k, svc, r.Scheme)
	return svc
}
func (r *ReaperReconciler) clusterRoleForKubeScrub(k *infrav1alpha1.Reaper) *rbacv1.ClusterRole {
	ls := labelsForKubescrub(k.Name)
	cr := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: Namespace,
			Labels:    ls,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "nodes"},
				Verbs:     []string{"get", "list", "watch", "delete"},
			},
		},
	}
	ctrl.SetControllerReference(k, cr, r.Scheme)
	return cr
}
func (r *ReaperReconciler) serviceAccountForKubescrub(k *infrav1alpha1.Reaper) *corev1.ServiceAccount {
	ls := labelsForKubescrub(k.Name)
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: Namespace,
			Labels:    ls,
		},
	}
	ctrl.SetControllerReference(k, sa, r.Scheme)
	return sa
}
func (r *ReaperReconciler) deploymentForKubescrubWeb(k *infrav1alpha1.Reaper) *appsv1.Deployment {
	ls := labelsForKubescrub(k.Name + "-web")
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name + "-web",
			Namespace: Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           "docker.io/caseywylie/kubescrub-ui:0.0.1",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "kubescrub-web",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "http",
						}},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(k, dep, r.Scheme)
	return dep
}
func (r *ReaperReconciler) deploymentForKubescrub(k *infrav1alpha1.Reaper) *appsv1.Deployment {
	ls := labelsForKubescrub(k.Name)
	replicas := int32(1)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           "docker.io/caseywylie/kubescrub:v0.0.1",
						ImagePullPolicy: corev1.PullAlways,
						Name:            "kubescrub",
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/scrub/healthz",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 8080,
									},
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       5,
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/scrub/healthz",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 8080,
									},
								},
							},
							InitialDelaySeconds: 5,
							PeriodSeconds:       5,
						},

						Command: []string{"./kubescrub", "serve", "-p", "8080", "--theme", k.Spec.Theme, "--watch", k.Spec.Resources, "--namespaces", k.Spec.Namespaces, "--poll", k.Spec.Poll, "--pollInterval", k.Spec.PollInterval},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "http",
						}},
					}},
					RestartPolicy:      corev1.RestartPolicyAlways,
					ServiceAccountName: Name,
				},
			},
		},
	}
	// Set Reaper instance as the owner and controller
	ctrl.SetControllerReference(k, dep, r.Scheme)
	return dep
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReaperReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.Reaper{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
