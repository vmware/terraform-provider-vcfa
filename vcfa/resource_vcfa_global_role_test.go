//go:build role || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TODO: TM: Review whether this test should be skipped when an API Token or service account
// is provided instead of user + password, in test configuration
func TestAccVcfaGlobalRole(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	var globalRoleName = t.Name()
	var globalRoleUpdateName = t.Name() + "-update"
	var globalRoleDescription = "A long description containing some text."
	var globalRoleUpdateDescription = "A shorter description."

	var params = StringMap{
		"Org":                         testConfig.Tm.Org,
		"GlobalRoleName":              globalRoleName,
		"GlobalRoleUpdateName":        globalRoleUpdateName,
		"GlobalRoleDescription":       globalRoleDescription,
		"GlobalRoleUpdateDescription": globalRoleUpdateDescription,
		"FuncName":                    globalRoleName,
		"Tags":                        "tm role",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccGlobalRole, params)

	params["FuncName"] = globalRoleUpdateName
	params["GlobalRoleDescription"] = globalRoleUpdateDescription
	configTextUpdate := templateFill(testAccGlobalRoleUpdate, params)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION basic: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	resourceDef := "vcfa_global_role." + globalRoleName
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckGlobalRoleDestroy(resourceDef),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRoleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", globalRoleName),
					resource.TestCheckResourceAttr(resourceDef, "description", globalRoleDescription),
					resource.TestCheckResourceAttr(resourceDef, "publish_to_all_orgs", "false"),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "4"),
					resource.TestCheckResourceAttr(resourceDef, "org_ids.#", "1"),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalRoleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", globalRoleUpdateName),
					resource.TestCheckResourceAttr(resourceDef, "description", globalRoleUpdateDescription),
					resource.TestCheckResourceAttr(resourceDef, "publish_to_all_orgs", "true"),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "2"),
				),
			},
			{
				ResourceName:      resourceDef,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) { return globalRoleUpdateName, nil },
			},
		},
	})
}

func testAccCheckGlobalRoleExists(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaGlobalRole)
		}

		conn := testAccProvider.Meta().(ClientContainer).tmClient

		_, err := conn.Client.GetGlobalRoleById(rs.Primary.ID)
		return err
	}
}

func testAccCheckGlobalRoleDestroy(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaGlobalRole)
		}

		conn := testAccProvider.Meta().(ClientContainer).tmClient

		_, err := conn.Client.GetGlobalRoleById(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("%s not deleted yet", identifier)
		}
		return nil

	}
}

const testAccGlobalRole = `
resource "vcfa_org" "org1" {
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "{{.Org}}"
}

resource "vcfa_global_role" "{{.GlobalRoleName}}" {
  name        = "{{.GlobalRoleName}}"
  description = "{{.GlobalRoleDescription}}"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = false
  org_ids             = [ vcfa_org.org1.id ]
}
`

const testAccGlobalRoleUpdate = `
resource "vcfa_org" "org1" {
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "{{.Org}}"
}

resource "vcfa_global_role" "{{.GlobalRoleName}}" {
  name        = "{{.GlobalRoleUpdateName}}"
  description = "{{.GlobalRoleUpdateDescription}}"
  rights = [
    # "Content Library: View",
    # "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = true
}
`
