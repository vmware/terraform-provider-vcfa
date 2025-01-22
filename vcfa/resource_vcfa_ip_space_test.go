//go:build tm || ALL || functional

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaIpSpace(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	var params = StringMap{
		"Testname":   t.Name(),
		"VcenterRef": vCenterHclRef,
		"RegionId":   fmt.Sprintf("%s.id", regionHclRef),
		"RegionName": t.Name(),

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
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "description", "description test"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "external_scope", "12.12.0.0/16"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_subnet_size", "24"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_cidr_count", "1"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_ip_count", "1"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "internal_scope.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "internal_scope.*", map[string]string{
						"name": "scope1",
						"cidr": "10.0.0.0/24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "internal_scope.*", map[string]string{
						"cidr": "11.0.0.0/26",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "internal_scope.*", map[string]string{
						"name": "scope3",
						"cidr": "12.0.0.0/27",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					cachedIpSpaceId.testCheckCachedResourceFieldValue("vcfa_ip_space.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_ip_space.test", "id"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "status", "REALIZED"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "description", "description test - update"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "external_scope", "12.12.0.0/20"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_subnet_size", "25"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_cidr_count", "-1"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "default_quota_max_ip_count", "-1"),
					resource.TestCheckResourceAttr("vcfa_ip_space.test", "internal_scope.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcfa_ip_space.test", "internal_scope.*", map[string]string{
						"name": "scope3",
						"cidr": "12.0.0.0/27",
					}),
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

	postTestChecks(t)
}

const testAccVcfaIpSpaceStep1 = `
resource "vcfa_ip_space" "test" {
  name                          = "{{.Testname}}"
  description                   = "description test"
  region_id                     = {{.RegionId}}
  external_scope                = "12.12.0.0/16"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1

  internal_scope {
    name = "scope1"
    cidr = "10.0.0.0/24"
  }

  internal_scope {
    cidr = "11.0.0.0/26"
  }

  internal_scope {
    name = "scope3"
    cidr = "12.0.0.0/27"
  }
}
`

const testAccVcfaIpSpaceStep2 = `
resource "vcfa_ip_space" "test" {
  name                          = "{{.Testname}}-updated"
  description                   = "description test - update"
  region_id                     = {{.RegionId}}
  external_scope                = "12.12.0.0/20"
  default_quota_max_subnet_size = 25
  default_quota_max_cidr_count  = -1
  default_quota_max_ip_count    = -1

  internal_scope {
    name = "scope3"
    cidr = "12.0.0.0/27"
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
