// © Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

//go:build org || tm || ALL || functional

package vcfa

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaOrgLocalUser(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
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
	configText3 := templateFill(testAccVcfaLocalUserStep4DS, params)

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
					resource.TestCheckResourceAttr("vcfa_org_local_user.test", "username", "new-"+params["Username"].(string)),
					resource.TestCheckResourceAttr("vcfa_org_local_user.test", "password", "long-change-ME1"),
					resource.TestCheckResourceAttr("vcfa_org_local_user.test", "role_ids.#", "1"),
					resource.TestCheckResourceAttrPair("vcfa_org_local_user.test", "org_id", "vcfa_org.test", "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org_local_user.test", "username", params["Username"].(string)),
					resource.TestCheckResourceAttr("vcfa_org_local_user.test", "password", "long-change-ME1-MORE"),
					resource.TestCheckResourceAttr("vcfa_org_local_user.test", "role_ids.#", "2"),
					resource.TestCheckResourceAttrPair("vcfa_org_local_user.test", "org_id", "vcfa_org.test", "id"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_org_local_user.test", "data.vcfa_org_local_user.test", []string{"%", "password"}),
				),
			},
			{
				ResourceName:            "vcfa_org_local_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           params["Testname"].(string) + ImportSeparator + params["Username"].(string),
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
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
resource "vcfa_org_local_user" "test" {
  org_id    = vcfa_org.test.id
  role_ids  = [data.vcfa_role.org-admin.id]
  username  = "new-{{.Username}}"
  password  = "long-change-ME1"
}
`

const testAccVcfaLocalUserStep2 = testAccVcfaLocalUserPrerequisites + `
resource "vcfa_org_local_user" "test" {
  org_id    = vcfa_org.test.id
  role_ids  = [data.vcfa_role.org-user.id, data.vcfa_role.org-admin.id]
  username  = "{{.Username}}"
  password  = "long-change-ME1-MORE"
}
`

const testAccVcfaLocalUserStep4DS = testAccVcfaLocalUserStep2 + `
data "vcfa_org_local_user" "test" {
  org_id   = vcfa_org.test.id
  username = vcfa_org_local_user.test.username
}
`
