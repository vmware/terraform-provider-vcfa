//go:build tp || ALL || functional

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaTpSupervisorNamespace(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	ipSpaceHcl, ipSpaceHclRef := getIpSpaceHcl(t, regionHclRef, "1", "1")
	providerGatewayHcl, providerGatewayHclRef := getProviderGatewayHcl(t, regionHclRef, ipSpaceHclRef)

	// missing prerequisite -  Setup Region Quota and Regional Network

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

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	// debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	// debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	// debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)
	// debugPrintf("#[DEBUG] CONFIGURATION step5: %s\n", configText5)
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
		},
	})

	postTestChecks(t)
}
