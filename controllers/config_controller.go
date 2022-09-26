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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/oam-dev/vector-controller/api/v1alpha1"
)

const (
	configFinalizer = "vector.oam.dev/config-finalizer"
)

// ConfigReconciler reconciles a Config object
type ConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vector.oam.dev,resources=configs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vector.oam.dev,resources=configs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vector.oam.dev,resources=configs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Config object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	object := v1alpha1.Config{}
	var err error

	if err = r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &object); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !controllerutil.ContainsFinalizer(&object, configFinalizer) {
		patch := client.MergeFrom(object.DeepCopy())
		controllerutil.AddFinalizer(&object, configFinalizer)
		if err := r.Patch(ctx, &object, patch); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if !object.DeletionTimestamp.IsZero() {
		err := r.reconcileDelete(ctx, &object)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	switch object.Spec.Role {
	case v1alpha1.SidecarRoleType:
		err := createOrUpdateUniqueConfigmap(ctx, r.Client, object)
		if err != nil {
			msg := err.Error()
			object.Status.Message = msg
			if err := r.Status().Update(ctx, &object); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		object.Status.Message = "Successful set vector config to configmap"
		if err := r.Status().Update(ctx, &object); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	case v1alpha1.DaemonRoleType, v1alpha1.AggregatorType:
		err := createOrMergeConfigmap(ctx, r.Client, &object)
		if err != nil {
			msg := err.Error()
			object.Status.Message = msg
			if err := r.Status().Update(ctx, &object); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		object.Status.Message = "Successful merge vector config to targetConfigmap"
		if err := r.Status().Update(ctx, &object); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Config{}).
		Complete(r)
}

func (r *ConfigReconciler) reconcileDelete(ctx context.Context, config *v1alpha1.Config) error {
	switch config.Spec.Role {
	case v1alpha1.SidecarRoleType:
		cm := v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: config.Name, Namespace: config.Namespace}}
		err := r.Delete(ctx, &cm)
		if err != nil {
			return err
		}

	case v1alpha1.DaemonRoleType, v1alpha1.AggregatorType:
		err := deleteConfigFromConfigmap(ctx, r.Client, config)
		if err != nil {
			return err
		}
	}
	patch := client.MergeFrom(config.DeepCopy())
	controllerutil.RemoveFinalizer(config, configFinalizer)
	return r.Patch(ctx, config, patch)
}
