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
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/webhook-exporter/pkg/metrics"
)

const MutatingWebhookExporterType = "mutating"

// MutatingWebhookConfigurationReconciler reconciles a MutatingWebhookConfiguration object
type MutatingWebhookConfigurationReconciler struct {
	client.Client
	Log logr.Logger
}

//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations/finalizers,verbs=update

//On reconcile kubernetes attempts to bring clusters actual state to desired state.
// When this is triggered, this function collects multiple metrics about the MutatingWebhookConfiguration being reconciled.
// It doesn't do any manipulation on the resource itself.

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *MutatingWebhookConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("webhook", req.NamespacedName)

	configuration := &admissionregistrationv1.MutatingWebhookConfiguration{}

	err := r.Client.Get(ctx, req.NamespacedName, configuration)
	if apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	for _, webhook := range configuration.Webhooks {
		collector := metrics.Collector{
			Name: webhook.Name,
			Kind: MutatingWebhookExporterType,
		}

		err = collector.CollectWebhookMetrics(
			ctx,
			log,
			r.Client,
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
func (r *MutatingWebhookConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&admissionregistrationv1.MutatingWebhookConfiguration{}).
		Complete(r)
}
