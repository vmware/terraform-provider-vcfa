//go:build tm || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVcfaVirtualNetworkApplianceClusterDS(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)
	if !vcfaVersionAtLeast("9.1.0") {
		t.Skip("Virtual Network Appliance Clusters require VCFA 9.1.0 or later")
	}

	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	vCenterHcl, vCenterHclRef := getVCenterHcl(t, nsxManagerHclRef)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	vnaClusterHcl, vnaClusterHclRef := getVnaClusterHcl(t, regionHclRef)

	var params = StringMap{
		"Testname": t.Name(),
		"Tags":     "tm",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + vnaClusterHcl
	configText1 := templateFill(preRequisites+"# skip-binary-test: requires the NSX-T Terraform provider", params)
	params["FuncName"] = t.Name() + "-step1"

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		ExternalProviders: getNsxtExternalProviders(),
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(vnaClusterHclRef, "id", regexp.MustCompile(`^urn:vcloud:virtualNetworkApplianceCluster:`)),
					resource.TestCheckResourceAttr(vnaClusterHclRef, "name", t.Name()),
					resource.TestCheckResourceAttrPair(vnaClusterHclRef, "region_id", regionHclRef, "id"),
					resource.TestCheckResourceAttrSet(vnaClusterHclRef, "description"),
					resource.TestCheckResourceAttrSet(vnaClusterHclRef, "node_count"),
					resource.TestCheckResourceAttrSet(vnaClusterHclRef, "org_count"),
					resource.TestCheckResourceAttrSet(vnaClusterHclRef, "health_status"),
				),
			},
		},
	})
}
