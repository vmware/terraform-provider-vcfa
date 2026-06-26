//go:build vks || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkskubernetesrelease_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/vmware/terraform-provider-vcfa/internal/testutils"
	"github.com/vmware/terraform-provider-vcfa/internal/testutils/providertest"
)

// TestAccVcfaVksKubernetesReleaseDatasourceExternal exercises the read path of the
// vcfa_vks_kubernetes_release data source against a live environment.
func TestAccVcfaVksKubernetesReleaseDatasourceExternal(t *testing.T) {
	testutils.SkipIfSysAdmin(t)

	cfg := testutils.GetTestConfig(t)

	params := testutils.StringMap{
		"Project":               cfg.Vks.Project,
		"Namespace":             cfg.Vks.Namespace,
		"KubernetesReleaseName": cfg.Vks.KubernetesReleaseName,
	}
	testutils.TestParamsNotEmpty(t, params)

	configText := testutils.TemplateFill(t, testAccVcfaVksKubernetesReleaseDatasourceExternalConfig, params)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providertest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "id"),
					resource.TestCheckResourceAttr("data.vcfa_vks_kubernetes_release.test", "context.project", params["Project"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_kubernetes_release.test", "context.namespace", params["Namespace"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_kubernetes_release.test", "name", params["KubernetesReleaseName"].(string)),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "metadata.creation_timestamp"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "metadata.generation"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "metadata.labels.%"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "metadata.resource_version"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "metadata.uid"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "version"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "kubernetes.version"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "kubernetes.image_repository"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "kubernetes.etcd.%"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "kubernetes.pause.%"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_kubernetes_release.test", "kubernetes.coredns.%"),
					testutils.CheckAttrNonEmptySet("data.vcfa_vks_kubernetes_release.test", "os_images.#"),
					testutils.CheckAttrNonEmptySet("data.vcfa_vks_kubernetes_release.test", "bootstrap_packages.#"),
					testutils.CheckAttrNonEmptySet("data.vcfa_vks_kubernetes_release.test", "status.conditions.#"),
				),
			},
		},
	})
}

// testAccVcfaVksKubernetesReleaseDatasourceExternalConfig is the HCL template for the
// vcfa_vks_kubernetes_release data source.
const testAccVcfaVksKubernetesReleaseDatasourceExternalConfig = `
data "vcfa_vks_kubernetes_release" "test" {
  context = {
    project   = "{{.Project}}"
    namespace = "{{.Namespace}}"
  }
  name = "{{.KubernetesReleaseName}}"
}
`
