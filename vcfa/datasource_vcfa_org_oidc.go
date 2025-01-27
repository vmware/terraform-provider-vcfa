package vcfa

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// datasourceVcfaOrgOidc defines the data source that reads Open ID Connect (OIDC) settings from an Organization
func datasourceVcfaOrgOidc() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaOrgOidcRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s ID that has the %s settings", labelVcfaOrg, labelVcfaOidc),
			},
			"client_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Client ID used when talking to the %s Identity Provider", labelVcfaOidc),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Client Secret used when talking to the %s Identity Provider", labelVcfaOidc),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether %s authentication for the specified Organization is enabled or disabled", labelVcfaOidc),
			},
			"wellknown_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Endpoint from the %s Identity Provider that serves all the configuration values", labelVcfaOidc),
			},
			"issuer_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The issuer identifier of the %s Identity Provider", labelVcfaOidc),
			},
			"user_authorization_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The user authorization endpoint of the %s Identity Provider", labelVcfaOidc),
			},
			"access_token_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The access token endpoint of the %s Identity Provider", labelVcfaOidc),
			},
			"userinfo_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The user info endpoint of the %s Identity Provider", labelVcfaOidc),
			},
			"prefer_id_token": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the claims from 'userinfo_endpoint' and the ID Token are combined (true) or not (false)",
			},
			"max_clock_skew_seconds": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The maximum clock skew is the maximum allowable time difference between the client and server",
			},
			"scopes": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: fmt.Sprintf("A set of scopes used with the %s provider", labelVcfaOidc),
			},
			"claims_mapping": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: fmt.Sprintf("A single configuration block that contains the claim mappings used with the %s provider", labelVcfaOidc),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Email claim mapping",
						},
						"subject": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Subject claim mapping",
						},
						"last_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last name claim mapping",
						},
						"first_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "First name claim mapping",
						},
						"full_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Full name claim mapping",
						},
						"groups": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Groups claim mapping",
						},
						"roles": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Roles claim mapping",
						},
					},
				},
			},
			"key": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("One or more configuration blocks that contain the keys used with the %s Identity Provider", labelVcfaOidc),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "ID of the key",
						},
						"algorithm": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Algorithm of the key",
						},
						"certificate": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The certificate contents",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Expiration date for the certificate",
						},
					},
				},
			},
			"key_refresh_endpoint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint used to refresh the keys",
			},
			"key_refresh_period_hours": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The frequency of key refresh",
			},
			"key_refresh_strategy": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the strategy of key refresh",
			},
			"key_expire_duration_hours": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The expiration period of the key, only available if 'key_refresh_strategy=EXPIRE_AFTER'",
			},
			"ui_button_label": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The label of the UI button of the login screen. Only available since VCD 10.5.1",
			},
			"redirect_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Redirect URI for this org",
			},
		},
	}
}

func datasourceVcfaOrgOidcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaOrgOidcRead(ctx, d, meta, "datasource")
}
