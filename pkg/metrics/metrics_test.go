package metrics_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/giantswarm/webhook-exporter/pkg/metrics"
)

const kind = "test"

var _ = Describe("Metrics", func() {
	var replicasMetrics, pdbMetrics, validNamespaceSelectorsMetrics *prometheus.GaugeVec

	BeforeEach(func() {
		replicasMetrics = metrics.ReplicasInfo
		pdbMetrics = metrics.PodDisruptionBudgetInfo
		validNamespaceSelectorsMetrics = metrics.ValidNamespaceSelectors

		replicasMetrics.Reset()
		pdbMetrics.Reset()
		validNamespaceSelectorsMetrics.Reset()
	})

	Describe("Exporting metrics", func() {
		Context("Collection of metrics from 3 different webhooks", func() {

			It("should have 3 metrics for replicas Info from the different webhooks", func() {
				replicasMetrics.WithLabelValues("test.giantswarm.webhook.one", kind).Inc()
				replicasMetrics.WithLabelValues("test.giantswarm.webhook.two", kind).Inc()
				replicasMetrics.WithLabelValues("test.giantswarm.webhook.three", kind).Inc()

				Expect(3).To(Equal(testutil.CollectAndCount(replicasMetrics)))
			})
		})

		Context("for replicas", func() {
			It("Should have a metric for webhook with ten replicas", func() {
				var replicas float64 = 10
				replicasMetrics.WithLabelValues("test.giantswarm.webhook.one", kind).Set(replicas)

				Expect(float64(replicas)).To(Equal(testutil.ToFloat64(replicasMetrics.WithLabelValues("test.giantswarm.webhook.one", kind))))
			})
		})

		Context("Collection of pod disruption metrics for three different webhooks", func() {
			It("should have three metrics for pod disruption budgets", func() {
				pdbMetrics.WithLabelValues("test.giantswarm.webhook.one", kind).Inc()
				pdbMetrics.WithLabelValues("test.giantswarm.webhook.two", kind).Inc()
				pdbMetrics.WithLabelValues("test.giantswarm.webhook.three", kind).Inc()

				Expect(3).To(Equal(testutil.CollectAndCount(pdbMetrics)))
			})
		})

		Context("Collection of pod disruption metric values for a webhook", func() {
			It("Should have a metric of with value of three", func() {
				var pdbs float64 = 3
				pdbMetrics.WithLabelValues("test.giantswarm.webhook", kind).Set(pdbs)

				Expect(float64(pdbs)).To(Equal(testutil.ToFloat64(pdbMetrics.WithLabelValues("test.giantswarm.webhook", kind))))
			})
		})

		Context("Collection of metrics for valid namespaces", func() {
			It("should export metrics for three different webhooks", func() {
				validNamespaceSelectorsMetrics.WithLabelValues("test.giantswarm.webhook.one", kind).Inc()
				validNamespaceSelectorsMetrics.WithLabelValues("test.giantswarm.webhook.two", kind).Inc()
				validNamespaceSelectorsMetrics.WithLabelValues("test.giantswarm.webhook.three", kind).Inc()

				Expect(3).To(Equal(testutil.CollectAndCount(validNamespaceSelectorsMetrics)))
			})
		})

	})

})
