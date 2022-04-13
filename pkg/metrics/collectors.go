package metrics

import (
	appsv1 "k8s.io/api/apps/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Collector struct {
	Name string
	Kind string
}

const (
	kubeSystemNamespace = "kube-system"
	giantswarmNamespace = "giantswarm"
)

func (c Collector) CollectNamespaceSelectorMetrics(namespaceSelector *metav1.LabelSelector) {
	ValidNamespaceSelectors.WithLabelValues(c.Name, c.Kind).Set(0)

	if namespaceSelector == nil {
		return
	}

	selectors := namespaceSelector.MatchExpressions

	for _, selector := range selectors {
		if hasValidNamespaceSelector(selector) {
			ValidNamespaceSelectors.
				WithLabelValues(c.Name, c.Kind).Set(1)

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

func (c Collector) CollectDeploymentMetrics(deployments appsv1.DeploymentList) {
	var replicas float64 = 0

	for _, deployment := range deployments.Items {
		replicas = replicas + float64(*deployment.Spec.Replicas)
	}

	ReplicasInfo.
		WithLabelValues(c.Name, c.Kind).
		Set(replicas)
}

func (c Collector) CollectPDBMetrics(pdbs *policyv1.PodDisruptionBudgetList) {
	var minAvailablePods float64 = 0

	for _, pdb := range pdbs.Items {
		if pdb.Spec.MinAvailable != nil {
			c.setPBDMetricValue(float64(pdb.Spec.MinAvailable.IntVal))
			continue
		}

		c.setPBDMetricValue(minAvailablePods)
	}
}

func (c Collector) setPBDMetricValue(value float64) {
	PodDisruptionBudgetInfo.
		WithLabelValues(c.Name, c.Kind).
		Set(value)
}
