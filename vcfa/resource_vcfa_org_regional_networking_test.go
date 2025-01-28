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

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
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
