// © Broadcom. All Rights Reserved.
// The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcfaOrgLdap() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaOrgLdapRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Organization ID",
			},
			"ldap_mode": { // OrgLdapMode
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of LDAP settings (one of NONE, SYSTEM, CUSTOM)",
			},
			"custom_user_ou": { // CustomUsersOu
				Type:        schema.TypeString,
				Computed:    true,
				Description: "If ldap_mode is SYSTEM, specifies a LDAP attribute=value pair to use for OU (organizational unit)",
			},
			"custom_settings": { // CustomOrgLdapSettings
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Custom settings when `ldap_mode` is CUSTOM",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"server": { // Hostname
							Type:        schema.TypeString,
							Computed:    true,
							Description: "host name or IP of the LDAP server",
						},
						"port": { // Port
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port number for LDAP service",
						},
						"connector_type": { // ConnectorType
							Type:        schema.TypeString,
							Computed:    true,
							Description: "type of connector: one of OPEN_LDAP, ACTIVE_DIRECTORY",
						},
						"base_distinguished_name": { //SearchBase
							Type:        schema.TypeString,
							Computed:    true,
							Description: "LDAP search base",
						},
						"is_ssl": { // IsSsl
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the LDAP service requires an SSL connection",
						},
						"username": { // Username
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Username to use when logging in to LDAP, specified using LDAP attribute=value pairs (for example: cn="ldap-admin", c="example", dc="com")`,
						},
						"custom_ui_button_label": { // CustomUiButtonLabel
							Type:        schema.TypeString,
							Computed:    true,
							Description: "On the login screen, the custom label replaces the default label for this identity provider",
						},
						"user_attributes":  ldapUserAttributes(true),
						"group_attributes": ldapGroupAttributes(true),
					},
				},
			},
		},
	}
}

func datasourceVcfaOrgLdapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaOrgLdapRead(ctx, d, meta, "datasource", nil)
}
