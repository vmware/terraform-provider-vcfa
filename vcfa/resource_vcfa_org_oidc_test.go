//go:build org || ALL || functional

package vcfa

import (
	_ "embed"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

func TestAccVcfaOrgOidc(t *testing.T) {
	preTestChecks(t)
	oidcServerUrl := validateAndGetOidcServerUrl(t, testConfig)

	orgName1 := t.Name() + "1"
	orgName2 := t.Name() + "2"
	orgName3 := t.Name() + "3"
	oidcResource1 := "vcfa_org_oidc.oidc1"
	oidcResource2 := "vcfa_org_oidc.oidc2"
	oidcResource3 := "vcfa_org_oidc.oidc3"
	oidcData := "data.vcfa_org_oidc.oidc_data"

	var params = StringMap{
		"OrgName1":          orgName1,
		"OrgName2":          orgName2,
		"OrgName3":          orgName3,
		"WellKnownEndpoint": oidcServerUrl.String(),
		"FuncName":          t.Name() + "-Step1",
		"PreferIdToken":     "true",
		"UIButtonLabel":     "this is a test",
		"SkipBinary":        "# skip-binary-test: redundant test",
	}
	testParamsNotEmpty(t, params)
	skipIfNotSysAdmin(t)

	step1 := templateFill(testAccCheckVcfaOrgOidc, params)
	params["FuncName"] = t.Name() + "-Step2"
	step2 := templateFill(testAccCheckVcfaOrgOidc2, params)
	params["FuncName"] = t.Name() + "-Step3"
	params["SkipBinary"] = " "
	step3 := templateFill(testAccCheckVcfaOrgOidc3, params)
	if vcfaShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] Configuration Step 1: %s", step1)
	debugPrintf("#[DEBUG] Configuration Step 2: %s", step2)
	debugPrintf("#[DEBUG] Configuration Step 3: %s", step3)

	skip := false
	skipFunc := func() (bool, error) {
		return skip, nil
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		ErrorCheck: func(err error) error {
			if strings.Contains(err.Error(), "could not establish a connection") {
				skip = true
				fmt.Printf("skipping %s as the OIDC server is not responding: %s", t.Name(), err)
				return nil
			}
			return err
		},
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckOrgDestroy(orgName1),
			testAccCheckOrgDestroy(orgName2),
			testAccCheckOrgDestroy(orgName3),
		),
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcfaOrgExists("vcfa_org.org1"),
					testAccCheckVcfaOrgExists("vcfa_org.org2"),
					testAccCheckVcfaOrgExists("vcfa_org.org3"),

					resource.TestMatchResourceAttr(oidcResource1, "redirect_uri", regexp.MustCompile(fmt.Sprintf("(?i).*=tenant:%s", orgName1))),
					resource.TestCheckResourceAttr(oidcResource1, "client_id", "clientId"),
					resource.TestCheckResourceAttr(oidcResource1, "client_secret", "clientSecret"),
					resource.TestCheckResourceAttr(oidcResource1, "enabled", "true"),
					resource.TestCheckResourceAttr(oidcResource1, "wellknown_endpoint", params["WellKnownEndpoint"].(string)),
					resource.TestMatchResourceAttr(oidcResource1, "issuer_id", regexp.MustCompile(fmt.Sprintf("^%s://%s.*$", oidcServerUrl.Scheme, oidcServerUrl.Host))),
					resource.TestMatchResourceAttr(oidcResource1, "user_authorization_endpoint", regexp.MustCompile(fmt.Sprintf("^%s://%s.*$", oidcServerUrl.Scheme, oidcServerUrl.Host))),
					resource.TestMatchResourceAttr(oidcResource1, "access_token_endpoint", regexp.MustCompile(fmt.Sprintf("^%s://%s.*$", oidcServerUrl.Scheme, oidcServerUrl.Host))),
					resource.TestMatchResourceAttr(oidcResource1, "userinfo_endpoint", regexp.MustCompile(fmt.Sprintf("^%s://%s.*$", oidcServerUrl.Scheme, oidcServerUrl.Host))),
					resource.TestMatchResourceAttr(oidcResource1, "prefer_id_token", regexp.MustCompile("^true$")),
					resource.TestCheckResourceAttr(oidcResource1, "max_clock_skew_seconds", "60"),
					resource.TestMatchResourceAttr(oidcResource1, "scopes.#", regexp.MustCompile(`[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(oidcResource1, "claims_mapping.0.email"),
					resource.TestCheckResourceAttrSet(oidcResource1, "claims_mapping.0.subject"),
					resource.TestCheckResourceAttrSet(oidcResource1, "claims_mapping.0.last_name"),
					resource.TestCheckResourceAttrSet(oidcResource1, "claims_mapping.0.first_name"),
					resource.TestMatchResourceAttr(oidcResource1, "key.#", regexp.MustCompile(`[1-9][0-9]*`)),
					resource.TestMatchResourceAttr(oidcResource1, "ui_button_label", regexp.MustCompile("^this is a test$")),
				),
			},
			{
				Config:   step2,
				SkipFunc: skipFunc,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(oidcResource1, oidcResource2, []string{
						"id", "org_id", "redirect_uri", "wellknown_endpoint", "key_refresh_endpoint",
						"issuer_id", "claims_mapping.0.subject", "ui_button_label", "prefer_id_token",
					}),
					resource.TestCheckResourceAttr(oidcResource2, "issuer_id", "https://doesnotexist.broadcom.com"),
					resource.TestCheckResourceAttr(oidcResource2, "claims_mapping.0.subject", "foo"),
					resourceFieldsEqual(oidcResource1, oidcResource3, []string{
						"id", "org_id", "redirect_uri", "wellknown_endpoint", "key_refresh_endpoint", "key.0.expiration_date",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(oidcResource3, "key.*", map[string]string{
						"expiration_date": "2077-05-13",
					}),
				),
			},
			{
				Config:   step3,
				SkipFunc: skipFunc,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(oidcResource1, oidcData, nil),
				),
			},
			{
				SkipFunc:          skipFunc,
				ResourceName:      oidcResource1,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) { return orgName1, nil },
			},
		},
	})

	postTestChecks(t)
}

const testAccCheckVcfaOrgOidc = `
{{.SkipBinary}}
resource "vcfa_org" "org1" {
  name              = "{{.OrgName1}}"
  full_name         = "{{.OrgName1}}"
  description       = "{{.OrgName1}}"
  delete_force      = true
  delete_recursive  = true
}

resource "vcfa_org" "org2" {
  name              = "{{.OrgName2}}"
  full_name         = "{{.OrgName2}}"
  description       = "{{.OrgName2}}"
  delete_force      = true
  delete_recursive  = true
}

resource "vcfa_org" "org3" {
  name              = "{{.OrgName3}}"
  full_name         = "{{.OrgName3}}"
  description       = "{{.OrgName3}}"
  delete_force      = true
  delete_recursive  = true
}

resource "vcfa_org_oidc" "oidc1" {
  org_id                      = vcfa_org.org1.id
  enabled                     = true
  {{.PreferIdToken}}
  client_id                   = "clientId"
  client_secret               = "clientSecret"
  max_clock_skew_seconds      = 60
  wellknown_endpoint          = "{{.WellKnownEndpoint}}"
  {{.UIButtonLabel}}
}
`

const testAccCheckVcfaOrgOidc2 = testAccCheckVcfaOrgOidc + `
resource "vcfa_org_oidc" "oidc2" {
  org_id                 = vcfa_org.org2.id
  enabled                = true
  client_id              = "clientId"
  client_secret          = "clientSecret"
  max_clock_skew_seconds = 60
  wellknown_endpoint     = "{{.WellKnownEndpoint}}"
  issuer_id              = "https://doesnotexist.broadcom.com"
  claims_mapping {
	subject = "foo"
  }
}

resource "vcfa_org_oidc" "oidc3" {
  org_id                      = vcfa_org.org3.id
  enabled                     = vcfa_org_oidc.oidc1.enabled
  {{.PreferIdToken}}
  client_id                   = vcfa_org_oidc.oidc1.client_id
  client_secret               = vcfa_org_oidc.oidc1.client_secret
  max_clock_skew_seconds      = vcfa_org_oidc.oidc1.max_clock_skew_seconds
  issuer_id                   = vcfa_org_oidc.oidc1.issuer_id
  user_authorization_endpoint = vcfa_org_oidc.oidc1.user_authorization_endpoint
  access_token_endpoint       = vcfa_org_oidc.oidc1.access_token_endpoint
  userinfo_endpoint           = vcfa_org_oidc.oidc1.userinfo_endpoint
  scopes                      = vcfa_org_oidc.oidc1.scopes
  claims_mapping {
    email      = vcfa_org_oidc.oidc1.claims_mapping[0].email
    subject    = vcfa_org_oidc.oidc1.claims_mapping[0].subject
    last_name  = vcfa_org_oidc.oidc1.claims_mapping[0].last_name
    first_name = vcfa_org_oidc.oidc1.claims_mapping[0].first_name
    full_name  = vcfa_org_oidc.oidc1.claims_mapping[0].full_name
    groups     = vcfa_org_oidc.oidc1.claims_mapping[0].groups
    roles      = vcfa_org_oidc.oidc1.claims_mapping[0].roles
  }
  key {
    id              = tolist(vcfa_org_oidc.oidc1.key)[0].id
    algorithm       = tolist(vcfa_org_oidc.oidc1.key)[0].algorithm
    certificate     = tolist(vcfa_org_oidc.oidc1.key)[0].certificate
	expiration_date = "2077-05-13"
  }
  {{.UIButtonLabel}}
}
`

const testAccCheckVcfaOrgOidc3 = testAccCheckVcfaOrgOidc2 + `
data "vcfa_org_oidc" "oidc_data" {
  org_id = vcfa_org.org1.id
}
`

func validateAndGetOidcServerUrl(t *testing.T, testConfig TestConfig) *url.URL {
	if testConfig.Tm.OidcServer.Url == "" || testConfig.Tm.OidcServer.WellKnownEndpoint == "" {
		t.Skip("test requires OIDC configuration")
	}

	oidcServer, err := url.Parse(testConfig.Tm.OidcServer.Url)
	if err != nil {
		t.Skip(t.Name() + " requires OIDC Server URL and its well-known endpoint")
	}
	return oidcServer.JoinPath(testConfig.Tm.OidcServer.WellKnownEndpoint)
}
