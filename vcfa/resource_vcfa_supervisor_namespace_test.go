//go:build cci || ALL || functional

package vcfa

import (
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAccVcfaSupervisorNamespace(t *testing.T) {
	t.Skipf("Not enabled by default") // Uncomment to execute the test

	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	ref, err := url.Parse(testConfig.Provider.Url)
	if err != nil {
		t.Fatalf("failed parsing '%s' host: %s", testConfig.Provider.Url, err)
	}
	var params = StringMap{
		"Testname":           t.Name(),
		"SupervisorName":     testConfig.Tm.VcenterSupervisor,
		"SupervisorZoneName": testConfig.Tm.VcenterSupervisorZone,

		"StorageClass": testConfig.Tm.StorageClass,
		"ProjectName":  "tf-project",

		"OrgName":     "tf-org",
		"OrgUser":     "tflocal",
		"OrgPassword": "long-change-ME1",

		"VcenterUsername": testConfig.Tm.VcenterUsername,
		"VcenterPassword": testConfig.Tm.VcenterPassword,
		"VcenterUrl":      testConfig.Tm.VcenterUrl,
		"NsxUsername":     testConfig.Tm.NsxManagerUsername,
		"NsxPassword":     testConfig.Tm.NsxManagerPassword,
		"NsxUrl":          testConfig.Tm.NsxManagerUrl,

		"Supervisor":     testConfig.Tm.VcenterSupervisor,
		"RegionName":     testConfig.Tm.Region,
		"VpcName":        testConfig.Tm.Region + "-Default-VPC",
		"Tier0Gateway":   testConfig.Tm.NsxTier0Gateway,
		"NsxEdgeCluster": testConfig.Tm.NsxEdgeCluster,
		"RegionVmClass":  "best-effort-2xlarge",

		"Tags": "tm org regionQuota",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaSupervisorNamespaceStep1, params)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcfaSupervisorNamespaceStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcfaSupervisorNamespaceStep3DS, params)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcfaSupervisorNamespaceStep4, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)

	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// This test uses also a provider config block logged in as a Tenant user, so we can not only test that administrators
	// can create tenant library items, but also tenant users can. This is a function and not a map to be lazy evaluated, as
	// the given user is created after some testing steps.
	multipleFactories := func() map[string]func() (*schema.Provider, error) {
		return map[string]func() (*schema.Provider, error){
			"vcfa": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
			"vcfatenant": func() (*schema.Provider, error) {
				return testOrgProvider(params["OrgName"].(string), params["OrgUser"].(string), params["OrgPassword"].(string)), nil
			},
		}
	}

	// Before this test ends we need to clean up the clients cache, because we create an Org user
	// and use it to login with the provider. Using same credentials and org name could lead to errors if this user
	// remains cached.
	defer cachedVCDClients.reset()

	cachedNamespaceName := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ProviderFactories: testAccProviders,
				Config:            configText1,
				Check:             resource.ComposeTestCheckFunc(),
			},
			{
				ProviderFactories: multipleFactories(),
				PreConfig:         func() { createProject(t, params) }, //Setup project
				Config:            configText2,
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
				ProviderFactories: multipleFactories(),
				Config:            configText3,
				Check: resource.ComposeTestCheckFunc(
					// Data source does not have 'name_prefix' therefore field count (%) differs
					resourceFieldsEqual("data.vcfa_supervisor_namespace.test", "vcfa_supervisor_namespace.test", []string{"%"}),
				),
			},
			{
				ProviderFactories:       multipleFactories(),
				ResourceName:            "vcfa_supervisor_namespace.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return params["ProjectName"].(string) + ImportSeparator + cachedNamespaceName.fieldValue, nil
				},
			},
			{
				ProviderFactories: multipleFactories(),
				Config:            configText4,
				Check: resource.ComposeTestCheckFunc(
					cachedNamespaceName.testCheckCachedResourceFieldValuePattern("data.vcfa_kubeconfig.test-namespace", "id", fmt.Sprintf("%s:%%s:%s", params["OrgName"].(string), params["ProjectName"].(string))),
					cachedNamespaceName.testCheckCachedResourceFieldValuePattern("data.vcfa_kubeconfig.test-namespace", "context_name", fmt.Sprintf("%s:%%s:%s", params["OrgName"].(string), params["ProjectName"].(string))),
					resource.TestCheckResourceAttr("data.vcfa_kubeconfig.test-namespace", "insecure_skip_tls_verify", fmt.Sprintf("%t", testConfig.Provider.AllowInsecure)),
					resource.TestCheckResourceAttr("data.vcfa_kubeconfig.test-namespace", "user", fmt.Sprintf("%s:%s@%s", params["OrgName"].(string), params["OrgUser"].(string), ref.Host)),
					resource.TestCheckResourceAttrSet("data.vcfa_kubeconfig.test-namespace", "token"),
					resource.TestCheckResourceAttrSet("data.vcfa_kubeconfig.test-namespace", "kube_config_raw"),
				),
			},
			{
				// Applying step1 config that will remove namespace
				ProviderFactories: multipleFactories(),
				Config:            configText1,
				Check:             resource.ComposeTestCheckFunc(),
			},
			{
				// Namespace already removed, removing project using SDK and leaveing for Terarform to teardwon
				PreConfig:         func() { removeProject(t, params) },
				ProviderFactories: multipleFactories(),
				Config:            configText1,
				Check:             resource.ComposeTestCheckFunc(),
			},
		},
	})
}

