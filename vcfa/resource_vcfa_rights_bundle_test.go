//go:build tm || role || ALL || functional

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TODO: TM: Review whether this test should be skipped when an API Token or service account
// is provided instead of user + password, in test configuration
func TestAccVcfaRightsBundle(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	var rightsBundleName = t.Name()
	var rightsBundleUpdateName = t.Name() + "-update"
	var rightsBundleDescription = "A long description containing some text."
	var rightsBundleUpdateDescription = "A shorter description."

	var params = StringMap{
		"Org":                           testConfig.Tm.Org,
		"RightsBundleName":              rightsBundleName,
		"RightsBundleUpdateName":        rightsBundleUpdateName,
		"RightsBundleDescription":       rightsBundleDescription,
		"RightsBundleUpdateDescription": rightsBundleUpdateDescription,
		"FuncName":                      rightsBundleName,
		"Tags":                          "tm role",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccRightsBundle, params)

	params["FuncName"] = rightsBundleUpdateName
	params["RightsBundleDescription"] = rightsBundleUpdateDescription
	configTextUpdate := templateFill(testAccRightsBundleUpdate, params)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION basic: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	resourceDef := "vcfa_rights_bundle." + rightsBundleName
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRightsBundleDestroy(resourceDef),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRightsBundleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", rightsBundleName),
					resource.TestCheckResourceAttr(resourceDef, "description", rightsBundleDescription),
					resource.TestCheckResourceAttr(resourceDef, "publish_to_all_orgs", "false"),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "4"),
					resource.TestCheckResourceAttr(resourceDef, "org_ids.#", "1"),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRightsBundleExists(resourceDef),
					resource.TestCheckResourceAttr(resourceDef, "name", rightsBundleUpdateName),
					resource.TestCheckResourceAttr(resourceDef, "description", rightsBundleUpdateDescription),
					resource.TestCheckResourceAttr(resourceDef, "publish_to_all_orgs", "true"),
					resource.TestCheckResourceAttr(resourceDef, "rights.#", "2"),
				),
			},
			{
				ResourceName:      resourceDef,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) { return rightsBundleUpdateName, nil },
			},
		},
	})
}

func testAccCheckRightsBundleExists(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaRightsBundle)
		}

		conn := testAccProvider.Meta().(ClientContainer).tmClient

		_, err := conn.Client.GetRightsBundleById(rs.Primary.ID)
		return err
	}
}

func testAccCheckRightsBundleDestroy(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaRightsBundle)
		}

		conn := testAccProvider.Meta().(ClientContainer).tmClient

		_, err := conn.Client.GetRightsBundleById(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("%s not deleted yet", identifier)
		}
		return nil

	}
}

const testAccRightsBundle = `
resource "vcfa_org" "org1" {
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "{{.Org}}"
}

resource "vcfa_rights_bundle" "{{.RightsBundleName}}" {
  name        = "{{.RightsBundleName}}"
  description = "{{.RightsBundleDescription}}"
  rights = [
    "Content Library: View",
    "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = false
  org_ids                = [ vcfa_org.org1.id ]
}
`

const testAccRightsBundleUpdate = `
resource "vcfa_org" "org1" {
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "{{.Org}}"
}

resource "vcfa_rights_bundle" "{{.RightsBundleName}}" {
  name        = "{{.RightsBundleUpdateName}}"
  description = "{{.RightsBundleUpdateDescription}}"
  rights = [
    # "Content Library: View",
    # "Content Library Item: View",
    "Group / User: View",
    "IP Blocks: View",
  ]
  publish_to_all_orgs = true
}
`
