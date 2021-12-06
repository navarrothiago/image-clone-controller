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
	"strings"
	"sync"

	"github.com/go-logr/logr"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/navarrothiago/image-clone-controller/config"
	"github.com/navarrothiago/image-clone-controller/imagecloner"
	"github.com/navarrothiago/image-clone-controller/objects"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ImageCloneReconciler reconciles a ImageClone object
type ImageCloneReconciler struct {
	// Client can be used to retrieve objects from the APIServer.
	Client client.Client

	// Object holds the bind type for the process
	Object objects.Object

	// Cfg controller configs
	Cfg config.Config

	// Cloner is the image clone object from source repository to target repository
	Cloner imagecloner.Cloner

	logr.Logger
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
func (r *ImageCloneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info(`starting reconcile`)

	// Fetch the controller
	rs := r.Object.Get()
	err := r.Client.Get(ctx, req.NamespacedName, rs)
	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch %s: %+v", r.Object.Name(), err)
	}

	newCopy := r.Object.NewCopy()
	wg := sync.WaitGroup{}
	errorChan := make(chan error, len(r.Object.Containers()))

	for i, container := range r.Object.Containers() {
		newName, isChanged, err := r.generateNewImageName(container.Image)
		if err != nil {
			return reconcile.Result{}, err
		}

		if isChanged {
			newCopy.OverrideImage(i, newName.Name())

			wg.Add(1)
			go func(imageName string) {
				defer wg.Done()

				source, _ := name.ParseReference(imageName)
				if source.Identifier() != `latest` {
					_, exist := r.Cloner.IsExistInClones(ctx, newName)
					if exist {
						return
					}
				}

				err := r.Cloner.Clone(ctx, source, newName)
				if err != nil {
					errorChan <- err
				}
			}(container.Image)
		}
	}
	wg.Wait()
	close(errorChan)

	var errs []string
	for err := range errorChan {
		errs = append(errs, err.Error())
	}
	if len(errs) > 0 {
		return reconcile.Result{}, fmt.Errorf("error clonning the docker images: %v", strings.Join(errs, " | "))
	}

	patchObject := client.StrategicMergeFrom(rs)

	// Patch data object
	// NOTE: if used Update instead of patch, it will conflict with the parallel changes and output errors
	err = r.Client.Patch(ctx, newCopy.Get(), patchObject) //err = r.client.Update(ctx, rs)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not write %s: %+v", r.Object.Name(), err)
	}

	logger.Info(`reconcile completed`)

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
// func (r *ImageCloneReconciler) SetupWithManager(mgr ctrl.Manager, ob objects.Object, config config.Config, cloner imagecloner.Cloner) error {
func (r *ImageCloneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New(r.Object.Name(), mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		r.Error(err, "unable to set up individual controller")
		return err
	}

	// Watch received Object and based on the given predicate
	err = c.Watch(&source.Kind{Type: r.Object.Get()}, &handler.EnqueueRequestForObject{},
		predicate.And(
			predicate.NewPredicateFuncs(func(object client.Object) bool {
				switch object.GetNamespace() {
				// TODO (navarrothiago) Get namespace from config file.
				case "kube-system", "kubernetes-dashboard", "image-clone-controller-system":
					return false
				}
				return true
			}),
			predicate.Funcs{
				DeleteFunc: func(event event.DeleteEvent) bool {
					return false
				},
				GenericFunc: func(event event.GenericEvent) bool {
					return false
				},
			},
		),
	)

	if err != nil {
		r.Error(err, "unable to watch "+r.Object.Get().GetName())
		return err
	}
	return nil
}

func (r *ImageCloneReconciler) generateNewImageName(imageName string) (n name.Reference, isChanged bool, err error) {
	source, err := name.ParseReference(imageName)
	if err != nil {
		return nil, false, err
	}
	r.Info(fmt.Sprintf("source image: %s", source.String()))
	if strings.Contains(source.String(), r.Cfg.DockerRegistry+"/"+r.Cfg.DockerUsername) {
		r.Info(fmt.Sprintf("source image already exists in %s/%s registry ", r.Cfg.DockerUsername, r.Cfg.DockerRegistry))
		return source, false, nil
	}

	target, err := name.ParseReference(r.Cfg.DockerUsername+"/"+strings.ReplaceAll(source.Context().RepositoryStr(), "/", "_")+":"+source.Identifier(), name.WithDefaultRegistry(r.Cfg.DockerRegistry))
	if err != nil {
		return nil, false, err
	}

	return target, true, nil
}
