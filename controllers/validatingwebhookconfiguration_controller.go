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

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/webhook-exporter/pkg/metrics"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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
	logger := log.FromContext(ctx)

	fmt.Println(req.Name)
	webhookConfiguration := &admissionregistrationv1.ValidatingWebhookConfiguration{}

	err := r.Client.Get(ctx, req.NamespacedName, webhookConfiguration)
	if apierrors.IsNotFound(err) {
		//TODO: Set everything to 0

		return ctrl.Result{}, err
	}

	for _, validationWebhook := range webhookConfiguration.Webhooks {
		r.collectWebhookMetrics(ctx, logger, validationWebhook)
	}

	return ctrl.Result{}, nil
}

//Probably a better name is scrapeWebhook
func (r *ValidatingWebhookConfigurationReconciler) collectWebhookMetrics(ctx context.Context, logger logr.Logger, validationWebhook admissionregistrationv1.ValidatingWebhook) error {
  webhookName := validationWebhook.Name
	clientConfig := validationWebhook.ClientConfig
	serviceName := clientConfig.Service.Name
	namespace := clientConfig.Service.Namespace

	service, err := r.K8sClient.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})

  logger.Info(fmt.Sprintf("Found the service named %s for the webhook %s", serviceName, webhookName))
	//TODO: Handle what happens if a service isn't found. This is likely to be the case when the webhook lives outside the cluster. Zero out every metric
	if err != nil {
		return err
	}

	selector := labels.FormatLabels(service.Spec.Selector)

	deployments, err := r.K8sClient.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		logger.Error(microerror.Mask(err), "Error fetching webhook deployement")
	}

	//TODO: Handle the cases where there are no deployments

	//TODO: Handle cases where we have multiple deployements. If that's possible
	logger.Info(fmt.Sprintf("Checking replicas number of replicas for %s", serviceName))
	collectDeploymentInfo(webhookName, *deployments)

	//Pod Disruption Budget
	logger.Info(fmt.Sprintf("Checking the pod disruption budget for %s", serviceName))
	pdbs, err := r.K8sClient.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		logger.Error(err, fmt.Sprintf("Error getting the pod disruption bugdet for %s", webhookName))
	}

	collectPDBInfo(webhookName, *pdbs)

	return nil
}

func collectDeploymentInfo(webhookName string, deployments appsv1.DeploymentList) {
	var replicas float64 = 0

	if len(deployments.Items) > 0 {
		deployment := deployments.Items[0]
		replicas = float64(*deployment.Spec.Replicas)
	}

	metrics.ReplicasInfo.
		WithLabelValues(webhookName, "Validation Webhook").
		Set(replicas)
}

func collectPDBInfo(webhookName string, pdbs policyv1.PodDisruptionBudgetList) {
	var minAvailablePods float64 = 0

	if len(pdbs.Items) > 0 {
		pdb := pdbs.Items[0]
		minAvailablePods = float64(pdb.Spec.MinAvailable.IntVal)
	}

	metrics.PodDisruptionBudgetInfo.
		WithLabelValues(webhookName, "Validation Webhook").
		Set(minAvailablePods)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ValidatingWebhookConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&admissionregistrationv1.ValidatingWebhookConfiguration{}).
		Complete(r)
}
