//go:build vks || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vksclusterclass_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/vmware/terraform-provider-vcfa/internal/testutils"
	"github.com/vmware/terraform-provider-vcfa/internal/testutils/providertest"
)

// TestAccVcfaVksClusterClassDatasourceExternal exercises the read path of the
// vcfa_vks_cluster_class data source against a live environment.
func TestAccVcfaVksClusterClassDatasourceExternal(t *testing.T) {
	testutils.SkipIfSysAdmin(t)

	cfg := testutils.GetTestConfig(t)

	params := testutils.StringMap{
		"Project":               cfg.Vks.Project,
		"Namespace":             cfg.Vks.Namespace,
		"ClusterClassName":      cfg.Vks.ClusterClassName,
		"ClusterClassNamespace": cfg.Vks.ClusterClassNamespace,
		"System":                "true",
	}
	if params["ClusterClassNamespace"].(string) == "vmware-system-vks-public" {
		params["System"] = "true"
	}
	testutils.TestParamsNotEmpty(t, params)

	configText := testutils.TemplateFill(t, testAccVcfaVksClusterClassDatasourceExternalConfig, params)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providertest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_class.test", "id"),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster_class.test", "context.project", params["Project"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster_class.test", "context.namespace", params["Namespace"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster_class.test", "name", params["ClusterClassName"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster_class.test", "system", params["System"].(string)),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_class.test", "metadata.creation_timestamp"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_class.test", "metadata.generation"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_class.test", "metadata.labels.%"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_class.test", "metadata.resource_version"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_class.test", "metadata.uid"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_class.test", "infrastructure.template_ref.name"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_class.test", "control_plane.template_ref.name"),
					testutils.CheckAttrNonEmptySet("data.vcfa_vks_cluster_class.test", "workers.machine_deployments.#"),
					testutils.CheckAttrNonEmptySet("data.vcfa_vks_cluster_class.test", "status.conditions.#"),
					testutils.CheckAttrNonEmptySet("data.vcfa_vks_cluster_class.test", "status.variables.#"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_class.test", "status.observed_generation"),
				),
			},
		},
	})
}

// testAccVcfaVksClusterClassDatasourceExternalConfig is the HCL template for the
// vcfa_vks_cluster_class data source.
const testAccVcfaVksClusterClassDatasourceExternalConfig = `
data "vcfa_vks_cluster_class" "test" {
  context = {
    project   = "{{.Project}}"
    namespace = "{{.Namespace}}"
  }
  name   = "{{.ClusterClassName}}"
  system = {{.System}}
}
`
