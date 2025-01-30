//go:build org || tm || ALL || functional

package vcfa

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaLocalUser(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Testname": t.Name(),
		"Username": "testlocaluser",

		"Tags": "tm org",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaLocalUserStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcfaLocalUserStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcfaLocalUserStep3DS, params)

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
					resource.TestCheckResourceAttr("vcfa_local_user.test", "username", "new-"+params["Username"].(string)),
					resource.TestCheckResourceAttr("vcfa_local_user.test", "password", "CHANGE-ME"),
					resource.TestCheckResourceAttrPair("vcfa_local_user.test", "role_id", "data.vcfa_role.org-admin", "id"),
					resource.TestCheckResourceAttrPair("vcfa_local_user.test", "org_id", "vcfa_org.test", "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_local_user.test", "username", params["Username"].(string)),
					resource.TestCheckResourceAttr("vcfa_local_user.test", "password", "CHANGE-ME-MORE"),
					resource.TestCheckResourceAttrPair("vcfa_local_user.test", "role_id", "data.vcfa_role.org-user", "id"),
					resource.TestCheckResourceAttrPair("vcfa_local_user.test", "org_id", "vcfa_org.test", "id"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_local_user.test", "data.vcfa_local_user.test", []string{"%", "password"}),
				),
			},
			{
				ResourceName:            "vcfa_local_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           params["Testname"].(string) + ImportSeparator + params["Username"].(string),
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaLocalUserPrerequisites = `
resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}

data "vcfa_role" "org-admin" {
  org_id = vcfa_org.test.id
  name   = "Organization Administrator"
}

data "vcfa_role" "org-user" {
  org_id = vcfa_org.test.id
  name   = "Organization User"
}
`

const testAccVcfaLocalUserStep1 = testAccVcfaLocalUserPrerequisites + `
resource "vcfa_local_user" "test" {
  org_id   = vcfa_org.test.id
  role_id  = data.vcfa_role.org-admin.id
  username = "new-{{.Username}}"
  password = "CHANGE-ME"
}
`

const testAccVcfaLocalUserStep2 = testAccVcfaLocalUserPrerequisites + `
resource "vcfa_local_user" "test" {
  org_id   = vcfa_org.test.id
  role_id  = data.vcfa_role.org-user.id
  username = "{{.Username}}"
  password = "CHANGE-ME-MORE"
}
`

const testAccVcfaLocalUserStep3DS = testAccVcfaLocalUserStep2 + `
data "vcfa_local_user" "test" {
  org_id   = vcfa_org.test.id
  username = vcfa_local_user.test.username
}
`
