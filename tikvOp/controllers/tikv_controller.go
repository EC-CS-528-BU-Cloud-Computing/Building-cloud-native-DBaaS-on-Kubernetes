/*
Copyright 2022.

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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"

	dbaasv1alpha1 "tikvOp/api/v1alpha1"
	"tikvOp/pkg/spawn"
)

// TikvReconciler reconciles a Tikv object
type TikvReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const defaultTikvImage string = "pingcap/tikv:latest"

//+kubebuilder:rbac:groups=dbaas.cs528,resources=tikvs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dbaas.cs528,resources=tikvs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dbaas.cs528,resources=tikvs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tikv object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *TikvReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("iKV reconcile")

	// TODO(user): your logic here
	instance := &dbaasv1alpha1.Tikv{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			// tikv not found, could have been deleted after
			// reconcile request, don't requeue
			return ctrl.Result{}, nil
		}

		// error reading the object, requeue the request
		return ctrl.Result{}, err
	}

	if instance.Status.Phase == "" {
		instance.Status.Phase = dbaasv1alpha1.PhasePending
	}

	if instance.Spec.Imagename == "" {
		instance.Spec.Imagename = defaultTikvImage
	}

	if instance.Spec.HealthCheckInterval == 0 {
		instance.Spec.HealthCheckInterval = 5
	}

	switch instance.Status.Phase {
	case dbaasv1alpha1.PhasePending:
		logger.Info("Phase: PENDING for creation, now will create")
		instance.Status.Phase = dbaasv1alpha1.PhaseCreating
	case dbaasv1alpha1.PhaseCreating:
		logger.Info("Phase: CREATING")
		pod := spawn.NewPodForCR(instance)
		err := ctrl.SetControllerReference(instance, pod, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}

		tikvPod := &corev1.Pod{}

		//Check if pod already exist
		err = r.Get(context.TODO(), req.NamespacedName, tikvPod)
		if err != nil && errors.IsNotFound(err) {
			//Pod does not exist create the pod
			err = r.Create(context.TODO(), pod)
			if err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Pod Created successfully", "name", pod.Name)
			instance.Status.Phase = dbaasv1alpha1.PhaseRunning
			err = r.UpdateInstanceStatus(&ctx, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: time.Duration(instance.Spec.HealthCheckInterval) * time.Second}, nil
		} else if err != nil {
			// requeue with err
			logger.Error(err, "cannot create pod")
			return ctrl.Result{}, err
		} else if tikvPod.Status.Phase == corev1.PodFailed {
			// pod errored out, need to recreate
			logger.Info("TiKV pod failed", "reason", tikvPod.Status.Reason, "message", tikvPod.Status.Message)
			// delete failed pod
			err = r.Delete(context.TODO(), tikvPod)
			if err != nil {
				return ctrl.Result{}, err
			}
			instance.Status.Phase = dbaasv1alpha1.PhasePending
		} else {
			//Pod already exist and running, do nothing
			return ctrl.Result{}, nil
		}
	case dbaasv1alpha1.PhaseRunning:
		logger.Info("Phase: RUNNING")

		tidbPod := &corev1.Pod{}

		//Get current pod
		err = r.Get(context.TODO(), req.NamespacedName, tidbPod)
		if err != nil && errors.IsNotFound(err) {
			//Smh the pod disappeared create the pod
			pod := spawn.NewPodForCR(instance)
			err := ctrl.SetControllerReference(instance, pod, r.Scheme)
			if err != nil {
				return ctrl.Result{}, err
			}
			err = r.Create(context.TODO(), pod)
			if err != nil {
				return ctrl.Result{}, err
			}
			logger.Info("Pod Created successfully", "name", pod.Name)
			instance.Status.Phase = dbaasv1alpha1.PhaseRunning
			err = r.UpdateInstanceStatus(&ctx, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: time.Duration(instance.Spec.HealthCheckInterval) * time.Second}, nil
		} else if err != nil {
			// requeue with err
			logger.Error(err, "cannot create pod")
			return ctrl.Result{}, err
		} else if tidbPod.Status.Phase == corev1.PodFailed {
			// pod errored out, need to recreate
			logger.Info("TiKV pod failed", "reason", tidbPod.Status.Reason, "message", tidbPod.Status.Message)
			// delete failed pod
			err = r.Delete(context.TODO(), tidbPod)
			if err != nil {
				return ctrl.Result{}, err
			}
			instance.Status.Phase = dbaasv1alpha1.PhasePending
		} else {
			//Pod already exist and running, do nothing
			return ctrl.Result{RequeueAfter: time.Duration(instance.Spec.HealthCheckInterval) * time.Second}, nil
		}
	default:
		logger.Info("ERROR: Unknown phase")
		return ctrl.Result{}, nil
	}

	// update status
	err = r.UpdateInstanceStatus(&ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TikvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbaasv1alpha1.Tikv{}).
		Complete(r)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TikvReconciler) UpdateInstanceStatus(ctx *context.Context, instance *dbaasv1alpha1.Tikv) error {
	return r.Status().Update(context.TODO(), instance)
}
