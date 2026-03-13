//go:build tm || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVcfaIpSpace(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	vCenterHcl, vCenterHclRef := getVCenterHcl(t, nsxManagerHclRef)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)

	k8sCompliantName := strings.ReplaceAll(strings.ToLower(t.Name()), "_", "-")

	var params = StringMap{
		"Testname":      k8sCompliantName,
		"VcenterRef":    vCenterHclRef,
		"RegionId":      fmt.Sprintf("%s.id", regionHclRef),
		"RegionName":    t.Name(),
		"IpScopePrefix": hashAndTakeFirstSix(testConfig.Tm.VcenterUrl),

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
	configText1 := templateFill(preRequisites+testAccVcfaIpSpaceStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcfaIpSpaceStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaIpSpaceStep3DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedIpSpaceId := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					cachedIpSpaceId.cacheTestResourceFieldValue("vcfa_ip_space.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_ip_space.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_ip_space.test", "status"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "name", k8sCompliantName),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "description", "description test"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_subnet_size", "24"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_cidr_count", "1"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_ip_count", "1"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "provider_visibility_only", "true"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "cidr_blocks.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "cidr_blocks.*", map[string]string{
						"name": params["IpScopePrefix"].(string) + "-1",
						"cidr": "10.0.0.0/24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "cidr_blocks.*", map[string]string{
						"cidr": "11.0.0.0/26",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "cidr_blocks.*", map[string]string{
						"name": params["IpScopePrefix"].(string) + "-3",
						"cidr": "12.0.0.0/27",
					}),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "ip_address_ranges.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "ip_address_ranges.*", map[string]string{
						"start_ip_address": "13.0.0.1",
						"end_ip_address":   "13.0.0.255",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "ip_address_ranges.*", map[string]string{
						"start_ip_address": "14.0.0.1",
						"end_ip_address":   "14.0.0.255",
					}),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "reserved_ip_address_ranges.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "reserved_ip_address_ranges.*", map[string]string{
						"start_ip_address": "14.0.0.1",
						"end_ip_address":   "14.0.0.10",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					cachedIpSpaceId.testCheckCachedResourceFieldValue("vcfa_ip_space.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_ip_space.test", "id"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "status", "REALIZED"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "name", k8sCompliantName+"-updated"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "description", "description test - update"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_subnet_size", "25"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_cidr_count", "-1"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_ip_count", "-1"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "provider_visibility_only", "false"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "cidr_blocks.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "internal_scope.*", map[string]string{
						"name": params["IpScopePrefix"].(string) + "-3",
						"cidr": "12.0.0.0/27",
					}),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "ip_address_ranges.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "ip_address_ranges.*", map[string]string{
						"start_ip_address": "14.0.0.1",
						"end_ip_address":   "14.0.0.255",
					}),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "reserved_ip_address_ranges.#", "0"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_ip_space.test", "data.vcfa_ip_space.test", nil),
				),
			},
			{
				ResourceName:      "vcfa_ip_space.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.Tm.Region + ImportSeparator + params["Testname"].(string) + "-updated",
			},
		},
	})
}

const testAccVcfaIpSpaceStep1 = `
resource "vcfa_ip_space" "test" {
  name                          = "{{.Testname}}"
  description                   = "description test"
  region_id                     = {{.RegionId}}
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1
  provider_visibility_only      = true

  cidr_blocks {
    name = "{{.IpScopePrefix}}-1"
    cidr = "10.0.0.0/24"
  }

  cidr_blocks {
    cidr = "11.0.0.0/26"
  }

  cidr_blocks {
    name = "{{.IpScopePrefix}}-3"
    cidr = "12.0.0.0/27"
  }

  ip_address_ranges {
    start_ip_address = "13.0.0.1"
    end_ip_address   = "13.0.0.255"
  }

  ip_address_ranges {
    start_ip_address = "14.0.0.1"
    end_ip_address   = "14.0.0.255"
  }

 reserved_ip_address_ranges {
    start_ip_address = "14.0.0.1"
    end_ip_address   = "14.0.0.10"
  }
}
`

const testAccVcfaIpSpaceStep2 = `
resource "vcfa_ip_space" "test" {
  name                          = "{{.Testname}}-updated"
  description                   = "description test - update"
  region_id                     = {{.RegionId}}
  default_quota_max_subnet_size = 25
  default_quota_max_cidr_count  = -1
  default_quota_max_ip_count    = -1

  cidr_blocks {
    name = "{{.IpScopePrefix}}-3"
    cidr = "12.0.0.0/27"
  }

  ip_address_ranges {
    start_ip_address = "14.0.0.1"
    end_ip_address   = "14.0.0.255"
  }
}
`

const testAccVcfaIpSpaceStep3DS = testAccVcfaIpSpaceStep2 + `
data "vcfa_ip_space" "test" {
  name      = vcfa_ip_space.test.name
  region_id = {{.RegionId}}

  depends_on = [ vcfa_ip_space.test ]
}
`

func hashAndTakeFirstSix(str string) string {
	hash := sha256.Sum256([]byte(str))
	hashStr := hex.EncodeToString(hash[:])
	if len(hashStr) < 6 {
		hashStr = fmt.Sprintf("%06s", hashStr)
	}
	return hashStr[:6]
}
