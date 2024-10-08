// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package admission

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	greenhousev1alpha1 "github.com/cloudoperators/greenhouse/pkg/apis/greenhouse/v1alpha1"
	"github.com/cloudoperators/greenhouse/pkg/test"
)

func TestWebhooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Suite")
}

var _ = BeforeSuite(func() {
	test.RegisterWebhook("pluginDefinitionWebhook", SetupPluginDefinitionWebhookWithManager)
	test.RegisterWebhook("pluginWebhook", SetupPluginWebhookWithManager)
	test.RegisterWebhook("pluginPresetWebhook", SetupPluginPresetWebhookWithManager)
	test.RegisterWebhook("clusterValidation", SetupClusterWebhookWithManager)
	test.RegisterWebhook("secretsWebhook", SetupSecretWebhookWithManager)
	test.RegisterWebhook("teamsWebhook", SetupTeamWebhookWithManager)
	test.RegisterWebhook("roleWebhook", SetupTeamRoleWebhookWithManager)
	test.RegisterWebhook("rolebindingWebhook", setupRoleBindingWebhookForTest)
	test.TestBeforeSuite()

	setupOrgResources()
})

var _ = AfterSuite(func() {
	test.TestAfterSuite()
})

const (
	testclustername = "test-cluster"
)

var (
	testcluster = &greenhousev1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testclustername,
			Namespace: test.TestNamespace,
		},
		Spec: greenhousev1alpha1.ClusterSpec{
			AccessMode: greenhousev1alpha1.ClusterAccessModeDirect,
		},
	}
)

// setupOrgResources creates necessary static org resources for the tests
func setupOrgResources() {
	Expect(test.K8sClient.Create(test.Ctx, testcluster)).To(Succeed())
}
