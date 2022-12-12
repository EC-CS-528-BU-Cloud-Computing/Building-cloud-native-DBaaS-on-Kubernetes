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

// TidbReconciler reconciles a Tidb object
type TidbReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logr.Logger
}

const defaultTidbImage string = "pingcap/tidb:latest"

//+kubebuilder:rbac:groups=tidb-cluster.dbaas,resources=tidbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tidb-cluster.dbaas,resources=tidbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tidb-cluster.dbaas,resources=tidbs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tidb object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *TidbReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = log.FromContext(ctx)
	r.logger.Info("Tidb reconcile")

	// TODO(user): your logic here
	tidbInstance := &tidbclusterv1.Tidb{}
	err := r.Get(context.TODO(), req.NamespacedName, tidbInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// tikv not found, could have been deleted after
			// reconcile request, don't requeue
			r.logger.Info("TiDB instance not found, could be deleted, NOT REQUEUED")
			return ctrl.Result{}, nil
		}

		// error reading the object, requeue the request
		r.logger.Info("TiDB instance read err, REQUEUED")
		return ctrl.Result{}, err
	}

	if tidbInstance.Status.Phase == "" {
		tidbInstance.Status.Phase = tidbclusterv1.PhasePending_TiDB
	}

	if tidbInstance.Spec.Imagename == "" {
		tidbInstance.Spec.Imagename = defaultTidbImage
	}

	if tidbInstance.Spec.HealthCheckInterval == 0 {
		tidbInstance.Spec.HealthCheckInterval = 5
	}

	switch tidbInstance.Status.Phase {
	case tidbclusterv1.PhasePending_TiDB:
		//Check if pd is running
		if !r.isPdRunning(ctx) {
			r.logger.Info("TIDB: Pd instance not ready")
			return ctrl.Result{RequeueAfter: time.Duration(tidbInstance.Spec.HealthCheckInterval) * time.Second}, nil
		} else {
			r.logger.Info("Phase: Pd is running, TiDB PENDING for creation, now will create")
			tidbInstance.Status.Phase = tidbclusterv1.PhaseCreating_TiDB
		}
	case tidbclusterv1.PhaseCreating_TiDB:
		r.logger.Info("Phase: TIDB CREATING")
		pod := spawn.NewTidbPod(tidbInstance)
		err := ctrl.SetControllerReference(tidbInstance, pod, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}

		tidbPod := &corev1.Pod{}

		//Check if pod already exist
		err = r.Get(context.TODO(), req.NamespacedName, tidbPod)
		if err != nil && errors.IsNotFound(err) {
			//Pod does not exist create the pod
			err = r.Create(context.TODO(), pod)
			if err != nil {
				return ctrl.Result{}, err
			}
			r.logger.Info("Tidb Pod Created successfully", "name", pod.Name)
			tidbInstance.Status.Phase = tidbclusterv1.PhaseRunning_TiDB
			err = r.updateTidbInstanceStatus(&ctx, tidbInstance)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: time.Duration(tidbInstance.Spec.HealthCheckInterval) * time.Second}, nil
		} else if err != nil {
			// requeue with err
			r.logger.Error(err, "cannot create tidb pod")
			return ctrl.Result{}, err
		} else if tidbPod.Status.Phase == corev1.PodFailed {
			// pod errored out, need to recreate
			r.logger.Info("TiDB pod failed", "reason", tidbPod.Status.Reason, "message", tidbPod.Status.Message)
			// delete failed pod
			err = r.Delete(context.TODO(), tidbPod)
			if err != nil {
				return ctrl.Result{}, err
			}
			tidbInstance.Status.Phase = tidbclusterv1.PhasePending_TiDB
		} else {
			//Pod already exist and running, do nothing
			return ctrl.Result{}, nil
		}
	case tidbclusterv1.PhaseRunning_TiDB:
		r.logger.Info("Phase: TIDB RUNNING")

		tidbPod := &corev1.Pod{}

		//Get current pod
		err = r.Get(context.TODO(), req.NamespacedName, tidbPod)
		if err != nil && errors.IsNotFound(err) {
			//Smh the pod disappeared create the pod
			r.logger.Info("TiDB pod disappeared in RUNNING phase, return to PENDING")
			tidbInstance.Status.Phase = tidbclusterv1.PhasePending_TiDB
		} else if err != nil {
			// requeue with err
			r.logger.Error(err, "cannot create TiDB pod")
			return ctrl.Result{}, err
		} else if tidbPod.Status.Phase == corev1.PodFailed {
			// pod errored out, need to recreate
			r.logger.Info("TiDB pod failed", "reason", tidbPod.Status.Reason, "message", tidbPod.Status.Message)
			// delete failed pod
			err = r.Delete(context.TODO(), tidbPod)
			if err != nil {
				return ctrl.Result{}, err
			}
			tidbInstance.Status.Phase = tidbclusterv1.PhasePending_TiDB
		} else {
			//Pod already exist and running, do nothing
			return ctrl.Result{RequeueAfter: time.Duration(tidbInstance.Spec.HealthCheckInterval) * time.Second}, nil
		}
	default:
		r.logger.Info("ERROR: Unknown phase")
		return ctrl.Result{}, nil
	}

	// update status
	err = r.updateTidbInstanceStatus(&ctx, tidbInstance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TidbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tidbclusterv1.Tidb{}).
		Complete(r)
}

// Update the status of tidb instance
func (r *TidbReconciler) updateTidbInstanceStatus(ctx *context.Context, tidbInstance *tidbclusterv1.Tidb) error {
	return r.Status().Update(context.TODO(), tidbInstance)
}

// Check pd running
func (r *TidbReconciler) isPdRunning(ctx context.Context) bool {
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
}
