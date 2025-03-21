//go:build tm || ALL || functional

package vcfa

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	registeredPriorityTests = append(registeredPriorityTests,
		priorityTest{Name: "TestAccVcfaNsxManager", Test: TestAccVcfaNsxManager},
	)
}

func TestAccVcfaNsxManager(t *testing.T) {
	testName := "TestAccVcfaNsxManager"
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	if !testConfig.Tm.CreateNsxManager {
		t.Skipf("Skipping NSX Manager creation")
	}

	var params = StringMap{
		"Testname": testName,
		"Username": testConfig.Tm.NsxManagerUsername,
		"Password": testConfig.Tm.NsxManagerPassword,
		"Url":      testConfig.Tm.NsxManagerUrl,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = testName
	configText1 := templateFill(testAccVcfaNsxManagerStep1, params)
	params["FuncName"] = testName + "-step2"
	configText2 := templateFill(testAccVcfaNsxManagerStep2, params)
	params["FuncName"] = testName + "-step3"
	configText3 := templateFill(testAccVcfaNsxManagerStep3DS, params)

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
					resource.TestMatchResourceAttr("vcfa_nsx_manager.test", "id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:`)),
					resource.TestMatchResourceAttr("vcfa_nsx_manager.test", "href", regexp.MustCompile(`api/admin/extension/nsxtManagers/`)),
					resource.TestCheckResourceAttr("vcfa_nsx_manager.test", "name", params["Testname"].(string)),
					resource.TestCheckResourceAttr("vcfa_nsx_manager.test", "description", "terraform test"),
					resource.TestCheckResourceAttrSet("vcfa_nsx_manager.test", "status"),
					resource.TestCheckResourceAttr("vcfa_nsx_manager.test", "url", params["Url"].(string)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcfa_nsx_manager.test", "id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:`)),
					resource.TestMatchResourceAttr("vcfa_nsx_manager.test", "href", regexp.MustCompile(`api/admin/extension/nsxtManagers/`)),
					resource.TestCheckResourceAttr("vcfa_nsx_manager.test", "name", params["Testname"].(string)),
					resource.TestCheckResourceAttr("vcfa_nsx_manager.test", "description", ""),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_nsx_manager.test", "data.vcfa_nsx_manager.test", []string{"%", "auto_trust_certificate", "password"}),
				),
			},
			{
				ResourceName:            "vcfa_nsx_manager.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           params["Testname"].(string),
				ImportStateVerifyIgnore: []string{"auto_trust_certificate", "password"},
			},
		},
	})
}

const testAccVcfaNsxManagerStep1 = `
# skip-binary-test: NSX Manager configuration is tested in next tests
resource "vcfa_nsx_manager" "test" {
  name                   = "{{.Testname}}"
  description            = "terraform test"
  username               = "{{.Username}}"
  password               = "{{.Password}}"
  url                    = "{{.Url}}"
  auto_trust_certificate = true
}
`
const testAccVcfaNsxManagerStep2 = `
# skip-binary-test: NSX Manager configuration is tested in next tests
resource "vcfa_nsx_manager" "test" {
  name                   = "{{.Testname}}"
  description            = ""
  username               = "{{.Username}}"
  password               = "{{.Password}}"
  url                    = "{{.Url}}"
  auto_trust_certificate = true
}
`

const testAccVcfaNsxManagerStep3DS = testAccVcfaNsxManagerStep1 + `
# skip-binary-test: NSX Manager configuration is tested in next tests
data "vcfa_nsx_manager" "test" {
  name = vcfa_nsx_manager.test.name
}
`
