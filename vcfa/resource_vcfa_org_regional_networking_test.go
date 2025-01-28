//go:build tm || org || ALL || functional

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaOrgRegionalNetworking(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	ipSpaceHcl, ipSpaceHclRef := getIpSpaceHcl(t, regionHclRef, "1", "1")
	providerGatewayHcl, providerGatewayHclRef := getProviderGatewayHcl(t, regionHclRef, ipSpaceHclRef)

	var params = StringMap{
		"Testname":          t.Name(),
		"RegionId":          fmt.Sprintf("%s.id", regionHclRef),
		"ProviderGatewayId": fmt.Sprintf("%s.id", providerGatewayHclRef),
		"EdgeClusterName":   testConfig.Tm.NsxEdgeCluster,
		"Tags":              "tm org",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + ipSpaceHcl + providerGatewayHcl
	configText1 := templateFill(preRequisites+testAccVcfaOrgRegionalNetworkingStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcfaOrgRegionalNetworkingStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaOrgRegionalNetworkingStep3DS, params)
	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(preRequisites+testAccVcfaOrgRegionalNetworkingStep4VpcQos, params)
	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(preRequisites+testAccVcfaOrgRegionalNetworkingStep5VpcQos, params)

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
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "status"),
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "org_id"),
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "provider_gateway_id"),
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "region_id"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking.test", "name", t.Name()),
					// Edge Cluster ID was not specified, but it is automatically picked
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "edge_cluster_id"),
				),
			},
			{ // Update - only name and edge_cluster_id can be updated
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "status"),
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "org_id"),
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "provider_gateway_id"),
					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "region_id"),

					resource.TestCheckResourceAttrSet("vcfa_org_regional_networking.test", "edge_cluster_id"),
					resource.TestCheckResourceAttrPair("vcfa_org_regional_networking.test", "edge_cluster_id", "data.vcfa_edge_cluster.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking.test", "name", t.Name()+"-upd"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_org_regional_networking.test", "data.vcfa_org_regional_networking.test", nil),
				),
			},
			{
				ResourceName:      "vcfa_org_regional_networking.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s%s%s", params["Testname"].(string), ImportSeparator, params["Testname"].(string)+"-upd"), // Org name and Region name
			},
			{ // Testing vcfa_org_regional_networking_vpc_qos - ensuring that the same QoS parameters are inherited from parent Edge Cluster
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					// Ensure that the same Edge Cluster is backing Org Regional Networking and that Egress and Ingress configurations are the same
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "edge_cluster_id", "data.vcfa_org_regional_networking_vpc_qos.test", "edge_cluster_id"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "ingress_committed_bandwidth_mbps", "data.vcfa_org_regional_networking_vpc_qos.test", "ingress_committed_bandwidth_mbps"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "ingress_burst_size_bytes", "data.vcfa_org_regional_networking_vpc_qos.test", "ingress_burst_size_bytes"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "egress_committed_bandwidth_mbps", "data.vcfa_org_regional_networking_vpc_qos.test", "egress_committed_bandwidth_mbps"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "egress_burst_size_bytes", "data.vcfa_org_regional_networking_vpc_qos.test", "egress_burst_size_bytes"),
				),
			},
			{ // Testing vcfa_org_regional_networking_vpc_qos - overriding default values to custom at VPC QoS level
				Config: configText5,
				Check: resource.ComposeTestCheckFunc(
					// Ensure that the same Edge Cluster is backing Org Regional Networking and that Egress and Ingress configurations are the same
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "edge_cluster_id", "data.vcfa_org_regional_networking_vpc_qos.test", "edge_cluster_id"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_vpc_qos.test", "ingress_committed_bandwidth_mbps", "14"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_vpc_qos.test", "ingress_burst_size_bytes", "15"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_vpc_qos.test", "egress_committed_bandwidth_mbps", "16"),
					resource.TestCheckResourceAttr("vcfa_org_regional_networking_vpc_qos.test", "egress_burst_size_bytes", "17"),
				),
			},
			{
				// vcfa_org_regional_networking_vpc_qos has the same path as vcfa_org_regional_networking, because VPC QoS is just a property
				ResourceName:      "vcfa_org_regional_networking_vpc_qos.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s%s%s", params["Testname"].(string), ImportSeparator, params["Testname"].(string)+"-upd"),
			},
			{ // Testing vcfa_org_regional_networking_vpc_qos - check that after removing the custom VPC QoS
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					// Ensure that the same Edge Cluster is backing Org Regional Networking and that Egress and Ingress configurations are the same
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "edge_cluster_id", "data.vcfa_org_regional_networking_vpc_qos.test", "edge_cluster_id"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "ingress_committed_bandwidth_mbps", "data.vcfa_org_regional_networking_vpc_qos.test", "ingress_committed_bandwidth_mbps"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "ingress_burst_size_bytes", "data.vcfa_org_regional_networking_vpc_qos.test", "ingress_burst_size_bytes"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "egress_committed_bandwidth_mbps", "data.vcfa_org_regional_networking_vpc_qos.test", "egress_committed_bandwidth_mbps"),
					resource.TestCheckResourceAttrPair("data.vcfa_edge_cluster_qos.test", "egress_burst_size_bytes", "data.vcfa_org_regional_networking_vpc_qos.test", "egress_burst_size_bytes"),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaOrgRegionalNetworkingPrerequisites = `
resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}

resource "vcfa_org_networking" "test" {
  org_id   = vcfa_org.test.id
  log_name = "tftest"
}
`

