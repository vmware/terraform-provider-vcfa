//go:build tm || contentlibrary || ALL || functional

package vcfa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TODO: TM: Test tenant Content Libraries. For that goal, we need to modify vcfa_org_region_quota to support
//       VM classes and storage classes

// TestAccVcfaContentLibraryProvider tests CRUD of a Content Library of type PROVIDER.
// It also tests vcfa_storage_class and vcfa_region_storage_policy data sources
func TestAccVcfaContentLibraryProvider(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)

	var params = StringMap{
		"Name":                t.Name(),
		"RegionId":            fmt.Sprintf("%s.id", regionHclRef),
		"RegionStoragePolicy": testConfig.Tm.StorageClass,
		"Tags":                "tm contentlibrary",
	}
	testParamsNotEmpty(t, params)

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	configText1 := templateFill(preRequisites+testAccVcfaContentLibraryProviderStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	params["Name"] = t.Name() + "Updated"
	configText2 := templateFill(preRequisites+testAccVcfaContentLibraryProviderStep1, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaContentLibraryProviderStep3, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcfa_content_library.cl"
	dsRegionStoragePolicy := "data.vcfa_region_storage_policy.sp"
	dsStorageClass := "data.vcfa_storage_class.sc"

	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					// Region Storage Policy
					resource.TestCheckResourceAttr(dsRegionStoragePolicy, "name", testConfig.Tm.StorageClass),
					resource.TestCheckResourceAttrPair(dsRegionStoragePolicy, "region_id", regionHclRef, "id"),
					resource.TestMatchResourceAttr(dsRegionStoragePolicy, "description", regexp.MustCompile(`.*`)),
					resource.TestCheckResourceAttr(dsRegionStoragePolicy, "status", ""),
					resource.TestCheckResourceAttrSet(dsRegionStoragePolicy, "storage_capacity_mb"),
					resource.TestCheckResourceAttrSet(dsRegionStoragePolicy, "storage_consumed_mb"),

					// Storage Class
					resource.TestCheckResourceAttr(dsStorageClass, "name", testConfig.Tm.StorageClass),
					resource.TestCheckResourceAttrPair(dsStorageClass, "region_id", regionHclRef, "id"),
					resource.TestCheckResourceAttrSet(dsStorageClass, "storage_capacity_mib"),
					resource.TestCheckResourceAttrSet(dsStorageClass, "storage_consumed_mib"),
					resource.TestMatchResourceAttr(dsStorageClass, "zone_ids.#", regexp.MustCompile("[0-9]+")),

					// Content Library
					cachedId.cacheTestResourceFieldValue(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestMatchResourceAttr(resourceName, "org_id", regexp.MustCompile("urn:vcloud:org:")),
					resource.TestCheckResourceAttr(resourceName, "description", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "storage_class_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_attach", "true"), // TODO: TM: Test with false
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "is_shared", "true"),        // TODO: TM: Test with false
					resource.TestCheckResourceAttr(resourceName, "is_subscribed", "false"),   // TODO: TM: Test with true
					resource.TestCheckResourceAttr(resourceName, "library_type", "PROVIDER"), // TODO: TM: Test with tenant catalog
					resource.TestCheckResourceAttr(resourceName, "subscription_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_number", "1"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()+"Updated"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual(resourceName, "data.vcfa_content_library.cl_ds", []string{
						"%", // Does not have delete_recursive, delete_force
						"delete_recursive",
						"delete_force",
					}),
				),
			},
			{
				ResourceName:      "vcfa_content_library.cl",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Name"].(string),
				ImportStateVerifyIgnore: []string{
					"delete_recursive",
					"delete_force",
				},
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaContentLibraryProviderStep1 = `
data "vcfa_region_storage_policy" "sp" {
  region_id = {{.RegionId}}
  name      = "{{.RegionStoragePolicy}}"
}

data "vcfa_storage_class" "sc" {
  region_id = {{.RegionId}}
  name      = "{{.RegionStoragePolicy}}"
}

resource "vcfa_content_library" "cl" {
  name = "{{.Name}}"
  description = "{{.Name}}"
  storage_class_ids = [
    data.vcfa_storage_class.sc.id
  ]
  delete_force = true
  delete_recursive = true
}
`

const testAccVcfaContentLibraryProviderStep3 = testAccVcfaContentLibraryProviderStep1 + `
data "vcfa_content_library" "cl_ds" {
  name = vcfa_content_library.cl.name
}
`
