package vcfa

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"log"
	"strings"
	"time"
)

const labelVcfaOidc = "OpenID Connect"

// resourceVcfaOrgOidc defines the resource that manages OpenID Connect (OIDC) settings for an Organization
func resourceVcfaOrgOidc() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcfaOrgOidcRead,
		CreateContext: resourceVcfaOrgOidcCreate,
		UpdateContext: resourceVcfaOrgOidcUpdate,
		DeleteContext: resourceVcfaOrgOidcDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgOidcImport,
		},
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("ID of the %s that will have the %s settings configured", labelVcfaOrg, labelVcfaOidc),
			},
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Client ID to use when talking to the %s Identity Provider", labelVcfaOidc),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Client Secret to use when talking to the %s Identity Provider", labelVcfaOidc),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: fmt.Sprintf("Enables or disables %s authentication for the specified Organization", labelVcfaOidc),
			},
			"wellknown_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Endpoint from the %s Identity Provider that serves all the configuration values", labelVcfaOidc),
			},
			"issuer_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: fmt.Sprintf("The issuer identifier of the %s Identity Provider. "+
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained issuer identifier", labelVcfaOidc),
				AtLeastOneOf: []string{"issuer_id", "wellknown_endpoint"},
			},
			"user_authorization_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: fmt.Sprintf("The user authorization endpoint of the %s Identity Provider. "+
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained user authorization endpoint", labelVcfaOidc),
				AtLeastOneOf: []string{"user_authorization_endpoint", "wellknown_endpoint"},
			},
			"access_token_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: fmt.Sprintf("The access token endpoint of the %s Identity Provider. "+
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained access token endpoint", labelVcfaOidc),
				AtLeastOneOf: []string{"access_token_endpoint", "wellknown_endpoint"},
			},
			"userinfo_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: fmt.Sprintf("The user info endpoint of the %s Identity Provider. "+
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained user info endpoint", labelVcfaOidc),
				AtLeastOneOf: []string{"userinfo_endpoint", "wellknown_endpoint"},
			},
			"prefer_id_token": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "If you want to combine claims from 'userinfo_endpoint' and the ID Token, set this to 'true'. " +
					"The identity providers do not provide all the required claims set in 'userinfo_endpoint'." +
					"By setting this argument to 'true', VCFA can fetch and consume claims from both sources",
			},
			"max_clock_skew_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
				Description: "The maximum clock skew is the maximum allowable time difference between the client and server. " +
					"This time compensates for any small-time differences in the timestamps when verifying tokens",
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			},
			"scopes": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: fmt.Sprintf("A set of scopes to use with the %s provider. "+
					"They are used to authorize access to user details, by defining the permissions that the access tokens have to access user information. "+
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained scopes", labelVcfaOidc),
			},
			"claims_mapping": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: fmt.Sprintf("A single configuration block that specifies the claim mappings to use with the %s provider. "+
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained claim mappings", labelVcfaOidc),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"email": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Email claim mapping",
						},
						"subject": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Subject claim mapping",
						},
						"last_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Last name claim mapping",
						},
						"first_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "First name claim mapping",
						},
						"full_name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Full name claim mapping",
						},
						"groups": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Groups claim mapping",
						},
						"roles": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true, // Can be obtained with "wellknown_endpoint"
							Description: "Roles claim mapping",
						},
					},
				},
			},
			"key": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: fmt.Sprintf("One or more configuration blocks that specify the keys to use with the %s Identity Provider. "+
					"If 'wellknown_endpoint' is set, this attribute overrides the obtained keys", labelVcfaOidc),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the key",
						},
						"algorithm": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Algorithm of the key, either RSA or EC",
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"RSA", "EC"}, false)),
						},
						"certificate": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The certificate contents",
						},
						"expiration_date": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Expiration date for the certificate",
							ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
								if i.(string) == "" {
									return nil // It's an optional value
								}
								_, err := time.Parse(time.DateOnly, i.(string))
								if err != nil {
									return diag.FromErr(err)
								}
								return nil
							},
						},
					},
				},
			},
			"key_refresh_endpoint": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Can be obtained with "wellknown_endpoint"
				Description: "Endpoint used to refresh the keys. If 'wellknown_endpoint' is set, then this argument" +
					"will override the obtained endpoint",
				RequiredWith: []string{"key_refresh_period_hours", "key_refresh_strategy"},
			},
			"key_refresh_period_hours": {
				Type:             schema.TypeInt,
				Optional:         true,
				Description:      "Defines the frequency of key refresh. Maximum is 720 hours",
				RequiredWith:     []string{"key_refresh_endpoint"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 720)),
			},
			"key_refresh_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "Defines the strategy of key refresh",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"ADD", "REPLACE", "EXPIRE_AFTER"}, false)),
				RequiredWith:     []string{"key_refresh_endpoint"},
			},
			"key_expire_duration_hours": {
				Type:             schema.TypeInt,
				Optional:         true,
				Description:      "Defines the expiration period of the key, only when 'key_refresh_strategy=EXPIRE_AFTER'. Maximum is 24 hours",
				RequiredWith:     []string{"key_refresh_endpoint", "key_refresh_strategy"},
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 24)),
			},
			"ui_button_label": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Customizes the label of the UI button of the login screen",
			},
			"redirect_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Redirect URI for this org",
			},
		},
	}
}

func resourceVcfaOrgOidcCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgId := d.Get("org_id").(string)

	org, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[%s %s] error searching for Org '%s': %s", labelVcfaOidc, operation, orgId, err)
	}

	// Runtime validations
	isWellKnownEndpointUsed := d.Get("wellknown_endpoint").(string) != ""
	scopes := d.Get("scopes").(*schema.Set).List()
	if !isWellKnownEndpointUsed && len(scopes) == 0 {
		return diag.Errorf("[%s %s] 'scopes' cannot be empty when a well-known endpoint is not used", labelVcfaOidc, operation)
	}

	if _, ok := d.GetOk("key_expire_duration_hours"); ok && d.Get("key_refresh_strategy") != "EXPIRE_AFTER" {
		return diag.Errorf("[%s %s] 'key_expire_duration_hours' can only be used when 'key_refresh_strategy=EXPIRE_AFTER', but key_refresh_strategy=%s", labelVcfaOidc, operation, d.Get("key_refresh_strategy"))
	}
	if _, ok := d.GetOk("key_expire_duration_hours"); !ok && d.Get("key_refresh_strategy") == "EXPIRE_AFTER" {
		return diag.Errorf("[%s %s] 'key_refresh_strategy=EXPIRE_AFTER' requires 'key_expire_duration_hours' to be set", labelVcfaOidc, operation)
	}
	// End of validations

	settings := types.OrgOAuthSettings{
		IssuerId:                   d.Get("issuer_id").(string),
		Enabled:                    d.Get("enabled").(bool),
		ClientId:                   d.Get("client_id").(string),
		ClientSecret:               d.Get("client_secret").(string),
		UserAuthorizationEndpoint:  d.Get("user_authorization_endpoint").(string),
		AccessTokenEndpoint:        d.Get("access_token_endpoint").(string),
		UserInfoEndpoint:           d.Get("userinfo_endpoint").(string),
		MaxClockSkew:               d.Get("max_clock_skew_seconds").(int),
		JwksUri:                    d.Get("key_refresh_endpoint").(string),
		AutoRefreshKey:             d.Get("key_refresh_endpoint").(string) != "" && d.Get("key_refresh_strategy").(string) != "",
		KeyRefreshStrategy:         d.Get("key_refresh_strategy").(string),
		KeyRefreshFrequencyInHours: d.Get("key_refresh_period_hours").(int),
		WellKnownEndpoint:          d.Get("wellknown_endpoint").(string),
		Scope:                      convertTypeListToSliceOfStrings(scopes),
		EnableIdTokenClaims:        addrOf(d.Get("prefer_id_token").(bool)),
		CustomUiButtonLabel:        addrOf(d.Get("ui_button_label").(string)),
	}

	// Key configurations: OAuthKeyConfigurations
	keyList := d.Get("key").(*schema.Set).List()
	if len(keyList) == 0 && !isWellKnownEndpointUsed {
		return diag.Errorf("[%s %s] error reading keys, either set a 'key' block or set 'wellknown_endpoint' to obtain this information", labelVcfaOidc, operation)
	}
	if len(keyList) > 0 {
		oAuthKeyConfigurations := make([]types.OAuthKeyConfiguration, len(keyList))
		for i, k := range keyList {
			key := k.(map[string]interface{})
			oAuthKeyConfigurations[i] = types.OAuthKeyConfiguration{
				KeyId:     key["id"].(string),
				Algorithm: key["algorithm"].(string),
				Key:       key["certificate"].(string),
			}
			if key["expiration_date"].(string) != "" {
				t, err := time.Parse(time.DateOnly, key["expiration_date"].(string))
				if err != nil {
					return diag.Errorf("wrong expiration date set in configuration for key '%s': %s", key["id"].(string), err)
				}
				oAuthKeyConfigurations[i].ExpirationDate = t.Format(time.RFC3339)
			}
		}
		settings.OAuthKeyConfigurations = &types.OAuthKeyConfigurationsList{
			OAuthKeyConfiguration: oAuthKeyConfigurations,
		}
	}

	// Claims mapping: OIDCAttributeMapping: Subject, Email, Full name, First name and Last name are mandatory
	claimsMapping := d.Get("claims_mapping").([]interface{})
	if len(claimsMapping) == 0 && !isWellKnownEndpointUsed {
		return diag.Errorf("[%s %s] error reading claims, either set a 'claims_mapping' block or set 'wellknown_endpoint' to obtain this information", labelVcfaOidc, operation)
	}
	if len(claimsMapping) > 0 {
		var oidcAttributeMapping types.OIDCAttributeMapping
		mappingEntry := claimsMapping[0].(map[string]interface{})
		oidcAttributeMapping.SubjectAttributeName = mappingEntry["subject"].(string)
		oidcAttributeMapping.EmailAttributeName = mappingEntry["email"].(string)
		oidcAttributeMapping.FullNameAttributeName = mappingEntry["full_name"].(string)
		oidcAttributeMapping.FirstNameAttributeName = mappingEntry["first_name"].(string)
		oidcAttributeMapping.LastNameAttributeName = mappingEntry["last_name"].(string)
		oidcAttributeMapping.GroupsAttributeName = mappingEntry["groups"].(string)
		oidcAttributeMapping.RolesAttributeName = mappingEntry["roles"].(string)
		settings.OIDCAttributeMapping = &oidcAttributeMapping
	}

	_, err = setOIDCSettings(org, settings)
	if err != nil {
		return diag.Errorf("[%s %s] Could not set OIDC settings: %s", labelVcfaOidc, operation, err)
	}

	return resourceVcfaOrgOidcRead(ctx, d, meta)
}
func resourceVcfaOrgOidcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcfaOrgOidcCreateOrUpdate(ctx, d, meta, "create")
}
func resourceVcfaOrgOidcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcfaOrgOidcCreateOrUpdate(ctx, d, meta, "update")
}

func resourceVcfaOrgOidcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaOrgOidcRead(ctx, d, meta, "resource")
}

func genericVcfaOrgOidcRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)

	adminOrg, err := vcdClient.GetAdminOrgByNameOrId(orgId)
	if govcd.ContainsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] unable to find Organization '%s' %s settings: %s. Removing from state", orgId, labelVcfaOidc, err)
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("[%s read] unable to find Organization '%s': %s", labelVcfaOidc, orgId, err)
	}

	settings, err := adminOrg.GetOpenIdConnectSettings()
	if err != nil {
		return diag.Errorf("[%s read] unable to read Organization '%s' OIDC settings: %s", labelVcfaOidc, orgId, err)
	}

	dSet(d, "client_id", settings.ClientId)
	dSet(d, "client_secret", settings.ClientSecret)
	dSet(d, "enabled", settings.Enabled)
	dSet(d, "wellknown_endpoint", settings.WellKnownEndpoint)
	dSet(d, "issuer_id", settings.IssuerId)
	dSet(d, "user_authorization_endpoint", settings.UserAuthorizationEndpoint)
	dSet(d, "access_token_endpoint", settings.AccessTokenEndpoint)
	dSet(d, "userinfo_endpoint", settings.UserInfoEndpoint)
	dSet(d, "max_clock_skew_seconds", settings.MaxClockSkew)
	err = d.Set("scopes", settings.Scope)
	if err != nil {
		return diag.FromErr(err)
	}
	if settings.OIDCAttributeMapping != nil {
		claims := make([]interface{}, 1)
		claim := map[string]interface{}{}
		claim["email"] = settings.OIDCAttributeMapping.EmailAttributeName
		claim["subject"] = settings.OIDCAttributeMapping.SubjectAttributeName
		claim["last_name"] = settings.OIDCAttributeMapping.LastNameAttributeName
		claim["first_name"] = settings.OIDCAttributeMapping.FirstNameAttributeName
		claim["full_name"] = settings.OIDCAttributeMapping.FullNameAttributeName
		claim["groups"] = settings.OIDCAttributeMapping.GroupsAttributeName
		claim["roles"] = settings.OIDCAttributeMapping.RolesAttributeName
		claims[0] = claim
		err = d.Set("claims_mapping", claims)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if settings.OAuthKeyConfigurations != nil {
		keyConfigurations := settings.OAuthKeyConfigurations.OAuthKeyConfiguration
		keyConfigs := make([]map[string]interface{}, len(keyConfigurations))
		for i, keyConfig := range keyConfigurations {
			key := map[string]interface{}{}
			key["id"] = keyConfig.KeyId
			key["algorithm"] = keyConfig.Algorithm
			key["certificate"] = keyConfig.Key
			if keyConfig.ExpirationDate != "" {
				t, err := time.Parse(time.RFC3339, keyConfig.ExpirationDate)
				if err != nil {
					return diag.Errorf("wrong expiration date received for key '%s': %s", keyConfig.KeyId, err)
				}
				key["expiration_date"] = t.Format(time.DateOnly)
			}
			keyConfigs[i] = key
		}
		err = d.Set("key", keyConfigs)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	dSet(d, "key_refresh_endpoint", settings.JwksUri)
	dSet(d, "key_refresh_period_hours", settings.KeyRefreshFrequencyInHours)
	dSet(d, "key_refresh_strategy", settings.KeyRefreshStrategy)
	dSet(d, "key_expire_duration_hours", settings.KeyExpireDurationInHours)
	dSet(d, "redirect_uri", settings.OrgRedirectUri)

	if settings.EnableIdTokenClaims != nil {
		dSet(d, "prefer_id_token", *settings.EnableIdTokenClaims)
	}
	if settings.CustomUiButtonLabel != nil {
		dSet(d, "ui_button_label", *settings.CustomUiButtonLabel)
	}

	d.SetId(adminOrg.AdminOrg.ID)

	return nil
}

func resourceVcfaOrgOidcDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	orgId := d.Get("org_id").(string)

	adminOrg, err := vcdClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[%s delete] error searching for Organization '%s': %s", labelVcfaOidc, orgId, err)
	}

	err = adminOrg.DeleteOpenIdConnectSettings()
	if err != nil {
		return diag.Errorf("[%s delete] error deleting OIDC settings for Organization '%s': %s", labelVcfaOidc, orgId, err)
	}

	return nil
}

