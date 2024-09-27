// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	greenhouseapis "github.com/cloudoperators/greenhouse/pkg/apis"
	greenhousev1alpha1 "github.com/cloudoperators/greenhouse/pkg/apis/greenhouse/v1alpha1"
	"github.com/cloudoperators/greenhouse/pkg/clientutil"
	"github.com/cloudoperators/greenhouse/pkg/common"
	"github.com/cloudoperators/greenhouse/pkg/helm"
	"github.com/cloudoperators/greenhouse/pkg/test"
)

// Test environment.
var (
	remoteKubeConfig []byte
	remoteEnvTest    *envtest.Environment
	remoteK8sClient  client.Client
)

// Test stimuli.
var (
	testPlugin = &greenhousev1alpha1.Plugin{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Plugin",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-plugindefinition",
			Namespace: test.TestNamespace,
		},
		Spec: greenhousev1alpha1.PluginSpec{
			ClusterName:      "test-cluster",
			PluginDefinition: "test-plugindefinition",
			ReleaseNamespace: test.TestNamespace,
		},
	}

	testPluginwithSR = &greenhousev1alpha1.Plugin{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Plugin",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-plugin-secretref",
			Namespace: test.TestNamespace,
		},
		Spec: greenhousev1alpha1.PluginSpec{
			PluginDefinition: "test-plugindefinition",
			ClusterName:      "test-cluster",
			OptionValues: []greenhousev1alpha1.PluginOptionValue{
				{
					Name: "secretValue",
					ValueFrom: &greenhousev1alpha1.ValueFromSource{
						Secret: &greenhousev1alpha1.SecretKeyReference{
							Name: "test-secret",
							Key:  "test-key",
						},
					},
				},
			},
		},
	}

	// A PluginConfig in the central cluster, test namespace with a release in the remote cluster, made-up-namespace.
	testPluginInDifferentNamespace = &greenhousev1alpha1.Plugin{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-plugin-in-made-up-namespace",
			Namespace: test.TestNamespace,
		},
		Spec: greenhousev1alpha1.PluginSpec{
			PluginDefinition: testPluginDefinition.GetName(),
			ClusterName:      testCluster.GetName(),
			ReleaseNamespace: "made-up-namespace",
		},
	}

	testPluginWithCRDs = &greenhousev1alpha1.Plugin{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Plugin",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-plugin-crd",
			Namespace: test.TestNamespace,
		},
		Spec: greenhousev1alpha1.PluginSpec{
			ClusterName:      "test-cluster",
			PluginDefinition: "test-plugindefinition-crd",
			ReleaseNamespace: test.TestNamespace,
		},
	}

	testPluginWithExposedService = &greenhousev1alpha1.Plugin{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Plugin",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-plugin-exposed",
			Namespace: test.TestNamespace,
		},
		Spec: greenhousev1alpha1.PluginSpec{
			ClusterName:      "test-cluster",
			PluginDefinition: "test-plugindefinition-exposed",
			ReleaseNamespace: test.TestNamespace,
		},
	}

	testSecret = corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.GroupName,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: test.TestNamespace,
		},
		Data: map[string][]byte{
			"test-key": []byte("secret-value"),
		},
	}

	testPluginDefinition = &greenhousev1alpha1.PluginDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PluginDefinition",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-plugindefinition",
			Namespace: corev1.NamespaceDefault,
		},
		Spec: greenhousev1alpha1.PluginDefinitionSpec{
			Description: "Testplugin",
			Version:     "1.0.0",
			HelmChart: &greenhousev1alpha1.HelmChartReference{
				Name:       "./../../test/fixtures/myChart",
				Repository: "dummy",
				Version:    "1.0.0",
			},
		},
	}

	testPluginWithHelmChartCRDs = &greenhousev1alpha1.PluginDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: test.TestNamespace,
			Name:      "test-plugindefinition-crd",
		},
		Spec: greenhousev1alpha1.PluginDefinitionSpec{
			Version: "1.0.0",
			HelmChart: &greenhousev1alpha1.HelmChartReference{
				Name:       "./../../test/fixtures/myChartWithCRDs",
				Repository: "dummy",
				Version:    "1.0.0",
			},
		},
	}

	pluginDefinitionWithExposedService = &greenhousev1alpha1.PluginDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: test.TestNamespace,
			Name:      "test-plugindefinition-exposed",
		},
		Spec: greenhousev1alpha1.PluginDefinitionSpec{
			Version: "1.0.0",
			HelmChart: &greenhousev1alpha1.HelmChartReference{
				Name:       "./../../test/fixtures/chartWithExposedService",
				Repository: "dummy",
				Version:    "1.3.0",
			},
		},
	}

	testCluster = &greenhousev1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: greenhousev1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: test.TestNamespace,
		},
		Spec: greenhousev1alpha1.ClusterSpec{
			AccessMode: greenhousev1alpha1.ClusterAccessModeDirect,
		},
	}

	testClusterK8sSecret = corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: corev1.GroupName,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: test.TestNamespace,
		},
		Type: greenhouseapis.SecretTypeKubeConfig,
	}
)

