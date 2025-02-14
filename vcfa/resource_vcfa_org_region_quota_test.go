//go:build tm || org || regionQuota || ALL || functional

package vcfa

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TODO: TM: Improve this test so it has more than 1 Region Storage Policy
func TestAccVcfaOrgRegionQuota(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	vmClassesHcl, vmClassesRefs := getRegionVmClassesHcl(t, regionHclRef)
	var params = StringMap{
		"Testname":           t.Name(),
		"SupervisorName":     testConfig.Tm.VcenterSupervisor,
		"SupervisorZoneName": testConfig.Tm.VcenterSupervisorZone,
		"VcenterRef":         vCenterHclRef,
		"RegionId":           fmt.Sprintf("%s.id", regionHclRef),
		"RegionVmClassRefs":  strings.Join(vmClassesRefs, ".id,\n    ") + ".id",
		"StorageClass":       testConfig.Tm.StorageClass,
		"StorageLimitMib":    "100",
		"Tags":               "tm org regionQuota",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + vmClassesHcl
	configText1 := templateFill(preRequisites+testAccVcfaOrgRegionQuotaStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	params["RegionVmClassRefs"] = fmt.Sprintf("%s.id", vmClassesRefs[0]) // There is always at least one VM class in config
	configText2 := templateFill(preRequisites+testAccVcfaOrgRegionQuotaStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaOrgRegionQuotaStep3DS, params)

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
					resource.TestCheckResourceAttrSet("vcfa_org_region_quota.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "name", fmt.Sprintf("%s_%s", params["Testname"], testConfig.Tm.Region)), // Name is a combination of Org name + Region name
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "status", "READY"),
					resource.TestCheckResourceAttrPair("vcfa_org_region_quota.test", "org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttrPair("vcfa_org_region_quota.test", "region_id", regionHclRef, "id"),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "supervisor_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("vcfa_org_region_quota.test", "supervisor_ids.*", "data.vcfa_supervisor.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "zone_resource_allocations.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_org_region_quota.test", "zone_resource_allocations.*", map[string]string{
						"region_zone_name":       testConfig.Tm.VcenterSupervisorZone,
						"cpu_limit_mhz":          "2000",
						"cpu_reservation_mhz":    "100",
						"memory_limit_mib":       "1024",
						"memory_reservation_mib": "512",
					}),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "region_vm_class_ids.#", fmt.Sprintf("%d", len(vmClassesRefs))),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_org_region_quota.test", "region_storage_policy.*", map[string]string{
						"storage_limit_mib": "100",
						"name":              params["StorageClass"].(string),
					}),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "region_storage_policy.#", "1"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_org_region_quota.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "name", fmt.Sprintf("%s_%s", params["Testname"], testConfig.Tm.Region)), // Name is a combination of Org name + Region name
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "status", "READY"),
					resource.TestCheckResourceAttrPair("vcfa_org_region_quota.test", "org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttrPair("vcfa_org_region_quota.test", "region_id", regionHclRef, "id"),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "supervisor_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("vcfa_org_region_quota.test", "supervisor_ids.*", "data.vcfa_supervisor.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "zone_resource_allocations.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_org_region_quota.test", "zone_resource_allocations.*", map[string]string{
						"region_zone_name":       testConfig.Tm.VcenterSupervisorZone,
						"cpu_limit_mhz":          "1900",
						"cpu_reservation_mhz":    "90",
						"memory_limit_mib":       "500",
						"memory_reservation_mib": "200",
					}),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "region_vm_class_ids.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_org_region_quota.test", "region_storage_policy.*", map[string]string{
						"storage_limit_mib": "77",
						"name":              params["StorageClass"].(string),
					}),
					resource.TestCheckResourceAttr("vcfa_org_region_quota.test", "region_storage_policy.#", "1"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_org_region_quota.test", "data.vcfa_org_region_quota.test", nil),
				),
			},
			{
				ResourceName:      "vcfa_org_region_quota.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s%s%s", params["Testname"], ImportSeparator, testConfig.Tm.Region), // Org name and Region name
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaOrgRegionQuotaStep1 = `
data "vcfa_supervisor" "test" {
  name       = "{{.SupervisorName}}"
  vcenter_id = {{.VcenterRef}}.id
}

data "vcfa_region_zone" "test" {
  region_id = {{.RegionId}}
  name      = "{{.SupervisorZoneName}}"
}

data "vcfa_region_storage_policy" "sp" {
  name      = "{{.StorageClass}}"
  region_id = {{.RegionId}}
}

resource "vcfa_org_region_quota" "test" {
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
  region_vm_class_ids = [
    {{.RegionVmClassRefs}}
  ]
  region_storage_policy {
    region_storage_policy_id = data.vcfa_region_storage_policy.sp.id
    storage_limit_mib        = 100
  }
}

resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}
`

const testAccVcfaOrgRegionQuotaStep2 = `
data "vcfa_supervisor" "test" {
  name       = "{{.SupervisorName}}"
  vcenter_id = {{.VcenterRef}}.id
  depends_on = [{{.VcenterRef}}]
}

data "vcfa_region_zone" "test" {
  region_id = {{.RegionId}}
  name      = "{{.SupervisorZoneName}}"
}

data "vcfa_region_storage_policy" "sp" {
  name      = "{{.StorageClass}}"
  region_id = {{.RegionId}}
}

resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}

resource "vcfa_org_region_quota" "test" {
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
  region_vm_class_ids = [
    {{.RegionVmClassRefs}}
  ]
  region_storage_policy {
    region_storage_policy_id = data.vcfa_region_storage_policy.sp.id
    storage_limit_mib        = 77
  }
}
`

const testAccVcfaOrgRegionQuotaStep3DS = testAccVcfaOrgRegionQuotaStep2 + `
data "vcfa_org_region_quota" "test" {
  org_id    = vcfa_org.test.id
  region_id = {{.RegionId}}
}
`
