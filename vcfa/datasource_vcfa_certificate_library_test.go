//go:build certificate || ALL || functional

package vcfa

import (
	"fmt"
	"testing"

	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcfaLibraryCertificateDS tests that existing certificate can be fetched
func TestAccVcfaLibraryCertificateDS(t *testing.T) {
	preTestChecks(t)

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCFAConnection(false)

	certificates, err := getAvailableCertificate(vcdClient)
	if err != nil {
		t.Skip("No suitable certificates found for this test")
		return
	}
	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.Tm.Org,
		"Alias":       certificates[0].CertificateLibrary.Alias,
		"Id":          certificates[0].CertificateLibrary.Id,
		"AliasSystem": certificates[1].CertificateLibrary.Alias,
		"IdSystem":    certificates[1].CertificateLibrary.Id,
	}
	testParamsNotEmpty(t, params)

	template := testAccVcfaLibraryCertificateOrgDS
	// add test part when test is run by System admin
	if vcdClient.Client.IsSysAdmin {
		template = template + testAccVcfaLibraryCertificateSysDS
	}

	configText1 := templateFill(template, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	checkFunctions := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("data.vcfa_library_certificate.existing", "alias", certificates[0].CertificateLibrary.Alias),
		resource.TestCheckResourceAttr("data.vcfa_library_certificate.existing", "id", certificates[0].CertificateLibrary.Id),
		resource.TestCheckResourceAttr("data.vcfa_library_certificate.existing", "description", certificates[0].CertificateLibrary.Description),
		resource.TestCheckResourceAttr("data.vcfa_library_certificate.existing", "certificate", certificates[0].CertificateLibrary.Certificate),
		resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingById", "alias", certificates[0].CertificateLibrary.Alias),
		resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingById", "id", certificates[0].CertificateLibrary.Id),
		resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingById", "description", certificates[0].CertificateLibrary.Description),
		resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingById", "certificate", certificates[0].CertificateLibrary.Certificate),
	}

	// add test part when test is run by System admin
	if vcdClient.Client.IsSysAdmin {
		sysCheckFunctions := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingSystem", "alias", certificates[1].CertificateLibrary.Alias),
			resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingSystem", "id", certificates[1].CertificateLibrary.Id),
			resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingSystem", "description", certificates[1].CertificateLibrary.Description),
			resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingSystem", "certificate", certificates[1].CertificateLibrary.Certificate),
			resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingSystemById", "alias", certificates[1].CertificateLibrary.Alias),
			resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingSystemById", "id", certificates[1].CertificateLibrary.Id),
			resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingSystemById", "description", certificates[1].CertificateLibrary.Description),
			resource.TestCheckResourceAttr("data.vcfa_library_certificate.existingSystemById", "certificate", certificates[1].CertificateLibrary.Certificate),
		}
		fmt.Printf("Sys admin part added \n")
		checkFunctions = append(checkFunctions, sysCheckFunctions...)
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check:  resource.ComposeAggregateTestCheckFunc(checkFunctions...),
			},
		},
	})
	postTestChecks(t)
}

// getAvailableCertificate fetches one available certificate to use in data source tests
func getAvailableCertificate(vcdClient *VCDClient) ([]*govcd.Certificate, error) {
	err := ProviderAuthenticate(vcdClient.VCDClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken, testConfig.Provider.ApiTokenFile, testConfig.Provider.ServiceAccountTokenFile)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %v", err)
	}

	adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.Tm.Org)
	if err != nil {
		return nil, fmt.Errorf("org not found : %s", err)
	}

	certificates, err := adminOrg.GetAllCertificatesFromLibrary(nil)
	if err != nil {
		return nil, fmt.Errorf("get all certificates failed : %s", err)
	}
	if len(certificates) == 0 {
		return nil, fmt.Errorf("no certificate found in org %v", testConfig.Tm.Org)
	}

	certificatesInSystem, err := vcdClient.Client.GetAllCertificatesFromLibrary(nil)
	if err != nil {
		return nil, fmt.Errorf("get all certificates from Sys failed : %s", err)
	}
	if len(certificatesInSystem) == 0 {
		return nil, fmt.Errorf("no certificate found in System")
	}

	return []*govcd.Certificate{certificates[0], certificatesInSystem[0]}, nil
}

const testAccVcfaLibraryCertificateOrgDS = `
data "vcfa_library_certificate" "existing" {
  org    = "{{.Org}}"
  alias  = "{{.Alias}}"
}

data "vcfa_library_certificate" "existingById" {
  org = "{{.Org}}"
  id  = "{{.Id}}"
}
`

const testAccVcfaLibraryCertificateSysDS = `
data "vcfa_library_certificate" "existingSystem" {
  org    = "System"
  alias  = "{{.AliasSystem}}"
}

data "vcfa_library_certificate" "existingSystemById" {
  org = "System"
  id  = "{{.IdSystem}}"
}
`
