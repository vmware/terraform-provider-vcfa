//go:build certificate || ALL || functional

package vcfa

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdLibraryCertificateResource tests that certificate can add to library
func TestAccVcdLibraryCertificateResource(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// This test requires access to VCFA before filling templates
	// Thus it won't run in the short test
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	if testConfig.Certificates.Certificate1Path == "" || testConfig.Certificates.Certificate2Path == "" ||
		testConfig.Certificates.Certificate2PrivateKeyPath == "" || testConfig.Certificates.Certificate2Pass == "" {
		t.Skip("Variables Certificates.Certificate1Path, Certificates.Certificate2Path2, " +
			"Certificates.Certificate2PrivateKeyPath, Certificates.Certificate2Pass must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                      testConfig.Tm.Org,
		"Alias":                    "TestAccVcdLibraryCertificateResource",
		"AliasUpdate":              "TestAccVcdLibraryCertificateResourceUpdated",
		"AliasSystem":              "TestAccVcdLibraryCertificateResourceSys",
		"AliasPrivate":             "TestAccVcdLibraryCertificateResourcePrivate",
		"AliasPrivateSystem":       "TestAccVcdLibraryCertificateResourcePrivateSys",
		"AliasPrivateSystemUpdate": "TestAccVcdLibraryCertificateResourcePrivateSysUpdated",
		"Certificate1Path":         testConfig.Certificates.Certificate1Path,
		"Certificate2Path":         testConfig.Certificates.Certificate2Path,
		"PrivateKey2":              testConfig.Certificates.Certificate2PrivateKeyPath,
		"PassPhrase":               testConfig.Certificates.Certificate2Pass,
		"Description1":             "myDescription 1",
		"Description1Update":       "myDescription 1 updated",
		"Description2":             "myDescription 2",
		"Description3":             "myDescription 3",
		"Description4":             "myDescription 4",
		"Description4Update":       "myDescription 4 updated",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdLibraryCertificateResource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-update"
	configText2 := templateFill(testAccVcdLibraryCertificateResourceUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	resourceAddressOrgCert := "vcfa_library_certificate.orgCertificate"
	resourceAddressOrgPrivateCert := "vcfa_library_certificate.OrgWithPrivateCertificate"
	resourceAddressSysCert := "vcfa_library_certificate.sysCertificate"
	resourceAddressSysPrivateCert := "vcfa_library_certificate.sysCertificateWithPrivate"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressOrgCert, "alias", params["Alias"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressOrgCert, "description", params["Description1"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgCert, "certificate", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressOrgPrivateCert, "alias", params["AliasPrivate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgPrivateCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressOrgPrivateCert, "description", params["Description2"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgPrivateCert, "certificate", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysCert, "alias", params["AliasSystem"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysCert, "description", params["Description3"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysCert, "certificate", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysPrivateCert, "alias", params["AliasPrivateSystem"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysPrivateCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysPrivateCert, "description", params["Description4"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysPrivateCert, "certificate", regexp.MustCompile(`^\S+`)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressOrgCert, "alias", params["AliasUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressOrgCert, "description", params["Description1Update"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgCert, "certificate", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysPrivateCert, "alias", params["AliasPrivateSystemUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysPrivateCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysPrivateCert, "description", params["Description4Update"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysPrivateCert, "certificate", regexp.MustCompile(`^\S+`)),
				),
			},
			{
				ResourceName:      resourceAddressOrgCert,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return testConfig.Tm.Org +
						ImportSeparator +
						params["AliasUpdate"].(string), nil
				},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdLibraryCertificateResource = `
resource "vcfa_library_certificate" "orgCertificate" {
  org         = "{{.Org}}"
  alias       = "{{.Alias}}"
  description = "{{.Description1}}"
  certificate = file("{{.Certificate1Path}}")
}

resource "vcfa_library_certificate" "OrgWithPrivateCertificate" {
  org                    = "{{.Org}}"
  alias                  = "{{.AliasPrivate}}"
  description            = "{{.Description2}}"
  certificate            = file("{{.Certificate2Path}}")
  private_key            = file("{{.PrivateKey2}}")
  private_key_passphrase = "{{.PassPhrase}}"
}

resource "vcfa_library_certificate" "sysCertificate" {
  org         = "System"
  alias       = "{{.AliasSystem}}"
  description = "{{.Description3}}"
  certificate = file("{{.Certificate1Path}}")
}

resource "vcfa_library_certificate" "sysCertificateWithPrivate" {
  org                    = "System"
  alias                  = "{{.AliasPrivateSystem}}"
  description            = "{{.Description4}}"
  certificate            = file("{{.Certificate2Path}}")
  private_key            = file("{{.PrivateKey2}}")
  private_key_passphrase = "{{.PassPhrase}}"
}
`

const testAccVcdLibraryCertificateResourceUpdate = `
resource "vcfa_library_certificate" "orgCertificate" {
  org         = "{{.Org}}"
  alias       = "{{.AliasUpdate}}"
  description = "{{.Description1Update}}"
  certificate = file("{{.Certificate1Path}}")
}

resource "vcfa_library_certificate" "sysCertificateWithPrivate" {
  org                    = "System"
  alias                  = "{{.AliasPrivateSystemUpdate}}"
  description            = "{{.Description4Update}}"
  certificate            = file("{{.Certificate2Path}}")
  private_key            = file("{{.PrivateKey2}}")
  private_key_passphrase = "{{.PassPhrase}}"
}
`
