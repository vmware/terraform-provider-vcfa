//go:build tm || org || vdc || ALL || functional

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaOrgVdc(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	var params = StringMap{
		"Testname":           t.Name(),
		"SupervisorName":     testConfig.Tm.VcenterSupervisor,
		"SupervisorZoneName": testConfig.Tm.VcenterSupervisorZone,
		"VcenterRef":         vCenterHclRef,
		"RegionId":           fmt.Sprintf("%s.id", regionHclRef),
		"Tags":               "tm org vdc",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl
	configText1 := templateFill(preRequisites+testAccVcfaOrgVdcStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcfaOrgVdcStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaOrgVdcStep3DS, params)

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
					resource.TestCheckResourceAttrSet("vcfa_org_vdc.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_vdc.test", "name", fmt.Sprintf("%s_%s", params["Testname"], testConfig.Tm.Region)), // Name is a combination of Org name + Region name
					resource.TestCheckResourceAttr("vcfa_org_vdc.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcfa_org_vdc.test", "status", "READY"),
					resource.TestCheckResourceAttrPair("vcfa_org_vdc.test", "org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttrPair("vcfa_org_vdc.test", "region_id", regionHclRef, "id"),
					resource.TestCheckResourceAttr("vcfa_org_vdc.test", "supervisor_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("vcfa_org_vdc.test", "supervisor_ids.*", "data.vcfa_supervisor.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_vdc.test", "zone_resource_allocations.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_org_vdc.test", "zone_resource_allocations.*", map[string]string{
						"region_zone_name":       testConfig.Tm.VcenterSupervisorZone,
						"cpu_limit_mhz":          "2000",
						"cpu_reservation_mhz":    "100",
						"memory_limit_mib":       "1024",
						"memory_reservation_mib": "512",
					}),
				),
			},
			// TODO: TM: Update throws a NullPointerException when trying to modify Region Zone allocations
			//{
			//	Config: configText2,
			//	Check: resource.ComposeTestCheckFunc(
			//		resource.TestCheckResourceAttrSet("vcfa_org_vdc.test", "id"),
			//		resource.TestCheckResourceAttr("vcfa_org_vdc.test", "name", fmt.Sprintf("%s_%s", params["Testname"], testConfig.Tm.Region)), // Name is a combination of Org name + Region name
			//		resource.TestCheckResourceAttr("vcfa_org_vdc.test", "is_enabled", "true"),
			//		resource.TestCheckResourceAttr("vcfa_org_vdc.test", "status", "READY"),
			//		resource.TestCheckResourceAttrPair("vcfa_org_vdc.test", "org_id", "vcfa_org.test", "id"),
			//		resource.TestCheckResourceAttrPair("vcfa_org_vdc.test", "region_id", "vcfa_region.region", "id"),
			//		resource.TestCheckResourceAttr("vcfa_org_vdc.test", "supervisor_ids.#", "1"),
			//		resource.TestCheckTypeSetElemAttrPair("vcfa_org_vdc.test", "supervisor_ids.*", "data.vcfa_supervisor.test", "id"),
			//		resource.TestCheckResourceAttr("vcfa_org_vdc.test", "zone_resource_allocations.#", "1"),
			//		resource.TestCheckTypeSetElemNestedAttrs("vcfa_org_vdc.test", "zone_resource_allocations.*", map[string]string{
			//			"region_zone_name":       testConfig.Tm.VcenterSupervisorZone,
			//			"cpu_limit_mhz":          "1900",
			//			"cpu_reservation_mhz":    "90",
			//			"memory_limit_mib":       "500",
			//			"memory_reservation_mib": "200",
			//		}),
			//	),
			//},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_org_vdc.test", "data.vcfa_org_vdc.test", []string{
						"is_enabled", // TODO: TM: is_enabled is always returned as false
					}),
				),
			},
			{
				ResourceName:      "vcfa_org_vdc.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s%s%s", params["Testname"], ImportSeparator, testConfig.Tm.Region), // Org name and Region name
				ImportStateVerifyIgnore: []string{
					"is_enabled", // TODO: TM: field is not populated on read
				},
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaOrgVdcStep1 = `
data "vcfa_supervisor" "test" {
  name       = "{{.SupervisorName}}"
  vcenter_id = {{.VcenterRef}}.id
}

data "vcfa_region_zone" "test" {
  region_id = {{.RegionId}}
  name      = "{{.SupervisorZoneName}}"
}

resource "vcfa_org_vdc" "test" {
  org_id         = vcfa_org.test.id
  region_id      = {{.RegionId}}
  supervisor_ids = [data.vcfa_supervisor.test.id]
  zone_resource_allocations {
    region_zone_id         = data.vcfa_region_zone.test.id
    cpu_limit_mhz          = 2000
    cpu_reservation_mhz    = 100
    memory_limit_mib       = 1024
    memory_reservation_mib = 512
  }
}

resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}
`

const testAccVcfaOrgVdcStep2 = `
data "vcfa_supervisor" "test" {
  name       = "{{.SupervisorName}}"
  vcenter_id = {{.VcenterRef}}.id
  depends_on = [{{.VcenterRef}}]
}

data "vcfa_region_zone" "test" {
  region_id = {{.RegionId}}
  name      = "{{.SupervisorZoneName}}"
}

resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}

resource "vcfa_org_vdc" "test" {
  org_id         = vcfa_org.test.id
  region_id      = {{.RegionId}}
  supervisor_ids = [data.vcfa_supervisor.test.id]
  zone_resource_allocations {
    region_zone_id         = data.vcfa_region_zone.test.id
    cpu_limit_mhz          = 1900
    cpu_reservation_mhz    = 90
    memory_limit_mib       = 500
    memory_reservation_mib = 200
  }
}
`

// TODO: TM: Change to testAccVcfaOrgVdcStep2 when Update is fixed
const testAccVcfaOrgVdcStep3DS = testAccVcfaOrgVdcStep1 + `
data "vcfa_org_vdc" "test" {
  org_id    = vcfa_org.test.id
  region_id = {{.RegionId}}
}
`
