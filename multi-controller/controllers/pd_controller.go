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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"

	tidbclusterv1 "cluster-operator/api/v1"
	"cluster-operator/pkg/spawn"
)

// PdReconciler reconciles a Pd object
type PdReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logr.Logger
}

const defaultPdImage string = "pingcap/pd:latest"

//+kubebuilder:rbac:groups=tidb-cluster.dbaas,resources=pds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tidb-cluster.dbaas,resources=pds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tidb-cluster.dbaas,resources=pds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pd object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = log.FromContext(ctx)
	r.logger.Info("PD reconcile, ", "namespace: ", req.Namespace, "name: ", req.Name)
	//r.logger.Info(string(req.NamespacedName))

	// TODO(user): your logic here
	pdInstance := &tidbclusterv1.Pd{}
	err := r.Get(context.TODO(), req.NamespacedName, pdInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			// pd not found, could have been deleted after
			// reconcile request, don't requeue
			r.logger.Info("PD instance not found, could be deleted, NOT REQUEUED")
			return ctrl.Result{}, nil
		}

		// error reading the object, requeue the request
		r.logger.Info("PD instance read err, REQUEUED")
		return ctrl.Result{}, err
	}

	if pdInstance.Status.Phase == "" {
		pdInstance.Status.Phase = tidbclusterv1.PhaseCreating_PD_SVC
	}

	if pdInstance.Spec.Imagename == "" {
		pdInstance.Spec.Imagename = defaultPdImage
	}

	if pdInstance.Spec.HealthCheckInterval == 0 {
		pdInstance.Spec.HealthCheckInterval = 5
	}

	switch pdInstance.Status.Phase {
	case tidbclusterv1.PhaseCreating_PD_SVC:
		r.logger.Info("Phase: PD CREATING SVC")
		pdSVC := spawn.CreatePdSVC(pdInstance)
		err := ctrl.SetControllerReference(pdInstance, pdSVC, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}

		pdExistingSVC := &corev1.Service{}

		//Check if service already exist
		err = r.Get(context.TODO(), req.NamespacedName, pdExistingSVC)

		if err != nil && errors.IsNotFound(err) {
			//svc does not exist
			err = r.Create(context.TODO(), pdSVC)
			//failed to create svc, requeue with error
			if err != nil {
				return ctrl.Result{}, err
			}
			r.logger.Info("PD SVC Created successfully", "name", pdSVC.Name)
			pdInstance.Status.Phase = tidbclusterv1.PhaseCreating_PD_Pod
			err = r.updatePDInstanceStatus(&ctx, pdInstance)
			if err != nil {
				return ctrl.Result{}, err
			}
			//Start creating pod after 1 second
			return ctrl.Result{RequeueAfter: time.Second}, nil
		} else if err != nil {
			// requeue with err
			r.logger.Error(err, "cannot create pd svc")
			return ctrl.Result{}, err
		} else {
			// Service exist, proceed to pod creation
			pdInstance.Status.Phase = tidbclusterv1.PhaseCreating_PD_Pod
		}
	case tidbclusterv1.PhaseCreating_PD_Pod:
		r.logger.Info("Phase: PD CREATING Pod")
		pdPod := spawn.NewPdPod(pdInstance)
		err := ctrl.SetControllerReference(pdInstance, pdPod, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}

		pdExistingPod := &corev1.Pod{}

		//Check if pod already exist
		err = r.Get(context.TODO(), req.NamespacedName, pdExistingPod)

		if err != nil && errors.IsNotFound(err) {
			//pod does not exist
			err = r.Create(context.TODO(), pdPod)
			//failed to create pod, requeue with error
			if err != nil {
				return ctrl.Result{}, err
			}
			r.logger.Info("PD Pod Created successfully", "name", pdPod.Name)
			pdInstance.Status.Phase = tidbclusterv1.PhaseRunning_PD
			err = r.updatePDInstanceStatus(&ctx, pdInstance)
			if err != nil {
				return ctrl.Result{}, err
			}
			//Start creating pod after health check interval
			return ctrl.Result{RequeueAfter: time.Duration(pdInstance.Spec.HealthCheckInterval) * time.Second}, nil
		} else if err != nil {
			// requeue with err
			r.logger.Error(err, "cannot create pd pod, REQUEUE")
			return ctrl.Result{}, err
		} else if pdExistingPod.Status.Phase == corev1.PodFailed {
			// pod errored out, need to recreate
			r.logger.Info("PD pod failed", "reason", pdExistingPod.Status.Reason, "message", pdExistingPod.Status.Message)
			// delete failed pod
			err = r.Delete(context.TODO(), pdExistingPod)
			if err != nil {
				return ctrl.Result{}, err
			}
			pdInstance.Status.Phase = tidbclusterv1.PhaseCreating_PD_Pod
		} else {
			//Pod already exist and running, do nothing
			return ctrl.Result{}, nil
		}
	case tidbclusterv1.PhaseRunning_PD:
		r.logger.Info("Phase: PD RUNNING")

		pdPod := &corev1.Pod{}

		//Get current pod
		err = r.Get(context.TODO(), req.NamespacedName, pdPod)
		if err != nil && errors.IsNotFound(err) {
			//Smh the pod disappeared create the pod
			/*
				pod := spawn.NewTidbPod(instance)
				err := ctrl.SetControllerReference(instance, pod, r.Scheme)
				if err != nil {
					return ctrl.Result{}, err
				}
				err = r.Create(context.TODO(), pod)
				if err != nil {
					return ctrl.Result{}, err
				}
				logger.Info("Pod Created successfully", "name", pod.Name)
				instance.Status.Phase = tidbclusterv1.PhaseRunning
				err = r.UpdateInstanceStatus(&ctx, instance)
				if err != nil {
					return ctrl.Result{}, err
				}
				return ctrl.Result{RequeueAfter: time.Duration(instance.Spec.HealthCheckInterval) * time.Second}, nil
			*/
			r.logger.Info("PD pod disappeared in RUNNING phase, return to POD creating")
			pdInstance.Status.Phase = tidbclusterv1.PhaseCreating_PD_Pod
			err = r.updatePDInstanceStatus(&ctx, pdInstance)
			if err != nil {
				return ctrl.Result{}, err
			}
			//Start creating pod after health check interval
			return ctrl.Result{RequeueAfter: time.Second}, nil
		} else if err != nil {
			// requeue with err
			r.logger.Error(err, "cannot create PD pod")
			return ctrl.Result{}, err
		} else if pdPod.Status.Phase == corev1.PodFailed {
			// pod errored out, need to recreate
			r.logger.Info("PD pod failed", "reason", pdPod.Status.Reason, "message", pdPod.Status.Message)
			// delete failed pod
			err = r.Delete(context.TODO(), pdPod)
			if err != nil {
				return ctrl.Result{}, err
			}
			pdInstance.Status.Phase = tidbclusterv1.PhaseCreating_PD_Pod
		} else {
			//Pod already exist and running, do nothing
			return ctrl.Result{RequeueAfter: time.Duration(pdInstance.Spec.HealthCheckInterval) * time.Second}, nil
		}
	default:
		r.logger.Info("ERROR: Unknown phase")
		return ctrl.Result{}, nil
	}
	//TODO: think about service creation logic
	err = r.updatePDInstanceStatus(&ctx, pdInstance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tidbclusterv1.Pd{}).
		Complete(r)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PdReconciler) updatePDInstanceStatus(ctx *context.Context, pdInstance *tidbclusterv1.Pd) error {
	return r.Status().Update(context.TODO(), pdInstance)
}
