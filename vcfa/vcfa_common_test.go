//go:build api || functional || tm || ALL

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

// getVCenterHcl gets a vCenter data source as first returned parameter and its HCL reference as second one,
// only if a vCenter is already configured in VCFA. Otherwise, it returns a vCenter resource HCL as first returned parameter
// and its HCL reference as second one, only if "createVCenter=true" in the testing configuration
func getVCenterHcl(t *testing.T) (string, string) {
	tmClient := createTemporaryVCFAConnection(false)
	vc, err := tmClient.GetVCenterByUrl(testConfig.Tm.VcenterUrl)
	if err == nil {
		return `
data "vcfa_vcenter" "vc" {
  name = "` + vc.VSphereVCenter.Name + `"
}
`, "data.vcfa_vcenter.vc"
	}
	if !govcd.ContainsNotFound(err) {
		t.Fatal(err)
		return "", ""
	}
	if !testConfig.Tm.CreateVcenter {
		t.Skip("vCenter is not configured and configuration is not allowed in config file")
		return "", ""
	}
	return `
resource "vcfa_vcenter" "vc" {
  name                     = "` + t.Name() + `"
  url                      = "` + testConfig.Tm.VcenterUrl + `"
  auto_trust_certificate   = true
  refresh_vcenter_on_read  = true
  refresh_policies_on_read = false
  username                 = "` + testConfig.Tm.VcenterUsername + `"
  password                 = "` + testConfig.Tm.VcenterPassword + `"
  is_enabled               = true
}
`, "vcfa_vcenter.vc"
}

// getNsxManagerHcl gets a NSX Manager data source as first returned parameter and its HCL reference as second one,
// only if a NSX Manager is already configured in VCFA. Otherwise, it returns a NSX Manager resource HCL as first returned parameter
// and its HCL reference as second one, only if "createNsxManager=true" in the testing configuration
func getNsxManagerHcl(t *testing.T) (string, string) {
	tmClient := createTemporaryVCFAConnection(false)
	nsxtManager, err := tmClient.GetNsxtManagerOpenApiByUrl(testConfig.Tm.NsxManagerUrl)
	if err == nil {
		return `
data "vcfa_nsx_manager" "nsx_manager" {
  name = "` + nsxtManager.NsxtManagerOpenApi.Name + `"
}
`, "data.vcfa_nsx_manager.nsx_manager"
	}
	if !govcd.ContainsNotFound(err) {
		t.Fatal(err)
		return "", ""
	}
	if !testConfig.Tm.CreateNsxManager {
		t.Skip("NSX Manager is not configured and configuration is not allowed in config file")
		return "", ""
	}
	return `
resource "vcfa_nsx_manager" "nsx_manager" {
  name                   = "` + t.Name() + `"
  description            = "` + t.Name() + `"
  username               = "` + testConfig.Tm.NsxManagerUsername + `"
  password               = "` + testConfig.Tm.NsxManagerPassword + `"
  url                    = "` + testConfig.Tm.NsxManagerUrl + `"
  network_provider_scope = ""
  auto_trust_certificate = true
}

`, "vcfa_nsx_manager.nsx_manager"
}

// getRegionHcl gets a Region data source as first returned parameter and its HCL reference as second one,
// only if a Region is already configured in VCFA. Otherwise, it returns a Region resource HCL as first returned parameter
// and its HCL reference as second one, only if "createRegion=true" in the testing configuration
func getRegionHcl(t *testing.T, vCenterHclRef, nsxManagerHclRef string) (string, string) {
	if testConfig.Tm.Region == "" {
		t.Fatalf("the property tm.region is required but it is not present in testing JSON")
	}
	tmClient := createTemporaryVCFAConnection(false)
	region, err := tmClient.GetRegionByName(testConfig.Tm.Region)
	if err == nil {
		return `
data "vcfa_region" "region" {
  name = "` + region.Region.Name + `"
}
`, "data.vcfa_region.region"
	}
	if !govcd.ContainsNotFound(err) {
		t.Fatal(err)
		return "", ""
	}
	if !testConfig.Tm.CreateRegion {
		t.Skip("Region is not configured and configuration is not allowed in config file")
		return "", ""
	}
	return `
data "vcfa_supervisor" "supervisor" {
  name       = "` + testConfig.Tm.VcenterSupervisor + `"
  vcenter_id = ` + vCenterHclRef + `.id
  depends_on = [` + vCenterHclRef + `]
}

resource "vcfa_region" "region" {
  name                 = "` + testConfig.Tm.Region + `"
  description          = "` + testConfig.Tm.Region + `"
  nsx_manager_id       = ` + nsxManagerHclRef + `.id
  supervisor_ids       = [data.vcfa_supervisor.supervisor.id]
  storage_policy_names = ["` + testConfig.Tm.VcenterStorageProfile + `"]
}
`, "vcfa_region.region"
}

