package metrics

import (
	"context"
	"fmt"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus/testutil"
	v1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	webhookName string = "test.giantswarm.webhook"
	webhookKind string = "test"

	minAvailablePods int = 2
)

var replicas int32 = 3
var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
	}

})

var _ = Describe("Collector", func() {
	var collector Collector

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	BeforeEach(func() {
		collector = Collector{
			Name: webhookName,
			Kind: webhookKind,
		}

		cfg, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).NotTo(BeNil())

		k8sClient, err := kubernetes.NewForConfig(cfg)

		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient).NotTo(BeNil())
	})

	It(fmt.Sprintf("should have %d replicas", replicas), func() {
		deployments := getDeployments()

		collector.collectDeploymentMetrics(deployments)

		Expect(float64(replicas)).To(Equal(testutil.ToFloat64(ReplicasInfo.WithLabelValues(webhookName, webhookKind))))
	})

	It("should have two pods avialable minimum", func() {
		pdbs := getPdbs()

		collector.collectPDBMetrics(&pdbs)
		Expect(float64(minAvailablePods)).To(Equal(testutil.ToFloat64(PodDisruptionBudgetInfo.WithLabelValues(webhookName, webhookKind))))
	})

	It("should have one valid namespace selector", func() {
		selectors := getNamespaceSelectors()

		collector.collectNamespaceSelectorMetrics(&selectors)
		Expect(float64(1)).To(Equal(testutil.ToFloat64(ValidNamespaceSelectors.WithLabelValues(webhookName, webhookKind))))

	})

})

func getWebhook() v1.MutatingWebhook {
	return v1.MutatingWebhook{
		Name:                    "test.giantswarm.webhook",
		ClientConfig:            v1.WebhookClientConfig{URL: new(string), Service: &v1.ServiceReference{}, CABundle: []byte{}},
		Rules:                   []v1.RuleWithOperations{},
		NamespaceSelector:       &metav1.LabelSelector{},
		ObjectSelector:          &metav1.LabelSelector{},
		AdmissionReviewVersions: []string{},
	}
}

func getDeployments() appsv1.DeploymentList {
	deployment := getDeployment()

	return appsv1.DeploymentList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items:    []appsv1.Deployment{*deployment},
	}
}

func getDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/appsv1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webhookName,
			Namespace: corev1.NamespaceDefault,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{},
					RestartPolicy: "always",
				},
			},
		},
	}
}

func getPdbs() policyv1.PodDisruptionBudgetList {
	pdb := getPDB()

	return policyv1.PodDisruptionBudgetList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []policyv1.PodDisruptionBudget{
			pdb,
		},
	}
}

func getPDB() policyv1.PodDisruptionBudget {
	minAvailable := intstr.FromInt(minAvailablePods)

	return policyv1.PodDisruptionBudget{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
		},
	}
}

func getNamespaceSelectors() metav1.LabelSelector {
	expression := getMatchExpressions()
	return metav1.LabelSelector{
		MatchLabels: map[string]string{},
		MatchExpressions: []metav1.LabelSelectorRequirement{
			expression,
		},
	}
}

func getMatchExpressions() metav1.LabelSelectorRequirement {
	return metav1.LabelSelectorRequirement{
		Key:      "name",
		Operator: "NotIn",
		Values:   []string{"kube-system", "giantswarm"},
	}
}

func createDeployment(ctx context.Context, k8sclient kubernetes.Clientset) error {
	deploymentsClient := k8sclient.AppsV1().Deployments(corev1.NamespaceDefault)
	deployment := getDeployment()

	_, err := deploymentsClient.Create(ctx, deployment, metav1.CreateOptions{})

	return err
}

var _ = AfterSuite(func() {
	Expect(testEnv.Stop()).To(Succeed())
})
