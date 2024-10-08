// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	greenhouseapis "github.com/cloudoperators/greenhouse/pkg/apis"
	greenhousev1alpha1 "github.com/cloudoperators/greenhouse/pkg/apis/greenhouse/v1alpha1"
	"github.com/cloudoperators/greenhouse/pkg/clientutil"
	"github.com/cloudoperators/greenhouse/pkg/test"
)

const (
	pluginPresetName           = "test-pluginpreset"
	pluginPresetDefinitionName = "preset-plugindefinition"

	releaseNamespace = "test-namespace"

	clusterA = "cluster-a"
	clusterB = "cluster-b"
	clusterC = "cluster-c"
)

var (
	pluginPresetRemoteKubeConfig []byte
	pluginPresetK8sClient        client.Client
	pluginPresetRemote           *envtest.Environment

	pluginPresetDefinition = &greenhousev1alpha1.PluginDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PluginDefinition",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pluginPresetDefinitionName,
		},
		Spec: greenhousev1alpha1.PluginDefinitionSpec{
			Description: "Testplugin",
			Version:     "1.0.0",
			HelmChart: &greenhousev1alpha1.HelmChartReference{
				Name:       "./../../test/fixtures/myChart",
				Repository: "dummy",
				Version:    "1.0.0",
			},
			Options: []greenhousev1alpha1.PluginOption{
				{
					Name:        "myRequiredOption",
					Description: "This is my required test plugin option",
					Required:    true,
					Type:        greenhousev1alpha1.PluginOptionTypeString,
				},
			},
		},
	}
)

