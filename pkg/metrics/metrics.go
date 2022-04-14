package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	metricNamespace = "webhook_exporter"
	metricSubsystem = "webhooks"
)

var (
	infoLabels = []string{"webhook", "webhook_type"}

	ReplicasInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "replicas",
			Help:      "A metric showing the number of replicas a webhook has",
		},
		infoLabels,
	)

	PodDisruptionBudgetInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "pod_disruption_budget",
			Help:      "A metric showing the minimum avalaible value in the pdb for a webhook",
		},
		infoLabels,
	)
)

var (
	ValidNamespaceSelectors = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "valid_namespace_selectors",
			Help: `
			   A metric showing the number of webhooks that a valid.
			   That is webhooks that follow the folloing rules:
			     namespaceSelector:
			       matchExpressions:
			       - key: name
			         operator: NotIn
			         values: ["kube-system"]
			   `,
		},
		infoLabels,
	)
)

func init() {
	metrics.Registry.MustRegister(
		ReplicasInfo,
		PodDisruptionBudgetInfo,
		ValidNamespaceSelectors,
	)
}
