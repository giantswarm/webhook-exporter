package metrics

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Collector struct {
	Name string
	Kind string
}

const (
	kubeSystemNamespace = "kube-system"
	giantswarmNamespace = "giantswarm"
)

func (collector Collector) CollectWebhookMetrics(
	ctx context.Context,
	log logr.Logger,
	k8sClient client.Client,
	clientConfig admissionregistrationv1.WebhookClientConfig,
	namespaceSelector metav1.LabelSelector,
) error {
	webhookName := collector.Name
	serviceName := clientConfig.Service.Name
	namespace := clientConfig.Service.Namespace

	service := &corev1.Service{}
	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      serviceName,
	}, service)
	log.Info("Found the webhook service", "webhook", webhookName, "service", serviceName)

	if err != nil {
		collector.setValueOfAllMetrics(0)

		return err
	}

	var selector client.MatchingLabels = service.Spec.Selector
	opts := []client.ListOption{
		selector,
	}

	deployments := &appsv1.DeploymentList{}
	err = k8sClient.List(ctx, deployments, opts...)
	if err != nil {
		log.Error(microerror.Mask(err), "Error fetching webhook deployment", "webhook", webhookName)
		return err
	}

	log.Info("Checking number of replicas", "webhook", webhookName)
	collector.collectDeploymentMetrics(*deployments)

	log.Info("Checking pod disruption budget", "webhook", webhookName)
	pdbs := &policyv1.PodDisruptionBudgetList{}
	if err = k8sClient.List(ctx, pdbs, opts...); err != nil {
		log.Error(err, "Error fetching pod disruption bugdet", "webhook", webhookName)
		return err
	}

	collector.collectPDBMetrics(pdbs)

	log.Info("Checking the namespace selector", "webhook", webhookName)
	collector.collectNamespaceSelectorMetrics(&namespaceSelector)

	return nil
}

func (c Collector) collectNamespaceSelectorMetrics(namespaceSelector *metav1.LabelSelector) {
	ValidNamespaceSelectors.WithLabelValues(c.Name, c.Kind).Set(0)

	if namespaceSelector == nil {
		return
	}

	selectors := namespaceSelector.MatchExpressions

	for _, selector := range selectors {
		if hasValidNamespaceSelector(selector) {
			ValidNamespaceSelectors.WithLabelValues(c.Name, c.Kind).Set(1)

			return
		}
	}

}

func hasValidNamespaceSelector(selector metav1.LabelSelectorRequirement) bool {
	return selector.Key == "name" &&
		selector.Operator == metav1.LabelSelectorOpNotIn &&
		contains(selector.Values, kubeSystemNamespace) &&
		contains(selector.Values, giantswarmNamespace)
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}

	return false
}

func (c Collector) collectDeploymentMetrics(deployments appsv1.DeploymentList) {
	var replicas float64 = 0

	for _, deployment := range deployments.Items {
		replicas = replicas + float64(*deployment.Spec.Replicas)
	}

	ReplicasInfo.WithLabelValues(c.Name, c.Kind).Set(replicas)
}

func (c Collector) collectPDBMetrics(pdbs *policyv1.PodDisruptionBudgetList) {
	var minAvailablePods float64 = 0

	PodDisruptionBudgetInfo.WithLabelValues(c.Name, c.Kind).Set(minAvailablePods)

	for _, pdb := range pdbs.Items {
		if pdb.Spec.MinAvailable != nil {
			PodDisruptionBudgetInfo.
				WithLabelValues(c.Name, c.Kind).
				Set(float64(pdb.Spec.MinAvailable.IntVal))

			continue
		}

		PodDisruptionBudgetInfo.WithLabelValues(c.Name, c.Kind).Set(minAvailablePods)
	}
}

func (c Collector) setValueOfAllMetrics(value float64) {
	ValidNamespaceSelectors.WithLabelValues(c.Name, c.Kind).Set(value)
	PodDisruptionBudgetInfo.WithLabelValues(c.Name, c.Kind).Set(value)
	ReplicasInfo.WithLabelValues(c.Name, c.Kind).Set(value)
}
