// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const providerLdapId = "Provider LDAP"

func resourceVcfaProviderLdap() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaProviderLdapCreate,
		ReadContext:   resourceVcfaProviderLdapRead,
		UpdateContext: resourceVcfaProviderLdapUpdate,
		DeleteContext: resourceVcfaProviderLdapDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaProviderLdapImport,
		},
		Schema: map[string]*schema.Schema{
			"server": { // HostName
				Type:        schema.TypeString,
				Required:    true,
				Description: "host name or IP of the LDAP server",
			},
			"port": { // Port
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Port number for LDAP service",
			},
			"base_distinguished_name": { //SearchBase
				Type:        schema.TypeString,
				Required:    true,
				Description: "LDAP search base",
			},
			"connector_type": { // ConnectorType
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of connector: one of OPEN_LDAP, ACTIVE_DIRECTORY",
				ValidateFunc: validation.StringInSlice([]string{"OPEN_LDAP", "ACTIVE_DIRECTORY"}, false),
			},
			"is_ssl": { // IsSsl
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "True if the LDAP service requires an SSL connection",
			},
			"auto_trust_certificate": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Defines if the LDAP certificate should automatically be trusted, only makes sense if 'is_ssl=true'",
			},
			"username": { // UserName
				Type:     schema.TypeString,
				Optional: true,
				Description: `Username to use when logging in to LDAP, specified using LDAP attribute=value ` +
					`pairs (for example: cn="ldap-admin", c="example", dc="com")`,
			},
			"password": { // Password
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Description: `Password for the user identified by UserName. This value is never returned back. ` +
					`It is inspected on create and modify. ` +
					`On modify, the absence of this element indicates that the password should not be changed`,
			},
			"user_attributes":  ldapUserAttributes(false),  // UserAttributes
			"group_attributes": ldapGroupAttributes(false), // GroupAttributes
			"custom_ui_button_label": { // CustomUiButtonLabel
				Type:        schema.TypeString,
				Optional:    true,
				Description: "If you provide a custom button label, on the login screen, the custom label replaces the default label for this identity provider",
			},
		},
	}
}

func resourceVcfaProviderLdapCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	settings, err := getTmLdapSettingsType(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = tmClient.TmLdapConfigure(settings, d.Get("auto_trust_certificate").(bool))
	if err != nil {
		return diag.Errorf("error configuring LDAP: %s", err)
	}
	return resourceVcfaProviderLdapRead(ctx, d, meta)
}

func resourceVcfaProviderLdapCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcfaProviderLdapCreateOrUpdate(ctx, d, meta)
}

func resourceVcfaProviderLdapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaLdapRead(ctx, d, meta, "resource")
}

func genericVcfaLdapRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	config, err := tmClient.TmGetLdapConfiguration()
	if err != nil {
		return diag.Errorf("error getting LDAP settings: %s", err)
	}

	err = saveTmLdapSettingsInState(d, config, origin)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVcfaProviderLdapUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcfaProviderLdapCreateOrUpdate(ctx, d, meta)
}

func resourceVcfaProviderLdapDelete(_ context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	err := tmClient.TmLdapDisable()
	if err != nil {
		return diag.Errorf("error disabling LDAP: %s", err)
	}
	return nil
}

func resourceVcfaProviderLdapImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// This is a no-op as read comes after and nothing is needed
	d.SetId(providerLdapId)
	return []*schema.ResourceData{d}, nil
}

