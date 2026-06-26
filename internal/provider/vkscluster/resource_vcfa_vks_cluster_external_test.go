//go:build vks || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vkscluster_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/vmware/terraform-provider-vcfa/internal/testutils"
	"github.com/vmware/terraform-provider-vcfa/internal/testutils/providertest"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

// TestAccVcfaVksClusterResourceExternal exercises the full lifecycle
// (create → update → datasource read → import → destroy) of the vcfa_vks_cluster
// resource against a live environment.
func TestAccVcfaVksClusterResourceExternal(t *testing.T) {
	testutils.SkipIfSysAdmin(t)

	cfg := testutils.GetTestConfig(t)

	// Kubernetes resource names must be lowercase DNS subdomains.
	clusterName := strings.ReplaceAll(strings.ToLower(t.Name()), "_", "-")

	// Compute the updated replica count by incrementing the configured value.
	workerReplicas, err := strconv.Atoi(cfg.Vks.WorkerReplicas)
	if err != nil {
		t.Fatalf("vks.workerReplicas %q is not a valid integer: %s", cfg.Vks.WorkerReplicas, err)
	}
	workerReplicasUpdated := strconv.Itoa(workerReplicas + 1)

	params := testutils.StringMap{
		"Project":     cfg.Vks.Project,
		"Namespace":   cfg.Vks.Namespace,
		"ClusterName": clusterName,

		"ClusterClassName":      cfg.Vks.ClusterClassName,
		"ClusterClassNamespace": cfg.Vks.ClusterClassNamespace,
		"KubernetesVersion":     cfg.Vks.KubernetesVersion,
		"ServicesCidr":          cfg.Vks.ServicesCidr,
		"VmClass":               cfg.Vks.VmClass,
		"StorageClass":          cfg.Vks.StorageClass,
		"ControlPlaneReplicas":  cfg.Vks.ControlPlaneReplicas,
		"WorkerReplicas":        cfg.Vks.WorkerReplicas,
		"WorkerReplicasUpdated": workerReplicasUpdated,
	}
	testutils.TestParamsNotEmpty(t, params)

	configText1 := testutils.TemplateFill(t, testAccVcfaVksClusterExternalConfig, params)
	params["FuncName"] = t.Name() + "-update"
	configText2 := testutils.TemplateFill(t, testAccVcfaVksClusterExternalConfigUpdate, params)
	params["FuncName"] = t.Name() + "-ds"
	configText3 := testutils.TemplateFill(t, testAccVcfaVksClusterExternalConfigWithDatasource, params)

	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providertest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create and verify all configured fields.
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_vks_cluster.test", "id"),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "context.project", params["Project"].(string)),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "context.namespace", params["Namespace"].(string)),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "cluster_class.name", params["ClusterClassName"].(string)),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "cluster_class.namespace", params["ClusterClassNamespace"].(string)),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "version", params["KubernetesVersion"].(string)),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "cluster_network.services.cidr_blocks.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcfa_vks_cluster.test", "cluster_network.services.cidr_blocks.*", params["ServicesCidr"].(string)),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "variables.#", "2"),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "control_plane.replicas", params["ControlPlaneReplicas"].(string)),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "machine_deployments.#", "1"),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "machine_deployments.0.class", "node-pool"),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "machine_deployments.0.name", "default"),
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "machine_deployments.0.replicas", params["WorkerReplicas"].(string)),
				),
			},
			// Step 2: scale worker nodes and verify the change is reflected.
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_vks_cluster.test", "machine_deployments.0.replicas", params["WorkerReplicasUpdated"].(string)),
				),
			},
			// Step 3: verify the data source reflects the updated state.
			// We use explicit checks rather than ResourceFieldsEqual because the
			// datasource schema exposes many computed-only fields (availability_gates,
			// control_plane_endpoint, status sub-fields, etc.) that the resource does
			// not track, making a full field-equality comparison impractical.
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					// vcfa_vks_cluster datasource checks.
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster.test", "id"),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "context.project", params["Project"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "context.namespace", params["Namespace"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "name", clusterName),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "cluster_class.name", params["ClusterClassName"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "cluster_class.namespace", params["ClusterClassNamespace"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "version", params["KubernetesVersion"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "cluster_network.services.cidr_blocks.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.vcfa_vks_cluster.test", "cluster_network.services.cidr_blocks.*", params["ServicesCidr"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "variables.#", "2"),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "control_plane.replicas", params["ControlPlaneReplicas"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "machine_deployments.#", "1"),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "machine_deployments.0.class", "node-pool"),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "machine_deployments.0.name", "default"),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster.test", "machine_deployments.0.replicas", params["WorkerReplicasUpdated"].(string)),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster.test", "metadata.creation_timestamp"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster.test", "metadata.generation"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster.test", "metadata.labels.%"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster.test", "metadata.resource_version"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster.test", "metadata.uid"),
					testutils.CheckAttrNonEmptySet("data.vcfa_vks_cluster.test", "status.conditions.#"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster.test", "status.phase"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster.test", "status.observed_generation"),

					// vcfa_vks_cluster_kubeconfig datasource checks.
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_kubeconfig.test", "id"),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster_kubeconfig.test", "context.project", params["Project"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster_kubeconfig.test", "context.namespace", params["Namespace"].(string)),
					resource.TestCheckResourceAttr("data.vcfa_vks_cluster_kubeconfig.test", "name", clusterName),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_kubeconfig.test", "host"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_kubeconfig.test", "kube_config_raw"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_kubeconfig.test", "context_name"),
					resource.TestCheckResourceAttrSet("data.vcfa_vks_cluster_kubeconfig.test", "user"),
				),
			},
			// Step 4: import and verify the state round-trips cleanly.
			{
				ResourceName:      "vcfa_vks_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return params["Project"].(string) + vcfa.ImportSeparator + params["Namespace"].(string) + vcfa.ImportSeparator + params["ClusterName"].(string), nil
				},
				ImportStateVerifyIgnore: []string{
					"dry_run_validation",     // local-only
					"wait_for",               // local-only
					"timeouts",               // local-only
					"metadata",               // computed-only
					"availability_gates",     // computed-only
					"control_plane_endpoint", // computed-only
					"status",                 // computed-only
				},
			},
		},
	})
}