const testAccVcfaSupervisorNamespaceStep1 = `
resource "vcfa_vcenter" "vc" {
  provider = vcfa

  name                       = "{{.Testname}}"
  url                        = "{{.VcenterUrl}}"
  auto_trust_certificate     = true
  refresh_vcenter_on_create  = true
  refresh_policies_on_create = true
  username                   = "{{.VcenterUsername}}"
  password                   = "{{.VcenterPassword}}"
  is_enabled                 = true
  nsx_manager_id             = vcfa_nsx_manager.nsx_manager.id
}

resource "vcfa_nsx_manager" "nsx_manager" {
  provider = vcfa

  name                   = "{{.Testname}}"
  description            = "description"
  username               = "{{.NsxUsername}}"
  password               = "{{.NsxPassword}}"
  url                    = "{{.NsxUrl}}"
  auto_trust_certificate = true
}



data "vcfa_supervisor" "test" {
  provider = vcfa

  name       = "{{.Supervisor}}"
  vcenter_id = vcfa_vcenter.vc.id
  depends_on = [vcfa_vcenter.vc]
}

resource "vcfa_region" "region" {
  provider = vcfa

  name                 = "{{.RegionName}}"
  description          = "test-region"
  nsx_manager_id       = vcfa_nsx_manager.nsx_manager.id
  supervisor_ids       = [data.vcfa_supervisor.test.id]
  storage_policy_names = ["{{.StorageClass}}"]
}

data "vcfa_region_vm_class" "region_vm_class0" {
  provider = vcfa

  region_id = vcfa_region.region.id
  name      = "{{.RegionVmClass}}"
}


resource "vcfa_ip_space" "test-1" {
  provider = vcfa

  name                          = "{{.Testname}}"
  description                   = "Made using Terraform"
  region_id                     = vcfa_region.region.id
  external_scope                = "43.12.1.0/30"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1
  internal_scope {
    cidr = "32.0.1.0/24"
  }
}
	
data "vcfa_tier0_gateway" "test" {
  provider = vcfa

  region_id = vcfa_region.region.id 
  name      = "{{.Tier0Gateway}}"
}

resource "vcfa_provider_gateway" "test" {
  provider = vcfa

  name                  = "{{.Testname}}"
  region_id             = vcfa_region.region.id
  nsxt_tier0_gateway_id = data.vcfa_tier0_gateway.test.id
  ip_space_ids          = [vcfa_ip_space.test-1.id]
}


data "vcfa_region_zone" "test" {
  provider = vcfa

  region_id = vcfa_region.region.id
  name      = "{{.SupervisorZoneName}}"
}

data "vcfa_region_storage_policy" "sp" {
  provider = vcfa

  name      = "{{.StorageClass}}"
  region_id = vcfa_region.region.id
}

resource "vcfa_org_region_quota" "test" {
  provider = vcfa

  org_id         = vcfa_org.test.id
  region_id      = vcfa_region.region.id
  supervisor_ids = [data.vcfa_supervisor.test.id]
  zone_resource_allocations {
    region_zone_id         = data.vcfa_region_zone.test.id
    cpu_limit_mhz          = 2000
    cpu_reservation_mhz    = 1000
    memory_limit_mib       = 1024
    memory_reservation_mib = 512
  }
  region_vm_class_ids = [
    data.vcfa_region_vm_class.region_vm_class0.id
  ]
  region_storage_policy {
    region_storage_policy_id = data.vcfa_region_storage_policy.sp.id
    storage_limit_mib        = 100
  }
}

resource "vcfa_org" "test" {
  provider = vcfa

  name         = "{{.OrgName}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}

data "vcfa_role" "org-admin" {
  provider = vcfa

  org_id   = vcfa_org.test.id
  name     = "Organization Administrator"
}

resource "vcfa_org_local_user" "user" {
  provider = vcfa
  org_id   = vcfa_org.test.id
  role_ids = [data.vcfa_role.org-admin.id]
  username = "{{.OrgUser}}"
  password = "{{.OrgPassword}}"
}

resource "vcfa_org_networking" "test" {
  provider = vcfa

  org_id   = vcfa_org.test.id
  log_name = "tftestsn"
}

data "vcfa_edge_cluster" "test" {
  provider = vcfa

  name      = "{{.NsxEdgeCluster}}"
  region_id = vcfa_region.region.id
}

resource "vcfa_org_regional_networking" "test" {
  provider = vcfa

  name                = "{{.Testname}}"
  org_id              = vcfa_org.test.id
  provider_gateway_id = vcfa_provider_gateway.test.id
  region_id           = vcfa_region.region.id
  edge_cluster_id     = data.vcfa_edge_cluster.test.id

  depends_on = [vcfa_org_networking.test]
}
`

