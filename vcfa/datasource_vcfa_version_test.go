//go:build ALL || tm || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVcfaVersion(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	tmClient := createSystemTemporaryVCFAConnection()
	currentVersion, err := tmClient.Client.GetVcdShortVersion()
	if err != nil {
		t.Fatalf("could not get VCFA version: %s", err)
	}

	apiVersion, err := tmClient.VCDClient.Client.MaxSupportedVersion()
	if err != nil {
		t.Fatalf("could not get VCFA API version: %s", err)
	}

	var params = StringMap{
		"SkipBinaryTest": " ",
		"Condition":      ">= 99.99.99",
		"FailIfNotMatch": "false",
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVcfaTmVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s", step1)

	params["FuncName"] = t.Name() + "-step2"
	params["FailIfNotMatch"] = "true"
	params["SkipBinaryTest"] = "# skip-binary-test - This one triggers an error"
	step2 := templateFill(testAccVcfaTmVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s", step2)

	params["FuncName"] = t.Name() + "-step3"
	params["Condition"] = "= " + currentVersion
	params["SkipBinaryTest"] = " "
	step3 := templateFill(testAccVcfaTmVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s", step3)

	params["FuncName"] = t.Name() + "-step4"
	versionTokens := strings.Split(currentVersion, ".")
	params["Condition"] = fmt.Sprintf("~> %s.%s", versionTokens[0], versionTokens[1])
	step4 := templateFill(testAccVcfaTmVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s", step4)

	params["FuncName"] = t.Name() + "-step5"
	params["Condition"] = "!= 10.3.0"
	step5 := templateFill(testAccVcfaTmVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step5: %s", step5)

	params["FuncName"] = t.Name() + "-step6"
	params["Condition"] = " " // Not used, but illustrates the point of this check
	params["FailIfNotMatch"] = " "
	step6 := templateFill(testAccVcfaTmVersionWithoutArguments, params)
	debugPrintf("#[DEBUG] CONFIGURATION step6: %s", step6)

	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcfa_version.version", "id", fmt.Sprintf("version='%s',condition='>= 99.99.99',fail_if_not_match='false'", currentVersion)),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "version", currentVersion),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "api_version", apiVersion),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "matches_condition", "false"),
				),
			},
			{
				Config:      step2,
				ExpectError: regexp.MustCompile(fmt.Sprintf(`the VCFA version '%s' doesn't match the version constraint '>= 99.99.99'`, currentVersion)),
			},
			{
				Config: step3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcfa_version.version", "id", fmt.Sprintf("version='%s',condition='= %s',fail_if_not_match='true'", currentVersion, currentVersion)),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "version", currentVersion),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "api_version", apiVersion),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "matches_condition", "true"),
				),
			},
			{
				Config: step4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcfa_version.version", "id", fmt.Sprintf("version='%s',condition='~> %s.%s',fail_if_not_match='true'", currentVersion, versionTokens[0], versionTokens[1])),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "version", currentVersion),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "api_version", apiVersion),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "matches_condition", "true"),
				),
			},
			{
				Config: step5,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcfa_version.version", "id", fmt.Sprintf("version='%s',condition='!= 10.3.0',fail_if_not_match='true'", currentVersion)),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "version", currentVersion),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "api_version", apiVersion),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "matches_condition", "true"),
				),
			},
			{
				Config: step6,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcfa_version.version", "id", fmt.Sprintf("version='%s',condition='',fail_if_not_match='false'", currentVersion)),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "version", currentVersion),
					resource.TestCheckResourceAttr("data.vcfa_version.version", "api_version", apiVersion),
					resource.TestCheckNoResourceAttr("data.vcfa_version.version", "matches_condition"),
				),
			},
		},
	})
}

const testAccVcfaTmVersion = `
{{.SkipBinaryTest}}
data "vcfa_version" "version" {
  condition         = "{{.Condition}}"
  fail_if_not_match = {{.FailIfNotMatch}}
}
`

const testAccVcfaTmVersionWithoutArguments = `
data "vcfa_version" "version" {
}
`