// testAccVcfaVksClusterExternalConfig is the Step 1 (create) HCL template.
const testAccVcfaVksClusterExternalConfig = `
resource "vcfa_vks_cluster" "test" {
  context = {
    project   = "{{.Project}}"
    namespace = "{{.Namespace}}"
  }

  name = "{{.ClusterName}}"

  wait_for = {
    available = false
    deleted   = true
  }

  cluster_class = {
    name      = "{{.ClusterClassName}}"
    namespace = "{{.ClusterClassNamespace}}"
  }
  version = "{{.KubernetesVersion}}"

  cluster_network = {
    services = {
      cidr_blocks = ["{{.ServicesCidr}}"]
    }
  }

  variables = [
    {
      name  = "vmClass"
      value = "{{.VmClass}}"
    },
    {
      name  = "storageClass"
      value = "{{.StorageClass}}"
    },
  ]

  control_plane = {
    replicas = {{.ControlPlaneReplicas}}
  }

  machine_deployments = [
    {
      class    = "node-pool"
      name     = "default"
      replicas = {{.WorkerReplicas}}
    },
  ]
}
`

// testAccVcfaVksClusterExternalConfigUpdate is the Step 2 (update) HCL template.
// It scales the worker node count to WorkerReplicasUpdated.
const testAccVcfaVksClusterExternalConfigUpdate = `
resource "vcfa_vks_cluster" "test" {
  context = {
    project   = "{{.Project}}"
    namespace = "{{.Namespace}}"
  }

  name = "{{.ClusterName}}"

  wait_for = {
    available = true
    deleted   = true
  }

  cluster_class = {
    name      = "{{.ClusterClassName}}"
    namespace = "{{.ClusterClassNamespace}}"
  }
  version = "{{.KubernetesVersion}}"

  cluster_network = {
    services = {
      cidr_blocks = ["{{.ServicesCidr}}"]
    }
  }

  variables = [
    {
      name  = "vmClass"
      value = "{{.VmClass}}"
    },
    {
      name  = "storageClass"
      value = "{{.StorageClass}}"
    },
  ]

  control_plane = {
    replicas = {{.ControlPlaneReplicas}}
  }

  machine_deployments = [
    {
      class    = "node-pool"
      name     = "default"
      replicas = {{.WorkerReplicasUpdated}}
    },
  ]
}
`

// testAccVcfaVksClusterExternalConfigWithDatasource adds a vcfa_vks_cluster and a
// vcfa_vks_cluster_kubeconfig data source to the Step 2 configuration so Step 3
// can verify both datasources read back the correct state. The cluster must be
// Available before this step runs (guaranteed by wait_for.available = true in
// Step 2) because the kubeconfig secret is only created once the cluster is
// provisioned.
const testAccVcfaVksClusterExternalConfigWithDatasource = testAccVcfaVksClusterExternalConfigUpdate + `
data "vcfa_vks_cluster" "test" {
  context = {
    project   = "{{.Project}}"
    namespace = "{{.Namespace}}"
  }

  name = vcfa_vks_cluster.test.name
}

data "vcfa_vks_cluster_kubeconfig" "test" {
  context = {
    project   = "{{.Project}}"
    namespace = "{{.Namespace}}"
  }

  name = vcfa_vks_cluster.test.name
}
`
