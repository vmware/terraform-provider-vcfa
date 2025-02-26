//go:build tm || contentlibrary || ALL || functional

package vcfa

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcfaContentLibraryItemProvider tests Content Library Items in a "PROVIDER" type Content Library
func TestAccVcfaContentLibraryItemProvider(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	contentLibraryHcl, contentLibraryHclRef := getContentLibraryHcl(t, regionHclRef, "")

	var params = StringMap{
		"Name":              t.Name(),
		"ContentLibraryRef": fmt.Sprintf("%s.id", contentLibraryHclRef),
		"OvaPath":           "../test-resources/test_vapp_template.ova",
		"IsoPath":           "../test-resources/test.iso",
		"OvfPaths":          "\"../test-resources/test_vapp_template_ovf/descriptor.ovf\", \"../test-resources/test_vapp_template_ovf/yVMFromVcd-disk1.vmdk\", ",
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
					resource.TestMatchResourceAttr(cli1, "owner_org_id", regexp.MustCompile("urn:vcloud:org:")),
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
					resource.TestMatchResourceAttr(cli2, "owner_org_id", regexp.MustCompile("urn:vcloud:org:")),
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
					resource.TestMatchResourceAttr(cli3, "owner_org_id", regexp.MustCompile("urn:vcloud:org:")),
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
				ImportStateId:           fmt.Sprintf("%s%s%s", testConfig.Tm.ContentLibrary, ImportSeparator, params["Name"].(string)),
				ImportStateVerifyIgnore: []string{"file_path", "upload_piece_size"}, // file_path and upload_piece_size cannot be obtained during imports, that's why it's Optional
			},
		},
	})

	postTestChecks(t)
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
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	vmClassesHcl, vmClassesRefs := getRegionVmClassesHcl(t, regionHclRef)
	// The Content Library for tenants must depend on the Region Quota, as it contains the Storage Policies required
	// to create libraries in the Organization
	contentLibraryHcl, contentLibraryHclRef := getContentLibraryHcl(t, regionHclRef, "vcfa_org_region_quota.test.org_id")

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
		"OvaPath":             "../test-resources/test_vapp_template.ova",
		"IsoPath":             "../test-resources/test.iso",
		"OvfPaths":            "\"../test-resources/test_vapp_template_ovf/descriptor.ovf\", \"../test-resources/test_vapp_template_ovf/yVMFromVcd-disk1.vmdk\", ",
		"Tags":                "tm contentlibrary",
	}
	testParamsNotEmpty(t, params)

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + vmClassesHcl + contentLibraryHcl

	configText1 := templateFill(preRequisites+testAccVcfaContentLibraryItemTenantStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	params["Name"] = t.Name() + "Updated"
	configText2 := templateFill(preRequisites+testAccVcfaContentLibraryItemTenantStep1, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaContentLibraryItemTenantStep3, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cli1 := "vcfa_content_library_item.cli1"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(cli1, "name", t.Name()+"1"),
					resource.TestCheckResourceAttr(cli1, "description", t.Name()+"1"),
					resource.TestCheckResourceAttrPair(cli1, "content_library_id", contentLibraryHclRef, "id"),
					resource.TestCheckResourceAttrSet(cli1, "creation_date"),
					resource.TestCheckResourceAttr(cli1, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cli1, "is_published", "false"),
					resource.TestCheckResourceAttrSet(cli1, "image_identifier"),
					resource.TestCheckResourceAttr(cli1, "item_type", "TEMPLATE"),
					resource.TestCheckResourceAttrPair(cli1, "owner_org_id", "vcfa_org_region_quota.test", "org_id"),
					resource.TestCheckResourceAttr(cli1, "status", "READY"),
					resource.TestCheckResourceAttr(cli1, "last_successful_sync", ""),
					resource.TestCheckResourceAttr(cli1, "version", "1"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(cli1, "name", t.Name()+"Updated1"),
					resource.TestCheckResourceAttr(cli1, "description", t.Name()+"Updated1"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// file_path and upload_piece_size cannot be obtained during reads, that's why it does not appear in data source schema
					resourceFieldsEqual(cli1, "data.vcfa_content_library_item.cli1_ds", []string{"file_path", "upload_piece_size", "%"}),
				),
			},
			{
				ResourceName:            cli1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           fmt.Sprintf("%s%s%s%s%s", testConfig.Tm.Org, ImportSeparator, testConfig.Tm.ContentLibrary, ImportSeparator, params["Name"].(string)),
				ImportStateVerifyIgnore: []string{"file_path", "upload_piece_size"}, // file_path and upload_piece_size cannot be obtained during imports, that's why it's Optional
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaContentLibraryItemTenantStep1 = testAccVcfaContentLibraryTenantPrerequisites + `
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

const testAccVcfaContentLibraryItemTenantStep3 = testAccVcfaContentLibraryItemTenantStep1 + `
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