var _ = Describe("PluginPreset Controller Lifecycle", Ordered, func() {
	BeforeAll(func() {
		By("creating a test PluginDefinition")
		err := test.K8sClient.Create(test.Ctx, pluginPresetDefinition)
		Expect(err).ToNot(HaveOccurred(), "failed to create test PluginDefinition")

		By("bootstrapping the remote cluster")
		_, pluginPresetK8sClient, pluginPresetRemote, pluginPresetRemoteKubeConfig = test.StartControlPlane("6886", false, false)

		// kubeConfigController ensures the namespace within the remote cluster -- we have to create it
		By("creating the namespace on the cluster")
		remoteRestClientGetter := clientutil.NewRestClientGetterFromBytes(pluginPresetRemoteKubeConfig, releaseNamespace, clientutil.WithPersistentConfig())
		remoteK8sClient, err := clientutil.NewK8sClientFromRestClientGetter(remoteRestClientGetter)
		Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating the k8s client")
		err = remoteK8sClient.Create(test.Ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: releaseNamespace}})
		Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating the namespace")

		By("creating two test clusters for the same remote environment")
		for _, clusterName := range []string{clusterA, clusterB} {
			err := test.K8sClient.Create(test.Ctx, cluster(clusterName))
			Expect(err).Should(Succeed(), "failed to create test cluster: "+clusterName)

			By("creating a secret with a valid kubeconfig for a remote cluster")
			secretObj := clusterSecret(clusterName)
			secretObj.Data = map[string][]byte{
				greenhouseapis.KubeConfigKey: pluginPresetRemoteKubeConfig,
			}
			Expect(test.K8sClient.Create(test.Ctx, secretObj)).Should(Succeed())
		}
	})

	AfterAll(func() {
		err := pluginPresetRemote.Stop()
		Expect(err).
			NotTo(HaveOccurred(), "there must be no error stopping the remote environment")
	})

	It("should reconcile a PluginPreset", func() {
		By("creating a PluginPreset")
		testPluginPreset := pluginPreset(pluginPresetName, clusterA)
		err := test.K8sClient.Create(test.Ctx, testPluginPreset)
		Expect(err).ToNot(HaveOccurred(), "failed to create test PluginPreset")

		By("ensuring a Plugin has been created")
		expPluginName := types.NamespacedName{Name: pluginPresetName + "-" + clusterA, Namespace: test.TestNamespace}
		expPlugin := &greenhousev1alpha1.Plugin{}
		Eventually(func() error {
			return test.K8sClient.Get(test.Ctx, expPluginName, expPlugin)
		}).Should(Succeed(), "the Plugin should be created")

		Expect(expPlugin.Labels).To(HaveKeyWithValue(greenhouseapis.LabelKeyPluginPreset, pluginPresetName), "the Plugin should be labeled as managed by the PluginPreset")

		By("modifying the Plugin and ensuring it is reconciled")
		_, err = clientutil.CreateOrPatch(test.Ctx, test.K8sClient, expPlugin, func() error {
			// add a new option value that is not specified by the PluginPreset
			opt := greenhousev1alpha1.PluginOptionValue{Name: "option1", Value: test.MustReturnJSONFor("value1")}
			expPlugin.Spec.OptionValues = append(expPlugin.Spec.OptionValues, opt)
			return nil
		})
		Expect(err).NotTo(HaveOccurred(), "failed to update Plugin")

		Eventually(func(g Gomega) {
			err := test.K8sClient.Get(test.Ctx, expPluginName, expPlugin)
			g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error getting Plugin")
			g.Expect(expPlugin.Spec.OptionValues).ToNot(ContainElement(greenhousev1alpha1.PluginOptionValue{Name: "option1", Value: test.MustReturnJSONFor("value1")}), "the Plugin should be reconciled")
		}).Should(Succeed(), "the Plugin should be reconciled")

		By("manually creating a Plugin with OwnerReference but cluster not matching the selector")
		pluginNotExp := plugin(clusterB, expPlugin.OwnerReferences)
		Expect(test.K8sClient.Create(test.Ctx, pluginNotExp)).Should(Succeed(), "failed to create test Plugin")

		Eventually(func(g Gomega) error {
			err := test.K8sClient.Get(test.Ctx, types.NamespacedName{Name: pluginNotExp.Name, Namespace: pluginNotExp.Namespace}, pluginNotExp)
			g.Expect(err).To(HaveOccurred(), "there should be an error getting the Plugin")
			return client.IgnoreNotFound(err)
		}).Should(Succeed(), "the Plugin should be deleted")

		By("removing the preset label from the Plugin")
		_, err = clientutil.CreateOrPatch(test.Ctx, test.K8sClient, expPlugin, func() error {
			delete(expPlugin.Labels, greenhouseapis.LabelKeyPluginPreset)
			return controllerutil.RemoveControllerReference(testPluginPreset, expPlugin, test.K8sClient.Scheme())
		})
		Expect(err).NotTo(HaveOccurred(), "failed to update Plugin")

		Eventually(func(g Gomega) bool {
			err := test.K8sClient.Get(test.Ctx, types.NamespacedName{Namespace: testPluginPreset.Namespace, Name: testPluginPreset.Name}, testPluginPreset)
			g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error getting PluginPreset")
			return testPluginPreset.Status.StatusConditions.GetConditionByType(greenhousev1alpha1.PluginSkippedCondition).IsTrue()
		}).Should(BeTrue(), "PluginPreset should have the SkippedCondition set to true")

		By("re-adding the preset label to the Plugin")
		_, err = clientutil.CreateOrPatch(test.Ctx, test.K8sClient, expPlugin, func() error {
			expPlugin.Labels[greenhouseapis.LabelKeyPluginPreset] = testPluginPreset.Name
			return controllerutil.SetControllerReference(testPluginPreset, expPlugin, test.K8sClient.Scheme())
		})
		Expect(err).NotTo(HaveOccurred(), "failed to update Plugin")

		By("deleting the PluginPreset to ensure all Plugins are deleted")
		Expect(test.K8sClient.Delete(test.Ctx, testPluginPreset)).Should(Succeed(), "failed to delete test PluginPreset")
		Eventually(func(g Gomega) error {
			err := test.K8sClient.Get(test.Ctx, expPluginName, pluginNotExp)
			g.Expect(err).To(HaveOccurred(), "there should be an error getting the Plugin")
			return client.IgnoreNotFound(err)
		}).Should(Succeed(), "the Plugin should be deleted")
	})

	It("should reconcile a PluginPreset on cluster changes", func() {
		By("creating a PluginPreset")
		testPluginPreset := pluginPreset(pluginPresetName, clusterA)
		err := test.K8sClient.Create(test.Ctx, testPluginPreset)
		Expect(err).ToNot(HaveOccurred(), "failed to create test PluginPreset")

		By("onboarding another cluster")
		err = test.K8sClient.Create(test.Ctx, cluster(clusterC))
		Expect(err).Should(Succeed(), "failed to create test cluster: "+clusterC)
		secretObj := clusterSecret(clusterC)
		secretObj.Data = map[string][]byte{
			greenhouseapis.KubeConfigKey: pluginPresetRemoteKubeConfig,
		}
		Expect(test.K8sClient.Create(test.Ctx, secretObj)).Should(Succeed())

		By("making clusterC match the clusterSelector")
		pluginList := &greenhousev1alpha1.PluginList{}
		Eventually(func(g Gomega) {
			err = test.K8sClient.List(test.Ctx, pluginList, client.MatchingLabels{greenhouseapis.LabelKeyPluginPreset: pluginPresetName})
			g.Expect(err).NotTo(HaveOccurred(), "failed to list Plugins")
			g.Expect(pluginList.Items).To(HaveLen(1), "there should be only one Plugin")
		}).Should(Succeed(), "there should be a Plugin created for the Preset")

		cluster := greenhousev1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterC,
				Namespace: test.TestNamespace,
			},
		}
		_, err = clientutil.CreateOrPatch(test.Ctx, test.K8sClient, &cluster, func() error {
			cluster.Labels = map[string]string{"cluster": clusterA}
			return nil
		})
		Expect(err).NotTo(HaveOccurred(), "failed to update Cluster labels")
		Eventually(func(g Gomega) {
			err = test.K8sClient.List(test.Ctx, pluginList, client.MatchingLabels{greenhouseapis.LabelKeyPluginPreset: pluginPresetName})
			g.Expect(err).NotTo(HaveOccurred(), "failed to list Plugins")
			g.Expect(pluginList.Items).To(HaveLen(2), "there should be two Plugins")
		}).Should(Succeed(), "the PluginPreset should have noticed the ClusterLabel change")

		By("deleting clusterC to ensure the Plugin is deleted")
		Expect(test.K8sClient.Delete(test.Ctx, &cluster)).To(Succeed(), "failed to delete ClusterC")

		Eventually(func(g Gomega) {
			err = test.K8sClient.List(test.Ctx, pluginList, client.MatchingLabels{greenhouseapis.LabelKeyPluginPreset: pluginPresetName})
			g.Expect(err).NotTo(HaveOccurred(), "failed to list Plugins")
			g.Expect(pluginList.Items).To(HaveLen(1), "there should be only one Plugin")
		}).Should(Succeed(), "the PluginPreset should have removed the Plugin for the deleted Cluster")

		By("deleting the PluginPreset to ensure all Plugins are deleted")
		Expect(test.K8sClient.Delete(test.Ctx, testPluginPreset)).Should(Succeed(), "failed to delete test PluginPreset")
		Eventually(func(g Gomega) {
			err = test.K8sClient.List(test.Ctx, pluginList, client.MatchingLabels{greenhouseapis.LabelKeyPluginPreset: pluginPresetName})
			g.Expect(err).NotTo(HaveOccurred(), "failed to list Plugins")
			g.Expect(pluginList.Items).To(BeEmpty(), "all plugins for the Preset should be deleted")
		}).Should(Succeed(), "plugins for the pluginPreset should be deleted")
	})

	It("should set the Status NotReady if ClusterSelector does not match", func() {
		// Create a PluginPreset with a ClusterSelector that does not match any cluster
		pluginPreset := pluginPreset("not-ready", "non-existing-cluster")
		Expect(test.K8sClient.Create(test.Ctx, pluginPreset)).Should(Succeed(), "failed to create test PluginPreset")

		Eventually(func(g Gomega) {
			err := test.K8sClient.Get(test.Ctx, types.NamespacedName{Name: "not-ready", Namespace: pluginPreset.Namespace}, pluginPreset)
			g.Expect(err).ShouldNot(HaveOccurred(), "unexpected error getting PluginPreset")
			g.Expect(pluginPreset.Status.StatusConditions.Conditions).NotTo(BeNil(), "the PluginPreset should have a StatusConditions")
			g.Expect(pluginPreset.Status.StatusConditions.GetConditionByType(greenhousev1alpha1.ClusterListEmpty).IsTrue()).Should(BeTrue(), "PluginPreset should have the ClusterListEmptyCondition set to true")
			g.Expect(pluginPreset.Status.StatusConditions.GetConditionByType(greenhousev1alpha1.ReadyCondition).IsFalse()).Should(BeTrue(), "PluginPreset should have the ReadyCondition set to false")
		}).Should(Succeed(), "the PluginPreset should be reconciled")
	})
})