// resourceVcfaOrgOidcImport is responsible for importing the resource.
// The only parameter needed is the Org identifier, which could be either the Org name or its ID
func resourceVcfaOrgOidcImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	orgNameOrId := d.Id()

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgByNameOrId(orgNameOrId)
	if err != nil {
		return nil, fmt.Errorf("[%s import] error searching for Organization '%s': %s", labelVcfaOidc, orgNameOrId, err)
	}

	dSet(d, "org_id", adminOrg.AdminOrg.ID)

	d.SetId(adminOrg.AdminOrg.ID)
	return []*schema.ResourceData{d}, nil
}

// setOIDCSettings sets the given OIDC settings for the given Organization. It does this operation
// with some tries to avoid failures due to network glitches.
func setOIDCSettings(adminOrg *govcd.AdminOrg, settings types.OrgOAuthSettings) (*types.OrgOAuthSettings, error) {
	tries := 0
	var newSettings *types.OrgOAuthSettings
	var err error
	for tries < 5 {
		tries++
		newSettings, err = adminOrg.SetOpenIdConnectSettings(settings)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "could not establish a connection") || strings.Contains(err.Error(), "connect timed out") {
			time.Sleep(10 * time.Second)
		}
	}
	if err != nil {
		return nil, err
	}
	return newSettings, nil
}
