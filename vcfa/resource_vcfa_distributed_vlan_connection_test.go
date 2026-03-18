//go:build tm || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVcfaDistributedVlanConnection(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	vCenterHcl, vCenterHclRef := getVCenterHcl(t, nsxManagerHclRef)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	ipSpaceHcl, ipSpaceHclRef := getIpSpaceHcl(t, regionHclRef, "1", "1")

	k8sCompliantName := strings.ReplaceAll(strings.ToLower(t.Name()), "_", "-")

	var params = StringMap{
		"Testname":   k8sCompliantName,
		"VcenterRef": vCenterHclRef,
		"RegionId":   fmt.Sprintf("%s.id", regionHclRef),
		"RegionName": t.Name(),
		"IpSpaceId":  fmt.Sprintf("%s.id", ipSpaceHclRef),

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + ipSpaceHcl
	configText1 := templateFill(preRequisites+testAccVcfaDistributedVlanConnectionStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcfaDistributedVlanConnectionStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaDistributedVlanConnectionStep3DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedDistributedVlanConnectionId := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					cachedDistributedVlanConnectionId.cacheTestResourceFieldValue("vcfa_distributed_vlan_connection.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_distributed_vlan_connection.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_distributed_vlan_connection.test", "backing_id"),
					resource.TestCheckResourceAttrSet("vcfa_distributed_vlan_connection.test", "status"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "name", k8sCompliantName),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "description", "description test"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "gateway_cidr", "32.0.1.1/24"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "subnet_exclusive", "false"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "vlan_id", "100"),
					resource.TestCheckResourceAttrPair("vcfa_distributed_vlan_connection.test", "ip_space_id", ipSpaceHclRef, "id"),

					resource.TestCheckResourceAttrSet("vcfa_distributed_vlan_connection.test2", "id"),
					resource.TestCheckResourceAttrSet("vcfa_distributed_vlan_connection.test2", "backing_id"),
					resource.TestCheckResourceAttrSet("vcfa_distributed_vlan_connection.test2", "status"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test2", "name", k8sCompliantName+"-2"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test2", "description", "description test - 2"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test2", "gateway_cidr", "10.0.0.1/24"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test2", "subnet_exclusive", "true"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test2", "vlan_id", "110"),
					resource.TestCheckResourceAttrSet("vcfa_distributed_vlan_connection.test2", "ip_space_id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					cachedDistributedVlanConnectionId.testCheckCachedResourceFieldValue("vcfa_distributed_vlan_connection.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_distributed_vlan_connection.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_distributed_vlan_connection.test", "backing_id"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "status", "REALIZED"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "name", k8sCompliantName+"-updated"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "description", "description test - update"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "gateway_cidr", "32.0.1.1/24"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "subnet_exclusive", "false"),
					resource.TestCheckResourceAttr("vcfa_distributed_vlan_connection.test", "vlan_id", "100"),
					resource.TestCheckResourceAttrPair("vcfa_distributed_vlan_connection.test", "ip_space_id", ipSpaceHclRef, "id"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_distributed_vlan_connection.test", "data.vcfa_distributed_vlan_connection.test", nil),
				),
			},
			{
				ResourceName:      "vcfa_distributed_vlan_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.Tm.Region + ImportSeparator + params["Testname"].(string) + "-updated",
			},
		},
	})
}

const testAccVcfaDistributedVlanConnectionStep1 = `
resource "vcfa_distributed_vlan_connection" "test" {
  name             = "{{.Testname}}"
  description      = "description test"
  region_id        = {{.RegionId}}
  gateway_cidr     = "32.0.1.1/24"
  ip_space_id      = {{.IpSpaceId}}
  subnet_exclusive = false
  vlan_id          = 100
}

resource "vcfa_distributed_vlan_connection" "test2" {
  name             = "{{.Testname}}-2"
  description      = "description test - 2"
  region_id        = {{.RegionId}}
  gateway_cidr     = "10.0.0.1/24"
  subnet_exclusive = true
  vlan_id          = 110
}
`

const testAccVcfaDistributedVlanConnectionStep2 = `
resource "vcfa_distributed_vlan_connection" "test" {
  name             = "{{.Testname}}-updated"
  description      = "description test - update"
  region_id        = {{.RegionId}}
  gateway_cidr     = "32.0.1.1/24"
  ip_space_id      = {{.IpSpaceId}}
  subnet_exclusive = false
  vlan_id          = 100
}
`

const testAccVcfaDistributedVlanConnectionStep3DS = testAccVcfaDistributedVlanConnectionStep2 + `
data "vcfa_distributed_vlan_connection" "test" {
  name       = vcfa_distributed_vlan_connection.test.name
  region_id  = {{.RegionId}}

  depends_on = [ vcfa_distributed_vlan_connection.test ]
}
`
