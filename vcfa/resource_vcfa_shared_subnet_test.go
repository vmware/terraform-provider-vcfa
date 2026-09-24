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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
			{
				// The backend intentionally leaves the Regional Network Settings it auto-created for
				// this Shared Subnet in place once the last Distributed VLAN Connection under it is
				// removed, as the setting may still host other shared network resources. Dropping
				// "vcfa_shared_subnet.test" from the config here destroys it, then the orphaned setting
				// is deleted directly - otherwise the Region above fails to be destroyed once this test
				// finishes.
				Config: preRequisites,
				Check: func(s *terraform.State) error {
					return deleteOrphanedRegionalNetworkingSetting(s, regionHclRef)
				},
			},
		},
	})
}

// deleteOrphanedRegionalNetworkingSetting deletes the Regional Network Settings that the backend
// auto-creates for a Region's default consumption Org the first time a Shared Subnet is added to it.
// Since 9.2, removing the last Distributed VLAN Connection under a Shared Subnet no longer deletes
// that setting, so it must be cleaned up explicitly here or the Region identified by
// regionResourceAddr cannot be destroyed once this test finishes.
func deleteOrphanedRegionalNetworkingSetting(s *terraform.State, regionResourceAddr string) error {
	rs, ok := s.RootModule().Resources[regionResourceAddr]
	if !ok {
		return fmt.Errorf("could not find %s in state", regionResourceAddr)
	}
	regionId := rs.Primary.ID

	tmClient, err := getTestVCFAFromJson(testConfig)
	if err != nil {
		return fmt.Errorf("error getting a client: %s", err)
	}
	err = ProviderAuthenticate(tmClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token,
		testConfig.Provider.SysOrg, testConfig.Provider.ApiToken, testConfig.Provider.ApiTokenFile, testConfig.Provider.ServiceAccountTokenFile)
	if err != nil {
		return fmt.Errorf("error authenticating: %s", err)
	}

	all, err := tmClient.GetAllTmRegionalNetworkingSettings(nil)
	if err != nil {
		return fmt.Errorf("error retrieving %s: %s", labelVcfaRegionalNetworkingSetting, err)
	}
	for _, one := range all {
		if one.TmRegionalNetworkingSetting.RegionRef.ID != regionId {
			continue
		}
		if err := one.Delete(); err != nil {
			return fmt.Errorf("error deleting orphaned %s '%s': %s", labelVcfaRegionalNetworkingSetting, one.TmRegionalNetworkingSetting.Name, err)
		}
	}
	return nil
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
