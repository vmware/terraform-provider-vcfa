//go:build certificate || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccVcfaCertificateResource tests certificate libraries. At least two certificates must be provided in the
// testing configuration
func TestAccVcfaCertificateResource(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)

	skipIfNotSysAdmin(t)

	if len(testConfig.Tm.Certificates) < 2 {
		t.Skip("there must be at least two certificates in tm.certificates from test configuration")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                      testConfig.Tm.Org,
		"Alias":                    "TestAccVcfaLibraryCertificateResource",
		"AliasUpdate":              "TestAccVcfaLibraryCertificateResourceUpdated",
		"AliasSystem":              "TestAccVcfaLibraryCertificateResourceSys",
		"AliasPrivate":             "TestAccVcfaLibraryCertificateResourcePrivate",
		"AliasPrivateSystem":       "TestAccVcfaLibraryCertificateResourcePrivateSys",
		"AliasPrivateSystemUpdate": "TestAccVcfaLibraryCertificateResourcePrivateSysUpdated",
		"Certificate1Path":         testConfig.Tm.Certificates[0].Path,
		"Certificate2Path":         testConfig.Tm.Certificates[1].Path,
		"PrivateKey2":              testConfig.Tm.Certificates[1].PrivateKeyPath,
		"PassPhrase":               testConfig.Tm.Certificates[1].Password,
		"Description1":             "myDescription 1",
		"Description1Update":       "myDescription 1 updated",
		"Description2":             "myDescription 2",
		"Description3":             "myDescription 3",
		"Description4":             "myDescription 4",
		"Description4Update":       "myDescription 4 updated",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaLibraryCertificateResource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-update"
	configText2 := templateFill(testAccVcfaLibraryCertificateResourceUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-ds"
	configText3 := templateFill(testAccVcfaLibraryCertificateDatasource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddressOrgCert := "vcfa_certificate.orgCertificate"
	resourceAddressOrgPrivateCert := "vcfa_certificate.OrgWithPrivateCertificate"
	resourceAddressSysCert := "vcfa_certificate.sysCertificate"
	resourceAddressSysPrivateCert := "vcfa_certificate.sysCertificateWithPrivate"

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
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(resourceAddressOrgCert, "data.vcfa_certificate.existing", []string{"%", "private_key", "private_key_passphrase"}),
					resourceFieldsEqual(resourceAddressOrgCert, "data.vcfa_certificate.existingById", []string{"%", "private_key", "private_key_passphrase"}),
					resourceFieldsEqual(resourceAddressSysPrivateCert, "data.vcfa_certificate.existingSystem", []string{"%", "private_key", "private_key_passphrase"}),
					resourceFieldsEqual(resourceAddressSysPrivateCert, "data.vcfa_certificate.existingSystemById", []string{"%", "private_key", "private_key_passphrase"}),
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
}

const testAccVcfaLibraryCertificateResource = `
resource "vcfa_org" "org1" {
  name              = "{{.Org}}"
  display_name      = "{{.Org}}"
  description       = "{{.Org}}"
}

data "vcfa_org" "system" {
  name = "System"
}

resource "vcfa_certificate" "orgCertificate" {
  org_id      = vcfa_org.org1.id
  alias       = "{{.Alias}}"
  description = "{{.Description1}}"
  certificate = file("{{.Certificate1Path}}")
}

resource "vcfa_certificate" "OrgWithPrivateCertificate" {
  org_id                 = vcfa_org.org1.id
  alias                  = "{{.AliasPrivate}}"
  description            = "{{.Description2}}"
  certificate            = file("{{.Certificate2Path}}")
  private_key            = file("{{.PrivateKey2}}")
  private_key_passphrase = "{{.PassPhrase}}"
}

resource "vcfa_certificate" "sysCertificate" {
  org_id      = data.vcfa_org.system.id
  alias       = "{{.AliasSystem}}"
  description = "{{.Description3}}"
  certificate = file("{{.Certificate1Path}}")
}

resource "vcfa_certificate" "sysCertificateWithPrivate" {
  org_id                 = data.vcfa_org.system.id
  alias                  = "{{.AliasPrivateSystem}}"
  description            = "{{.Description4}}"
  certificate            = file("{{.Certificate2Path}}")
  private_key            = file("{{.PrivateKey2}}")
  private_key_passphrase = "{{.PassPhrase}}"
}
`

const testAccVcfaLibraryCertificateResourceUpdate = `
resource "vcfa_org" "org1" {
  name              = "{{.Org}}"
  display_name      = "{{.Org}}"
  description       = "{{.Org}}"
}

data "vcfa_org" "system" {
  name = "System"
}

resource "vcfa_certificate" "orgCertificate" {
  org_id      = vcfa_org.org1.id
  alias       = "{{.AliasUpdate}}"
  description = "{{.Description1Update}}"
  certificate = file("{{.Certificate1Path}}")
}

resource "vcfa_certificate" "sysCertificateWithPrivate" {
  org_id                 = data.vcfa_org.system.id
  alias                  = "{{.AliasPrivateSystemUpdate}}"
  description            = "{{.Description4Update}}"
  certificate            = file("{{.Certificate2Path}}")
  private_key            = file("{{.PrivateKey2}}")
  private_key_passphrase = "{{.PassPhrase}}"
}
`

const testAccVcfaLibraryCertificateDatasource = testAccVcfaLibraryCertificateResourceUpdate + `
data "vcfa_certificate" "existing" {
  org_id = vcfa_org.org1.id
  alias  = vcfa_certificate.orgCertificate.alias
}

data "vcfa_certificate" "existingById" {
  org_id = vcfa_org.org1.id
  id     = vcfa_certificate.orgCertificate.id
}

data "vcfa_certificate" "existingSystem" {
  org_id = data.vcfa_org.system.id
  alias  = vcfa_certificate.sysCertificateWithPrivate.alias
}

data "vcfa_certificate" "existingSystemById" {
  org_id = data.vcfa_org.system.id
  id     = vcfa_certificate.sysCertificateWithPrivate.id
}
`
