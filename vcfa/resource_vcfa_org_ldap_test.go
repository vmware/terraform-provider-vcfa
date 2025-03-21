//go:build ldap || org || ALL || functional

package vcfa

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccVcfaOrgLdap tests LDAP configuration against an LDAP server with the given configuration
func TestAccVcfaOrgLdap(t *testing.T) {
	preTestChecks(t)
	defer postTestChecks(t)
	skipIfNotSysAdmin(t)

	if testConfig.Ldap.Host == "" || testConfig.Ldap.Username == "" || testConfig.Ldap.Password == "" || testConfig.Ldap.Type == "" ||
		testConfig.Ldap.Port == 0 || testConfig.Ldap.BaseDistinguishedName == "" {
		t.Skip("LDAP testing configuration is required")
	}

	var params = StringMap{
		"Org":                       testConfig.Tm.Org,
		"LdapServer":                testConfig.Ldap.Host,
		"LdapPort":                  testConfig.Ldap.Port,
		"LdapIsSsl":                 testConfig.Ldap.IsSsl,
		"LdapUsername":              testConfig.Ldap.Username,
		"LdapPassword":              testConfig.Ldap.Password,
		"LdapType":                  testConfig.Ldap.Type,
		"LdapBaseDistinguishedName": testConfig.Ldap.BaseDistinguishedName,
		"CustomUiLabel":             "custom_ui_button_label  = \"Hello there\"",
		"Tags":                      "ldap org",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcfaOrgLdap, params)

	params["FuncName"] = t.Name() + "-step2"
	params["CustomUiLabel"] = " "
	configText2 := templateFill(testAccVcfaOrgLdap, params)

	params["FuncName"] = t.Name() + "-DS"
	configTextDS := templateFill(testAccVcfaOrgLdapDS, params)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION Resource for Organization LDAP Step 1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION Resource for Organization LDAP Step 2: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION Data source: %s\n", configTextDS)

	orgDef := "vcfa_org.org1"
	ldapResourceDef := "vcfa_org_ldap.ldap"
	ldapDatasourceDef := "data.vcfa_org_ldap.ldap-ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrgLdapExists(ldapResourceDef),
					resource.TestCheckResourceAttr(orgDef, "name", params["Org"].(string)),
					resource.TestCheckResourceAttr(ldapResourceDef, "ldap_mode", "CUSTOM"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.server", testConfig.Ldap.Host),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.port", fmt.Sprintf("%d", testConfig.Ldap.Port)),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.is_ssl", fmt.Sprintf("%t", testConfig.Ldap.IsSsl)),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.base_distinguished_name", testConfig.Ldap.BaseDistinguishedName),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.connector_type", testConfig.Ldap.Type),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.custom_ui_button_label", "Hello there"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.password", testConfig.Ldap.Password),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.object_class", "user"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.unique_identifier", "objectGuid"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.display_name", "displayName"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.username", "sAMAccountName"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.given_name", "givenName"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.surname", "sn"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.telephone", "telephoneNumber"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.group_membership_identifier", "dn"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.email", "mail"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.group_back_link_identifier", "tokenGroups"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.group_attributes.0.name", "cn"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.group_attributes.0.object_class", "group"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.group_attributes.0.membership", "member"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.group_attributes.0.unique_identifier", "objectGuid"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.group_attributes.0.group_membership_identifier", "dn"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.group_attributes.0.group_back_link_identifier", "objectSid"),
					resource.TestCheckResourceAttrPair(orgDef, "id", ldapResourceDef, "org_id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrgLdapExists(ldapResourceDef),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.custom_ui_button_label", ""),
					resource.TestCheckResourceAttrPair(orgDef, "id", ldapResourceDef, "org_id"),
				),
			},
			{
				Config: configTextDS,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrgLdapExists(ldapResourceDef),
					resourceFieldsEqual(ldapResourceDef, ldapDatasourceDef, []string{"%", "auto_trust_certificate", "custom_settings.0.%", "custom_settings.0.password"}),
				),
			},
			{
				ResourceName:            ldapResourceDef,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       func(state *terraform.State) (string, error) { return testConfig.Tm.Org, nil },
				ImportStateVerifyIgnore: []string{"auto_trust_certificate"},
			},
		},
	})
}

func testAccCheckOrgLdapExists(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaOrg)
		}

		conn := testAccProvider.Meta().(ClientContainer).tmClient

		tmOrg, err := conn.GetTmOrgById(rs.Primary.ID)
		if err != nil {
			return err
		}
		config, err := tmOrg.GetLdapConfiguration()
		if err != nil {
			return err
		}
		if config.OrgLdapMode == "NONE" {
			return fmt.Errorf("resource %s not configured", identifier)
		}
		return nil
	}
}

const testAccVcfaOrgLdap = `
resource "vcfa_org" "org1" {
  name         = "{{.Org}}"
  display_name = "{{.Org}}"
  description  = "{{.Org}}"
}

resource "vcfa_org_ldap" "ldap" {
  org_id                 = vcfa_org.org1.id
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
    {{.CustomUiLabel}}

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

const testAccVcfaOrgLdapDS = testAccVcfaOrgLdap + `
data "vcfa_org_ldap" "ldap-ds" {
  org_id = vcfa_org.org1.id
  depends_on = [vcfa_org_ldap.ldap]
}
`
