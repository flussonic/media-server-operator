/*
Copyright 2024.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	// "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// "sigs.k8s.io/controller-runtime/pkg/log"
	mediav1 "github.com/flussonic/media-server-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/lithammer/shortuuid/v3"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// MediaServerReconciler reconciles a MediaServer object
type MediaServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=media.flussonic.com,resources=mediaservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=media.flussonic.com,resources=mediaservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=media.flussonic.com,resources=mediaservers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *MediaServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)
	log := r.Log.WithValues("MediaServer", req.NamespacedName, "ReconcileId", shortuuid.New())

	// TODO(user): your logic here
	log.Info("Processing MediaServerReconciler")

	mediaServer := &mediav1.MediaServer{}
	err := r.Client.Get(ctx, req.NamespacedName, mediaServer)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("MediaServer resource is not found, ignoring further work")
			return ctrl.Result{}, nil
		}
		log.Error(err, "error getting MediaServer")
		return ctrl.Result{}, err
	}

	result, err := r.deployServiceAccount(ctx, mediaServer)
	if err != nil || result.Requeue {
		return result, err
	}

	result, err = r.deployDaemonSet(ctx, mediaServer)
	if err != nil || result.Requeue {
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *MediaServerReconciler) deployServiceAccount(ctx context.Context, ms *mediav1.MediaServer) (ctrl.Result, error) {
	serviceAccountName := ms.Name + "-sa"
	roleName := ms.Name + "-role"
	roleBindingName := ms.Name + "-rb"

	sa1 := &corev1.ServiceAccount{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: serviceAccountName, Namespace: ms.Namespace}, sa1)
	if err != nil && errors.IsNotFound(err) {
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceAccountName,
				Namespace: ms.Namespace,
			},
			// Secrets: []corev1.ObjectReference{{
			// 	Name: ms.Name,
			// 	Namespace: ms.Namespace,
			// }},
		}
		ctrl.SetControllerReference(ms, sa, r.Scheme)
		err = r.Client.Create(ctx, sa)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	role1 := &rbacv1.Role{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: roleName, Namespace: ms.Namespace}, role1)
	if err != nil && errors.IsNotFound(err) {
		role := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      roleName,
				Namespace: ms.Namespace,
			},
			Rules: []rbacv1.PolicyRule{
				{
					Verbs:         []string{"get", "update", "patch"},
					Resources:     []string{"secrets"},
					ResourceNames: []string{ms.Name + "-license-storage"},
					APIGroups:     []string{""},
				},
			},
		}
		ctrl.SetControllerReference(ms, role, r.Scheme)
		err = r.Client.Create(ctx, role)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	rb1 := &rbacv1.RoleBinding{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: roleBindingName, Namespace: ms.Namespace}, rb1)
	if err != nil && errors.IsNotFound(err) {
		rb := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      roleBindingName,
				Namespace: ms.Namespace,
			},
			Subjects: []rbacv1.Subject{{
				Kind: "ServiceAccount",
				Name: serviceAccountName,
			}},
			RoleRef: rbacv1.RoleRef{
				Kind:     "Role",
				Name:     roleName,
				APIGroup: "rbac.authorization.k8s.io",
			},
		}
		ctrl.SetControllerReference(ms, rb, r.Scheme)
		err = r.Client.Create(ctx, rb)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func dsEnv(ms *mediav1.MediaServer) []corev1.EnvVar {
	env := []corev1.EnvVar{{
		Name: "FLUSSONIC_HOSTNAME",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "spec.nodeName",
			},
		},
	}}

	env = append(env, corev1.EnvVar{
		Name:  "FLUSSONIC_SECRETS_STORAGE",
		Value: "k8s://" + ms.Name + "-license-storage",
	})

	env = append(env, corev1.EnvVar{
		Name:  "DO_NOT_DO_NET_TUNING",
		Value: "true",
	})

	env = append(env, ms.Spec.PodEnvVariables...)
	return env
}

func (r *MediaServerReconciler) deployDaemonSet(ctx context.Context, ms *mediav1.MediaServer) (ctrl.Result, error) {
	secretName := ms.Name + "-license-storage"
	configMapName := ms.Name + "-configmap"
	appLabel := ms.Name + "-streamer"
	configVolName := ms.Name + "-configvol"
	daemonSetName := ms.Name + "-streamer"

	ls1 := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: ms.Namespace}, ls1)
	if err != nil && errors.IsNotFound(err) {
		ls := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: ms.Namespace,
			},
			StringData: map[string]string{
				"initial": "value",
			},
		}
		ctrl.SetControllerReference(ms, ls, r.Scheme)
		err := r.Client.Create(ctx, ls)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	configData := map[string]string{
		"k8s.conf": "http 80 {api false;} http 81;",
	}
	if ms.Spec.ConfigExtra != nil {
		for k, v := range ms.Spec.ConfigExtra {
			configData[k] = v
		}
	}

	cm1 := &corev1.ConfigMap{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: ms.Namespace}, cm1)
	if err != nil && errors.IsNotFound(err) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: ms.Namespace,
			},
			Data: configData,
		}
		ctrl.SetControllerReference(ms, cm, r.Scheme)
		err := r.Client.Create(ctx, cm)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	cm1.Data = configData
	err = r.Client.Update(ctx, cm1)
	if err != nil {
		return ctrl.Result{}, err
	}

	labels := map[string]string{"app": appLabel}

	dataPort := corev1.ContainerPort{
		Name:          "data",
		ContainerPort: 80,
		HostPort:      ms.Spec.HostPort,
	}
	apiPort := corev1.ContainerPort{
		Name:          "api",
		ContainerPort: 81,
		HostPort:      ms.Spec.AdminHostPort,
	}

	env := dsEnv(ms)
	defaultMode := int32(0440)
	optional := false

	configMapping := []corev1.KeyToPath{}
	streamerVolumeMounts := []corev1.VolumeMount{}
	streamerVolumes := []corev1.Volume{{
		Name: configVolName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
				Items:       configMapping,
				DefaultMode: &defaultMode,
				Optional:    &optional,
			},
		},
	}}

	for k, _ := range configData {
		configMapping = append(configMapping, corev1.KeyToPath{
			Key:  k,
			Path: k,
		})
		streamerVolumeMounts = append(streamerVolumeMounts, corev1.VolumeMount{
			Name:      configVolName,
			MountPath: "/etc/flussonic/flussonic.conf.d/" + k,
			SubPath:   k,
			ReadOnly:  true,
		})
	}

	for _, volume := range ms.Spec.Volumes {
		streamerVolumeMounts = append(streamerVolumeMounts, corev1.VolumeMount{
			Name:      volume.Name,
			MountPath: volume.MountPath,
		})
		streamerVolumes = append(streamerVolumes, corev1.Volume{
			Name:         volume.Name,
			VolumeSource: volume.VolumeSource,
		})

	}

	spec := corev1.PodSpec{
		Volumes:            streamerVolumes,
		NodeSelector:       ms.Spec.NodeSelector,
		ServiceAccountName: ms.Name + "-sa",
		Containers: []corev1.Container{{
			Name:            "mediaserver",
			Image:           ms.Spec.Image,
			ImagePullPolicy: "IfNotPresent",
			Env:             env,
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/streamer/api/v3/monitoring/liveness",
						Port: intstr.FromInt(81),
					},
				},
				InitialDelaySeconds: 10,
				PeriodSeconds:       3,
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/streamer/api/v3/monitoring/readiness",
						Port: intstr.FromInt(81),
					},
				},
				InitialDelaySeconds: 2,
				PeriodSeconds:       2,
			},
			StartupProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/streamer/api/v3/monitoring/readiness",
						Port: intstr.FromInt(81),
					},
				},
				InitialDelaySeconds: 2,
				PeriodSeconds:       2,
				FailureThreshold:    30,
			},
			Ports: []corev1.ContainerPort{
				dataPort,
				apiPort,
			},
			VolumeMounts: streamerVolumeMounts,
		}},
	}

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemonSetName,
			Namespace: ms.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: spec,
			},
		},
	}

	ds1 := &appsv1.DaemonSet{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: daemonSetName, Namespace: ms.Namespace}, ds1)
	if err != nil && errors.IsNotFound(err) {
		ctrl.SetControllerReference(ms, ds, r.Scheme)
		err = r.Client.Create(ctx, ds)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}
	ds1.Spec.Template.Spec.Volumes[0].VolumeSource.ConfigMap.Items = configMapping
	ds1.Spec.Template.Spec.NodeSelector = ms.Spec.NodeSelector
	ds1.Spec.Template.Spec.Volumes = streamerVolumes
	ds1.Spec.Template.Spec.Containers[0].Image = ms.Spec.Image
	ds1.Spec.Template.Spec.Containers[0].Ports[0].HostPort = ms.Spec.HostPort
	ds1.Spec.Template.Spec.Containers[0].Ports[1].HostPort = ms.Spec.AdminHostPort
	ds1.Spec.Template.Spec.Containers[0].VolumeMounts = streamerVolumeMounts
	ds1.Spec.Template.Spec.Containers[0].Env = env
	err = r.Client.Update(ctx, ds1)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MediaServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mediav1.MediaServer{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
