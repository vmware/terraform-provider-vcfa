// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcfaLdap() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaLdapRead,
		Schema: map[string]*schema.Schema{
			"server": { // HostName
				Type:        schema.TypeString,
				Computed:    true,
				Description: "host name or IP of the LDAP server",
			},
			"port": { // Port
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Port number for LDAP service",
			},
			"base_distinguished_name": { //SearchBase
				Type:        schema.TypeString,
				Computed:    true,
				Description: "LDAP search base",
			},
			"connector_type": { // ConnectorType
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of connector: one of OPEN_LDAP, ACTIVE_DIRECTORY",
			},
			"is_ssl": { // IsSsl
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the LDAP service requires an SSL connection",
			},
			"username": { // UserName
				Type:     schema.TypeString,
				Computed: true,
				Description: `Username to use when logging in to LDAP, specified using LDAP attribute=value ` +
					`pairs (for example: cn="ldap-admin", c="example", dc="com")`,
			},
			"user_attributes":  ldapUserAttributes(true),  // UserAttributes
			"group_attributes": ldapGroupAttributes(true), // GroupAttributes
			"custom_ui_button_label": { // CustomUiButtonLabel
				Type:        schema.TypeString,
				Computed:    true,
				Description: "If you provide a custom button label, on the login screen, the custom label replaces the default label for this identity provider",
			},
		},
	}
}

func datasourceVcfaLdapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaLdapRead(ctx, d, meta, "datasource")
}
