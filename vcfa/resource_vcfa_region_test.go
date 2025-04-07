/*
 * // © Broadcom. All Rights Reserved.
 * // The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
 * // SPDX-License-Identifier: MPL-2.0
 */

//go:build tm || region || ALL || functional

package vcfa

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaRegion(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	vCenterHcl, vCenterHclRef := getVCenterHcl(t, nsxManagerHclRef)

	var params = StringMap{
		"Testname":   t.Name(),
		"RegionName": strings.ToLower(t.Name()), // to match 'rfc1123LabelNameRegex'

		"VcenterRefId":    fmt.Sprintf("%s.id", vCenterHclRef),
		"NsxManagerRefId": fmt.Sprintf("%s.id", nsxManagerHclRef),

		"NsxManagerUsername": testConfig.Tm.NsxManagerUsername,
		"NsxManagerPassword": testConfig.Tm.NsxManagerPassword,
		"NsxManagerUrl":      testConfig.Tm.NsxManagerUrl,

		"VcenterUsername":       testConfig.Tm.VcenterUsername,
		"VcenterPassword":       testConfig.Tm.VcenterPassword,
		"VcenterUrl":            testConfig.Tm.VcenterUrl,
		"VcenterStorageProfile": testConfig.Tm.VcenterStorageProfile,
		"VcenterSupervisor":     testConfig.Tm.VcenterSupervisor,
		"VcenterSupervisorZone": testConfig.Tm.VcenterSupervisorZone,

		"Tags": "tm region",
	}
	testParamsNotEmpty(t, params)

	prerequisites := vCenterHcl + nsxManagerHcl

	configText1 := templateFill(prerequisites+testAccVcfaRegionStep1, params)
	params["FuncName"] = t.Name() + "-step1"
	configText2 := templateFill(prerequisites+testAccVcfaRegionStep2, params)
	params["FuncName"] = t.Name() + "-step2"
	configText3 := templateFill(prerequisites+testAccVcfaRegionStep3DS, params)
	params["FuncName"] = t.Name() + "-step3"

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)

	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedRegionId := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_region.test", "id"),
					cachedRegionId.cacheTestResourceFieldValue("vcfa_region.test", "id"),
					resource.TestCheckResourceAttr("vcfa_region.test", "name", params["RegionName"].(string)),
					resource.TestCheckResourceAttr("vcfa_region.test", "description", "Terraform description"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "cpu_reservation_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "memory_reservation_capacity_mib"),
					resource.TestCheckResourceAttr("vcfa_region.test", "status", "READY"),
					resource.TestCheckResourceAttr("vcfa_region.test", "storage_policy_names.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcfa_region.test", "storage_policy_names.*", testConfig.Tm.VcenterStorageProfile),

					resource.TestCheckResourceAttrSet("data.vcfa_supervisor.test", "id"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "id"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "cpu_used_mhz"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "memory_used_mib"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_region.test", "id"),
					cachedRegionId.testCheckCachedResourceFieldValue("vcfa_region.test", "id"),
					resource.TestCheckResourceAttr("vcfa_region.test", "name", params["RegionName"].(string)),
					resource.TestCheckResourceAttr("vcfa_region.test", "description", "Terraform description updated"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "cpu_reservation_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "memory_reservation_capacity_mib"),
					resource.TestCheckResourceAttr("vcfa_region.test", "status", "READY"),
					resource.TestCheckResourceAttr("vcfa_region.test", "storage_policy_names.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcfa_region.test", "storage_policy_names.*", testConfig.Tm.VcenterStorageProfile),

					resource.TestCheckResourceAttrSet("data.vcfa_supervisor.test", "id"),

					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "id"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "cpu_used_mhz"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "memory_used_mib"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_region.test", "data.vcfa_region.test",
						[]string{"memory_reservation_capacity_mib", "memory_capacity_mib"}, // these values fluctuate
					),
				),
			},
			{
				ResourceName:      "vcfa_region.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["RegionName"].(string),
				ImportStateVerifyIgnore: []string{
					"memory_capacity_mib",             // Field value slightly fluctuates over test duration
					"memory_reservation_capacity_mib", // Field value slightly fluctuates over test duration
				},
			},
		},
	})
}

const testAccVcfaRegionPrerequisites = `
data "vcfa_supervisor" "test" {
  name       = "{{.VcenterSupervisor}}"
  vcenter_id = {{.VcenterRefId}}
}

data "vcfa_supervisor_zone" "test" {
  supervisor_id = data.vcfa_supervisor.test.id
  name          = "{{.VcenterSupervisorZone}}"
}
`

const testAccVcfaRegionStep1 = testAccVcfaRegionPrerequisites + `
resource "vcfa_region" "test" {
  name                 = "{{.RegionName}}"
  description          = "Terraform description"
  nsx_manager_id       = {{.NsxManagerRefId}}
  supervisor_ids       = [data.vcfa_supervisor.test.id]
  storage_policy_names = ["{{.VcenterStorageProfile}}"]
}
`

const testAccVcfaRegionStep2 = testAccVcfaRegionPrerequisites + `
# skip-binary-test: update test
resource "vcfa_region" "test" {
  name                 = "{{.RegionName}}"
  description          = "Terraform description updated"
  nsx_manager_id       = {{.NsxManagerRefId}}
  supervisor_ids       = [data.vcfa_supervisor.test.id]
  storage_policy_names = ["{{.VcenterStorageProfile}}"]
}
`

const testAccVcfaRegionStep3DS = testAccVcfaRegionStep2 + `
data "vcfa_region" "test" {
  name = vcfa_region.test.name
}
`
