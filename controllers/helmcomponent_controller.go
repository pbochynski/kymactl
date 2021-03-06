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
	"fmt"
	"time"

	"golang.org/x/time/rate"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"

	inventoryv1alpha1 "github.com/kyma-incubator/kymactl/api/v1alpha1"
	"github.com/kyma-incubator/kymactl/manifests"
	"github.com/kyma-incubator/kymactl/pkg/helm"
)

// HelmComponentReconciler reconciles a HelmComponent object
type HelmComponentReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	manifests map[string]string
}

//+kubebuilder:rbac:groups=inventory.kyma-project.io,resources=helmcomponents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inventory.kyma-project.io,resources=helmcomponents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inventory.kyma-project.io,resources=helmcomponents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HelmComponent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *HelmComponentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var helmComponent inventoryv1alpha1.HelmComponent
	log.V(2).Info("Helm reconciliation started")
	if err := r.Get(ctx, req.NamespacedName, &helmComponent); err != nil {
		log.Info("unable to fetch HelmComponent")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	prevStatus := helmComponent.Status.Status
	requeue := time.Duration(len(helmComponent.Spec.ComponentName)) * time.Second
	switch prevStatus {
	case "pending":
		helmComponent.Status.Status = "started"
	case "started":
		helmComponent.Status.Status = "failing"
	case "failing":
		helmComponent.Status.Status = "retrying"
	case "retrying":
		helmComponent.Status.Status = "success"
	case "success":
		requeue = 0 * time.Second
	default:
		helmComponent.Status.Status = "pending"
		requeue = 1 * time.Second
	}

	log.V(2).Info("Reconciliation", "status", helmComponent.Status.Status, "requeue", requeue)
	if helmComponent.Status.Status != prevStatus {
		manifest := r.manifests[helmComponent.Spec.ComponentName]
		if manifest == "" {
			renderer := helm.NewGenericRenderer(manifests.FS, "charts/"+helmComponent.Spec.ComponentName, helmComponent.Spec.ComponentName, helmComponent.Spec.Namespace)
			renderer.Run()
			manifest, err := renderer.RenderManifest("")
			if err != nil {
				log.Error(fmt.Errorf("Rendering error"), "Cannot render chart")
			}
			log.Info("New manifest rendered")
			r.manifests[helmComponent.Spec.ComponentName] = manifest
		}
		if err := r.Status().Update(ctx, &helmComponent); err != nil {
			return ctrl.Result{}, err
		}
	}
	if requeue > 0*time.Second {
		return ctrl.Result{RequeueAfter: requeue}, nil
	}
	return ctrl.Result{}, nil
}

func CustomRateLimiter() ratelimiter.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(1*time.Second, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(150), 200)})
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelmComponentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.manifests = make(map[string]string)
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.HelmComponent{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10, RateLimiter: CustomRateLimiter()}).
		Complete(r)
}