// clusterSecret returns the secret for a cluster.
func clusterSecret(clusterName string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.GroupName,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-" + clusterName,
			Namespace: test.TestNamespace,
		},
		Type: greenhouseapis.SecretTypeKubeConfig,
	}
}

// cluster returns a cluster object with the given name.
func cluster(name string) *greenhousev1alpha1.Cluster {
	return &greenhousev1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: test.TestNamespace,
			Labels: map[string]string{
				"cluster": name,
				"foo":     "bar",
			},
		},
		Spec: greenhousev1alpha1.ClusterSpec{
			AccessMode: greenhousev1alpha1.ClusterAccessModeDirect,
		},
	}
}

func plugin(clusterName string, ownerRefs []metav1.OwnerReference) *greenhousev1alpha1.Plugin {
	return &greenhousev1alpha1.Plugin{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Plugin",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pluginPresetName + "-" + clusterName,
			Namespace: test.TestNamespace,
			Labels: map[string]string{
				greenhouseapis.LabelKeyPluginPreset: pluginPresetName,
			},
			OwnerReferences: ownerRefs,
		},
		Spec: greenhousev1alpha1.PluginSpec{
			ClusterName:      clusterB,
			PluginDefinition: pluginPresetDefinitionName,
			OptionValues: []greenhousev1alpha1.PluginOptionValue{
				{
					Name:  "myRequiredOption",
					Value: test.MustReturnJSONFor("myValue"),
				},
			},
		},
	}
}

func pluginPreset(name, selectorValue string) *greenhousev1alpha1.PluginPreset {
	return &greenhousev1alpha1.PluginPreset{
		TypeMeta: metav1.TypeMeta{
			Kind:       greenhousev1alpha1.PluginPresetKind,
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: test.TestNamespace,
		},
		Spec: greenhousev1alpha1.PluginPresetSpec{
			Plugin: greenhousev1alpha1.PluginSpec{
				PluginDefinition: pluginPresetDefinitionName,
				ReleaseNamespace: releaseNamespace,
				OptionValues: []greenhousev1alpha1.PluginOptionValue{
					{
						Name:  "myRequiredOption",
						Value: test.MustReturnJSONFor("myValue"),
					},
				},
			},
			ClusterSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cluster": selectorValue,
				},
			},
		},
	}
}
