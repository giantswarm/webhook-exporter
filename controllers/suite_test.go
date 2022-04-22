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
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

const (
	webhookName string = "test.giantswarm.webhook"

	serviceName      string = "test-giantswarm-webhook"
	minAvailablePods int    = 2
)

var (
	testEnv *envtest.Environment
	ctx     context.Context
	cancel  context.CancelFunc

	timeout  = time.Second * 20
	interval = time.Millisecond * 250

	replicas int32 = 3
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Tests")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{})
	Expect(err).NotTo(HaveOccurred(), "failed to create manager")

	ctx, cancel = context.WithCancel(context.TODO())
	k8sClient, err := kubernetes.NewForConfig(cfg)

	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	err = createMutatingWebhook(ctx, k8sClient)
	Expect(err).NotTo(HaveOccurred())

	err = createValidatingWebhook(ctx, k8sClient)
	Expect(err).NotTo(HaveOccurred())

	err = createService(ctx, k8sClient)
	Expect(err).NotTo(HaveOccurred())

	err = createDeployment(ctx, k8sClient)
	Expect(err).NotTo(HaveOccurred())

	err = createPDB(ctx, k8sClient)
	Expect(err).NotTo(HaveOccurred())

	mutatingController := &MutatingWebhookConfigurationReconciler{
		Client: mgr.GetClient(),
		Log:    logf.Log,

		K8sClient: k8sClient,
	}

	validatingController := &ValidatingWebhookConfigurationReconciler{
		Client: mgr.GetClient(),
		Log:    logf.Log,

		K8sClient: k8sClient,
	}

	err = mutatingController.SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred(), "failed to setup mutating controller")

	err = validatingController.SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred(), "failed to setup validating controller")

	go func() {
		err := mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred(), "failed to start manager")
	}()
})

var _ = AfterSuite(func() {
	cancel()
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func getMutatingWebhook() *v1.MutatingWebhookConfiguration {
	var port int32 = 3500
	var sideEffects = v1.SideEffectClassNone
	expression := getMatchExpressions()
	webhook := v1.MutatingWebhook{
		Name: webhookName,
		ClientConfig: v1.WebhookClientConfig{
			Service: &v1.ServiceReference{
				Namespace: corev1.NamespaceDefault,
				Name:      serviceName,
				Port:      &port,
			}, CABundle: []byte{}},
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{},
			MatchExpressions: []metav1.LabelSelectorRequirement{
				expression,
			},
		},
		ObjectSelector:          &metav1.LabelSelector{},
		SideEffects:             &sideEffects,
		AdmissionReviewVersions: []string{"v1", "v1beta1"},
	}

	return &v1.MutatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:   webhookName,
			Labels: map[string]string{},
		},
		Webhooks: []v1.MutatingWebhook{
			webhook,
		},
	}
}

func getValidatingWebhook() *v1.ValidatingWebhookConfiguration {
	var port int32 = 3500
	var sideEffects = v1.SideEffectClassNone
	expression := getMatchExpressions()
	webhook := v1.ValidatingWebhook{
		Name: webhookName,
		ClientConfig: v1.WebhookClientConfig{
			Service: &v1.ServiceReference{
				Namespace: corev1.NamespaceDefault,
				Name:      serviceName,
				Port:      &port,
			}, CABundle: []byte{}},
		NamespaceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{},
			MatchExpressions: []metav1.LabelSelectorRequirement{
				expression,
			},
		},
		ObjectSelector:          &metav1.LabelSelector{},
		SideEffects:             &sideEffects,
		AdmissionReviewVersions: []string{"v1", "v1beta1"},
	}

	return &v1.ValidatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:   webhookName,
			Labels: map[string]string{},
		},
		Webhooks: []v1.ValidatingWebhook{
			webhook,
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

func getService() *corev1.Service {
	port := corev1.ServicePort{
		Name:     serviceName,
		Protocol: "TCP",
		Port:     3500,
		TargetPort: intstr.IntOrString{
			Type:   0,
			IntVal: 3500,
			StrVal: "3500",
		},
		NodePort: 0,
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				port,
			},
			Selector: map[string]string{
				"app": "test",
			},
		},
	}
}

func getPDB() *policyv1.PodDisruptionBudget {
	minAvailable := intstr.FromInt(minAvailablePods)

	return &policyv1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
			Labels: map[string]string{
				"app": "test",
			},
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
		},
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
			Labels: map[string]string{
				"app": "test",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 3500,
								},
							},
						},
					},
					RestartPolicy: "Always",
				},
			},
		},
	}
}

func createMutatingWebhook(ctx context.Context, k8sClient *kubernetes.Clientset) error {
	webhookClient := k8sClient.AdmissionregistrationV1().MutatingWebhookConfigurations()

	webhook := getMutatingWebhook()

	_, err := webhookClient.Create(ctx, webhook, metav1.CreateOptions{})

	return err
}

func createValidatingWebhook(ctx context.Context, k8sClient *kubernetes.Clientset) error {
	webhookClient := k8sClient.AdmissionregistrationV1().ValidatingWebhookConfigurations()

	webhook := getValidatingWebhook()

	_, err := webhookClient.Create(ctx, webhook, metav1.CreateOptions{})

	return err
}

func createService(ctx context.Context, k8sClient *kubernetes.Clientset) error {
	serviceClient := k8sClient.CoreV1().Services(corev1.NamespaceDefault)

	service := getService()
	_, err := serviceClient.Create(ctx, service, metav1.CreateOptions{})

	return err
}

func createDeployment(ctx context.Context, k8sClient *kubernetes.Clientset) error {
	deploymentsClient := k8sClient.AppsV1().Deployments(corev1.NamespaceDefault)
	deployment := getDeployment()

	_, err := deploymentsClient.Create(ctx, deployment, metav1.CreateOptions{})

	return err
}

func createPDB(ctx context.Context, k8sClient *kubernetes.Clientset) error {
	pdbClient := k8sClient.PolicyV1().PodDisruptionBudgets(corev1.NamespaceDefault)

	pdb := getPDB()
	_, err := pdbClient.Create(ctx, pdb, metav1.CreateOptions{})

	return err
}