func getTmLdapSettingsType(d *schema.ResourceData) (*types.TmLdapSettings, error) {
	settings := &types.TmLdapSettings{
		HostName:                d.Get("server").(string),
		Port:                    d.Get("port").(int),
		IsSsl:                   d.Get("is_ssl").(bool),
		SearchBase:              d.Get("base_distinguished_name").(string),
		UserName:                d.Get("username").(string),
		Password:                d.Get("password").(string),
		AuthenticationMechanism: "SIMPLE", // Only SIMPLE is allowed in UI
		ConnectorType:           d.Get("connector_type").(string),

		// Same values as UI:
		PageSize:      200,
		MaxResults:    200,
		MaxUserGroups: 1015,
	}

	rawUserAttributes := d.Get("user_attributes").([]interface{})[0].(map[string]interface{}) // Guaranteed as it's Required
	settings.UserAttributes = &types.LdapUserAttributesType{
		ObjectClass:               rawUserAttributes["object_class"].(string),
		ObjectIdentifier:          rawUserAttributes["unique_identifier"].(string),
		UserName:                  rawUserAttributes["username"].(string),
		Email:                     rawUserAttributes["email"].(string),
		FullName:                  rawUserAttributes["display_name"].(string),
		GivenName:                 rawUserAttributes["given_name"].(string),
		Surname:                   rawUserAttributes["surname"].(string),
		Telephone:                 rawUserAttributes["telephone"].(string),
		GroupMembershipIdentifier: rawUserAttributes["group_membership_identifier"].(string),
		GroupBackLinkIdentifier:   rawUserAttributes["group_back_link_identifier"].(string),
	}
	rawGroupAttributes := d.Get("group_attributes").([]interface{})[0].(map[string]interface{}) // Guaranteed as it's Required
	settings.GroupAttributes = &types.LdapGroupAttributesType{
		ObjectClass:          rawGroupAttributes["object_class"].(string),
		ObjectIdentifier:     rawGroupAttributes["unique_identifier"].(string),
		GroupName:            rawGroupAttributes["name"].(string),
		Membership:           rawGroupAttributes["membership"].(string),
		MembershipIdentifier: rawGroupAttributes["group_membership_identifier"].(string),
		BackLinkIdentifier:   rawGroupAttributes["group_back_link_identifier"].(string),
	}

	if uiLabel, ok := d.GetOk("custom_ui_button_label"); ok {
		settings.CustomUiButtonLabel = addrOf(uiLabel.(string))
	}

	return settings, nil
}

func saveTmLdapSettingsInState(d *schema.ResourceData, config *types.TmLdapSettings, origin string) error {
	d.SetId(providerLdapId) // We don't need an ID
	dSet(d, "server", config.HostName)
	dSet(d, "port", config.Port)
	dSet(d, "base_distinguished_name", config.SearchBase)
	dSet(d, "connector_type", config.ConnectorType)
	dSet(d, "is_ssl", config.IsSsl)
	dSet(d, "username", config.UserName)
	if config.CustomUiButtonLabel != nil {
		dSet(d, "custom_ui_button_label", *config.CustomUiButtonLabel)
	}
	if config.UserAttributes != nil {
		err := d.Set("user_attributes", []map[string]interface{}{
			{
				"object_class":                config.UserAttributes.ObjectClass,
				"unique_identifier":           config.UserAttributes.ObjectIdentifier,
				"username":                    config.UserAttributes.UserName,
				"email":                       config.UserAttributes.Email,
				"display_name":                config.UserAttributes.FullName,
				"given_name":                  config.UserAttributes.GivenName,
				"surname":                     config.UserAttributes.Surname,
				"telephone":                   config.UserAttributes.Telephone,
				"group_membership_identifier": config.UserAttributes.GroupMembershipIdentifier,
				"group_back_link_identifier":  config.UserAttributes.GroupBackLinkIdentifier,
			},
		})
		if err != nil {
			return fmt.Errorf("error setting LDAP user_attributes: %s", err)
		}
	}
	if config.GroupAttributes != nil {
		err := d.Set("group_attributes", []map[string]interface{}{
			{
				"object_class":                config.GroupAttributes.ObjectClass,
				"unique_identifier":           config.GroupAttributes.ObjectIdentifier,
				"name":                        config.GroupAttributes.GroupName,
				"membership":                  config.GroupAttributes.Membership,
				"group_membership_identifier": config.GroupAttributes.MembershipIdentifier,
				"group_back_link_identifier":  config.GroupAttributes.BackLinkIdentifier,
			},
		})
		if err != nil {
			return fmt.Errorf("error setting LDAP group_attributes: %s", err)
		}
	}
	return nil
}
