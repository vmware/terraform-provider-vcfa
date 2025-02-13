//go:build cci || ALL || functional

package vcfa

import (
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcfaSupervisorNamespace(t *testing.T) {
	preTestChecks(t)
	skipIfSysAdmin(t)

	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	ref, err := url.Parse(testConfig.Provider.Url)
	if err != nil {
		t.Fatalf("failed parsing '%s' host: %s", testConfig.Provider.Url, err)
	}
	var params = StringMap{
		"Testname":           t.Name(),
		"ProjectName":        "tf-project",
		"RegionName":         testConfig.Cci.Region,
		"VpcName":            testConfig.Cci.Vpc,
		"StorageClassName":   testConfig.Cci.StoragePolicy,
		"SupervisorZoneName": testConfig.Cci.SupervisorZone,

		"Tags": "cci",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaSupervisorNamespaceStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcfaSupervisorNamespaceStep2DS, params)
	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcfaSupervisorNamespaceStep4KubeConfig, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)

	cachedNamespaceName := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{

				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_project.test", "id", params["ProjectName"].(string)),
					resource.TestCheckResourceAttr("vcfa_project.test", "name", params["ProjectName"].(string)),
					resource.TestCheckResourceAttr("vcfa_project.test", "description", "description"),

					resource.TestMatchResourceAttr("vcfa_supervisor_namespace.test", "id", regexp.MustCompile(`^tf-project:terraform-test`)),
					resource.TestMatchResourceAttr("vcfa_supervisor_namespace.test", "name", regexp.MustCompile(`^terraform-test`)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "description", "Supervisor Namespace created by Terraform"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "region_name", params["RegionName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "vpc_name", params["VpcName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_initial_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_initial_class_config_overrides.#", "1"),
					cachedNamespaceName.cacheTestResourceFieldValue("vcfa_supervisor_namespace.test", "name"), // capturing computed 'name' to use for other test steps
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
					return params["ProjectName"].(string) + ImportSeparator + cachedNamespaceName.fieldValue, nil
				},
			},
			{

				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					cachedNamespaceName.testCheckCachedResourceFieldValuePattern("data.vcfa_kube_config.test-namespace", "id", fmt.Sprintf("%s:%%s:%s", testConfig.Org.Name, params["ProjectName"].(string))),
					cachedNamespaceName.testCheckCachedResourceFieldValuePattern("data.vcfa_kube_config.test-namespace", "context_name", fmt.Sprintf("%s:%%s:%s", testConfig.Org.Name, params["ProjectName"].(string))),
					resource.TestCheckResourceAttr("data.vcfa_kube_config.test-namespace", "insecure_skip_tls_verify", fmt.Sprintf("%t", testConfig.Provider.AllowInsecure)),
					resource.TestCheckResourceAttr("data.vcfa_kube_config.test-namespace", "user", fmt.Sprintf("%s:%s@%s", testConfig.Org.Name, testConfig.Org.User, ref.Host)),
					resource.TestCheckResourceAttrSet("data.vcfa_kube_config.test-namespace", "token"),
					resource.TestCheckResourceAttrSet("data.vcfa_kube_config.test-namespace", "kube_config_raw"),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaSupervisorNamespaceStep1 = `
resource "vcfa_project" "test" {
  name        = "{{.ProjectName}}"
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

const testAccVcfaSupervisorNamespaceStep4KubeConfig = testAccVcfaSupervisorNamespaceStep1 + `
data "vcfa_supervisor_namespace" "test" {
  name         = vcfa_supervisor_namespace.test.name
  project_name = vcfa_supervisor_namespace.test.project_name
}

data "vcfa_kube_config" "test-namespace" {
  project_name              = vcfa_supervisor_namespace.test.project_name
  supervisor_namespace_name = vcfa_supervisor_namespace.test.name
}
`
