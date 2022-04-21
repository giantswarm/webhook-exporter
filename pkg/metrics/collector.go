package metrics

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/go-logr/logr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
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
	k8sClient *kubernetes.Clientset,
	clientConfig admissionregistrationv1.WebhookClientConfig,
	namespaceSelector metav1.LabelSelector,
) error {
	webhookName := collector.Name
	serviceName := clientConfig.Service.Name
	namespace := clientConfig.Service.Namespace

	service, err := k8sClient.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	log.Info(fmt.Sprintf("Found the service named %s for the webhook %s", serviceName, webhookName))

	if err != nil {
		collector.setValueOfAllMetrics(0)

		return err
	}

	selector := labels.FormatLabels(service.Spec.Selector)

	deployments, err := k8sClient.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		log.Error(microerror.Mask(err), "Error fetching webhook deployement")
		return err
	}

	log.Info(fmt.Sprintf("Checking replicas number of replicas for %s", serviceName))
	collector.collectDeploymentMetrics(*deployments)

	log.Info(fmt.Sprintf("Checking the pod disruption budget for %s", serviceName))
	pdbs, err := k8sClient.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		log.Error(err, fmt.Sprintf("Error getting the pod disruption bugdet for %s", webhookName))
		return err
	}

	collector.collectPDBMetrics(pdbs)

	log.Info(fmt.Sprintf("Checking the namespace selector for %s", webhookName))
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
