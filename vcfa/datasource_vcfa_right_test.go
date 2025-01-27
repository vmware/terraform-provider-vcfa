//go:build tm || rights || ALL || functional

package vcfa

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaRight(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Name": "Organization: Edit Limits",
		"Tags": "tm rights",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcfaRight, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	datasourceName := "data.vcfa_right.right"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(datasourceName, "id", regexp.MustCompile(`^urn:vcloud:right:.+$`)),
					resource.TestCheckResourceAttr(datasourceName, "name", params["Name"].(string)),
					resource.TestCheckResourceAttr(datasourceName, "bundle_key", "RIGHT_ORG_OPERATIONS_LIMIT_EDIT"),
					resource.TestMatchResourceAttr(datasourceName, "category_id", regexp.MustCompile(`^urn:vcloud:rightsCategory:.+$`)),
					resource.TestCheckResourceAttr(datasourceName, "description", "Organization: Edit Limits"),
					resource.TestCheckResourceAttr(datasourceName, "implied_rights.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "right_type", "MODIFY"),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaRight = `
data "vcfa_right" "right" {
  name = "{{.Name}}"
}
`
