//go:build tm || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package project_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/vmware/terraform-provider-vcfa/internal/testutils"
	"github.com/vmware/terraform-provider-vcfa/internal/testutils/providertest"
)

// TestAccVcfaProject exercises the full lifecycle of the vcfa_project resource
// against a live environment. It creates an Org with an LDAP configuration and a local user, and
// manages a Project in that Org, as the local user, with up to three kinds of members: the local user,
// and a user and, optionally, a group of the LDAP server. The steps create the Project, update the description
// and the roles, read it back with the data source, import it, and finally unassign every member.
func TestAccVcfaProject(t *testing.T) {
	testutils.SkipIfShort(t)

	cfg := testutils.GetTestConfig(t)
	ldap := cfg.Ldap
	if ldap.Host == "" || ldap.Username == "" || ldap.Password == "" || ldap.Type == "" || ldap.Port == 0 ||
		ldap.BaseDistinguishedName == "" || ldap.ImportUser == "" {
		t.Skip("LDAP testing configuration, including an import user, is required")
	}

	// A system administrator gets a new Org with LDAP and a local user, and manages the Project with
	// a provider logged in to that Org. An Org user manages it in its own Org, which must already
	// have the LDAP configuration, using the user as the local member.
	sysAdmin := testutils.UsingSysAdmin(cfg)

	// Project names must be lowercase DNS labels.
	projectName := strings.ReplaceAll(strings.ToLower(t.Name()), "_", "-")

	// The LDAP group is optional
	expectedMembers := "2"
	groupMember := func(role string) string { return "" }
	groupResource := ""
	groupDependency := ""
	if ldap.ImportGroup != "" {
		groupResource = testAccVcfaProjectLdapGroup
		groupDependency = ", vcfa_org_external_group.ldap"
		expectedMembers = "3"
		groupMember = func(role string) string {
			return `{ kind = "Group", name = "` + ldap.ImportGroup + `", role = "` + role + `" },`
		}
	}

	params := testutils.StringMap{
		"ProjectName": projectName,
		"Description": "Project created by " + t.Name(),
		"LocalRole":   "admin",
		"LdapRole":    "view",

		"Org":           projectName,
		"OrgUser":       "tflocal",
		"OrgPassword":   "long-change-ME1",
		"Provider":      "provider = vcfatenant",
		"LocalUser":     "vcfa_org_local_user.test.username",
		"OrgRef":        "vcfa_org.test",
		"DependsOn":     "vcfa_org_ldap.test, vcfa_org_external_user.ldap" + groupDependency,
		"LdapDependsOn": "vcfa_org_ldap.test",
		"VcfaUrl":       cfg.Provider.Url,
		"Insecure":      cfg.Provider.AllowInsecure,

		"LdapServer":                ldap.Host,
		"LdapPort":                  ldap.Port,
		"LdapIsSsl":                 ldap.IsSsl,
		"LdapUsername":              ldap.Username,
		"LdapPassword":              ldap.Password,
		"LdapType":                  ldap.Type,
		"LdapBaseDistinguishedName": ldap.BaseDistinguishedName,
		"LdapUser":                  ldap.ImportUser,
		"LdapGroupMember":           groupMember("view"),
		"LdapGroup":                 ldap.ImportGroup,

		"Tags": "cci ldap",
	}

	prerequisites := testAccVcfaProjectPrerequisites + testAccVcfaProjectLdapUser + groupResource
	tenantProvider := testAccVcfaProjectTenantProvider
	localUser := "tflocal"
	if !sysAdmin {
		// The group of an Org user is read from the Org, which must already have it imported.
		if ldap.ImportGroup != "" {
			groupResource = testAccVcfaProjectLdapGroupData
			groupDependency = ", data.vcfa_org_external_group.ldap"
		}
		prerequisites, tenantProvider = testAccVcfaProjectOrgData+testAccVcfaProjectLdapUser+groupResource, ""
		localUser = cfg.Provider.User
		params["Provider"] = ""
		params["LocalUser"] = `"` + localUser + `"`
		params["OrgRef"] = "data.vcfa_org.test"
		params["Org"] = cfg.Org.Name
		params["DependsOn"] = "vcfa_org_external_user.ldap" + groupDependency
		params["LdapDependsOn"] = "data.vcfa_org.test"
	}

	// The tenant provider logs in when Terraform plans, so it can only be used once the Org and its
	// local user exist: the first step creates just them.
	configText0 := testutils.TemplateFill(t, prerequisites, params)

	params["FuncName"] = t.Name() + "-create"
	configText1 := testutils.TemplateFill(t, prerequisites+tenantProvider+testAccVcfaProjectMembersConfig, params)

	params["FuncName"] = t.Name() + "-update"
	params["Description"] = "Project updated by " + t.Name()
	params["LocalRole"] = "view"
	params["LdapRole"] = "edit"
	params["LdapGroupMember"] = groupMember("edit")
	configText2 := testutils.TemplateFill(t, prerequisites+tenantProvider+testAccVcfaProjectMembersConfig, params)

	params["FuncName"] = t.Name() + "-ds"
	configText3 := testutils.TemplateFill(t, prerequisites+tenantProvider+testAccVcfaProjectMembersConfig+testAccVcfaProjectDataSourceConfig, params)

	params["FuncName"] = t.Name() + "-no-members"
	configText4 := testutils.TemplateFill(t, prerequisites+tenantProvider+testAccVcfaProjectNoMembersConfig, params)

	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	testutils.DebugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)

	members := func(resourceName, localRole, ldapRole string) resource.TestCheckFunc {
		check := resource.ComposeTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "members.#", expectedMembers),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "members.*", map[string]string{
				"kind": "User", "name": localUser, "role": localRole,
			}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "members.*", map[string]string{
				"kind": "User", "name": ldap.ImportUser, "role": ldapRole,
			}),
		)
		if ldap.ImportGroup != "" {
			check = resource.ComposeTestCheckFunc(check, resource.TestCheckTypeSetElemNestedAttrs(resourceName, "members.*", map[string]string{
				"kind": "Group", "name": ldap.ImportGroup, "role": ldapRole,
			}))
		}
		return check
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providertest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 0: create the Org, its LDAP configuration and the local user.
			{
				SkipFunc: func() (bool, error) { return !sysAdmin, nil },
				Config:   configText0,
				Check:    resource.TestCheckResourceAttrSet("vcfa_org_local_user.test", "id"),
			},
			// Step 1: create the Project with a local user, an LDAP user and an LDAP group.
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_project.test", "id", projectName),
					resource.TestCheckResourceAttr("vcfa_project.test", "name", projectName),
					resource.TestCheckResourceAttr("vcfa_project.test", "description", "Project created by "+t.Name()),
					resource.TestCheckResourceAttrSet("vcfa_project.test", "uid"),
					members("vcfa_project.test", "admin", "view"),
				),
			},
			// Step 2: update the description and the roles.
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_project.test", "description", "Project updated by "+t.Name()),
					members("vcfa_project.test", "view", "edit"),
				),
			},
			// Step 3: verify the data source reads back what the resource manages.
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.vcfa_project.test", "id", "vcfa_project.test", "id"),
					resource.TestCheckResourceAttrPair("data.vcfa_project.test", "uid", "vcfa_project.test", "uid"),
					resource.TestCheckResourceAttrPair("data.vcfa_project.test", "description", "vcfa_project.test", "description"),
					members("data.vcfa_project.test", "view", "edit"),
				),
			},
			// Step 4: import and verify the state, members included, round-trips cleanly.
			{
				ResourceName:      "vcfa_project.test",
				ImportState:       true,
				ImportStateId:     projectName,
				ImportStateVerify: true,
			},
			// Step 5: an empty member list unassigns every user and group.
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcfa_project.test", "members.#", "0"),
				),
			},
		},
	})
}

