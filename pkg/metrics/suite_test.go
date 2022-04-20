package metrics_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics Test")
}

// var _ = BeforeSuite(func() {
// 	Expect(os.Setenv("TEST_ASSET_KUBE_APISERVER", "../testbin/bin/kube-apiserver")).To(Succeed())
// 	Expect(os.Setenv("TEST_ASSET_ETCD", "../testbin/bin/etcd")).To(Succeed())
// 	Expect(os.Setenv("TEST_ASSET_KUBECTL", "../testbin/bin/kubectl")).To(Succeed())
// })

// var _ = AfterSuite(func() {
//     Expect(os.Unsetenv("TEST_ASSET_KUBE_APISERVER")).To(Succeed())
//     Expect(os.Unsetenv("TEST_ASSET_ETCD")).To(Succeed())
//     Expect(os.Unsetenv("TEST_ASSET_KUBECTL")).To(Succeed())
// })
