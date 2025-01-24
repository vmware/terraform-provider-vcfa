//go:build tm || ALL || functional

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaEdgeCluster(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	var params = StringMap{
		"Testname":        t.Name(),
		"VcenterRef":      vCenterHclRef,
		"RegionId":        fmt.Sprintf("%s.id", regionHclRef),
		"RegionName":      t.Name(),
		"EdgeClusterName": testConfig.Tm.NsxEdgeCluster,
		"SyncBeforeRead":  "true",

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl
	configText1 := templateFill(preRequisites+testAccVcfaEdgeClusterQosStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	params["SyncBeforeRead"] = "false" // TODO TM - Sync'ing resets QoS policy to unlimited (-1)
	configText2 := templateFill(preRequisites+testAccVcfaEdgeClusterQosStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaEdgeClusterQosStep3, params)
	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(preRequisites+testAccVcfaEdgeClusterQosStep4, params)
	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(preRequisites+testAccVcfaEdgeClusterQosStep5, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)
	debugPrintf("#[DEBUG] CONFIGURATION step5: %s\n", configText5)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "id"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster_qos.demo", "id"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.demo", "id", "data.vcfa_edge_cluster.demo", "id"),
					resource.TestCheckResourceAttr("data.vcfa_edge_cluster.demo", "name", testConfig.Tm.NsxEdgeCluster),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "region_id"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "node_count"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "org_count"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "vpc_count"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "average_cpu_usage_percentage"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "average_memory_usage_percentage"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "health_status"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "status"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "deployment_type"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "id"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster_qos.demo", "id"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.demo", "id", "data.vcfa_edge_cluster.demo", "id"),
					resource.TestCheckResourceAttr("data.vcfa_edge_cluster.demo", "name", testConfig.Tm.NsxEdgeCluster),

					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "egress_committed_bandwidth_mbps", "1"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "egress_burst_size_bytes", "2"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "ingress_committed_bandwidth_mbps", "3"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "ingress_burst_size_bytes", "4"),
				),
			},
			{
				RefreshState: true, // ensuring that data source is reloaded with latest data that is configured in the resource
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster.demo", "id"),
					resource.TestCheckResourceAttrSet("data.vcfa_edge_cluster_qos.demo", "id"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.demo", "id", "data.vcfa_edge_cluster.demo", "id"),
					resource.TestCheckResourceAttr("data.vcfa_edge_cluster.demo", "name", testConfig.Tm.NsxEdgeCluster),

					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "egress_committed_bandwidth_mbps", "1"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "egress_burst_size_bytes", "2"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "ingress_committed_bandwidth_mbps", "3"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "ingress_burst_size_bytes", "4"),
					resourceFieldsEqual("data.vcfa_edge_cluster_qos.demo", "vcfa_edge_cluster_qos.demo", nil),
				),
			},
			{
				ResourceName:      "vcfa_edge_cluster_qos.demo",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.Tm.Region + ImportSeparator + params["EdgeClusterName"].(string),
			},
			{
				// Ensuring that the resource is removed (therefore QoS settings must be unset)
				Config: configText3,
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				// Checking that the data source reflects empty QoS values (delete of resource removes Qos Settings)
				// The refresh is required
				RefreshState: true, // ensuring that data source is reloaded with latest data that is configured in the resource
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcfa_edge_cluster_qos.demo2", "egress_committed_bandwidth_mbps", "-1"),
					resource.TestCheckResourceAttr("data.vcfa_edge_cluster_qos.demo2", "egress_burst_size_bytes", "-1"),
					resource.TestCheckResourceAttr("data.vcfa_edge_cluster_qos.demo2", "ingress_committed_bandwidth_mbps", "-1"),
					resource.TestCheckResourceAttr("data.vcfa_edge_cluster_qos.demo2", "ingress_burst_size_bytes", "-1"),
				),
			},
			{
				// Ensuring that the resource is removed (therefore QoS settings must be unset)
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "egress_committed_bandwidth_mbps", "7"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "egress_burst_size_bytes", "8"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "ingress_committed_bandwidth_mbps", "-1"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "ingress_burst_size_bytes", "-1"),
				),
			},
			{
				// Ensuring that the resource is removed (therefore QoS settings must be unset)
				Config: configText5,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "egress_committed_bandwidth_mbps", "-1"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "egress_burst_size_bytes", "-1"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "ingress_committed_bandwidth_mbps", "5"),
					resource.TestCheckResourceAttr("vcfa_edge_cluster_qos.demo", "ingress_burst_size_bytes", "6"),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaEdgeClusterQosStep1 = `
data "vcfa_edge_cluster" "demo" {
  name             = "{{.EdgeClusterName}}"
  region_id        = {{.RegionId}}
  sync_before_read = {{.SyncBeforeRead}}
}

data "vcfa_edge_cluster_qos" "demo" {
  edge_cluster_id = data.vcfa_edge_cluster.demo.id
}  
`

const testAccVcfaEdgeClusterQosStep2 = testAccVcfaEdgeClusterQosStep1 + `
resource "vcfa_edge_cluster_qos" "demo" {
  edge_cluster_id = data.vcfa_edge_cluster.demo.id

  egress_committed_bandwidth_mbps  = 1
  egress_burst_size_bytes          = 2
  ingress_committed_bandwidth_mbps = 3
  ingress_burst_size_bytes         = 4
}
`

const testAccVcfaEdgeClusterQosStep3 = `
data "vcfa_edge_cluster" "demo" {
  name             = "{{.EdgeClusterName}}"
  region_id        = {{.RegionId}}
  sync_before_read = {{.SyncBeforeRead}}
}

data "vcfa_edge_cluster_qos" "demo2" {
  edge_cluster_id = data.vcfa_edge_cluster.demo.id
}  
`

// egress only
const testAccVcfaEdgeClusterQosStep4 = testAccVcfaEdgeClusterQosStep1 + `
resource "vcfa_edge_cluster_qos" "demo" {
  edge_cluster_id = data.vcfa_edge_cluster.demo.id

  egress_committed_bandwidth_mbps  = 7
  egress_burst_size_bytes          = 8

}
`

// ingress only
const testAccVcfaEdgeClusterQosStep5 = testAccVcfaEdgeClusterQosStep1 + `
resource "vcfa_edge_cluster_qos" "demo" {
  edge_cluster_id = data.vcfa_edge_cluster.demo.id

  ingress_committed_bandwidth_mbps  = 5
  ingress_burst_size_bytes          = 6

}
`
