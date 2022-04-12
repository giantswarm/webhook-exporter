package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

//Need a metric for the number of webhooks with not enough replicas
//Need a metric for the number of webhooks with no good pdbs
//Namespace selector
//Operator selector

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

func init() {
	prometheus.MustRegister(
		ReplicasInfo,
		PodDisruptionBudgetInfo,
	)
}
