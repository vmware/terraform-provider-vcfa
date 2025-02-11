//go:build cci || ALL || functional

package vcfa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcfaSupervisorNamespace(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"Testname":           t.Name(),
		"ProjectName":        "tf-project",
		"RegionName":         "terraform-demo",
		"VpcName":            fmt.Sprintf("%s-Default-VPC", "terraform-demo"), // region-name + '-Default-VPC' suffix
		"StorageClassName":   "vSAN Default Storage Policy",
		"SupervisorZoneName": "vcfa-gen-wl-vc08-cl1-zone1",

		"Tags": "cci",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaSupervisorNamespaceStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcfaSupervisorNamespaceStep2DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)

	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedProjectName := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{

				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_project.test", "id", "test-project"),
					resource.TestCheckResourceAttr("vcfa_project.test", "name", "test-project"),
					resource.TestCheckResourceAttr("vcfa_project.test", "description", "description"),

					resource.TestMatchResourceAttr("vcfa_supervisor_namespace.test", "id", regexp.MustCompile(`^test-project:terraform-test`)),
					resource.TestMatchResourceAttr("vcfa_supervisor_namespace.test", "name", regexp.MustCompile(`^terraform-test`)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "description", "Supervisor Namespace created by Terraform"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "region_name", params["RegionName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "vpc_name", params["VpcName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_initial_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_initial_class_config_overrides.#", "1"),
					cachedProjectName.cacheTestResourceFieldValue("vcfa_supervisor_namespace.test", "name"), // capturing computed 'name' to use for import test
				),
			},
			{

				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					// Data source does not have 'name_prefix' therefore field count (%) differs
					resourceFieldsEqual("data.vcfa_supervisor_namespace.test", "vcfa_supervisor_namespace.test", []string{"%"}),
				),
			},
			{
				ResourceName:            "vcfa_supervisor_namespace.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "test-project" + ImportSeparator + cachedProjectName.fieldValue, nil
				},
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaSupervisorNamespaceStep1 = `
resource "vcfa_project" "test" {
  name        = "test-project"
  description = "description"
}

resource "vcfa_supervisor_namespace" "test" {
  name_prefix  = "terraform-test"
  project_name = vcfa_project.test.name
  class_name   = "small"
  description  = "Supervisor Namespace created by Terraform"
  region_name  = "{{.RegionName}}"
  vpc_name     = "{{.VpcName}}"

  storage_classes_initial_class_config_overrides {
    limit_mib = 200
    name      = "{{.StorageClassName}}"
  }

  zones_initial_class_config_overrides {
    cpu_limit_mhz          = 100
    cpu_reservation_mhz    = 1
    memory_limit_mib       = 200
    memory_reservation_mib = 2
    name                   = "{{.SupervisorZoneName}}"
  }
}
`

const testAccVcfaSupervisorNamespaceStep2DS = testAccVcfaSupervisorNamespaceStep1 + `
data "vcfa_supervisor_namespace" "test" {
  name         = vcfa_supervisor_namespace.test.name
  project_name = vcfa_supervisor_namespace.test.project_name
}
`
