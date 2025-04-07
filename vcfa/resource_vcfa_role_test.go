//go:build tm || role || ALL || functional

// © Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"testing"

	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TODO: TM: Review whether this test should be skipped when an API Token or service account
// is provided instead of user + password, in test configuration
func TestAccVcfaRole(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	var roleName = t.Name()
	var roleUpdateName = t.Name() + "-update"
	var roleDescription = "A long description containing some text."
	var roleUpdateDescription = "A shorter description."

	var params = StringMap{
		"Org":                   testConfig.Tm.Org,
		"RoleName":              roleName,
		"RoleUpdateName":        roleUpdateName,
		"RoleDescription":       roleDescription,
		"RoleUpdateDescription": roleUpdateDescription,
		"FuncName":              roleName,
		"Tags":                  "tm role",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccRole, params)

	params["FuncName"] = roleUpdateName
	params["roleDescription"] = roleUpdateDescription
	configTextUpdate := templateFill(testAccRoleUpdate, params)

	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION basic: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	resourceRole := "vcfa_role." + roleName
	resourceOrg := "vcfa_org.org1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRoleDestroy(resourceOrg, resourceRole),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(resourceOrg, resourceRole),
					resource.TestCheckResourceAttr(resourceRole, "name", roleName),
					resource.TestCheckResourceAttr(resourceRole, "rights.#", "4"),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(resourceOrg, resourceRole),
					resource.TestCheckResourceAttr(resourceRole, "name", roleUpdateName),
					resource.TestCheckResourceAttr(resourceRole, "description", roleUpdateDescription),
					resource.TestCheckResourceAttr(resourceRole, "rights.#", "2"),
				),
			},
			{
				ResourceName:      resourceRole,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return testConfig.Tm.Org + ImportSeparator + roleUpdateName, nil
				},
			},
		},
	})
}

func testAccCheckRoleExists(orgId, roleId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get Org
		rsOrg, ok := s.RootModule().Resources[orgId]
		if !ok {
			return fmt.Errorf("not found: %s", orgId)
		}
		if rsOrg.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaOrg)
		}

		// Get Role
		rsRole, ok := s.RootModule().Resources[roleId]
		if !ok {
			return fmt.Errorf("not found: %s", roleId)
		}

		if rsRole.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaRole)
		}

		conn := testAccProvider.Meta().(ClientContainer).tmClient

		org, err := conn.GetAdminOrgById(rsOrg.Primary.ID)
		if err != nil {
			return err
		}
		_, err = org.GetRoleById(rsRole.Primary.ID)
		return err
	}
}

func testAccCheckRoleDestroy(orgId, roleId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Get Org
		rsOrg, ok := s.RootModule().Resources[orgId]
		if !ok {
			return fmt.Errorf("not found: %s", orgId)
		}
		if rsOrg.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaOrg)
		}

		// Get Role
		rsRole, ok := s.RootModule().Resources[roleId]
		if !ok {
			return fmt.Errorf("not found: %s", roleId)
		}

		if rsRole.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaRole)
		}

		conn := testAccProvider.Meta().(ClientContainer).tmClient

		// TODO: TM: Change to tmClient.GetTmOrgById(orgId), requires implementing Role support for that type
		org, err := conn.GetAdminOrgById(rsOrg.Primary.ID)
		if err != nil {
			// TODO: TM: Would be nice to have a method to retrieve the role without an Org. This way we can check
			// the role is correctly destroyed. Otherwise, the Org gets deleted and we cannot check further with the
			// existing methods.
			if govcd.ContainsNotFound(err) {
				return nil
			}
			return err
		}
		_, err = org.GetRoleById(rsRole.Primary.ID)

		if err == nil {
			return fmt.Errorf("%s not deleted yet", roleId)
		}
		return nil

	}
}

const testAccRole = `
resource "vcfa_org" "org1" {
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "{{.Org}}"
}

resource "vcfa_role" "{{.RoleName}}" {
  org_id      = vcfa_org.org1.id
  name        = "{{.RoleName}}"
  description = "{{.RoleDescription}}"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
}
`

const testAccRoleUpdate = `
resource "vcfa_org" "org1" {
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "{{.Org}}"
}

resource "vcfa_role" "{{.RoleName}}" {
  org_id      = vcfa_org.org1.id
  name        = "{{.RoleUpdateName}}"
  description = "{{.RoleUpdateDescription}}"
  rights = [
    # "Content Library: View",
    # "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
}
`
