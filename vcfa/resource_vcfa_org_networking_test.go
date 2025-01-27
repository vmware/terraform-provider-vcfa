//go:build org || tm || ALL || functional

package vcfa

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaOrgNetworking(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Testname": t.Name(),
		"Tags":     "tm org",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaOrgNetworkingStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcfaOrgNetworkingStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcfaOrgNetworkingStep3DS, params)

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
					resource.TestCheckResourceAttrPair("vcfa_org.test", "id", "vcfa_org_networking.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_networking.test", "log_name", "l-one"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org.test", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttrPair("vcfa_org.test", "id", "vcfa_org_networking.test", "id"),
					resource.TestCheckResourceAttr("vcfa_org_networking.test", "log_name", "l-one-u"),
				),
			},
			{
				ResourceName:      "vcfa_org_networking.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Testname"].(string), // Org name
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_org_networking.test", "data.vcfa_org_networking.test", nil),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaOrgNetworkingStep1 = `
resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}

resource "vcfa_org_networking" "test" {
  org_id   = vcfa_org.test.id
  log_name = "l-one"
}
`

const testAccVcfaOrgNetworkingStep2 = `
resource "vcfa_org" "test" {
  name         = "{{.Testname}}-updated"
  display_name = "terraform-test"
  description  = ""
  is_enabled   = false
}

resource "vcfa_org_networking" "test" {
  org_id   = vcfa_org.test.id
  log_name = "l-one-u"
}
`

const testAccVcfaOrgNetworkingStep3DS = testAccVcfaOrgNetworkingStep2 + `
data "vcfa_org_networking" "test" {
  org_id = vcfa_org.test.id
}
`
