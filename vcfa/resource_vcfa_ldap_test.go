//go:build ldap || ALL || functional

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccVcfaSystemLdap tests Provider (System) LDAP configuration against an LDAP server with the given configuration
func TestAccVcfaSystemLdap(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if testConfig.Ldap.Host == "" || testConfig.Ldap.Username == "" || testConfig.Ldap.Password == "" || testConfig.Ldap.Type == "" ||
		testConfig.Ldap.Port == 0 || testConfig.Ldap.BaseDistinguishedName == "" {
		t.Skip("LDAP testing configuration is required")
	}

	var params = StringMap{
		"LdapServer":                testConfig.Ldap.Host,
		"LdapPort":                  testConfig.Ldap.Port,
		"LdapIsSsl":                 testConfig.Ldap.IsSsl,
		"LdapUser":                  testConfig.Ldap.Username,
		"LdapPassword":              testConfig.Ldap.Password,
		"LdapType":                  testConfig.Ldap.Type,
		"LdapBaseDistinguishedName": testConfig.Ldap.BaseDistinguishedName,
		"Password":                  testConfig.Tm.VcenterPassword,
		"Tags":                      "ldap",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name()
	configText := templateFill(testAccVcfaLdap, params)

	params["FuncName"] = t.Name() + "-DS"
	configTextDS := templateFill(testAccVcfaLdapDS, params)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION Resource for LDAP: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION Data source: %s\n", configTextDS)

	ldapResourceDef := "vcfa_ldap.ldap"
	ldapDatasourceDef := "data.vcfa_ldap.ldap-ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// TODO: TM: Check LDAP is destroyed before Organization is
		// CheckDestroy:      testAccCheckOrgLdapDestroy(ldapResourceDef),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ldapResourceDef, "server", testConfig.Ldap.Host),
					resource.TestCheckResourceAttr(ldapResourceDef, "port", fmt.Sprintf("%d", testConfig.Ldap.Port)),
					resource.TestCheckResourceAttr(ldapResourceDef, "is_ssl", fmt.Sprintf("%t", testConfig.Ldap.IsSsl)),
					resource.TestCheckResourceAttr(ldapResourceDef, "base_distinguished_name", testConfig.Ldap.BaseDistinguishedName),
					resource.TestCheckResourceAttr(ldapResourceDef, "connector_type", testConfig.Ldap.Type),
					resource.TestCheckResourceAttr(ldapResourceDef, "password", ""), // Password is not returned
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.object_class", "user"),
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.unique_identifier", "objectGuid"),
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.display_name", "displayName"),
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.username", "sAMAccountName"),
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.given_name", "givenName"),
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.surname", "sn"),
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.telephone", "telephoneNumber"),
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.group_membership_identifier", "dn"),
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.email", "mail"),
					resource.TestCheckResourceAttr(ldapResourceDef, "user_attributes.0.group_back_link_identifier", "tokenGroups"),
					resource.TestCheckResourceAttr(ldapResourceDef, "group_attributes.0.name", "cn"),
					resource.TestCheckResourceAttr(ldapResourceDef, "group_attributes.0.object_class", "group"),
					resource.TestCheckResourceAttr(ldapResourceDef, "group_attributes.0.membership", "member"),
					resource.TestCheckResourceAttr(ldapResourceDef, "group_attributes.0.unique_identifier", "objectGuid"),
					resource.TestCheckResourceAttr(ldapResourceDef, "group_attributes.0.group_membership_identifier", "dn"),
					resource.TestCheckResourceAttr(ldapResourceDef, "group_attributes.0.group_back_link_identifier", "objectSid"),
				),
			},
			{
				Config: configTextDS,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrgLdapExists(ldapResourceDef),
					resourceFieldsEqual(ldapResourceDef, ldapDatasourceDef, []string{}),
				),
			},
			{
				ResourceName:      ldapResourceDef,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) { return testConfig.Tm.Org, nil },
			},
		},
	})
	postTestChecks(t)
}

const testAccVcfaLdap = `
resource "vcfa_ldap" "ldap" {
  auto_trust_certificate  = true
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
  
  lifecycle {
    # password value does not get returned by GET
    ignore_changes = [custom_settings[0].password]
  }
}
`

const testAccVcfaLdapDS = testAccVcfaOrgLdap + `
data "vcfa_ldap" "ldap-ds" {
  depends_on = [vcfa_ldap.ldap]
}
`
