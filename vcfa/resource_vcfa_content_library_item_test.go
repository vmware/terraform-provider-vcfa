//go:build tm || contentlibrary || ALL || functional

package vcfa

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var contentLibraryItemTestingResourcePaths = []string{
	"../test-resources/test_vapp_template.ova",
	"../test-resources/test.iso",
	"../test-resources/test_vapp_template_ovf/descriptor.ovf",
	"../test-resources/test_vapp_template_ovf/disk1.vmdk"}

// getTestingResourcesAbsolutePaths returns the absolute paths to the testing resources
// required to test. Absolute paths are required when running binary tests.
func getTestingResourcesAbsolutePaths(t *testing.T, relativePaths []string) []string {
	absPaths := make([]string, len(relativePaths))
	for i, s := range relativePaths {
		absPath, err := filepath.Abs(s)
		if err != nil {
			t.Fatal(err)
		}
		absPaths[i] = absPath
	}
	return absPaths
}

// TestAccVcfaContentLibraryItemProvider tests Content Library Items in a "PROVIDER" type Content Library
func TestAccVcfaContentLibraryItemProvider(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	vCenterHcl, vCenterHclRef := getVCenterHcl(t, nsxManagerHclRef)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	contentLibraryHcl, contentLibraryHclRef := getContentLibraryHcl(t, regionHclRef, "")

	itemPaths := getTestingResourcesAbsolutePaths(t, contentLibraryItemTestingResourcePaths)
	var params = StringMap{
		"Name":              t.Name(),
		"ContentLibraryRef": fmt.Sprintf("%s.id", contentLibraryHclRef),
		"OvaPath":           itemPaths[0],
		"IsoPath":           itemPaths[1],
		"OvfPaths":          fmt.Sprintf("\"%s\", \"%s\"", itemPaths[2], itemPaths[3]),
		"Tags":              "tm contentlibrary",
	}
	testParamsNotEmpty(t, params)

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + contentLibraryHcl

	configText1 := templateFill(preRequisites+testAccVcfaContentLibraryItemProviderStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	params["Name"] = t.Name() + "Updated"
	configText2 := templateFill(preRequisites+testAccVcfaContentLibraryItemProviderStep1, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaContentLibraryItemProviderStep3, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cli1 := "vcfa_content_library_item.cli1"
	cli2 := "vcfa_content_library_item.cli2"
	cli3 := "vcfa_content_library_item.cli3"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// CLI 1: OVA
					resource.TestCheckResourceAttr(cli1, "name", t.Name()+"1"),
					resource.TestCheckResourceAttr(cli1, "description", t.Name()+"1"),
					resource.TestCheckResourceAttrPair(cli1, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli1, "creation_date"),
					resource.TestCheckResourceAttr(cli1, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli1, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli1, "image_identifier"),
					resource.TestCheckResourceAttr(cli1, "item_type", "TEMPLATE"),
					resource.TestCheckResourceAttrPair(cli1, "owner_org_id", "data.vcfa_org.system", "id"),
					resource.TestCheckResourceAttr(cli1, "status", "READY"),
					resource.TestCheckResourceAttr(cli1, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli1, "version", "1"),

					// CLI 2: ISO
					resource.TestCheckResourceAttr(cli2, "name", t.Name()+"2"),
					resource.TestCheckResourceAttr(cli2, "description", t.Name()+"2"),
					resource.TestCheckResourceAttrPair(cli2, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli2, "creation_date"),
					resource.TestCheckResourceAttr(cli2, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli2, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli2, "image_identifier"),
					resource.TestCheckResourceAttr(cli2, "item_type", "ISO"),
					resource.TestCheckResourceAttrPair(cli2, "owner_org_id", "data.vcfa_org.system", "id"),
					resource.TestCheckResourceAttr(cli2, "status", "READY"),
					resource.TestCheckResourceAttr(cli2, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli2, "version", "1"),

					// CLI 3: OVF
					resource.TestCheckResourceAttr(cli3, "name", t.Name()+"3"),
					resource.TestCheckResourceAttr(cli3, "description", t.Name()+"3"),
					resource.TestCheckResourceAttrPair(cli3, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli3, "creation_date"),
					resource.TestCheckResourceAttr(cli3, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli3, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli3, "image_identifier"),
					resource.TestCheckResourceAttr(cli3, "item_type", "TEMPLATE"),
					resource.TestCheckResourceAttrPair(cli3, "owner_org_id", "data.vcfa_org.system", "id"),
					resource.TestCheckResourceAttr(cli3, "status", "READY"),
					resource.TestCheckResourceAttr(cli3, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli3, "version", "1"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// CLI 1: OVA
					resource.TestCheckResourceAttr(cli1, "name", t.Name()+"Updated1"),
					resource.TestCheckResourceAttr(cli1, "description", t.Name()+"Updated1"),

					// CLI 2: ISO
					resource.TestCheckResourceAttr(cli2, "name", t.Name()+"Updated2"),
					resource.TestCheckResourceAttr(cli2, "description", t.Name()+"Updated2"),

					// CLI 3: OVF
					resource.TestCheckResourceAttr(cli3, "name", t.Name()+"Updated3"),
					resource.TestCheckResourceAttr(cli3, "description", t.Name()+"Updated3"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// file_paths and upload_piece_size cannot be obtained during reads, that's why it does not appear in data source schema
					resourceFieldsEqual(cli1, "data.vcfa_content_library_item.cli1_ds", []string{"file_paths.#", "file_paths.0", "upload_piece_size", "%"}),
					resourceFieldsEqual(cli2, "data.vcfa_content_library_item.cli2_ds", []string{"file_paths.#", "file_paths.0", "upload_piece_size", "%"}),
					resourceFieldsEqual(cli3, "data.vcfa_content_library_item.cli3_ds", []string{"file_paths.#", "file_paths.0", "file_paths.1", "upload_piece_size", "%"}),
				),
			},
			{
				ResourceName:            cli1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           fmt.Sprintf("System%s%s%s%s", ImportSeparator, testConfig.Tm.ContentLibrary, ImportSeparator, params["Name"].(string)+"1"),
				ImportStateVerifyIgnore: []string{"file_paths.#", "file_paths.0", "upload_piece_size", "%"}, // file_paths and upload_piece_size cannot be obtained during imports, that's why it's Optional
			},
		},
	})
}

