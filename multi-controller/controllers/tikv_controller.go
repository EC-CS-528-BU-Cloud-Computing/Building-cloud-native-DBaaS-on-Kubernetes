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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"

	tidbclusterv1 "cluster-operator/api/v1"
	"cluster-operator/pkg/spawn"
)

// TikvReconciler reconciles a Tikv object
type TikvReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logr.Logger
}

const defaultTikvImage string = "pingcap/tikv:latest"

//+kubebuilder:rbac:groups=tidb-cluster.dbaas,resources=tikvs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tidb-cluster.dbaas,resources=tikvs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tidb-cluster.dbaas,resources=tikvs/finalizers,verbs=update

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
	r.logger = log.FromContext(ctx)
	r.logger.Info("Tikv reconcile")

	// TODO(user): your logic here
	instance := &tidbclusterv1.Tikv{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// tikv not found, could have been deleted after
			// reconcile request, don't requeue
			r.logger.Info("TiKV instance not found, could be deleted, NOT REQUEUED")
			return ctrl.Result{}, nil
		}

		// error reading the object, requeue the request
		r.logger.Info("TiKV instance read err, REQUEUED")
		return ctrl.Result{}, err
	}

	if instance.Status.Phase == "" {
		instance.Status.Phase = tidbclusterv1.PhasePending_TiKV
	}

	if instance.Spec.Imagename == "" {
		instance.Spec.Imagename = defaultTikvImage
	}

	if instance.Spec.HealthCheckInterval == 0 {
		instance.Spec.HealthCheckInterval = 5
	}

	switch instance.Status.Phase {
	case tidbclusterv1.PhasePending_TiKV:
		r.logger.Info("Phase: TIKV PENDING")
		//Check if pd is running
		if !r.isPdRunning(ctx) {
			r.logger.Info("Pd instance not ready")
			return ctrl.Result{RequeueAfter: time.Duration(instance.Spec.HealthCheckInterval) * time.Second}, nil
		} else {
			r.logger.Info("Phase: Pd is running, Tikv PENDING for creation, now will create")
			instance.Status.Phase = tidbclusterv1.PhaseCreating_TiKV
		}
	case tidbclusterv1.PhaseCreating_TiKV:
		r.logger.Info("Phase: TIKV CREATING")
		pod := spawn.NewTikvPod(instance)
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
			r.logger.Info("Pod Created successfully", "name", pod.Name)
			instance.Status.Phase = tidbclusterv1.PhaseRunning_TiKV
			err = r.updateTikvInstanceStatus(&ctx, instance)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: time.Duration(instance.Spec.HealthCheckInterval) * time.Second}, nil
		} else if err != nil {
			// requeue with err
			r.logger.Error(err, "cannot create pod")
			return ctrl.Result{}, err
		} else if tikvPod.Status.Phase == corev1.PodFailed {
			// pod errored out, need to recreate
			r.logger.Info("TiKV pod failed", "reason", tikvPod.Status.Reason, "message", tikvPod.Status.Message)
			// delete failed pod
			err = r.Delete(context.TODO(), tikvPod)
			if err != nil {
				return ctrl.Result{}, err
			}
			instance.Status.Phase = tidbclusterv1.PhasePending_TiKV
		} else {
			//Pod already exist and running, do nothing
			return ctrl.Result{}, nil
		}
	case tidbclusterv1.PhaseRunning_TiKV:
		r.logger.Info("Phase: TIKV RUNNING")

		tikvPod := &corev1.Pod{}

		//Get current pod
		err = r.Get(context.TODO(), req.NamespacedName, tikvPod)
		if err != nil && errors.IsNotFound(err) {
			r.logger.Info("TiKV pod disappeared in RUNNING phase, return to PENDING")
			instance.Status.Phase = tidbclusterv1.PhasePending_TiKV
		} else if err != nil {
			// requeue with err
			r.logger.Error(err, "cannot create TiKV pod")
			return ctrl.Result{}, err
		} else if tikvPod.Status.Phase == corev1.PodFailed {
			// pod errored out, need to recreate
			r.logger.Info("TiKV pod failed", "reason", tikvPod.Status.Reason, "message", tikvPod.Status.Message)
			// delete failed pod
			err = r.Delete(context.TODO(), tikvPod)
			if err != nil {
				return ctrl.Result{}, err
			}
			instance.Status.Phase = tidbclusterv1.PhasePending_TiKV
		} else {
			//Pod already exist and running, do nothing
			return ctrl.Result{RequeueAfter: time.Duration(instance.Spec.HealthCheckInterval) * time.Second}, nil
		}
	default:
		r.logger.Info("ERROR: Unknown phase")
		return ctrl.Result{}, nil
	}

	// update status
	err = r.updateTikvInstanceStatus(&ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TikvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tidbclusterv1.Tikv{}).
		Complete(r)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TikvReconciler) updateTikvInstanceStatus(ctx *context.Context, instance *tidbclusterv1.Tikv) error {
	return r.Status().Update(context.TODO(), instance)
}

// Check pd running
func (r *TikvReconciler) isPdRunning(ctx context.Context) bool {

	pdInstance := &tidbclusterv1.Pd{}
	pdNamespacedName := types.NamespacedName{
		Namespace: "default",
		Name:      "pd-sample",
	}
	err := r.Get(context.TODO(), pdNamespacedName, pdInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			r.logger.Info("Pd instance not found")
		} else {
			r.logger.Info("Error reading pd instance")
		}
		return false
	} else if pdInstance.Status.Phase != tidbclusterv1.PhaseRunning_PD {
		r.logger.Info("Pd instance not running")
		return false
	}
	return true

	//time.Sleep(10 * time.Second)
	//return true
}