// getRegionVmClassesHcl gets HCL code to fetch Region VM Classes given by the testing config
func getRegionVmClassesHcl(t *testing.T, regionRef string) (string, []string) {
	if len(testConfig.Tm.RegionVmClasses) == 0 {
		t.Fatalf("at least one Region VM Class is needed in the slice tm.regionVmClasses")
	}
	hcl := ``
	refs := make([]string, len(testConfig.Tm.RegionVmClasses))
	for i, class := range testConfig.Tm.RegionVmClasses {
		refs[i] = fmt.Sprintf("data.vcfa_region_vm_class.region_vm_class%d", i)
		hcl += fmt.Sprintf(`
data "vcfa_region_vm_class" "region_vm_class%d" {
  region_id = %s.id
  name      = "%s"
}
`, i, regionRef, class)
	}

	return hcl, refs
}

// getContentLibraryHcl gets a Content Library data source as first returned parameter and its HCL reference as second one,
// only if a Content Library is already configured in TM. Otherwise, it returns a Content Library resource HCL as first returned parameter
// and its HCL reference as second one
func getContentLibraryHcl(t *testing.T, regionHclRef string) (string, string) {
	if testConfig.Tm.ContentLibrary == "" {
		t.Fatalf("the property tm.contentLibrary is required but it is not present in testing JSON")
	}
	if testConfig.Tm.StorageClass == "" {
		t.Fatalf("the property tm.storageClass is required but it is not present in testing JSON")
	}
	tmClient := createTemporaryVCFAConnection(false)
	cl, err := tmClient.GetContentLibraryByName(testConfig.Tm.ContentLibrary, nil)
	if err == nil {
		return `
data "vcfa_content_library" "content_library" {
  name = "` + cl.ContentLibrary.Name + `"
}
`, "data.vcfa_content_library.content_library"
	}
	if !govcd.ContainsNotFound(err) {
		t.Fatal(err)
		return "", ""
	}
	return `
data "vcfa_storage_class" "storage_class" {
  region_id = ` + regionHclRef + `.id 
  name      = "` + testConfig.Tm.StorageClass + `"
}

resource "vcfa_content_library" "content_library" {
  name                 = "` + testConfig.Tm.ContentLibrary + `"
  description          = "` + testConfig.Tm.ContentLibrary + `"
  storage_class_ids    = [data.vcfa_storage_class.storage_class.id]
  delete_force         = true
  delete_recursive     = true
}
`, "vcfa_content_library.content_library"
}

func getIpSpaceHcl(t *testing.T, regionHclRef, nameSuffix, octet3 string) (string, string) {
	return `
resource "vcfa_ip_space" "test-` + nameSuffix + `" {
  name                          = "` + t.Name() + nameSuffix + `"
  description                   = "Made using Terraform"
  region_id                     = ` + regionHclRef + `.id
  external_scope                = "43.12.` + octet3 + `.0/30"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1
  internal_scope {
    name = "scope3"
    cidr = "32.0.` + octet3 + `.0/24"
  }
}
	`, `vcfa_ip_space.test-` + nameSuffix
}

func getProviderGatewayHcl(t *testing.T, regionHclRef, ipSpaceHclRef string) (string, string) {
	if testConfig.Tm.ProviderGateway == "" {
		t.Fatalf("the property tm.providerGateway is required but it is not present in testing JSON")
	}

	tmClient := createTemporaryVCFAConnection(false)
	pg, err := tmClient.GetTmProviderGatewayByName(testConfig.Tm.ProviderGateway)
	if err == nil {
		return `
data "vcfa_provider_gateway" "test" {
  name      = "` + pg.TmProviderGateway.Name + `"
  region_id = ` + regionHclRef + `.id
}
`, "data.vcfa_provider_gateway.test"
	}
	if !govcd.ContainsNotFound(err) {
		t.Fatal(err)
		return "", ""
	}
	return `
data "vcfa_tier0_gateway" "test" {
  region_id = ` + regionHclRef + `.id 
  name      = "` + testConfig.Tm.NsxTier0Gateway + `"
}

resource "vcfa_provider_gateway" "test" {
  name                  = "` + testConfig.Tm.ProviderGateway + `"
  description           = "getProviderGatewayHcl"
  region_id             = ` + regionHclRef + `.id
  nsxt_tier0_gateway_id = data.vcfa_tier0_gateway.test.id
  ip_space_ids          = [` + ipSpaceHclRef + `.id]
}
`, "vcfa_provider_gateway.test"
}

func testAccCheckOrgDestroy(orgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tmClient := testAccProvider.Meta().(ClientContainer).tmClient
		org, err := tmClient.GetTmOrgByName(orgName)
		if org != nil {
			return fmt.Errorf("%s %s was found", labelVcfaOrg, orgName)
		}
		if !govcd.ContainsNotFound(err) {
			return fmt.Errorf("%s %s was not destroyed: %s", labelVcfaOrg, orgName, err)
		}
		return nil
	}
}

func testAccCheckVcfaOrgExists(orgResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[orgResource]
		if !ok {
			return fmt.Errorf("not found: %s", orgResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaOrg)
		}

		conn := testAccProvider.Meta().(ClientContainer).tmClient
		orgName := rs.Primary.Attributes["name"]
		_, err := conn.VCDClient.GetTmOrgByName(orgName)
		if err != nil {
			return err
		}

		return nil
	}
}
