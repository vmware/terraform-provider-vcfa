//go:build tm || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVcfaProviderGateway(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	vCenterHcl, vCenterHclRef := getVCenterHcl(t, nsxManagerHclRef)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	ipSpace1Hcl, ipSpace1HclRef := getIpSpaceHcl(t, regionHclRef, "1", "1")
	ipSpace2Hcl, ipSpace2HclRef := getIpSpaceHcl(t, regionHclRef, "2", "2")

	var params = StringMap{
		"Testname":     t.Name(),
		"VcenterRef":   vCenterHclRef,
		"RegionId":     fmt.Sprintf("%s.id", regionHclRef),
		"RegionName":   t.Name(),
		"IpSpace1Id":   fmt.Sprintf("%s.id", ipSpace1HclRef),
		"IpSpace2Id":   fmt.Sprintf("%s.id", ipSpace2HclRef),
		"Tier0Gateway": testConfig.Tm.NsxTier0Gateway,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + ipSpace1Hcl + ipSpace2Hcl
	configText1 := templateFill(preRequisites+testAccVcfaProviderGatewayStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcfaProviderGatewayStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaProviderGatewayStep3, params)
	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(preRequisites+testAccVcfaProviderGatewayStep4DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedProviderGateway := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					cachedProviderGateway.cacheTestResourceFieldValue("vcfa_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_provider_gateway.test", "region_id"),
					resource.TestCheckResourceAttrSet("vcfa_provider_gateway.test", "tier0_gateway_id"),
					resource.TestCheckResourceAttr("vcfa_provider_gateway.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcfa_provider_gateway.test", "description", "Made using Terraform"),
					resource.TestCheckResourceAttr("vcfa_provider_gateway.test", "ip_space_ids.#", "2"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					cachedProviderGateway.testCheckCachedResourceFieldValue("vcfa_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_provider_gateway.test", "region_id"),
					resource.TestCheckResourceAttrSet("vcfa_provider_gateway.test", "tier0_gateway_id"),
					resource.TestCheckResourceAttr("vcfa_provider_gateway.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcfa_provider_gateway.test", "description", "Made using Terraform updated"),
					resource.TestCheckResourceAttr("vcfa_provider_gateway.test", "ip_space_ids.#", "1"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					cachedProviderGateway.testCheckCachedResourceFieldValue("vcfa_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_provider_gateway.test", "region_id"),
					resource.TestCheckResourceAttrSet("vcfa_provider_gateway.test", "tier0_gateway_id"),
					resource.TestCheckResourceAttr("vcfa_provider_gateway.test", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttr("vcfa_provider_gateway.test", "description", ""),
					resource.TestCheckResourceAttr("vcfa_provider_gateway.test", "ip_space_ids.#", "3"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_provider_gateway.test", "data.vcfa_provider_gateway.test", nil),
				),
			},
			{
				ResourceName:      "vcfa_provider_gateway.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.Tm.Region + ImportSeparator + params["Testname"].(string) + "-updated",
			},
		},
	})
}

const testAccVcfaProviderGatewayPrereqs = `
resource "vcfa_ip_space" "test" {
  name                          = "{{.Testname}}"
  description                   = "description test"
  region_id                     = {{.RegionId}}
  external_scope                = "12.12.0.0/30"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1

  internal_scope {
    cidr = "10.0.0.0/28"
  }
}

resource "vcfa_ip_space" "test2" {
  name                          = "{{.Testname}}-2"
  description                   = "description test"
  region_id                     = {{.RegionId}}
  external_scope                = "13.12.0.0/30"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1

  internal_scope {
    cidr = "9.0.0.0/28"
  }
}

data "vcfa_tier0_gateway" "test" {
  name      = "{{.Tier0Gateway}}"
  region_id = {{.RegionId}}
}
`

const testAccVcfaProviderGatewayStep1 = testAccVcfaProviderGatewayPrereqs + `
resource "vcfa_provider_gateway" "test" {
  name             = "{{.Testname}}"
  description      = "Made using Terraform"
  region_id        = {{.RegionId}}
  tier0_gateway_id = data.vcfa_tier0_gateway.test.id
  ip_space_ids     = [ vcfa_ip_space.test.id, vcfa_ip_space.test2.id ]
}
`

const testAccVcfaProviderGatewayStep2 = testAccVcfaProviderGatewayPrereqs + `
resource "vcfa_provider_gateway" "test" {
  name             = "{{.Testname}}"
  description      = "Made using Terraform updated"
  region_id        = {{.RegionId}}
  tier0_gateway_id = data.vcfa_tier0_gateway.test.id
  ip_space_ids     = [ vcfa_ip_space.test2.id ]
}
`

const testAccVcfaProviderGatewayStep3 = testAccVcfaProviderGatewayPrereqs + `
resource "vcfa_provider_gateway" "test" {
  name             = "{{.Testname}}-updated"
  region_id        = {{.RegionId}}
  tier0_gateway_id = data.vcfa_tier0_gateway.test.id
  ip_space_ids     = [ vcfa_ip_space.test2.id, vcfa_ip_space.test.id, {{.IpSpace1Id}} ]
}
`

const testAccVcfaProviderGatewayStep4DS = testAccVcfaProviderGatewayStep3 + `
data "vcfa_provider_gateway" "test" {
  name      = vcfa_provider_gateway.test.name
  region_id = {{.RegionId}}
}
`