const testAccVcfaContentLibraryItemProviderStep1 = `
resource "vcfa_content_library_item" "cli1" {
  name               = "{{.Name}}1"
  description        = "{{.Name}}1"
  content_library_id = {{.ContentLibraryRef}}
  file_paths         = ["{{.OvaPath}}"]
}

resource "vcfa_content_library_item" "cli2" {
  name               = "{{.Name}}2"
  description        = "{{.Name}}2"
  content_library_id = {{.ContentLibraryRef}}
  file_paths         = ["{{.IsoPath}}"]
}

resource "vcfa_content_library_item" "cli3" {
  name               = "{{.Name}}3"
  description        = "{{.Name}}3"
  content_library_id = {{.ContentLibraryRef}}
  file_paths         = [{{.OvfPaths}}]
}
`

const testAccVcfaContentLibraryItemProviderStep3 = testAccVcfaContentLibraryItemProviderStep1 + `
data "vcfa_content_library_item" "cli1_ds" {
  name               = vcfa_content_library_item.cli1.name
  content_library_id = vcfa_content_library_item.cli1.content_library_id
}
data "vcfa_content_library_item" "cli2_ds" {
  name               = vcfa_content_library_item.cli2.name
  content_library_id = vcfa_content_library_item.cli2.content_library_id
}
data "vcfa_content_library_item" "cli3_ds" {
  name               = vcfa_content_library_item.cli3.name
  content_library_id = vcfa_content_library_item.cli3.content_library_id
}
`

