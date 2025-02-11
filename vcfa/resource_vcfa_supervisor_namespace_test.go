//go:build cci || ALL || functional

package vcfa

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaTpSupervisorNamespace(t *testing.T) {
	preTestChecks(t)
	// skipIfNotSysAdmin(t)

	// vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	// nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	// regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	// ipSpaceHcl, ipSpaceHclRef := getIpSpaceHcl(t, regionHclRef, "1", "1")
	// providerGatewayHcl, providerGatewayHclRef := getProviderGatewayHcl(t, regionHclRef, ipSpaceHclRef)

	// missing prerequisite -  Setup Region Quota and Regional Network

	var params = StringMap{
		"Testname":    t.Name(),
		"ProjectName": "tf-project",
		// "RegionId":          fmt.Sprintf("%s.id", regionHclRef),
		// "ProviderGatewayId": fmt.Sprintf("%s.id", providerGatewayHclRef),
		// "EdgeClusterName":   testConfig.Tm.NsxEdgeCluster,
		"Tags": "tm cci",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	// skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	// configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	// params["FuncName"] = t.Name() + "-step0"

	// preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + ipSpaceHcl + providerGatewayHcl
	configText1 := templateFill(testAccVcfaTpSupervisorNamespaceStep1, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	// debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	// debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	// debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)
	// debugPrintf("#[DEBUG] CONFIGURATION step5: %s\n", configText5)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ExternalProviders: map[string]resource.ExternalProvider{
			"kubernetes": {
				VersionConstraint: "2.35.1",
				Source:            "hashicorp/kubernetes",
			},
		},
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{

				Config: configText1,
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcfaTpSupervisorNamespaceStep1 = `
data "vcfa_kube_config" "demo" {}

provider "kubernetes" {
  host     = data.vcfa_kube_config.demo.host
  insecure = data.vcfa_kube_config.demo.insecure_skip_tls_verify
  token    = data.vcfa_kube_config.demo.token
}

resource "kubernetes_manifest" "project" {
  #   id = "fake-id"
  manifest = {
    "apiVersion" = "project.cci.vmware.com/v1alpha1"
    "kind"       = "Project"
    "metadata" = {
      "name" = "{{.ProjectName}}"
    }
    "spec" = {
      "description" = "Project [{{.ProjectName}}] created by Terraform acceptance testing"
    }
  }
}
`