// Tests
var _ = Describe("HelmController reconciliation", Ordered, func() {
	BeforeAll(func() {
		err := test.K8sClient.Create(test.Ctx, testPluginDefinition)
		Expect(err).ToNot(HaveOccurred(), "there should be no error creating the pluginDefinition")

		By("bootstrapping remote cluster")
		bootstrapRemoteCluster()

		By("creating a cluster")
		Expect(test.K8sClient.Create(test.Ctx, testCluster)).Should(Succeed(), "there should be no error creating the cluster resource")

		// kubeConfigController ensures the namespace within the remote cluster -- we have to create it
		By("creating the namespace on the cluster")
		remoteRestClientGetter := clientutil.NewRestClientGetterFromBytes(remoteKubeConfig, testPlugin.Spec.ReleaseNamespace, clientutil.WithPersistentConfig())
		remoteClient, err := clientutil.NewK8sClientFromRestClientGetter(remoteRestClientGetter)
		Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating the k8s client")
		err = remoteClient.Create(test.Ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testPlugin.Spec.ReleaseNamespace}})
		Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating the namespace")

		By("creating a secret with a valid kubeconfig for a remote cluster")
		testClusterK8sSecret.Data = map[string][]byte{
			greenhouseapis.KubeConfigKey: remoteKubeConfig,
		}
		Expect(test.K8sClient.Create(test.Ctx, &testClusterK8sSecret)).Should(Succeed())
	})

	AfterAll(func() {
		err := remoteEnvTest.Stop()
		Expect(err).
			NotTo(HaveOccurred(), "there must be no error stopping the remote environment")
	})

	It("should correctly handle the plugin on a referenced cluster", func() {
		remoteRestClientGetter := clientutil.NewRestClientGetterFromBytes(remoteKubeConfig, testPlugin.Spec.ReleaseNamespace, clientutil.WithPersistentConfig())

		By("creating a plugin referencing the cluster")
		testPlugin.Spec.ClusterName = "test-cluster"
		Expect(test.K8sClient.Create(test.Ctx, testPlugin)).Should(Succeed(), "there should be no error updating the plugin")

		By("checking the ClusterAccessReadyCondition on the plugin")
		Eventually(func(g Gomega) bool {
			err := test.K8sClient.Get(test.Ctx, types.NamespacedName{Name: testPlugin.Name, Namespace: testPlugin.Namespace}, testPlugin)
			g.Expect(err).ShouldNot(HaveOccurred(), "there should be no error getting the plugin")
			clusterAccessReadyCondition := testPlugin.Status.GetConditionByType(greenhousev1alpha1.ClusterAccessReadyCondition)
			readyCondition := testPlugin.Status.GetConditionByType(greenhousev1alpha1.ReadyCondition)
			g.Expect(clusterAccessReadyCondition).ToNot(BeNil(), "the ClusterAccessReadyCondition should not be nil")
			g.Expect(readyCondition).ToNot(BeNil(), "the ReadyCondition should not be nil")
			g.Expect(testPlugin.Status.GetConditionByType(greenhousev1alpha1.ClusterAccessReadyCondition).IsFalse()).Should(BeTrue(), "the ClusterAccessReadyCondition should be false")
			g.Expect(testPlugin.Status.GetConditionByType(greenhousev1alpha1.ReadyCondition).IsFalse()).Should(BeTrue(), "the ReadyCondition should be false")
			return true
		}).Should(BeTrue(), "the ClusterAccessReadyCondition should be false")

		By("setting the ready condition on the test-cluster")
		testCluster.Status.StatusConditions.SetConditions(greenhousev1alpha1.TrueCondition(greenhousev1alpha1.ReadyCondition, "", ""))
		Expect(test.K8sClient.Status().Update(test.Ctx, testCluster)).Should(Succeed(), "there should be no error updating the cluster resource")

		By("triggering setting a label on the plugin to trigger reconciliation")
		testPlugin.Labels = map[string]string{"test": "label"}
		Expect(test.K8sClient.Update(test.Ctx, testPlugin)).Should(Succeed(), "there should be no error updating the plugin")

		By("checking the ClusterAccessReadyCondition on the plugin")
		Eventually(func(g Gomega) bool {
			err := test.K8sClient.Get(test.Ctx, types.NamespacedName{Name: testPlugin.Name, Namespace: testPlugin.Namespace}, testPlugin)
			g.Expect(err).ShouldNot(HaveOccurred(), "there should be no error getting the plugin")
			g.Expect(testPlugin.Status.GetConditionByType(greenhousev1alpha1.ClusterAccessReadyCondition).IsTrue()).Should(BeTrue(), "the ClusterAccessReadyCondition should be true")
			g.Expect(testPlugin.Status.GetConditionByType(greenhousev1alpha1.ReadyCondition).IsTrue()).Should(BeTrue(), "the ReadyCondition should be true")
			return true
		}).Should(BeTrue(), "the ClusterAccessReadyCondition should be true")

		By("checking the helm releases deployed to the remote cluster")
		helmConfig, err := helm.ExportNewHelmAction(remoteRestClientGetter, testPlugin.Spec.ReleaseNamespace)
		Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating helm config")
		listAction := action.NewList(helmConfig)

		Eventually(func() []*release.Release {
			releases, err := listAction.Run()
			Expect(err).ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
			return releases
		}).Should(ContainElement(gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{"Name": Equal("test-plugindefinition")}))), "the helm release should be deployed to the remote cluster")

		By("updating the plugin")
		_, err = clientutil.CreateOrPatch(test.Ctx, test.K8sClient, testPlugin, func() error {
			// this value enables the template of another pod
			testPlugin.Spec.OptionValues = append(testPlugin.Spec.OptionValues, greenhousev1alpha1.PluginOptionValue{Name: "enabled", Value: test.MustReturnJSONFor("true")})
			return nil
		})
		Expect(err).ShouldNot(HaveOccurred(), "there should be no error updating the plugin")
		By("checking the resources deployed to the remote cluster")
		Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating the k8s client")
		podID := types.NamespacedName{Name: "alpine-flag", Namespace: test.TestNamespace}
		pod := &corev1.Pod{}
		Eventually(func(g Gomega) bool {
			err := remoteK8sClient.Get(test.Ctx, podID, pod)
			if err != nil {
				g.Expect(err).ShouldNot(HaveOccurred(), "there should be no error retrieving the pod")
				return false
			}
			return true
		}).Should(BeTrue(), "the pod should have been created on the remote cluster")

		By("deleting the plugin")
		Expect(test.K8sClient.Delete(test.Ctx, testPlugin)).Should(Succeed(), "there should be no error deleting the plugin")

		By("checking the helm releases deployed to the remote cluster")
		Eventually(func() []*release.Release {
			releases, err := listAction.Run()
			Expect(err).ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
			return releases
		}).Should(BeEmpty(), "the helm release should be deleted from the remote cluster")
	})

	It("should correctly handle the plugin on a referenced cluster with a secret reference", func() {
		remoteRestClientGetter := clientutil.NewRestClientGetterFromBytes(remoteKubeConfig, testPlugin.Spec.ReleaseNamespace, clientutil.WithPersistentConfig())

		By("creating a secret holding the OptionValue referenced by the Plugin")
		Expect(test.K8sClient.Create(test.Ctx, &testSecret)).Should(Succeed())

		By("creating a plugin referencing the cluster")
		testPluginwithSR.Spec.ClusterName = "test-cluster"
		Expect(test.K8sClient.Create(test.Ctx, testPluginwithSR)).Should(Succeed(), "there should be no error updating the plugin")

		By("checking the helm releases deployed to the remote cluster")
		helmConfig, err := helm.ExportNewHelmAction(remoteRestClientGetter, testPluginwithSR.Namespace)
		Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating helm config")
		listAction := action.NewList(helmConfig)

		Eventually(func(g Gomega) []*release.Release {
			releases, err := listAction.Run()
			g.Expect(err).ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
			return releases
		}).Should(ContainElement(
			gstruct.PointTo(
				gstruct.MatchFields(
					gstruct.IgnoreExtras, gstruct.Fields{
						"Name":   Equal("test-plugin-secretref"),
						"Config": gstruct.MatchKeys(gstruct.IgnoreExtras, gstruct.Keys{"secretValue": Equal("secret-value")})}))), "the helm release should be deployed to the remote cluster")

		By("deleting the plugin")
		Expect(test.K8sClient.Delete(test.Ctx, testPluginwithSR)).Should(Succeed(), "there should be no error deleting the plugin")

		By("checking the helm releases deployed to the remote cluster")
		Eventually(func(g Gomega) []*release.Release {
			releases, err := listAction.Run()
			g.Expect(err).ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
			return releases
		}).Should(BeEmpty(), "the helm release should be deleted from the remote cluster")
	})

	It("should correctly handle the plugin on a referenced cluster with a different namespace", func() {
		Expect(testPluginInDifferentNamespace.GetNamespace()).
			Should(Equal(test.TestNamespace), "the namespace should be the test namespace")
		Expect(testPluginInDifferentNamespace.Spec.ReleaseNamespace).
			Should(Equal("made-up-namespace"), "the release namespace should be the made-up-namespace")

		By("creating a plugin referencing the cluster")
		Expect(test.K8sClient.Create(test.Ctx, testPluginInDifferentNamespace)).
			Should(Succeed(), "there should be no error creating the plugin")

		By("checking the helm releases deployed to the remote cluster in a different namespace")
		remoteRestClientGetter := clientutil.NewRestClientGetterFromBytes(
			remoteKubeConfig, testPluginInDifferentNamespace.Spec.ReleaseNamespace, clientutil.WithPersistentConfig(),
		)
		helmConfig, err := helm.ExportNewHelmAction(remoteRestClientGetter, testPluginInDifferentNamespace.Spec.ReleaseNamespace)
		Expect(err).
			ShouldNot(HaveOccurred(), "there should be no error creating helm config")

		Eventually(func(g Gomega) string {
			release, err := action.NewGet(helmConfig).Run(testPluginInDifferentNamespace.GetName())
			g.Expect(err).ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
			return release.Namespace
		}).Should(
			Equal(testPluginInDifferentNamespace.Spec.ReleaseNamespace),
			"the helm release should be deployed to the remote cluster in a different namespace",
		)

		By("checking the pod template without explicit namespace is deployed to the releaseNamespace")
		podName := types.NamespacedName{Name: "alpine", Namespace: testPluginInDifferentNamespace.Spec.ReleaseNamespace}
		Eventually(func(g Gomega) {
			pod := &corev1.Pod{}
			err := remoteK8sClient.Get(test.Ctx, podName, pod)
			g.Expect(err).NotTo(HaveOccurred(), "there should be no error getting the pod")
		}).Should(
			Succeed(),
			"the pod template without explicit namespace should be deployed to the releaseNamespace",
		)

		By("deleting the plugin")
		Expect(test.K8sClient.Delete(test.Ctx, testPluginInDifferentNamespace)).
			Should(Succeed(), "there should be no error deleting the plugin")

		By("checking the helm releases deployed to the remote cluster")
		Eventually(func(g Gomega) []*release.Release {
			releases, err := action.NewList(helmConfig).Run()
			g.Expect(err).
				ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
			return releases
		}).Should(BeEmpty(), "the helm release should be deleted from the remote cluster")
	})

	It("should re-create CRD if CRD was deleted", func() {
		By("creating plugin definition with CRDs")
		Expect(test.K8sClient.Create(test.Ctx, testPluginWithHelmChartCRDs)).To(Succeed(), "should create plugin definition")

		remoteRestClientGetter := clientutil.NewRestClientGetterFromBytes(remoteKubeConfig, testPluginWithCRDs.Spec.ReleaseNamespace, clientutil.WithPersistentConfig())

		By("creating test plugin referencing the cluster")
		testPluginWithCRDs.Spec.ClusterName = "test-cluster"
		Expect(test.K8sClient.Create(test.Ctx, testPluginWithCRDs)).
			Should(Succeed(), "there should be no error creating the plugin")

		By("checking the ClusterAccessReadyCondition on the plugin")
		Eventually(func(g Gomega) bool {
			err := test.K8sClient.Get(test.Ctx, types.NamespacedName{Name: testPluginWithCRDs.Name, Namespace: testPluginWithCRDs.Namespace}, testPluginWithCRDs)
			g.Expect(err).ShouldNot(HaveOccurred(), "there should be no error getting the plugin")
			clusterAccessReadyCondition := testPluginWithCRDs.Status.GetConditionByType(greenhousev1alpha1.ClusterAccessReadyCondition)
			readyCondition := testPluginWithCRDs.Status.GetConditionByType(greenhousev1alpha1.ReadyCondition)
			g.Expect(clusterAccessReadyCondition).ToNot(BeNil(), "the ClusterAccessReadyCondition should not be nil")
			g.Expect(readyCondition).ToNot(BeNil(), "the ReadyCondition should not be nil")
			return true
		}).Should(BeTrue(), "the ClusterAccessReadyCondition should be false")

		By("checking the helm releases deployed to the remote cluster")
		helmConfig, err := helm.ExportNewHelmAction(remoteRestClientGetter, testPluginWithCRDs.Spec.ReleaseNamespace)
		Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating helm config")
		listAction := action.NewList(helmConfig)
		Eventually(func() []*release.Release {
			releases, err := listAction.Run()
			Expect(err).ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
			return releases
		}).Should(ContainElement(gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{"Name": Equal(testPluginWithCRDs.Name)}))), "the helm release should be deployed to the remote cluster")

		By("checking if helm release exists")
		Eventually(func() bool {
			_, err := helm.GetReleaseForHelmChartFromPlugin(test.Ctx, remoteRestClientGetter, testPluginWithCRDs)
			return err == nil
		}).Should(BeTrue(), "release for helm chart should already exist")

		teamCRDName := "teams.greenhouse.sap"
		teamCRDKey := types.NamespacedName{Name: teamCRDName, Namespace: ""}

		By("Getting Team CRD from remote cluster")
		var teamCRD = &apiextensionsv1.CustomResourceDefinition{}
		Eventually(func(g Gomega) {
			g.Expect(remoteK8sClient.Get(test.Ctx, teamCRDKey, teamCRD)).To(Succeed(), "there must be no error getting the Team CRD")
			g.Expect(teamCRD.Name).To(Equal(teamCRDName), "created Team CRD should have the correct name")
		}).ShouldNot(HaveOccurred(), "Team CRD should be created on remote cluster")

		By("deleting Team CRD from the remote cluster")
		Eventually(func() error {
			return remoteK8sClient.Delete(test.Ctx, teamCRD)
		}).Should(Succeed(), "there must be no error deleting Team CRD")

		By("setting label on plugin to trigger reconciliation")
		// Get up-to-date version of plugin.
		err = test.K8sClient.Get(test.Ctx, types.NamespacedName{Name: testPluginWithCRDs.Name, Namespace: testPluginWithCRDs.Namespace}, testPluginWithCRDs)
		Expect(err).ToNot(HaveOccurred(), "there should be no error getting plugin")
		// Set a label on the plugin to trigger reconciliation.
		testPluginWithCRDs.Labels = map[string]string{"test": "label"}
		Expect(test.K8sClient.Update(test.Ctx, testPluginWithCRDs)).Should(Succeed(), "there should be no error updating the plugin")

		By("ensuring Team CRD was re-created in the remote cluster")
		Eventually(func(g Gomega) {
			var teamCRD = &apiextensionsv1.CustomResourceDefinition{}
			g.Expect(remoteK8sClient.Get(test.Ctx, teamCRDKey, teamCRD)).To(Succeed(), "there must be no error getting the Team CRD")
			g.Expect(teamCRD.Name).To(Equal(teamCRDName), "re-created Team CRD should have the correct name")
		}).Should(Succeed(), "Team CRD should be re-created")

		By("cleaning up test")
		By("deleting the plugin")
		Expect(test.K8sClient.Delete(test.Ctx, testPluginWithCRDs)).Should(Succeed(), "there should be no error deleting the plugin")

		By("checking the helm releases deployed to the remote cluster")
		Eventually(func() []*release.Release {
			releases, err := action.NewList(helmConfig).Run()
			Expect(err).ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
			return releases
		}).Should(BeEmpty(), "the helm release should be deleted from the remote cluster")
	})

	When("reconciling status for plugin with exposed service", func() {
		It("should generate exposed service URL", func() {
			common.DNSDomain = "example.com"

			By("creating plugin definition with exposed service")
			Expect(test.K8sClient.Create(test.Ctx, pluginDefinitionWithExposedService)).To(Succeed(), "should create plugin definition")

			testPluginWithExposedService1 := testPluginWithExposedService.DeepCopy()

			remoteRestClientGetter := clientutil.NewRestClientGetterFromBytes(remoteKubeConfig, testPluginWithExposedService1.Spec.ReleaseNamespace, clientutil.WithPersistentConfig())

			By("creating test plugin referencing the cluster")
			testPluginWithExposedService1.Spec.ClusterName = "test-cluster"
			Expect(test.K8sClient.Create(test.Ctx, testPluginWithExposedService1)).
				Should(Succeed(), "there should be no error creating the plugin")

			By("checking the helm releases deployed to the remote cluster")
			helmConfig, err := helm.ExportNewHelmAction(remoteRestClientGetter, testPluginWithExposedService1.Spec.ReleaseNamespace)
			Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating helm config")
			listAction := action.NewList(helmConfig)
			Eventually(func() []*release.Release {
				releases, err := listAction.Run()
				Expect(err).ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
				return releases
			}).Should(ContainElement(gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{"Name": Equal(testPluginWithExposedService1.Name)}))), "the helm release should be deployed to the remote cluster")

			By("checking plugin status")
			Eventually(func(g Gomega) {
				err = test.K8sClient.Get(test.Ctx, types.NamespacedName{Name: testPluginWithExposedService1.Name, Namespace: testPluginWithExposedService1.Namespace}, testPluginWithExposedService1)
				g.Expect(err).ToNot(HaveOccurred(), "there should be no error getting plugin")
				statusUpToDateCondition := testPluginWithExposedService1.Status.GetConditionByType(greenhousev1alpha1.StatusUpToDateCondition)
				g.Expect(statusUpToDateCondition.Status).To(Equal(metav1.ConditionTrue), "plugin status up to date condition should be set to true")
				g.Expect(testPluginWithExposedService1.Status.ExposedServices).ToNot(BeEmpty(), "exposed services in plugin status should not be empty")
				g.Expect(testPluginWithExposedService1.Status.ExposedServices).To(HaveLen(1), "there should be only one exposed service in plugin status")
				exposedServiceURL := ""
				for exposedServiceURL = range testPluginWithExposedService1.Status.ExposedServices {
					break
				}
				// URL pattern: $https://$service-$namespace-$cluster.$organisation.$basedomain
				g.Expect(exposedServiceURL).To(Equal("https://exposed-service--test-org--test-cluster.test-org.example.com"), "exposed service URL should be generated correctly")
			}).Should(Succeed(), "plugin should have correct status")

			By("cleaning up test")
			By("deleting the plugin")
			Expect(test.K8sClient.Delete(test.Ctx, testPluginWithExposedService1)).Should(Succeed(), "there should be no error deleting the plugin")

			By("checking the helm releases deployed to the remote cluster")
			Eventually(func() []*release.Release {
				releases, err := action.NewList(helmConfig).Run()
				Expect(err).ShouldNot(HaveOccurred(), "there should be no error listing helm releases")
				return releases
			}).Should(BeEmpty(), "the helm release should be deleted from the remote cluster")
		})

		It("should set error on status when cluster name is missing", func() {
			testPluginWithExposedService2 := testPluginWithExposedService.DeepCopy()

			By("checking greenhouse namespace")
			var greenhouseNamespace = new(corev1.Namespace)
			err := test.K8sClient.Get(test.Ctx, types.NamespacedName{Namespace: "", Name: "greenhouse"}, greenhouseNamespace)
			if err != nil {
				By("creating central cluster greenhouse namespace")
				greenhouseNamespace = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "greenhouse"}}
				err = test.K8sClient.Create(test.Ctx, greenhouseNamespace)
				Expect(err).ShouldNot(HaveOccurred(), "there should be no error creating the greenhouse namespace")
			}

			By("creating test plugin without ClusterName")
			// Deploy plugin to central cluster.
			testPluginWithExposedService2.Namespace = "greenhouse"
			testPluginWithExposedService2.Spec.ClusterName = ""
			Expect(test.K8sClient.Create(test.Ctx, testPluginWithExposedService2)).
				Should(Succeed(), "there should be no error creating the plugin")

			By("checking plugin status")
			Eventually(func(g Gomega) {
				err = test.K8sClient.Get(test.Ctx, types.NamespacedName{Name: testPluginWithExposedService2.Name, Namespace: testPluginWithExposedService2.Namespace}, testPluginWithExposedService2)
				g.Expect(err).ToNot(HaveOccurred(), "there should be no error getting plugin")
				statusUpToDateCondition := testPluginWithExposedService2.Status.GetConditionByType(greenhousev1alpha1.StatusUpToDateCondition)
				g.Expect(statusUpToDateCondition).ToNot(BeNil(), "status up to date condition should exist")
				g.Expect(statusUpToDateCondition.Status).To(Equal(metav1.ConditionFalse), "plugin status up to date condition should be set to false")
				g.Expect(statusUpToDateCondition.Message).To(ContainSubstring("plugin does not have ClusterName"), "plugin status up to date condition should have correct message")
				g.Expect(testPluginWithExposedService2.Status.ExposedServices).To(BeEmpty(), "exposed services in plugin status should be empty")
			}).Should(Succeed(), "plugin should have correct status")

			By("deleting the plugin")
			Expect(test.K8sClient.Delete(test.Ctx, testPluginWithExposedService2)).Should(Succeed(), "there should be no error deleting the plugin")
		})
	})
})

func bootstrapRemoteCluster() {
	_, remoteK8sClient, remoteEnvTest, remoteKubeConfig = test.StartControlPlane("6885", false, false)
}
