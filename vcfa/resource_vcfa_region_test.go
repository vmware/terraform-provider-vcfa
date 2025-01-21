//go:build tm || region || ALL || functional

package vcfa

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TODO: TM: the test has an update, but it just recreates the resource behind the scenes now
// as the API does not support update yet
func TestAccVcfaRegion(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Testname":           t.Name(),
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

	configText1 := templateFill(testAccVcfaRegionStep1, params)
	params["FuncName"] = t.Name() + "-step1"
	configText2 := templateFill(testAccVcfaRegionStep2, params)
	params["FuncName"] = t.Name() + "-step2"
	configText3 := templateFill(testAccVcfaRegionStep3DS, params)
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
					resource.TestMatchResourceAttr("vcfa_nsx_manager.test", "id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:`)),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "id"),
					cachedRegionId.cacheTestResourceFieldValue("vcfa_region.test", "id"),
					resource.TestCheckResourceAttr("vcfa_region.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcfa_region.test", "description", "Terraform description"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "cpu_reservation_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "memory_reservation_capacity_mib"),
					resource.TestCheckResourceAttr("vcfa_region.test", "status", "READY"),
					resource.TestCheckResourceAttr("vcfa_region.test", "storage_policy_names.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcfa_region.test", "storage_policy_names.*", testConfig.Tm.VcenterStorageProfile),

					resource.TestCheckResourceAttrSet("data.vcfa_supervisor.test", "id"),
					resource.TestCheckResourceAttrPair("data.vcfa_supervisor.test", "vcenter_id", "vcfa_vcenter.test", "id"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "id"),
					resource.TestCheckResourceAttrPair("data.vcfa_supervisor_zone.test", "vcenter_id", "vcfa_vcenter.test", "id"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "cpu_used_mhz"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "memory_used_mib"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcfa_nsx_manager.test", "id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:`)),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "id"),
					cachedRegionId.testCheckCachedResourceFieldValueChanged("vcfa_region.test", "id"),
					resource.TestCheckResourceAttr("vcfa_region.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcfa_region.test", "description", "Terraform description updated"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "cpu_reservation_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("vcfa_region.test", "memory_reservation_capacity_mib"),
					resource.TestCheckResourceAttr("vcfa_region.test", "status", "READY"),
					resource.TestCheckResourceAttr("vcfa_region.test", "storage_policy_names.#", "1"),
					resource.TestCheckTypeSetElemAttr("vcfa_region.test", "storage_policy_names.*", testConfig.Tm.VcenterStorageProfile),

					resource.TestCheckResourceAttrSet("data.vcfa_supervisor.test", "id"),
					resource.TestCheckResourceAttrPair("data.vcfa_supervisor.test", "vcenter_id", "vcfa_vcenter.test", "id"),

					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "id"),
					resource.TestCheckResourceAttrPair("data.vcfa_supervisor_zone.test", "vcenter_id", "vcfa_vcenter.test", "id"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "cpu_used_mhz"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("data.vcfa_supervisor_zone.test", "memory_used_mib"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_region.test", "data.vcfa_region.test", []string{
						"is_enabled", // TODO: TM: field is not populated on read
					}),
				),
			},
			{
				ResourceName:      "vcfa_region.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Testname"].(string),
				ImportStateVerifyIgnore: []string{
					"is_enabled", // TODO: TM: field is not populated on read
				},
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaRegionPrerequisites = `
resource "vcfa_nsx_manager" "test" {
  name                   = "{{.Testname}}"
  description            = "terraform test"
  username               = "{{.NsxManagerUsername}}"
  password               = "{{.NsxManagerPassword}}"
  url                    = "{{.NsxManagerUrl}}"
  network_provider_scope = ""
  auto_trust_certificate = true
}

resource "vcfa_vcenter" "test" {
  name                     = "{{.Testname}}"
  url                      = "{{.VcenterUrl}}"
  auto_trust_certificate   = true
  refresh_vcenter_on_read  = true
  refresh_policies_on_read = true
  username                 = "{{.VcenterUsername}}"
  password                 = "{{.VcenterPassword}}"
  is_enabled               = true
}

data "vcfa_supervisor" "test" {
  name       = "{{.VcenterSupervisor}}"
  vcenter_id = vcfa_vcenter.test.id

  depends_on = [vcfa_vcenter.test]
}

data "vcfa_supervisor_zone" "test" {
  supervisor_id = data.vcfa_supervisor.test.id
  name          = "{{.VcenterSupervisorZone}}"
}
`

const testAccVcfaRegionStep1 = testAccVcfaRegionPrerequisites + `
resource "vcfa_region" "test" {
  name                 = "{{.Testname}}"
  description          = "Terraform description"
  is_enabled           = true
  nsx_manager_id       = vcfa_nsx_manager.test.id
  supervisor_ids       = [data.vcfa_supervisor.test.id]
  storage_policy_names = ["{{.VcenterStorageProfile}}"]
}
`

const testAccVcfaRegionStep2 = testAccVcfaRegionPrerequisites + `
# skip-binary-test: update test
resource "vcfa_region" "test" {
  name                 = "{{.Testname}}"
  description          = "Terraform description updated"
  is_enabled           = true
  nsx_manager_id       = vcfa_nsx_manager.test.id
  supervisor_ids       = [data.vcfa_supervisor.test.id]
  storage_policy_names = ["{{.VcenterStorageProfile}}"]
}
`

const testAccVcfaRegionStep3DS = testAccVcfaRegionStep2 + `
data "vcfa_region" "test" {
  name = vcfa_region.test.name
}
`
