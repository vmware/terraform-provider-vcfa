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

func TestAccVcfaSharedSubnet(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	vCenterHcl, vCenterHclRef := getVCenterHcl(t, nsxManagerHclRef)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)

	k8sCompliantName := strings.ReplaceAll(strings.ToLower(t.Name()), "_", "-")

	var params = StringMap{
		"Testname":   k8sCompliantName,
		"VcenterRef": vCenterHclRef,
		"RegionId":   fmt.Sprintf("%s.id", regionHclRef),
		"RegionName": t.Name(),

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl
	configText1 := templateFill(preRequisites+testAccVcfaSharedSubnetStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcfaSharedSubnetStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcfaSharedSubnetStep3DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedSharedSubnetId := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					cachedSharedSubnetId.cacheTestResourceFieldValue("vcfa_shared_subnet.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_shared_subnet.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_shared_subnet.test", "backing_id"),
					resource.TestCheckResourceAttrSet("vcfa_shared_subnet.test", "ip_space_id"),
					resource.TestCheckResourceAttrSet("vcfa_shared_subnet.test", "status"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "name", k8sCompliantName),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "description", "description test"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "subnet_type", "VLAN"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "gateway_cidr", "10.0.0.1/24"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "vlan_id", "100"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					cachedSharedSubnetId.testCheckCachedResourceFieldValue("vcfa_shared_subnet.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_shared_subnet.test", "id"),
					resource.TestCheckResourceAttrSet("vcfa_shared_subnet.test", "backing_id"),
					resource.TestCheckResourceAttrSet("vcfa_shared_subnet.test", "ip_space_id"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "status", "REALIZED"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "name", k8sCompliantName+"-updated"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "description", "description test - update"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "subnet_type", "VLAN"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "gateway_cidr", "10.0.0.1/24"),
					resource.TestCheckResourceAttr("vcfa_shared_subnet.test", "vlan_id", "100"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcfa_shared_subnet.test", "data.vcfa_shared_subnet.test", nil),
				),
			},
			{
				ResourceName:      "vcfa_shared_subnet.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.Tm.Region + ImportSeparator + params["Testname"].(string) + "-updated",
			},
		},
	})
}

const testAccVcfaSharedSubnetStep1 = `
resource "vcfa_shared_subnet" "test" {
  name         = "{{.Testname}}"
  description  = "description test"
  region_id    = {{.RegionId}}
  subnet_type  = "VLAN"
  gateway_cidr = "10.0.0.1/24"
  vlan_id      = 100
}
`

const testAccVcfaSharedSubnetStep2 = `
resource "vcfa_shared_subnet" "test" {
  name         = "{{.Testname}}-updated"
  description  = "description test - update"
  region_id    = {{.RegionId}}
  subnet_type  = "VLAN"
  gateway_cidr = "10.0.0.1/24"
  vlan_id      = 100
}
`

const testAccVcfaSharedSubnetStep3DS = testAccVcfaSharedSubnetStep2 + `
data "vcfa_shared_subnet" "test" {
  name      = vcfa_shared_subnet.test.name
  region_id = {{.RegionId}}

  depends_on = [ vcfa_shared_subnet.test ]
}
`