// Project must be precreated before
const testAccVcfaSupervisorNamespaceStep2 = testAccVcfaSupervisorNamespaceStep1 + `
resource "vcfa_supervisor_namespace" "test" {
  provider = vcfatenant

  name_prefix  = "terraform-test"
  project_name = "{{.ProjectName}}"
  class_name   = "small"
  description  = "Supervisor Namespace created by Terraform"
  region_name  = "{{.RegionName}}"
  vpc_name     = "{{.VpcName}}"

  storage_classes_initial_class_config_overrides {
    limit     = "90Mi"
    name      = "{{.StorageClass}}"
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

const testAccVcfaSupervisorNamespaceStep3DS = testAccVcfaSupervisorNamespaceStep2 + `
data "vcfa_supervisor_namespace" "test" {
  provider = vcfatenant

  name         = vcfa_supervisor_namespace.test.name
  project_name = vcfa_supervisor_namespace.test.project_name
}
`

const testAccVcfaSupervisorNamespaceStep4 = testAccVcfaSupervisorNamespaceStep2 + `
data "vcfa_kubeconfig" "test-namespace" {
  provider = vcfatenant

  project_name              = vcfa_supervisor_namespace.test.project_name
  supervisor_namespace_name = vcfa_supervisor_namespace.test.name
}
`

func createProject(t *testing.T, params StringMap) {
	tmClient := createTemporaryOrgConnection(params["OrgName"].(string), params["OrgUser"].(string), params["OrgPassword"].(string))
	projectCfg := &ccitypes.Project{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.ProjectKind,
			APIVersion: ccitypes.ProjectAPI + "/" + ccitypes.ProjectVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: params["ProjectName"].(string),
		},
		Spec: ccitypes.ProjectSpec{
			Description: fmt.Sprintf("Terraform test project [%s]", params["ProjectName"].(string)),
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
}

func removeProject(t *testing.T, params StringMap) {
	tmClient := createTemporaryOrgConnection(params["OrgName"].(string), params["OrgUser"].(string), params["OrgPassword"].(string))

	projectAddr, err := tmClient.Client.GetEntityUrl(ccitypes.ProjectsURL, "/", params["ProjectName"].(string))
	if err != nil {
		t.Fatalf("error getting Project url: %s", err)
	}
	err = tmClient.Client.DeleteEntity(projectAddr, nil, nil)
	if err != nil {
		t.Fatalf("failed removing Project: %s", err)
	}
}
