/*
 * // © Broadcom. All Rights Reserved.
 * // The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
 * // SPDX-License-Identifier: MPL-2.0
 */

//go:build cci || ALL || functional

package vcfa

import (
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAccVcfaSupervisorNamespaceExternal(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
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

	// Setup project and defer cleanup
	cleanup := setupProject(t, params["ProjectName"].(string))
	defer cleanup()

	configText1 := templateFill(testAccVcfaSupervisorNamespaceExternalStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcfaSupervisorNamespaceExternalStep2DS, params)
	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcfaSupervisorNamespaceExternalStep4KubeConfig, params)

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
					resource.TestMatchResourceAttr("vcfa_supervisor_namespace.test", "id", regexp.MustCompile(fmt.Sprintf(`^%s:terraform-test`, params["ProjectName"].(string)))),
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
					cachedNamespaceName.testCheckCachedResourceFieldValuePattern("data.vcfa_kubeconfig.test-namespace", "id", fmt.Sprintf("%s:%%s:%s", testConfig.Org.Name, params["ProjectName"].(string))),
					cachedNamespaceName.testCheckCachedResourceFieldValuePattern("data.vcfa_kubeconfig.test-namespace", "context_name", fmt.Sprintf("%s:%%s:%s", testConfig.Org.Name, params["ProjectName"].(string))),
					resource.TestCheckResourceAttr("data.vcfa_kubeconfig.test-namespace", "insecure_skip_tls_verify", fmt.Sprintf("%t", testConfig.Provider.AllowInsecure)),
					resource.TestCheckResourceAttr("data.vcfa_kubeconfig.test-namespace", "user", fmt.Sprintf("%s:%s@%s", testConfig.Org.Name, testConfig.Org.User, ref.Host)),
					resource.TestCheckResourceAttrSet("data.vcfa_kubeconfig.test-namespace", "token"),
					resource.TestCheckResourceAttrSet("data.vcfa_kubeconfig.test-namespace", "kube_config_raw"),
				),
			},
		},
	})
}

const testAccVcfaSupervisorNamespaceExternalStep1 = `
resource "vcfa_supervisor_namespace" "test" {
  name_prefix  = "terraform-test"
  project_name = "{{.ProjectName}}"
  class_name   = "small"
  description  = "Supervisor Namespace created by Terraform"
  region_name  = "{{.RegionName}}"
  vpc_name     = "{{.VpcName}}"

  storage_classes_initial_class_config_overrides {
    limit     = "200Mi"
    name      = "{{.StorageClassName}}"
  }

  zones_initial_class_config_overrides {
    cpu_limit          = "100M"
    cpu_reservation    = "1M"
    memory_limit       = "200Mi"
    memory_reservation = "2Mi"
    name               = "{{.SupervisorZoneName}}"
  }
}
`

const testAccVcfaSupervisorNamespaceExternalStep2DS = testAccVcfaSupervisorNamespaceExternalStep1 + `
data "vcfa_supervisor_namespace" "test" {
  name         = vcfa_supervisor_namespace.test.name
  project_name = vcfa_supervisor_namespace.test.project_name
}
`

const testAccVcfaSupervisorNamespaceExternalStep4KubeConfig = testAccVcfaSupervisorNamespaceExternalStep1 + `
data "vcfa_supervisor_namespace" "test" {
  name         = vcfa_supervisor_namespace.test.name
  project_name = vcfa_supervisor_namespace.test.project_name
}

data "vcfa_kubeconfig" "test-namespace" {
  project_name              = vcfa_supervisor_namespace.test.project_name
  supervisor_namespace_name = vcfa_supervisor_namespace.test.name
}
`

func setupProject(t *testing.T, projectName string) func() {
	// setup project
	tmClient := createTemporaryVCFAConnection(false)

	projectCfg := &ccitypes.Project{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.ProjectKind,
			APIVersion: ccitypes.ProjectAPI + "/" + ccitypes.ProjectVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: projectName,
		},
		Spec: ccitypes.ProjectSpec{
			Description: fmt.Sprintf("Terraform test project [%s]", projectName),
		},
	}

	newProjectAddr, err := tmClient.Client.GetEntityUrl(ccitypes.ProjectsURL)
	if err != nil {
		t.Fatalf("error creating URL for new project")
	}

	newProject := &ccitypes.Project{}
	// Create
	err = tmClient.Client.PostEntity(newProjectAddr, nil, projectCfg, newProject, nil)
	if err != nil {
		t.Fatalf("error creating project %s: %s", projectCfg.Name, err)
	}

	// defer project cleanup
	return func() {
		projectAddr, err := tmClient.Client.GetEntityUrl(ccitypes.ProjectsURL, "/", projectCfg.Name)
		if err != nil {
			t.Fatalf("error getting Project url: %s", err)
		}
		err = tmClient.Client.DeleteEntity(projectAddr, nil, nil)
		if err != nil {
			t.Fatalf("failed removing Project: %s", err)
		}
	}
}
