//go:build cci || ALL || functional

// © Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcfaKubeConfig(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)

	ref, err := url.Parse(testConfig.Provider.Url)
	if err != nil {
		t.Fatalf("failed parsing '%s' host: %s", testConfig.Provider.Url, err)
	}
	var params = StringMap{
		"Testname": t.Name(),
		"Org":      testConfig.Org.Name,

		"Tags": "cci",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcfaKubeConfigStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)

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
					resource.TestCheckResourceAttr("data.vcfa_kubeconfig.test", "id", testConfig.Org.Name),
					resource.TestCheckResourceAttr("data.vcfa_kubeconfig.test", "context_name", testConfig.Org.Name),
					resource.TestCheckResourceAttr("data.vcfa_kubeconfig.test", "insecure_skip_tls_verify", fmt.Sprintf("%t", testConfig.Provider.AllowInsecure)),
					resource.TestCheckResourceAttr("data.vcfa_kubeconfig.test", "user", fmt.Sprintf("%s:%s@%s", testConfig.Org.Name, testConfig.Org.User, ref.Host)),
					resource.TestCheckResourceAttrSet("data.vcfa_kubeconfig.test", "token"),
					resource.TestCheckResourceAttrSet("data.vcfa_kubeconfig.test", "kube_config_raw"),
				),
			},
		},
	})
}

const testAccVcfaKubeConfigStep1 = `
data "vcfa_kubeconfig" "test" {}
`
