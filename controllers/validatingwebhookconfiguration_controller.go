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

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/webhook-exporter/pkg/metrics"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const ValidatingWebhookExporterType = "validating"

// ValidatingWebhookConfigurationReconciler reconciles a ValidatingWebhookConfiguration object
type ValidatingWebhookConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Log       logr.Logger
	K8sClient *kubernetes.Clientset
}

//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ValidatingWebhookConfiguration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ValidatingWebhookConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("webhook", req.NamespacedName)

	configuration := &admissionregistrationv1.ValidatingWebhookConfiguration{}

	err := r.Client.Get(ctx, req.NamespacedName, configuration)
	if apierrors.IsNotFound(err) {
		//TODO: Set everything to 0
		//TODO: Handle errors and logging properly
		return ctrl.Result{}, err
	}

	for _, webhook := range configuration.Webhooks {
		collector := metrics.Collector{
			Name: webhook.Name,
			Kind: ValidatingWebhookExporterType,
		}

		err = collector.CollectWebhookMetrics(
			ctx,
			log,
			r.K8sClient,
			webhook.ClientConfig,
			*webhook.NamespaceSelector,
		)

		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{
		Requeue:      true,
		RequeueAfter: time.Minute * 5,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ValidatingWebhookConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&admissionregistrationv1.ValidatingWebhookConfiguration{}).
		Complete(r)
}
