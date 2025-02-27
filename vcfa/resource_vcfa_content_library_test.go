//go:build tm || contentlibrary || ALL || functional

package vcfa

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

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
				Check: resource.ComposeAggregateTestCheckFunc(
					// Region Storage Policy
					resource.TestCheckResourceAttr(dsRegionStoragePolicy, "name", testConfig.Tm.StorageClass),
					resource.TestCheckResourceAttrPair(dsRegionStoragePolicy, "region_id", regionHclRef, "id"),
					resource.TestMatchResourceAttr(dsRegionStoragePolicy, "description", regexp.MustCompile(`.*`)),
					resource.TestCheckResourceAttr(dsRegionStoragePolicy, "status", "READY"),
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
					resource.TestCheckResourceAttr(resourceName, "auto_attach", "true"), // Always true for PROVIDER libraries
					resource.TestCheckResourceAttrSet(resourceName, "creation_date"),
					resource.TestCheckResourceAttr(resourceName, "is_shared", "true"),      // Always true for PROVIDER libraries
					resource.TestCheckResourceAttr(resourceName, "is_subscribed", "false"), // TODO: TM: Test with true
					resource.TestCheckResourceAttr(resourceName, "library_type", "PROVIDER"),
					resource.TestCheckResourceAttr(resourceName, "subscription_config.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "version_number", regexp.MustCompile("[1-9]")),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()+"Updated"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
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
  name        = "{{.Name}}"
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

// TestAccVcfaContentLibraryTenant tests CRUD of a Content Library of type TENANT.
func TestAccVcfaContentLibraryTenant(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	vmClassesHcl, vmClassesRefs := getRegionVmClassesHcl(t, regionHclRef)

	var params = StringMap{
		"Org":                 testConfig.Tm.Org,
		"Username":            "test-user",
		"Password":            "long-change-ME1",
		"Name":                t.Name(),
		"Name2":               t.Name() + "2",
		"Name3":               t.Name() + "3",
		"RegionId":            fmt.Sprintf("%s.id", regionHclRef),
		"SupervisorName":      testConfig.Tm.VcenterSupervisor,
		"SupervisorZoneName":  testConfig.Tm.VcenterSupervisorZone,
		"StorageClass":        testConfig.Tm.StorageClass,
		"VcenterRef":          vCenterHclRef,
		"RegionStoragePolicy": testConfig.Tm.StorageClass,
		"RegionVmClassRefs":   strings.Join(vmClassesRefs, ".id,\n    ") + ".id",
		"VcfaUrl":             testConfig.Provider.Url,
		"Tags":                "tm contentlibrary",
	}
	testParamsNotEmpty(t, params)

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + vmClassesHcl

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	configText1 := templateFill(preRequisites+testAccVcfaContentLibraryTenantStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	params["Name"] = t.Name() + "Updated"
	params["Name2"] = t.Name() + "2Updated"
	configText2 := templateFill(preRequisites+testAccVcfaContentLibraryTenantStep1, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaContentLibraryTenantStep3, params)
	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(preRequisites+testAccVcfaContentLibraryTenantStep4, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cl1 := "vcfa_content_library.cl1"
	cl2 := "vcfa_content_library.cl2"
	cl3 := "vcfa_content_library.cl3"

	cachedId := &testCachedFieldValue{}

	// This test uses also a provider config block logged in as a Tenant user, so we can not only test that administrators
	// can create tenant libraries, but also tenant users can. This is a function and not a map to be lazy evaluated, as
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

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProviderFactories: testAccProviders,
				Config:            configText0,
			},
			{
				ProviderFactories: testAccProviders,
				Config:            configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// First content library
					cachedId.cacheTestResourceFieldValue(cl1, "id"),
					resource.TestCheckResourceAttr(cl1, "name", t.Name()),
					resource.TestCheckResourceAttrPair(cl1, "org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttr(cl1, "description", t.Name()),
					resource.TestCheckResourceAttr(cl1, "storage_class_ids.#", "1"),
					resource.TestCheckResourceAttr(cl1, "auto_attach", "true"), // Defaults to true for TENANT libraries
					resource.TestCheckResourceAttrSet(cl1, "creation_date"),
					resource.TestCheckResourceAttr(cl1, "is_shared", "false"), // Always false for TENANT libraries
					resource.TestCheckResourceAttr(cl1, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cl1, "library_type", "TENANT"),
					resource.TestCheckResourceAttr(cl1, "subscription_config.#", "0"),
					resource.TestMatchResourceAttr(cl1, "version_number", regexp.MustCompile("[1-9]")),

					// Second content library
					resource.TestCheckResourceAttr(cl2, "name", t.Name()+"2"),
					resource.TestCheckResourceAttrPair(cl2, "org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttr(cl2, "description", t.Name()+"2"),
					resource.TestCheckResourceAttr(cl2, "storage_class_ids.#", "1"),
					resource.TestCheckResourceAttr(cl2, "auto_attach", "false"),
					resource.TestCheckResourceAttrSet(cl2, "creation_date"),
					resource.TestCheckResourceAttr(cl2, "is_shared", "false"), // Always false for TENANT libraries
					resource.TestCheckResourceAttr(cl2, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cl2, "library_type", "TENANT"),
					resource.TestCheckResourceAttr(cl2, "subscription_config.#", "0"),
					resource.TestMatchResourceAttr(cl2, "version_number", regexp.MustCompile("[1-9]")),
				),
			},
			{
				ProviderFactories: testAccProviders,
				Config:            configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(cl1, "id"),
					resource.TestCheckResourceAttr(cl1, "name", t.Name()+"Updated"),
					resource.TestCheckResourceAttr(cl2, "name", t.Name()+"2Updated"),
				),
			},
			{
				ProviderFactories: multipleFactories(),
				Config:            configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Tenant content library
					resource.TestCheckResourceAttr(cl3, "name", t.Name()+"3"),
					resource.TestCheckResourceAttrPair(cl3, "org_id", "vcfa_org.test", "id"),
					resource.TestCheckResourceAttr(cl3, "description", t.Name()+"3"),
					resource.TestCheckResourceAttr(cl3, "storage_class_ids.#", "1"),
					resource.TestCheckResourceAttr(cl3, "auto_attach", "true"),
					resource.TestCheckResourceAttrSet(cl3, "creation_date"),
					resource.TestCheckResourceAttr(cl3, "is_shared", "false"), // Always false for TENANT libraries
					resource.TestCheckResourceAttr(cl3, "is_subscribed", "false"),
					resource.TestCheckResourceAttr(cl3, "library_type", "TENANT"),
					resource.TestCheckResourceAttr(cl3, "subscription_config.#", "0"),
					resource.TestMatchResourceAttr(cl3, "version_number", regexp.MustCompile("[1-9]")),
				),
			},
			{
				ProviderFactories: multipleFactories(),
				Config:            configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(cl1, "data.vcfa_content_library.cl_ds1", []string{
						"%", // Does not have delete_recursive, delete_force
						"delete_recursive",
						"delete_force",
					}),
					resourceFieldsEqual(cl2, "data.vcfa_content_library.cl_ds2", []string{
						"%",
						"delete_recursive",
						"delete_force",
					}),
					resourceFieldsEqual(cl3, "data.vcfa_content_library.cl_ds3", []string{
						"%",
						"delete_recursive",
						"delete_force",
					}),
					resourceFieldsEqual(cl3, "data.vcfa_content_library.cl_ds3tenant", []string{
						"%",
						"delete_recursive",
						"delete_force",
					}),
				),
			},
			{
				ProviderFactories: multipleFactories(),
				ResourceName:      cl1,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     fmt.Sprintf("%s%s%s", params["Org"].(string), ImportSeparator, params["Name"].(string)),
				ImportStateVerifyIgnore: []string{
					"delete_recursive",
					"delete_force",
				},
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaContentLibraryTenantPrerequisites = `
resource "vcfa_org" "test" {
  provider     = vcfa
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "{{.Org}}"
}

data "vcfa_role" "org-admin" {
  provider = vcfa
  org_id   = vcfa_org.test.id
  name     = "Organization Administrator"
}

resource "vcfa_org_local_user" "user" {
  provider = vcfa
  org_id   = vcfa_org.test.id
  role_ids = [data.vcfa_role.org-admin.id]
  username = "{{.Username}}"
  password = "{{.Password}}"
}

data "vcfa_supervisor" "test" {
  provider   = vcfa
  name       = "{{.SupervisorName}}"
  vcenter_id = {{.VcenterRef}}.id
  depends_on = [{{.VcenterRef}}]
}

data "vcfa_region_zone" "test" {
  provider  = vcfa
  region_id = {{.RegionId}}
  name      = "{{.SupervisorZoneName}}"
}

data "vcfa_region_storage_policy" "sp" {
  provider  = vcfa
  name      = "{{.StorageClass}}"
  region_id = {{.RegionId}}
}

resource "vcfa_org_region_quota" "test" {
  provider       = vcfa
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
    storage_limit_mib        = 1024
  }

  # This explicit dependency avoids that the user from the org is deleted
  # before any child object from the Region Quota
  depends_on = [vcfa_org_local_user.user]
}
`

const testAccVcfaContentLibraryTenantStep1 = testAccVcfaContentLibraryTenantPrerequisites + `
data "vcfa_storage_class" "sc" {
  provider  = vcfa
  region_id = {{.RegionId}}
  name      = data.vcfa_region_storage_policy.sp.name
}

resource "vcfa_content_library" "cl1" {
  provider    = vcfa
  org_id      = vcfa_org_region_quota.test.org_id # Explicit dependency on Region Quota
  name        = "{{.Name}}"
  description = "{{.Name}}"
  storage_class_ids = [
    data.vcfa_storage_class.sc.id
  ]
  delete_recursive = true
}

resource "vcfa_content_library" "cl2" {
  provider    = vcfa
  org_id      = vcfa_org_region_quota.test.org_id # Explicit dependency on Region Quota
  name        = "{{.Name2}}"
  description = "{{.Name2}}"
  auto_attach = false
  storage_class_ids = [
    data.vcfa_storage_class.sc.id
  ]
  delete_force     = true # Should be ignored, otherwise it would fail
  delete_recursive = true
}
`

const testAccVcfaContentLibraryTenantStep3 = testAccVcfaContentLibraryTenantStep1 + `
# skip-binary-test: Requires an extra provider configuration block with a tenant user

data "vcfa_storage_class" "sc-tenant" {
  provider  = vcfatenant
  region_id = {{.RegionId}}
  name      = data.vcfa_region_storage_policy.sp.name
}

resource "vcfa_content_library" "cl3" {
  provider    = vcfatenant
  org_id      = vcfa_org.test.id
  name        = "{{.Name3}}"
  description = "{{.Name3}}"
  storage_class_ids = [
    data.vcfa_storage_class.sc-tenant.id
  ]
  delete_force     = true # Should be ignored, otherwise it would fail
  delete_recursive = true

  # Explicit dependency on Region Quota.
  # A real tenant user should not need to do something like this (a Region Quota should be already provisioned),
  # but as we created the Region Quota at same time, we need to guarantee dependencies
  # so they are removed correctly afterwards.
  # Also depends on the logged in user.
  depends_on = [vcfa_org_region_quota.test, vcfa_org_local_user.user]
}
`

const testAccVcfaContentLibraryTenantStep4 = testAccVcfaContentLibraryTenantStep3 + `
# skip-binary-test: Requires an extra provider configuration block with a tenant user

data "vcfa_content_library" "cl_ds1" {
  provider = vcfa
  org_id   = vcfa_org.test.id
  name     = vcfa_content_library.cl1.name
}

data "vcfa_content_library" "cl_ds2" {
  provider = vcfa
  org_id   = vcfa_org.test.id
  name     = vcfa_content_library.cl2.name
}

data "vcfa_content_library" "cl_ds3" {
  provider = vcfa
  org_id   = vcfa_org.test.id
  name     = vcfa_content_library.cl3.name
}

data "vcfa_content_library" "cl_ds3tenant" {
  provider = vcfatenant
  org_id   = vcfa_org.test.id
  name     = vcfa_content_library.cl3.name
}
`
