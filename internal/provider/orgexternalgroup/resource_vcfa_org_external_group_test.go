//go:build org || tm || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package orgexternalgroup_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/vmware/terraform-provider-vcfa/internal/testutils"
	"github.com/vmware/terraform-provider-vcfa/internal/testutils/providertest"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

// TestAccVcfaOrgExternalGroup creates an Org with LDAP, and imports an LDAP group into it with the
// resource, updates its role, reads it with the data source, and imports it.
func TestAccVcfaOrgExternalGroup(t *testing.T) {
	testutils.SkipIfNotSysAdmin(t)
	testutils.SkipIfShort(t)

	cfg := testutils.GetTestConfig(t)
	ldap := cfg.Ldap
	if ldap.Host == "" || ldap.Username == "" || ldap.Password == "" || ldap.Type == "" || ldap.Port == 0 ||
		ldap.BaseDistinguishedName == "" || ldap.ImportGroup == "" {
		t.Skip("LDAP testing configuration, including an import group, is required")
	}

	params := testutils.StringMap{
		"Testname":                  t.Name(),
		"LdapServer":                ldap.Host,
		"LdapPort":                  ldap.Port,
		"LdapIsSsl":                 ldap.IsSsl,
		"LdapUsername":              ldap.Username,
		"LdapPassword":              ldap.Password,
		"LdapBaseDistinguishedName": ldap.BaseDistinguishedName,
		"LdapType":                  ldap.Type,
		"GroupName":                 ldap.ImportGroup,

		"Tags": "cci ldap",
	}

	configText1 := testutils.TemplateFill(t, testAccVcfaOrgExternalGroupStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := testutils.TemplateFill(t, testAccVcfaOrgExternalGroupStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := testutils.TemplateFill(t, testAccVcfaOrgExternalGroupStep4DS, params)

	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providertest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org_external_group.test", "name", ldap.ImportGroup),
					resource.TestCheckResourceAttr("vcfa_org_external_group.test", "provider_type", "LDAP"),
					resource.TestCheckResourceAttr("vcfa_org_external_group.test", "role_ids.#", "1"),
					resource.TestCheckResourceAttrPair("vcfa_org_external_group.test", "org_id", "vcfa_org.test", "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_org_external_group.test", "name", ldap.ImportGroup),
					resource.TestCheckResourceAttr("vcfa_org_external_group.test", "provider_type", "LDAP"),
					resource.TestCheckResourceAttr("vcfa_org_external_group.test", "role_ids.#", "1"),
					resource.TestCheckResourceAttrPair("vcfa_org_external_group.test", "org_id", "vcfa_org.test", "id"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					testutils.ResourceFieldsEqual("vcfa_org_external_group.test", "data.vcfa_org_external_group.test", nil),
				),
			},
			{
				ResourceName:      "vcfa_org_external_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     t.Name() + vcfa.ImportSeparator + ldap.ImportGroup,
			},
		},
	})
}

const testAccVcfaOrgExternalGroupPrerequisites = `
resource "vcfa_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}

data "vcfa_role" "org-admin" {
  org_id = vcfa_org.test.id
  name   = "Organization Administrator"
}

data "vcfa_role" "org-user" {
  org_id = vcfa_org.test.id
  name   = "Organization User"
}

resource "vcfa_org_ldap" "ldap" {
  org_id                 = vcfa_org.test.id
  ldap_mode              = "CUSTOM"
  auto_trust_certificate = true

  custom_settings {
    server                  = "{{.LdapServer}}"
    port                    = {{.LdapPort}}
    is_ssl                  = {{.LdapIsSsl}}
    username                = "{{.LdapUsername}}"
    password                = "{{.LdapPassword}}"
    base_distinguished_name = "{{.LdapBaseDistinguishedName}}"
    connector_type          = "{{.LdapType}}"

    user_attributes {
      object_class                = "user"
      unique_identifier           = "objectGuid"
      display_name                = "displayName"
      username                    = "sAMAccountName"
      given_name                  = "givenName"
      surname                     = "sn"
      telephone                   = "telephoneNumber"
      group_membership_identifier = "dn"
      email                       = "mail"
      group_back_link_identifier  = "tokenGroups"
    }

    group_attributes {
      name                        = "cn"
      object_class                = "group"
      membership                  = "member"
      unique_identifier           = "objectGuid"
      group_membership_identifier = "dn"
      group_back_link_identifier  = "objectSid"
    }
  }
  lifecycle {
    # password value does not get returned by GET
    ignore_changes = [custom_settings[0].password]
  }
}
`

const testAccVcfaOrgExternalGroupStep1 = testAccVcfaOrgExternalGroupPrerequisites + `
resource "vcfa_org_external_group" "test" {
  org_id        = vcfa_org.test.id
  provider_type = "LDAP"
  role_ids      = [data.vcfa_role.org-admin.id]
  name          = "{{.GroupName}}"

  depends_on = [vcfa_org_ldap.ldap]
}
`

const testAccVcfaOrgExternalGroupStep2 = testAccVcfaOrgExternalGroupPrerequisites + `
resource "vcfa_org_external_group" "test" {
  org_id        = vcfa_org.test.id
  provider_type = "LDAP"
  role_ids      = [data.vcfa_role.org-user.id]
  name          = "{{.GroupName}}"

  depends_on = [vcfa_org_ldap.ldap]
}
`

const testAccVcfaOrgExternalGroupStep4DS = testAccVcfaOrgExternalGroupStep2 + `
data "vcfa_org_external_group" "test" {
  org_id = vcfa_org.test.id
  name   = vcfa_org_external_group.test.name
}
`