// TestAccVcfaContentLibraryItemTenant tests Content Library Items in a "TENANT" type Content Library
func TestAccVcfaContentLibraryItemTenant(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	vCenterHcl, vCenterHclRef := getVCenterHcl(t, nsxManagerHclRef)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	vmClassesHcl, vmClassesRefs := getRegionVmClassesHcl(t, regionHclRef)
	// The Content Library for tenants must depend on the Region Quota, as it contains the Storage Policies required
	// to create libraries in the Organization
	contentLibraryHcl, contentLibraryHclRef := getContentLibraryHcl(t, regionHclRef, "vcfa_org_region_quota.test.org_id")

	itemPaths := getTestingResourcesAbsolutePaths(t, contentLibraryItemTestingResourcePaths)
	var params = StringMap{
		"Org":                 testConfig.Tm.Org,
		"Username":            "test-user",
		"Password":            "long-change-ME1",
		"Name":                t.Name(),
		"RegionId":            fmt.Sprintf("%s.id", regionHclRef),
		"SupervisorName":      testConfig.Tm.VcenterSupervisor,
		"SupervisorZoneName":  testConfig.Tm.VcenterSupervisorZone,
		"StorageClass":        testConfig.Tm.StorageClass,
		"VcenterRef":          vCenterHclRef,
		"RegionStoragePolicy": testConfig.Tm.StorageClass,
		"RegionVmClassRefs":   strings.Join(vmClassesRefs, ".id,\n    ") + ".id",
		"VcfaUrl":             testConfig.Provider.Url,
		"ContentLibraryRef":   fmt.Sprintf("%s.id", contentLibraryHclRef),
		"OvaPath":             itemPaths[0],
		"IsoPath":             itemPaths[1],
		"OvfPaths":            fmt.Sprintf("\"%s\", \"%s\"", itemPaths[2], itemPaths[3]),
		"Tags":                "tm contentlibrary",
	}
	testParamsNotEmpty(t, params)

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + vmClassesHcl + contentLibraryHcl

	configText1 := templateFill(preRequisites+testAccVcfaContentLibraryTenantPrerequisites+testAccVcfaContentLibraryItemProviderStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	params["Name"] = t.Name() + "Updated"
	configText2 := templateFill(preRequisites+testAccVcfaContentLibraryTenantPrerequisites+testAccVcfaContentLibraryItemProviderStep1, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaContentLibraryTenantPrerequisites+testAccVcfaContentLibraryItemProviderStep3, params)
	params["FuncName"] = t.Name() + "-step4"
	params["Name"] = t.Name()
	configText4 := templateFill(preRequisites+testAccVcfaContentLibraryTenantPrerequisites+testAccVcfaContentLibraryItemTenantStep4, params)
	params["FuncName"] = t.Name() + "-step5"
	params["Name"] = t.Name() + "Updated"
	configText5 := templateFill(preRequisites+testAccVcfaContentLibraryTenantPrerequisites+testAccVcfaContentLibraryItemTenantStep4, params)
	params["FuncName"] = t.Name() + "-step6"
	configText6 := templateFill(preRequisites+testAccVcfaContentLibraryTenantPrerequisites+testAccVcfaContentLibraryItemTenantStep6, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)
	debugPrintf("#[DEBUG] CONFIGURATION step5: %s\n", configText5)
	debugPrintf("#[DEBUG] CONFIGURATION step6: %s\n", configText6)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cli1 := "vcfa_content_library_item.cli1"
	cli2 := "vcfa_content_library_item.cli2"
	cli3 := "vcfa_content_library_item.cli3"
	cli4 := "vcfa_content_library_item.cli4"
	cli5 := "vcfa_content_library_item.cli5"
	cli6 := "vcfa_content_library_item.cli6"

	// This test uses also a provider config block logged in as a Tenant user, so we can not only test that administrators
	// can create tenant library items, but also tenant users can. This is a function and not a map to be lazy evaluated, as
	// the given user is created after some testing steps.
	multipleFactories := func() map[string]func() (*schema.Provider, error) {
		return map[string]func() (*schema.Provider, error){
			"vcfa": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
			"vcfatenant": func() (*schema.Provider, error) {
				return testOrgProvider(testConfig.Tm.Org, params["Username"].(string), params["Password"].(string)), nil
			},
		}
	}

	// Before this test ends we need to clean up the clients cache, because we create an Org user
	// and use it to login with the provider. Using same credentials and org name could lead to errors if this user
	// remains cached.
	defer cachedVCDClients.reset()

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProviderFactories: testAccProviders,
				Config:            configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// CLI 1: OVA
					resource.TestCheckResourceAttr(cli1, "name", t.Name()+"1"),
					resource.TestCheckResourceAttr(cli1, "description", t.Name()+"1"),
					resource.TestCheckResourceAttrPair(cli1, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli1, "creation_date"),
					resource.TestCheckResourceAttr(cli1, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli1, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli1, "image_identifier"),
					resource.TestCheckResourceAttr(cli1, "item_type", "TEMPLATE"),
					resource.TestCheckResourceAttrPair(cli1, "owner_org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttr(cli1, "status", "READY"),
					resource.TestCheckResourceAttr(cli1, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli1, "version", "1"),

					// CLI 2: ISO
					resource.TestCheckResourceAttr(cli2, "name", t.Name()+"2"),
					resource.TestCheckResourceAttr(cli2, "description", t.Name()+"2"),
					resource.TestCheckResourceAttrPair(cli2, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli2, "creation_date"),
					resource.TestCheckResourceAttr(cli2, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli2, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli2, "image_identifier"),
					resource.TestCheckResourceAttr(cli2, "item_type", "ISO"),
					resource.TestCheckResourceAttrPair(cli1, "owner_org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttr(cli2, "status", "READY"),
					resource.TestCheckResourceAttr(cli2, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli2, "version", "1"),

					// CLI 3: OVF
					resource.TestCheckResourceAttr(cli3, "name", t.Name()+"3"),
					resource.TestCheckResourceAttr(cli3, "description", t.Name()+"3"),
					resource.TestCheckResourceAttrPair(cli3, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli3, "creation_date"),
					resource.TestCheckResourceAttr(cli3, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli3, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli3, "image_identifier"),
					resource.TestCheckResourceAttr(cli3, "item_type", "TEMPLATE"),
					resource.TestCheckResourceAttrPair(cli1, "owner_org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttr(cli3, "status", "READY"),
					resource.TestCheckResourceAttr(cli3, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli3, "version", "1"),
				),
			},
			{
				ProviderFactories: testAccProviders,
				Config:            configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(cli1, "name", t.Name()+"Updated1"),
					resource.TestCheckResourceAttr(cli1, "description", t.Name()+"Updated1"),
					resource.TestCheckResourceAttr(cli2, "name", t.Name()+"Updated2"),
					resource.TestCheckResourceAttr(cli2, "description", t.Name()+"Updated2"),
					resource.TestCheckResourceAttr(cli3, "name", t.Name()+"Updated3"),
					resource.TestCheckResourceAttr(cli3, "description", t.Name()+"Updated3"),
				),
			},
			{
				ProviderFactories: testAccProviders,
				Config:            configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// file_paths and upload_piece_size cannot be obtained during reads, that's why it does not appear in data source schema
					resourceFieldsEqual(cli1, "data.vcfa_content_library_item.cli1_ds", []string{"file_paths.#", "file_paths.0", "upload_piece_size", "%"}),
					resourceFieldsEqual(cli2, "data.vcfa_content_library_item.cli2_ds", []string{"file_paths.#", "file_paths.0", "upload_piece_size", "%"}),
					resourceFieldsEqual(cli3, "data.vcfa_content_library_item.cli3_ds", []string{"file_paths.#", "file_paths.0", "file_paths.1", "upload_piece_size", "%"}),
				),
			},
			{
				ProviderFactories:       testAccProviders,
				ResourceName:            cli1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           fmt.Sprintf("%s%s%s%s%s", testConfig.Tm.Org, ImportSeparator, testConfig.Tm.ContentLibrary, ImportSeparator, t.Name()+"Updated1"),
				ImportStateVerifyIgnore: []string{"file_paths.#", "file_paths.0", "upload_piece_size", "%"}, // file_paths and upload_piece_size cannot be obtained during imports, that's why it's Optional
			},
			{
				ProviderFactories: multipleFactories(),
				Config:            configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					// CLI 4: OVA
					resource.TestCheckResourceAttr(cli4, "name", t.Name()+"4"),
					resource.TestCheckResourceAttr(cli4, "description", t.Name()+"4"),
					resource.TestCheckResourceAttrPair(cli4, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli4, "creation_date"),
					resource.TestCheckResourceAttr(cli4, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli4, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli4, "image_identifier"),
					resource.TestCheckResourceAttr(cli4, "item_type", "TEMPLATE"),
					resource.TestCheckResourceAttrPair(cli4, "owner_org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttr(cli4, "status", "READY"),
					resource.TestCheckResourceAttr(cli4, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli4, "version", "1"),

					// CLI 5: ISO
					resource.TestCheckResourceAttr(cli5, "name", t.Name()+"5"),
					resource.TestCheckResourceAttr(cli5, "description", t.Name()+"5"),
					resource.TestCheckResourceAttrPair(cli5, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli5, "creation_date"),
					resource.TestCheckResourceAttr(cli5, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli5, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli5, "image_identifier"),
					resource.TestCheckResourceAttr(cli5, "item_type", "ISO"),
					resource.TestCheckResourceAttrPair(cli5, "owner_org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttr(cli5, "status", "READY"),
					resource.TestCheckResourceAttr(cli5, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli5, "version", "1"),

					// CLI 6: OVF
					resource.TestCheckResourceAttr(cli6, "name", t.Name()+"6"),
					resource.TestCheckResourceAttr(cli6, "description", t.Name()+"6"),
					resource.TestCheckResourceAttrPair(cli6, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli6, "creation_date"),
					resource.TestCheckResourceAttr(cli6, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli6, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli6, "image_identifier"),
					resource.TestCheckResourceAttr(cli6, "item_type", "TEMPLATE"),
					resource.TestCheckResourceAttrPair(cli6, "owner_org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttr(cli6, "status", "READY"),
					resource.TestCheckResourceAttr(cli6, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli6, "version", "1"),
				),
			},
			{
				ProviderFactories: multipleFactories(),
				Config:            configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(cli4, "name", t.Name()+"Updated4"),
					resource.TestCheckResourceAttr(cli4, "description", t.Name()+"Updated4"),
					resource.TestCheckResourceAttr(cli5, "name", t.Name()+"Updated5"),
					resource.TestCheckResourceAttr(cli5, "description", t.Name()+"Updated5"),
					resource.TestCheckResourceAttr(cli6, "name", t.Name()+"Updated6"),
					resource.TestCheckResourceAttr(cli6, "description", t.Name()+"Updated6"),
				),
			},
			{
				ProviderFactories: multipleFactories(),
				Config:            configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					// file_paths and upload_piece_size cannot be obtained during reads, that's why it does not appear in data source schema
					resourceFieldsEqual(cli4, "data.vcfa_content_library_item.cli4_ds", []string{"file_paths.#", "file_paths.0", "upload_piece_size", "%"}),
					resourceFieldsEqual(cli5, "data.vcfa_content_library_item.cli5_ds", []string{"file_paths.#", "file_paths.0", "upload_piece_size", "%"}),
					resourceFieldsEqual(cli6, "data.vcfa_content_library_item.cli6_ds", []string{"file_paths.#", "file_paths.0", "file_paths.1", "upload_piece_size", "%"}),
				),
			},
		},
	})
}

