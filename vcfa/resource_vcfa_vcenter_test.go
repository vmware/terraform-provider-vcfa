//go:build tm || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	registeredPriorityTests = append(registeredPriorityTests,
		priorityTest{Name: "TestAccVcfaVcenter", Test: TestAccVcfaVcenter},
		priorityTest{Name: "TestAccVcfaVcenterInvalid", Test: TestAccVcfaVcenterInvalid},
	)
}

func TestAccVcfaVcenter(t *testing.T) {
	testName := "TestAccVcfaVcenter" // Triggering the test at priority will create incorrect t.Name() value
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	if !testConfig.Tm.CreateVcenter {
		t.Skipf("Skipping vCenter creation")
	}

	var params = StringMap{
		"Org":             testConfig.Tm.Org,
		"VcenterUsername": testConfig.Tm.VcenterUsername,
		"VcenterPassword": testConfig.Tm.VcenterPassword,
		"VcenterUrl":      testConfig.Tm.VcenterUrl,
		"NsxUsername":     testConfig.Tm.NsxManagerUsername,
		"NsxPassword":     testConfig.Tm.NsxManagerPassword,
		"NsxUrl":          testConfig.Tm.NsxManagerUrl,

		"Testname": testName,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = testName
	configText1 := templateFill(testAccVcfaVcenterStep1, params)
	params["FuncName"] = testName + "-step2"
	configText2 := templateFill(testAccVcfaVcenterStep2, params)

	params["FuncName"] = testName + "-step3"
	configText3 := templateFill(testAccVcfaVcenterStep3, params)

	params["FuncName"] = testName + "-step4"
	configText4 := templateFill(testAccVcfaVcenterStep4DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)
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
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "id"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "name", testName),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "description", ""),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "has_proxy", "false"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "cluster_health_status"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "is_connected"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "connection_status"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "mode"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "uuid"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "vcenter_version"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "id"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "name", testName+"-rename"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "is_enabled", "false"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "description", "description from Terraform"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "has_proxy", "false"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "cluster_health_status"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "is_connected"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "connection_status"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "mode"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "uuid"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "vcenter_version"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "status", "READY"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "id"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "name", testName),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "description", ""),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "has_proxy", "false"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "cluster_health_status"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "is_connected"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "connection_status"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "mode"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "uuid"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "vcenter_version"),
				),
			},
			{
				ResourceName:            "vcfa_vcenter.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           params["Testname"].(string),
				ImportStateVerifyIgnore: []string{"password", "auto_trust_certificate", "refresh_vcenter_on_read", "refresh_policies_on_read", "refresh_vcenter_on_create", "refresh_policies_on_create", "nsx_manager_id"},
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcfa_vcenter.test", "vcfa_vcenter.test", []string{"%", "nsx_manager_id"}),
				),
			},
		},
	})
}

const testAccVcfaVcenterPrerequisites = `
# skip-binary-test: vCenter server configuration is tested in next tests
resource "vcfa_nsx_manager" "test" {
  name                   = "{{.Testname}}"
  description            = "terraform test"
  username               = "{{.NsxUsername}}"
  password               = "{{.NsxPassword}}"
  url                    = "{{.NsxUrl}}"
  auto_trust_certificate = true
}
`

const testAccVcfaVcenterStep1 = testAccVcfaVcenterPrerequisites + `
resource "vcfa_vcenter" "test" {
  name                     = "{{.Testname}}"
  url                      = "{{.VcenterUrl}}"
  auto_trust_certificate   = true
  refresh_vcenter_on_read  = true
  refresh_policies_on_read = true
  username                 = "{{.VcenterUsername}}"
  password                 = "{{.VcenterPassword}}"
  is_enabled               = true
  nsx_manager_id           = vcfa_nsx_manager.test.id
}
`

const testAccVcfaVcenterStep2 = testAccVcfaVcenterPrerequisites + `
resource "vcfa_vcenter" "test" {
  name                   = "{{.Testname}}-rename"
  description            = "description from Terraform"
  auto_trust_certificate = true
  url                    = "{{.VcenterUrl}}"
  username               = "{{.VcenterUsername}}"
  password               = "{{.VcenterPassword}}"
  is_enabled             = false
  nsx_manager_id         = vcfa_nsx_manager.test.id
}
`

const testAccVcfaVcenterStep3 = testAccVcfaVcenterPrerequisites + `
resource "vcfa_vcenter" "test" {
  name                   = "{{.Testname}}"
  url                    = "{{.VcenterUrl}}"
  auto_trust_certificate = true
  username               = "{{.VcenterUsername}}"
  password               = "{{.VcenterPassword}}"
  is_enabled             = true
  nsx_manager_id         = vcfa_nsx_manager.test.id
}
`

const testAccVcfaVcenterStep4DS = testAccVcfaVcenterStep3 + `
data "vcfa_vcenter" "test" {
  name = vcfa_vcenter.test.name
}
`

func TestAccVcfaVcenterInvalid(t *testing.T) {
	testName := "TestAccVcfaVcenterInvalid" // Triggering the test at priority will create incorrect t.Name() value
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	// test fails on purpose
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	var params = StringMap{
		"Org":             testConfig.Tm.Org,
		"VcenterUsername": testConfig.Tm.VcenterUsername,
		"VcenterPassword": "invalid",
		"VcenterUrl":      testConfig.Tm.VcenterUrl,
		"NsxUsername":     testConfig.Tm.NsxManagerUsername,
		"NsxPassword":     testConfig.Tm.NsxManagerPassword,
		"NsxUrl":          testConfig.Tm.NsxManagerUrl,
		"Testname":        testName,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = testName
	configText1 := templateFill(testAccVcfaVcenterStep1, params)
	params["FuncName"] = testName + "-step2"
	params["VcenterPassword"] = testConfig.Tm.VcenterPassword
	configText2 := templateFill(testAccVcfaVcenterStep1, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      configText1,
				ExpectError: regexp.MustCompile(`Failed to connect to the vCenter`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "id"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "name", testName),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "is_enabled", "true"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "cluster_health_status"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "is_connected"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "connection_status"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "mode"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "uuid"),
					resource.TestCheckResourceAttrSet("vcfa_vcenter.test", "vcenter_version"),
					resource.TestCheckResourceAttr("vcfa_vcenter.test", "status", "READY"),
				),
			},
		},
	})
}
