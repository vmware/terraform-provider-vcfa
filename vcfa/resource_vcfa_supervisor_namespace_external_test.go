//go:build cci || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
		"Testname":               t.Name(),
		"ProjectName":            "tf-project",
		"RegionName":             testConfig.Cci.Region,
		"VpcName":                testConfig.Cci.Vpc,
		"StorageClassName":       testConfig.Cci.StoragePolicy,
		"StorageLimit":           "200Mi",
		"StorageLimitUpdated":    "210Mi",
		"SupervisorZoneName":     testConfig.Cci.SupervisorZone,
		"ContentLibrary":         testConfig.Cci.ContentLibrary,
		"InfraPolicyName":        testConfig.Cci.InfraPolicyName,
		"SharedSubnetName":       testConfig.Cci.SharedSubnetName,
		"VmClass1":               testConfig.Cci.VmClass1,
		"VmClass2":               testConfig.Cci.VmClass2,
		"Description":            "Supervisor Namespace created by Terraform",
		"DescriptionUpdated":     "Supervisor Namespace updated by Terraform",
		"ZoneCpuLimit":           "100M",
		"ZoneCpuLimitUpdated":    "110M",
		"ZoneMemoryLimit":        "200Mi",
		"ZoneMemoryLimitUpdated": "210Mi",

		"Tags": "cci",
	}
	testParamsNotEmpty(t, params)

	// Setup project and defer cleanup
	cleanup := setupProject(t, params["ProjectName"].(string))
	defer cleanup()

	configText1 := templateFill(testAccVcfaSupervisorNamespaceExternalStep1, params)
	params["FuncName"] = t.Name() + "-step3"
	configText2 := templateFill(testAccVcfaSupervisorNamespaceExternalStep2Update, params)
	params["FuncName"] = t.Name() + "-step2"
	configText3 := templateFill(testAccVcfaSupervisorNamespaceExternalStep3DS, params)
	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccVcfaSupervisorNamespaceExternalStep5KubeConfig, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step5: %s\n", configText5)

	cachedNamespaceName := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{

				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcfa_supervisor_namespace.test", "id", regexp.MustCompile(fmt.Sprintf(`^%s:terraform-test`, params["ProjectName"].(string)))),
					resource.TestMatchResourceAttr("vcfa_supervisor_namespace.test", "name", regexp.MustCompile(`^terraform-test`)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "description", params["Description"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "region_name", params["RegionName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "vpc_name", params["VpcName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "infra_policy_names.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "shared_subnet_names.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_initial_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_class_config_overrides.0.limit", params["StorageLimit"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_class_config_overrides.0.name", params["StorageClassName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "vm_classes_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_class_config_overrides.0.cpu_limit", params["ZoneCpuLimit"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_class_config_overrides.0.memory_limit", params["ZoneMemoryLimit"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_initial_class_config_overrides.#", "1"),
					cachedNamespaceName.cacheTestResourceFieldValue("vcfa_supervisor_namespace.test", "name"), // capturing computed 'name' to use for other test steps
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "description", params["DescriptionUpdated"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "region_name", params["RegionName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "vpc_name", params["VpcName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "content_sources_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "infra_policy_names.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "shared_subnet_names.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_initial_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_class_config_overrides.0.limit", params["StorageLimitUpdated"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "storage_classes_class_config_overrides.0.name", params["StorageClassName"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "vm_classes_class_config_overrides.#", "2"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_class_config_overrides.#", "1"),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_class_config_overrides.0.cpu_limit", params["ZoneCpuLimitUpdated"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_class_config_overrides.0.memory_limit", params["ZoneMemoryLimitUpdated"].(string)),
					resource.TestCheckResourceAttr("vcfa_supervisor_namespace.test", "zones_initial_class_config_overrides.#", "1"),
				),
			},
			{
				Config: configText3,
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
				Config: configText5,
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
  name_prefix         = "terraform-test"
  project_name        = "{{.ProjectName}}"
  class_name          = "small"
  description         = "Supervisor Namespace created by Terraform"
  infra_policy_names  = [ "{{.InfraPolicyName}}" ]
  region_name         = "{{.RegionName}}"
  shared_subnet_names = [ "{{.SharedSubnetName}}" ]
  vpc_name            = "{{.VpcName}}"

  storage_classes_class_config_overrides {
    limit = "{{.StorageLimit}}"
    name  = "{{.StorageClassName}}"
  }

  vm_classes_class_config_overrides {
    name = "{{.VmClass1}}"
  }

  zones_class_config_overrides {
    cpu_limit          = "{{.ZoneCpuLimit}}"
    cpu_reservation    = "0M"
    memory_limit       = "{{.ZoneMemoryLimit}}"
    memory_reservation = "0Mi"
    name               = "{{.SupervisorZoneName}}"
  }
}
`

const testAccVcfaSupervisorNamespaceExternalStep2Update = `
resource "vcfa_supervisor_namespace" "test" {
  name_prefix         = "terraform-test"
  project_name        = "{{.ProjectName}}"
  class_name          = "small"
  description         = "{{.DescriptionUpdated}}"
  infra_policy_names  = [ "{{.InfraPolicyName}}" ]
  region_name         = "{{.RegionName}}"
  shared_subnet_names = [ "{{.SharedSubnetName}}" ]
  vpc_name            = "{{.VpcName}}"

  content_sources_class_config_overrides {
    name = "{{.ContentLibrary}}"
    type = "ContentLibrary"
  }

  storage_classes_class_config_overrides {
    limit = "{{.StorageLimitUpdated}}"
    name  = "{{.StorageClassName}}"
  }

  vm_classes_class_config_overrides {
    name = "{{.VmClass1}}"
  }

  vm_classes_class_config_overrides {
    name = "{{.VmClass2}}"
  }

  zones_class_config_overrides {
    cpu_limit          = "{{.ZoneCpuLimitUpdated}}"
    cpu_reservation    = "0M"
    memory_limit       = "{{.ZoneMemoryLimitUpdated}}"
    memory_reservation = "0Mi"
    name               = "{{.SupervisorZoneName}}"
  }
}
`

const testAccVcfaSupervisorNamespaceExternalStep3DS = testAccVcfaSupervisorNamespaceExternalStep2Update + `
data "vcfa_supervisor_namespace" "test" {
  name         = vcfa_supervisor_namespace.test.name
  project_name = vcfa_supervisor_namespace.test.project_name
}
`

const testAccVcfaSupervisorNamespaceExternalStep5KubeConfig = testAccVcfaSupervisorNamespaceExternalStep2Update + `
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
