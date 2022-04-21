package controllers

import (
	"fmt"
	"time"

	"github.com/giantswarm/webhook-exporter/pkg/metrics"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

const (
	webhookName string = "test.giantswarm.webhook"
	webhookKind string = MutatingWebhookExporterType

	serviceName      string = "test-giantswarm-webhook"
	minAvailablePods int    = 2
)

var replicas int32 = 3
var (
	timeout  = time.Second * 20
	duration = time.Second * 10
	interval = time.Millisecond * 250
)

var _ = Context("MutatatingWebhookConfiguration Controller", func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	Describe("Is Reconciled", func() {
		It(fmt.Sprintf("should have a deployment with %d replicas", replicas), func() {
			Eventually(func() int32 {
				found := testutil.ToFloat64(metrics.ReplicasInfo.WithLabelValues(webhookName, webhookKind))

				return int32(found)
			}, timeout, interval).Should(Equal(replicas))
		})

		It(fmt.Sprintf("should have a pod disruption budget with a minimum of %d available", minAvailablePods), func() {
			Eventually(func() int {
				found := testutil.ToFloat64(metrics.PodDisruptionBudgetInfo.WithLabelValues(webhookName, webhookKind))

				return int(found)
			}, timeout, interval).Should(Equal(minAvailablePods))
		})

		It("should have a valid namespace selector", func() {
			Eventually(func() int {
				found := testutil.ToFloat64(metrics.ValidNamespaceSelectors.WithLabelValues(webhookName, webhookKind))

				return int(found)
			}, timeout, interval).Should(Equal(1))
		})

	})
})