// Org with LDAP and a local Org Administrator.
const testAccVcfaProjectPrerequisites = `
resource "vcfa_org" "test" {
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "terraform test"
  is_enabled   = true
}

resource "vcfa_org_ldap" "test" {
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

data "vcfa_role" "org-admin" {
  org_id = vcfa_org.test.id
  name   = "Organization Administrator"
}

resource "vcfa_org_local_user" "test" {
  org_id   = vcfa_org.test.id
  role_ids = [data.vcfa_role.org-admin.id]
  username = "{{.OrgUser}}"
  password = "{{.OrgPassword}}"
}
`

// The Org of an Org user.
const testAccVcfaProjectOrgData = `
data "vcfa_org" "test" {
  name = "{{.Org}}"
}
`

// The user of the LDAP server must be imported into the Org before it can be a Project member.
const testAccVcfaProjectLdapUser = `
data "vcfa_role" "org-user" {
  org_id = {{.OrgRef}}.id
  name   = "Organization User"
}

resource "vcfa_org_external_user" "ldap" {
  org_id        = {{.OrgRef}}.id
  provider_type = "LDAP"
  role_ids      = [data.vcfa_role.org-user.id]
  username      = "{{.LdapUser}}"

  depends_on = [{{.LdapDependsOn}}]
}
`

// The group of the LDAP server, imported into the Org like the user.
const testAccVcfaProjectLdapGroup = `
resource "vcfa_org_external_group" "ldap" {
  org_id        = {{.OrgRef}}.id
  provider_type = "LDAP"
  role_ids      = [data.vcfa_role.org-user.id]
  name          = "{{.LdapGroup}}"

  depends_on = [{{.LdapDependsOn}}]
}
`

// The group of the LDAP server, already imported into the Org of an Org user.
const testAccVcfaProjectLdapGroupData = `
data "vcfa_org_external_group" "ldap" {
  org_id = {{.OrgRef}}.id
  name   = "{{.LdapGroup}}"
}
`

// Provider logged in as the local user of the Org, as Projects are tenant objects.
const testAccVcfaProjectTenantProvider = `
provider "vcfatenant" {
  user                 = "{{.OrgUser}}"
  password             = "{{.OrgPassword}}"
  auth_type            = "integrated"
  url                  = "{{.VcfaUrl}}"
  org                  = "{{.Org}}"
  allow_unverified_ssl = {{.Insecure}}
}
`

const testAccVcfaProjectMembersConfig = `
resource "vcfa_project" "test" {
  {{.Provider}}
  name        = "{{.ProjectName}}"
  description = "{{.Description}}"

  members = [
    { kind = "User", name = {{.LocalUser}}, role = "{{.LocalRole}}" },
    { kind = "User", name = "{{.LdapUser}}", role = "{{.LdapRole}}" },
    {{.LdapGroupMember}}
  ]

  depends_on = [{{.DependsOn}}]
}
`

const testAccVcfaProjectDataSourceConfig = `
data "vcfa_project" "test" {
  {{.Provider}}
  name     = vcfa_project.test.name

  depends_on = [vcfa_project.test]
}
`

const testAccVcfaProjectNoMembersConfig = `
resource "vcfa_project" "test" {
  {{.Provider}}
  name        = "{{.ProjectName}}"
  description = "{{.Description}}"
  members     = []

  depends_on = [{{.DependsOn}}]
}
`
