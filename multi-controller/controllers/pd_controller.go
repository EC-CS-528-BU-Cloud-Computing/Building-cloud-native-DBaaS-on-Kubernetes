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
	r.logger.Info("PD reconcile")

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
		r.logger.Info("PD svc created successfully")

		pdExistingSVC := &corev1.Service{}

		//Check if service already exist
		err = r.Get(context.TODO(), req.NamespacedName, pdExistingSVC)

		if err != nil {
			//svc does not exist
			if errors.IsNotFound(err) {
				err = r.Create(context.TODO(), pdSVC)
			}

			//failed to create svc, requeue with error
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	//TODO: think about service creation logic
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tidbclusterv1.Pd{}).
		Complete(r)
}