const testAccVcfaContentLibraryItemTenantStep4 = `
# skip-binary-test: Requires an extra provider configuration block with a tenant user

resource "vcfa_content_library_item" "cli4" {
  provider           = vcfatenant
  name               = "{{.Name}}4"
  description        = "{{.Name}}4"
  content_library_id = {{.ContentLibraryRef}}
  file_paths         = ["{{.OvaPath}}"]
}

resource "vcfa_content_library_item" "cli5" {
  provider           = vcfatenant
  name               = "{{.Name}}5"
  description        = "{{.Name}}5"
  content_library_id = {{.ContentLibraryRef}}
  file_paths         = ["{{.IsoPath}}"]
}

resource "vcfa_content_library_item" "cli6" {
  provider           = vcfatenant
  name               = "{{.Name}}6"
  description        = "{{.Name}}6"
  content_library_id = {{.ContentLibraryRef}}
  file_paths         = [{{.OvfPaths}}]
}
`

const testAccVcfaContentLibraryItemTenantStep6 = testAccVcfaContentLibraryItemTenantStep4 + `
# skip-binary-test: Requires an extra provider configuration block with a tenant user

data "vcfa_content_library_item" "cli4_ds" {
  provider           = vcfatenant
  name               = vcfa_content_library_item.cli4.name
  content_library_id = vcfa_content_library_item.cli4.content_library_id
}
data "vcfa_content_library_item" "cli5_ds" {
  provider           = vcfatenant
  name               = vcfa_content_library_item.cli5.name
  content_library_id = vcfa_content_library_item.cli5.content_library_id
}
data "vcfa_content_library_item" "cli6_ds" {
  provider           = vcfatenant
  name               = vcfa_content_library_item.cli6.name
  content_library_id = vcfa_content_library_item.cli6.content_library_id
}
`
