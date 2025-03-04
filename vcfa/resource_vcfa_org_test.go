//go:build org || tm || ALL || functional

package vcfa

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaOrg(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Testname": t.Name(),
		"Tags":     "tm org",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaOrgStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcfaOrgStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcfaOrgStep3DS, params)

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
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcfa_org.test", "display_name", "terraform-test"),
					resource.TestCheckResourceAttr("vcfa_org.test", "description", "terraform test"),
					resource.TestCheckResourceAttr("vcfa_org.test", "is_enabled", "true"),
					resource.TestMatchResourceAttr("vcfa_org.test", "managed_by_id", regexp.MustCompile("^urn:vcloud:org:")),
					resource.TestCheckResourceAttr("vcfa_org.test", "managed_by_name", "System"),
					resource.TestCheckResourceAttr("vcfa_org.test", "is_classic_tenant", "false"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org.test", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttr("vcfa_org.test", "display_name", "terraform-test"),
					resource.TestCheckResourceAttr("vcfa_org.test", "description", ""),
					resource.TestCheckResourceAttr("vcfa_org.test", "is_enabled", "false"),
					resource.TestMatchResourceAttr("vcfa_org.test", "managed_by_id", regexp.MustCompile("^urn:vcloud:org:")),
					resource.TestCheckResourceAttr("vcfa_org.test", "managed_by_name", "System"),
					resource.TestCheckResourceAttr("vcfa_org.test", "is_classic_tenant", "false"),

					// Test Organization settings
					resource.TestCheckResourceAttr("vcfa_org_settings.allow", "can_create_subscribed_libraries", "true"),
					resource.TestCheckResourceAttr("vcfa_org_settings.allow", "quarantine_content_library_items", "true"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_org.test", "data.vcfa_org.test", nil),

					// Settings are destroyed
					resource.TestCheckResourceAttr("data.vcfa_org_settings.allow_ds", "can_create_subscribed_libraries", "false"),
					resource.TestCheckResourceAttr("data.vcfa_org_settings.allow_ds", "quarantine_content_library_items", "false"),
				),
			},
			{
				ResourceName:      "vcfa_org.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Testname"].(string),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaOrgStep1 = `
resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}
`

const testAccVcfaOrgStep2 = `
resource "vcfa_org" "test" {
  name         = "{{.Testname}}-updated"
  display_name = "terraform-test"
  description  = ""
  is_enabled   = false
}

resource "vcfa_org_settings" "allow" {
  org_id                           = vcfa_org.test.id
  can_create_subscribed_libraries  = true
  quarantine_content_library_items = true
}
`

const testAccVcfaOrgStep3DS = testAccVcfaOrgStep1 + `
data "vcfa_org" "test" {
  name = vcfa_org.test.name
}

data "vcfa_org_settings" "allow_ds" {
  org_id = vcfa_org.test.id
}
`

// TestAccVcfaOrgClassicTenant tests a Tenant Manager Organization configured as "Classic Tenant"
func TestAccVcfaOrgClassicTenant(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Testname": t.Name(),
		"Tags":     "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaOrgClassicStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcfaOrgClassicStep2, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcfa_org.test", "display_name", "terraform-test"),
					resource.TestCheckResourceAttr("vcfa_org.test", "description", "terraform test"),
					resource.TestCheckResourceAttr("vcfa_org.test", "is_enabled", "true"),
					resource.TestMatchResourceAttr("vcfa_org.test", "managed_by_id", regexp.MustCompile("^urn:vcloud:org:")),
					resource.TestCheckResourceAttr("vcfa_org.test", "managed_by_name", "System"),
					resource.TestCheckResourceAttr("vcfa_org.test", "is_classic_tenant", "true"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_org.test", "data.vcfa_org.test", nil),
				),
			},
			{
				ResourceName:      "vcfa_org.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Testname"].(string),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaOrgClassicStep1 = `
resource "vcfa_org" "test" {
  name              = "{{.Testname}}"
  display_name      = "terraform-test"
  description       = "terraform test"
  is_enabled        = true
  is_classic_tenant = true
}

resource "vcfa_org" "test2" {
  name              = "{{.Testname}}2"
  display_name      = "terraform-test"
  description       = "terraform test"
  is_enabled        = true
  is_classic_tenant = true
}
`

const testAccVcfaOrgClassicStep2 = testAccVcfaOrgClassicStep1 + `
data "vcfa_org" "test" {
  name = vcfa_org.test.name
}
`
