package metrics_test

import (
	"github.com/giantswarm/webhook-exporter/pkg/metrics"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

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
		Context("Collection of metrics", func() {
			It("should have 3 metrics for replicas Info", func() {

				replicasMetrics.WithLabelValues("test.giantswarm.webhook.one", "test").Inc()
				replicasMetrics.WithLabelValues("test.giantswarm.webhook.two", "test").Inc()
				replicasMetrics.WithLabelValues("test.giantswarm.webhook.three", "test").Inc()

				Expect(3).To(Equal(testutil.CollectAndCount(replicasMetrics)))
			})
		})

		Context("for replicas", func() {
			It("Should be 10", func() {
				var replicas float64 = 10

				replicasMetrics.WithLabelValues("test.giantswarm.webhook.one", "test").Set(replicas)
				Expect(float64(replicas)).To(Equal(testutil.ToFloat64(replicasMetrics.WithLabelValues("test.giantswarm.webhook.one", "test"))))
			})
		})

		Context("Collection of pod disruption metrics", func() {
			It("should have three metrics for pod disruption budgets", func() {

				pdbMetrics.WithLabelValues("test.giantswarm.webhook.one", "test").Inc()
				pdbMetrics.WithLabelValues("test.giantswarm.webhook.two", "test").Inc()
				pdbMetrics.WithLabelValues("test.giantswarm.webhook.three", "test").Inc()

				Expect(3).To(Equal(testutil.CollectAndCount(pdbMetrics)))
			})
		})

		Context("Collection of pod disruption metric values", func() {
			It("Should have a metric of value three", func() {
				var pdbs float64 = 3

				pdbMetrics.WithLabelValues("test.giantswarm.webhook", "test").Set(pdbs)
				Expect(float64(pdbs)).To(Equal(testutil.ToFloat64(pdbMetrics.WithLabelValues("test.giantswarm.webhook", "test"))))
			})
		})

		Context("Collection of metrics for valid namespaces", func() {
			It("should export metrics for three different webhooks", func() {
				validNamespaceSelectorsMetrics.WithLabelValues("test.giantswarm.webhook.one", "test").Inc()
				validNamespaceSelectorsMetrics.WithLabelValues("test.giantswarm.webhook.two", "test").Inc()
				validNamespaceSelectorsMetrics.WithLabelValues("test.giantswarm.webhook.three", "test").Inc()

				Expect(3).To(Equal(testutil.CollectAndCount(validNamespaceSelectorsMetrics)))
			})
		})

	})

})
