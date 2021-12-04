/*
Copyright 2021.

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
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ImageCloneReconciler reconciles a ImageClone object
type ImageCloneReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=controllergroup.kubermatic.com,resources=imageclones,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controllergroup.kubermatic.com,resources=imageclones/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controllergroup.kubermatic.com,resources=imageclones/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ImageClone object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ImageCloneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// set up a convenient log object so we don't have to type request over and over again
	log := log.FromContext(ctx)

	// Fetch the ReplicaSet from the cache
	rs := &appsv1.ReplicaSet{}
	err := r.Client.Get(ctx, req.NamespacedName, rs)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not find ReplicaSet")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch ReplicaSet: %+v", err)
	}

	// Print the ReplicaSet
	log.Info("Reconciling ReplicaSet", "container name", rs.Spec.Template.Spec.Containers[0].Name)

	// Set the label if it is missing
	if rs.Labels == nil {
		rs.Labels = map[string]string{}
	}
	if rs.Labels["hello"] == "world" {
		return reconcile.Result{}, nil
	}

	// Update the ReplicaSet
	rs.Labels["hello"] = "world"
	err = r.Client.Update(ctx, rs)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not write ReplicaSet: %+v", err)
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImageCloneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Log.Info("Setting up controller")
	c, err := controller.New("image-clone-controller", mgr, controller.Options{
		Reconciler: &ImageCloneReconciler{Client: mgr.GetClient()},
	})
	if err != nil {
		r.Log.Error(err, "unable to set up individual controller")
		return err
	}

	// Watch ReplicaSets and enqueue ReplicaSet object key
	if err := c.Watch(&source.Kind{Type: &appsv1.ReplicaSet{}}, &handler.EnqueueRequestForObject{}); err != nil {
		r.Log.Error(err, "unable to watch ReplicaSets")
		return err
	}

	// Watch Pods and enqueue owning ReplicaSet key
	if err := c.Watch(&source.Kind{Type: &corev1.Pod{}},
		&handler.EnqueueRequestForOwner{OwnerType: &appsv1.ReplicaSet{}, IsController: true}); err != nil {
		r.Log.Error(err, "unable to watch Pods")
		return err
	}
	return nil

}
