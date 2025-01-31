//go:build ldap || org || ALL || functional

package vcfa

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcfaOrgLdap(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Org":        testConfig.Tm.Org,
		"LdapServer": regexp.MustCompile(`https?://`).ReplaceAllString(testConfig.Tm.VcenterUrl, ""),
		"Password":   testConfig.Tm.VcenterPassword,
		"Tags":       "ldap org",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name()
	configText := templateFill(testAccVcfaOrgLdap, params)

	// TODO: TM: Missing System test

	params["FuncName"] = t.Name() + "-DS"
	configTextDS := templateFill(testAccVcfaOrgLdapDS, params)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION Resource for Organization LDAP (Custom): %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION Data source: %s\n", configTextDS)

	orgDef := "vcfa_org.org1"
	ldapResourceDef := "vcfa_org_ldap.ldap"
	ldapDatasourceDef := "data.vcfa_org_ldap.ldap-ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// TODO: TM: Check LDAP is destroyed before Organization is
		// CheckDestroy:      testAccCheckOrgLdapDestroy(ldapResourceDef),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrgLdapExists(ldapResourceDef),
					resource.TestCheckResourceAttr(orgDef, "name", params["Org"].(string)),
					resource.TestCheckResourceAttr(ldapResourceDef, "ldap_mode", "CUSTOM"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.server", params["LdapServer"].(string)),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.authentication_method", "SIMPLE"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.connector_type", "OPEN_LDAP"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.user_attributes.0.object_class", "inetOrgPerson"),
					resource.TestCheckResourceAttr(ldapResourceDef, "custom_settings.0.group_attributes.0.object_class", "group"),
					resource.TestCheckResourceAttrPair(orgDef, "id", ldapResourceDef, "org_id"),
				),
			},
			{
				Config: configTextDS,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrgLdapExists(ldapResourceDef),
					resource.TestCheckResourceAttrPair(ldapResourceDef, "org_id", ldapDatasourceDef, "org_id"),
					resource.TestCheckResourceAttrPair(ldapResourceDef, "ldap_mode", ldapDatasourceDef, "ldap_mode"),
					resource.TestCheckResourceAttrPair(ldapResourceDef, "custom_settings.0.server", ldapDatasourceDef, "custom_settings.0.server"),
					resource.TestCheckResourceAttrPair(ldapResourceDef, "custom_settings.0.authentication_method", ldapDatasourceDef, "custom_settings.0.authentication_method"),
					resource.TestCheckResourceAttrPair(ldapResourceDef, "custom_settings.0.connector_type", ldapDatasourceDef, "custom_settings.0.connector_type"),
					resource.TestCheckResourceAttrPair(ldapResourceDef, "custom_settings.0.user_attributes.0.object_class", ldapDatasourceDef, "custom_settings.0.user_attributes.0.object_class"),
					resource.TestCheckResourceAttrPair(ldapResourceDef, "custom_settings.0.group_attributes.0.object_class", ldapDatasourceDef, "custom_settings.0.group_attributes.0.object_class"),
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

func testAccCheckOrgLdapExists(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no %s ID is set", labelVcfaOrg)
		}

		conn := testAccProvider.Meta().(*VCDClient)

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
  org_id    = vcfa_org.org1.id
  ldap_mode = "CUSTOM"
  custom_settings {
    server                  = "{{.LdapServer}}"
    port                    = 389
    is_ssl                  = false
    username                = "cn=Administrator,cn=Users,dc=vsphere,dc=local"
    password                = "{{.Password}}"
    authentication_method   = "SIMPLE"
    base_distinguished_name = "dc=vsphere,dc=local"
    connector_type          = "OPEN_LDAP"
    user_attributes {
      object_class                = "inetOrgPerson"
      unique_identifier           = "uid"
      display_name                = "cn"
      username                    = "uid"
      given_name                  = "givenName"
      surname                     = "sn"
      telephone                   = "telephoneNumber"
      group_membership_identifier = "dn"
      email                       = "mail"
    }
    group_attributes {
      name                        = "cn"
      object_class                = "group"
      membership                  = "member"
      unique_identifier           = "cn"
      group_membership_identifier = "dn"
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