const testAccVcfaOrgRegionalNetworkingStep1 = testAccVcfaOrgRegionalNetworkingPrerequisites + `
resource "vcfa_org_regional_networking" "test" {
  name                = "{{.Testname}}"
  org_id              = vcfa_org.test.id
  provider_gateway_id = {{.ProviderGatewayId}}
  region_id           = {{.RegionId}}

  depends_on = [vcfa_org_networking.test]
}
`

const testAccVcfaOrgRegionalNetworkingStep2 = testAccVcfaOrgRegionalNetworkingPrerequisites + `
data "vcfa_edge_cluster" "test" {
  name      = "{{.EdgeClusterName}}"
  region_id = {{.RegionId}}
}

resource "vcfa_org_regional_networking" "test" {
  name                = "{{.Testname}}-upd"
  org_id              = vcfa_org.test.id
  provider_gateway_id = {{.ProviderGatewayId}}
  region_id           = {{.RegionId}}
  edge_cluster_id     = data.vcfa_edge_cluster.test.id

  depends_on = [vcfa_org_networking.test]
}
`

const testAccVcfaOrgRegionalNetworkingStep3DS = testAccVcfaOrgRegionalNetworkingStep2 + `
data "vcfa_org_regional_networking" "test" {
  name   = vcfa_org_regional_networking.test.name
  org_id = vcfa_org.test.id
}
`

const testAccVcfaOrgRegionalNetworkingStep4VpcQos = testAccVcfaOrgRegionalNetworkingStep2 + `
data "vcfa_edge_cluster_qos" "test" {
  edge_cluster_id = data.vcfa_edge_cluster.test.id
}

data "vcfa_org_regional_networking_vpc_qos" "test" {
  org_regional_networking_id = vcfa_org_regional_networking.test.id
}

`

const testAccVcfaOrgRegionalNetworkingStep5VpcQos = testAccVcfaOrgRegionalNetworkingStep2 + `
data "vcfa_edge_cluster_qos" "test" {
  edge_cluster_id = data.vcfa_edge_cluster.test.id
}

data "vcfa_org_regional_networking_vpc_qos" "test" {
  org_regional_networking_id = vcfa_org_regional_networking.test.id
}

resource "vcfa_org_regional_networking_vpc_qos" "test" {
  org_regional_networking_id = vcfa_org_regional_networking.test.id
  ingress_committed_bandwidth_mbps = 14
  ingress_burst_size_bytes         = 15
  egress_committed_bandwidth_mbps  = 16
  egress_burst_size_bytes          = 17
}

`
