package vcfa

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

var ldapUserAttributes = func(isDatasource bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: !isDatasource,
		Computed: isDatasource,
		MaxItems: func() int {
			if isDatasource {
				return 0
			}
			return 1
		}(),
		Description: "LDAP user attributes",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"object_class": { // ObjectClass
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP objectClass of which imported users are members. For example, 'user' or 'person' attributes",
				},
				"unique_identifier": { // ObjectIdentifier
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use as the unique identifier for a user. For example, the 'objectGuid' attribute",
				},
				"username": { // Username
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use when looking up a user name to import. For example, 'userPrincipalName' or 'samAccountName' attributes",
				},
				"email": { // Email
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use for the user's email address. For example, the 'mail' attribute",
				},
				"display_name": { // FullName
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use for the user's full name. For example, the 'displayName' attribute",
				},
				"given_name": { // GivenName
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use for the user's given name. For example, the 'givenName' attribute",
				},
				"surname": { // Surname
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use for the user's surname. For example, the 'sn' attribute",
				},
				"telephone": { // Telephone
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use for the user's telephone number. For example, the 'telephoneNumber' attribute",
				},
				"group_membership_identifier": { // GroupMembershipIdentifier
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute that identifies a user as a member of a group. For example, the 'dn' attribute",
				},
				"group_back_link_identifier": { // GroupBackLinkIdentifier
					Type:        schema.TypeString,
					Optional:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute that returns the identifiers of all the groups of which the user is a member",
				},
			},
		},
	}
}
var ldapGroupAttributes = func(isDatasource bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: !isDatasource,
		Computed: isDatasource,
		MaxItems: func() int {
			if isDatasource {
				return 0
			}
			return 1
		}(),
		Description: "Group settings when `ldap_mode` is CUSTOM",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"object_class": { // ObjectClass
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP objectClass of which imported groups are members. For example, the 'group' attribute",
				},
				"unique_identifier": { // ObjectIdentifier
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use as the unique identifier for a group. For example, the 'objectGuid' attribute",
				},
				"name": { // GroupName
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use for the group name. For example, the 'cn' attribute",
				},
				"membership": { // Membership
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute to use when getting the members of a group. For example, the 'member' attribute",
				},
				"group_membership_identifier": { // MembershipIdentifier
					Type:        schema.TypeString,
					Required:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP attribute that identifies a group as a member of another group. For example, the 'dn' attribute",
				},
				"group_back_link_identifier": { // BackLinkIdentifier
					Type:        schema.TypeString,
					Optional:    !isDatasource,
					Computed:    isDatasource,
					Description: "LDAP group attribute used to identify a group member",
				},
			},
		},
	}
}

// resourceVcfaOrgLdap defines types.OrgLdapSettingsType
// The field names are the ones used in the GUI, with a comment to indicate which API field each one corresponds to
func resourceVcfaOrgLdap() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcfaOrgLdapRead,
		CreateContext: resourceVcfaOrgLdapCreate,
		UpdateContext: resourceVcfaOrgLdapUpdate,
		DeleteContext: resourceVcfaOrgLdapDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgLdapImport,
		},
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Organization ID",
			},
			"ldap_mode": { // OrgLdapMode
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of LDAP settings (one of NONE, SYSTEM, CUSTOM)",
				ValidateFunc: validation.StringInSlice([]string{types.LdapModeNone, types.LdapModeSystem, types.LdapModeCustom}, false),
			},
			"auto_trust_certificate": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: "Defines if the LDAP certificate should automatically be trusted, only makes sense if 'custom_settings.0.is_ssl=true'",
			},
			"custom_user_ou": { // CustomUsersOu
				Type:        schema.TypeString,
				Optional:    true,
				Description: "If ldap_mode is SYSTEM, specifies a LDAP attribute=value pair to use for OU (organizational unit)",
			},
			"custom_settings": { // CustomOrgLdapSettings
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Custom settings when `ldap_mode` is CUSTOM",
				Elem: &schema.Resource{
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
						"connector_type": { // ConnectorType
							Type:         schema.TypeString,
							Required:     true,
							Description:  "type of connector: one of OPEN_LDAP, ACTIVE_DIRECTORY",
							ValidateFunc: validation.StringInSlice([]string{"OPEN_LDAP", "ACTIVE_DIRECTORY"}, false),
						},
						"base_distinguished_name": { //SearchBase
							Type:        schema.TypeString,
							Optional:    true,
							Description: "LDAP search base",
						},
						"is_ssl": { // IsSsl
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "True if the LDAP service requires an SSL connection",
						},
						"username": { // Username
							Type:     schema.TypeString,
							Optional: true,
							Description: `Username to use when logging in to LDAP, specified using LDAP attribute=value ` +
								`pairs (for example: cn="ldap-admin", c="example", dc="com")`,
						},
						"password": { // Password
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							Description: `Password for the user identified by UserName. This value is never returned by GET. ` +
								`It is inspected on create and modify. ` +
								`On modify, the absence of this element indicates that the password should not be changed`,
						},
						"custom_ui_button_label": { // CustomUiButtonLabel
							Type:        schema.TypeString,
							Optional:    true,
							Description: "If you provide a custom button label, on the login screen, the custom label replaces the default label for this identity provider",
						},
						"user_attributes":  ldapUserAttributes(false),  // CustomOrgLdapSettings.UserAttributes
						"group_attributes": ldapGroupAttributes(false), // CustomOrgLdapSettings.GroupAttributes
					},
				},
			},
		},
	}
}

func resourceVcfaOrgLdapCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcfaOrgLdapCreateOrUpdate(ctx, d, meta, "resource")
}
func resourceVcfaOrgLdapRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaOrgLdapRead(ctx, d, meta, "resource", nil)
}
func resourceVcfaOrgLdapUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcfaOrgLdapCreateOrUpdate(ctx, d, meta, "resource")
}

func resourceVcfaOrgLdapCreateOrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	// Lock the Organization to serialize create/update operation and prevent side effects like bricked Organizations when
	// vcfa_org_settings is updated at the same time
	orgId := d.Get("org_id").(string)
	vcfa.kvLock(orgId)
	defer vcfa.kvUnlock(orgId)

	settings, err := fillOrgLdapSettings(d)
	if err != nil {
		return diag.Errorf("[Org LDAP %s] error collecting settings values: %s", origin, err)
	}

	tmClient := meta.(ClientContainer).tmClient
	org, err := tmClient.GetTmOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Org LDAP %s] error searching for Org %s: %s", origin, orgId, err)
	}

	_, err = org.LdapConfigure(settings, d.Get("auto_trust_certificate").(bool))
	if err != nil {
		return diag.Errorf("[Org LDAP %s] error setting org '%s' LDAP configuration: %s", origin, orgId, err)
	}

	return genericVcfaOrgLdapRead(ctx, d, meta, origin, settings)
}

func genericVcfaOrgLdapRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string, settings *types.OrgLdapSettingsType) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	orgId := d.Get("org_id").(string)

	tmOrg, err := tmClient.GetTmOrgById(orgId)
	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] unable to find Organization %s LDAP settings: %s. Removing from state", orgId, err)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("unable to find organization %s: %s", orgId, err)
	}

	config, err := tmOrg.GetLdapConfiguration()
	if err != nil {
		d.SetId("")
		return diag.Errorf("[Org LDAP read %s] error getting LDAP settings for Org %s: %s", origin, orgId, err)
	}

	dSet(d, "org_id", orgId)
	dSet(d, "ldap_mode", config.OrgLdapMode)
	dSet(d, "custom_user_ou", config.CustomUsersOu)
	d.SetId(tmOrg.TmOrg.ID)

	if config.OrgLdapMode == "CUSTOM" {
		customSettings := map[string]interface{}{
			"server":                  config.CustomOrgLdapSettings.HostName,
			"port":                    config.CustomOrgLdapSettings.Port,
			"connector_type":          config.CustomOrgLdapSettings.ConnectorType,
			"base_distinguished_name": config.CustomOrgLdapSettings.SearchBase,
			"is_ssl":                  config.CustomOrgLdapSettings.IsSsl,
			"username":                config.CustomOrgLdapSettings.Username,
			"user_attributes": []map[string]interface{}{
				{
					"object_class":                config.CustomOrgLdapSettings.UserAttributes.ObjectClass,
					"unique_identifier":           config.CustomOrgLdapSettings.UserAttributes.ObjectIdentifier,
					"username":                    config.CustomOrgLdapSettings.UserAttributes.Username,
					"email":                       config.CustomOrgLdapSettings.UserAttributes.Email,
					"display_name":                config.CustomOrgLdapSettings.UserAttributes.FullName,
					"given_name":                  config.CustomOrgLdapSettings.UserAttributes.GivenName,
					"surname":                     config.CustomOrgLdapSettings.UserAttributes.Surname,
					"telephone":                   config.CustomOrgLdapSettings.UserAttributes.Telephone,
					"group_membership_identifier": config.CustomOrgLdapSettings.UserAttributes.GroupMembershipIdentifier,
					"group_back_link_identifier":  config.CustomOrgLdapSettings.UserAttributes.GroupBackLinkIdentifier,
				},
			},
			"group_attributes": []map[string]interface{}{
				{
					"object_class":                config.CustomOrgLdapSettings.GroupAttributes.ObjectClass,
					"unique_identifier":           config.CustomOrgLdapSettings.GroupAttributes.ObjectIdentifier,
					"name":                        config.CustomOrgLdapSettings.GroupAttributes.GroupName,
					"membership":                  config.CustomOrgLdapSettings.GroupAttributes.Membership,
					"group_membership_identifier": config.CustomOrgLdapSettings.GroupAttributes.MembershipIdentifier,
					"group_back_link_identifier":  config.CustomOrgLdapSettings.GroupAttributes.BackLinkIdentifier,
				},
			},
		}

		if config.CustomOrgLdapSettings.CustomUiButtonLabel != nil && *config.CustomOrgLdapSettings.CustomUiButtonLabel != "" {
			customSettings["custom_ui_button_label"] = *config.CustomOrgLdapSettings.CustomUiButtonLabel
		}

		if origin == "resource" {
			if settings != nil && settings.CustomOrgLdapSettings != nil && settings.CustomOrgLdapSettings.Password != "" {
				// This will be true on Create and Update, as we pass the original LDAP settings as parameter to this function.
				// This way we can save the original password that the user set on create or update.
				customSettings["password"] = settings.CustomOrgLdapSettings.Password
			} else {
				// This happens on Reads. We don't have original settings (Read operations pass 'settings' parameter as nil).
				// In this case, we recover the password from state.
				oldSettings := d.Get("custom_settings").([]interface{})
				if len(oldSettings) > 0 {
					customSettings["password"] = oldSettings[0].(map[string]interface{})["password"]
				}
			}
		}

		err = d.Set("custom_settings", []map[string]interface{}{customSettings})
		if err != nil {
			return diag.Errorf("[Org LDAP read %s] error setting 'user_attributes' field: %s", origin, err)
		}
	}
	return nil
}

func resourceVcfaOrgLdapDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Lock the Organization to serialize delete operation and prevent side effects like bricked Organizations when
	// vcfa_org_settings is deleted at the same time
	orgId := d.Get("org_id").(string)
	vcfa.kvLock(orgId)
	defer vcfa.kvUnlock(orgId)

	tmClient := meta.(ClientContainer).tmClient

	tmOrg, err := tmClient.GetTmOrgById(orgId)
	if err != nil {
		return diag.Errorf("[Org LDAP delete] error searching for Org %s: %s", orgId, err)
	}
	return diag.FromErr(tmOrg.LdapDisable())
}

func fillOrgLdapSettings(d *schema.ResourceData) (*types.OrgLdapSettingsType, error) {
	settings := types.OrgLdapSettingsType{
		OrgLdapMode: d.Get("ldap_mode").(string),
	}

	if settings.OrgLdapMode == "SYSTEM" {
		settings.CustomUsersOu = d.Get("custom_user_ou").(string)
		return &settings, nil
	}

	if settings.OrgLdapMode != "CUSTOM" {
		return &settings, nil
	}
	customSettings := d.Get("custom_settings")
	if customSettings == nil {
		return nil, fmt.Errorf("custom_settings are empty with CUSTOM ldap_mode")
	}
	customSettingsList, ok := customSettings.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid custom settings: expected []interface{}")
	}
	customSettingsMap, ok := customSettingsList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid custom settings: expected map[string]interface{}")
	}

	settings.CustomOrgLdapSettings = &types.CustomOrgLdapSettings{
		HostName:                customSettingsMap["server"].(string),
		Port:                    customSettingsMap["port"].(int),
		IsSsl:                   customSettingsMap["is_ssl"].(bool),
		SearchBase:              customSettingsMap["base_distinguished_name"].(string),
		Username:                customSettingsMap["username"].(string),
		Password:                customSettingsMap["password"].(string),
		AuthenticationMechanism: "SIMPLE", // Only SIMPLE is allowed in UI
		ConnectorType:           customSettingsMap["connector_type"].(string),
	}

	rawUserAttributesList, okUserList := customSettingsMap["user_attributes"].([]interface{})
	rawGroupAttributesList, okGroupList := customSettingsMap["group_attributes"].([]interface{})
	if !okUserList || len(rawUserAttributesList) == 0 {
		return nil, fmt.Errorf("user_attributes settings are empty with CUSTOM ldap_mode")
	}
	if !okGroupList || len(rawGroupAttributesList) == 0 {
		return nil, fmt.Errorf("group_attributes settings are empty with CUSTOM ldap_mode")
	}
	userAttributesMap, okUser := rawUserAttributesList[0].(map[string]interface{})
	groupAttributesMap, okGroup := rawGroupAttributesList[0].(map[string]interface{})
	if !okUser || userAttributesMap == nil || len(userAttributesMap) == 0 {
		return nil, fmt.Errorf("user_attributes settings are empty with CUSTOM ldap_mode")
	}
	if !okGroup || groupAttributesMap == nil || len(groupAttributesMap) == 0 {
		return nil, fmt.Errorf("group_attributes settings are empty with CUSTOM ldap_mode")
	}
	settings.CustomOrgLdapSettings.UserAttributes = &types.OrgLdapUserAttributes{
		ObjectClass:               userAttributesMap["object_class"].(string),
		ObjectIdentifier:          userAttributesMap["unique_identifier"].(string),
		Username:                  userAttributesMap["username"].(string),
		Email:                     userAttributesMap["email"].(string),
		FullName:                  userAttributesMap["display_name"].(string),
		GivenName:                 userAttributesMap["given_name"].(string),
		Surname:                   userAttributesMap["surname"].(string),
		Telephone:                 userAttributesMap["telephone"].(string),
		GroupMembershipIdentifier: userAttributesMap["group_membership_identifier"].(string),
		GroupBackLinkIdentifier:   userAttributesMap["group_back_link_identifier"].(string),
	}
	settings.CustomOrgLdapSettings.GroupAttributes = &types.OrgLdapGroupAttributes{
		ObjectClass:          groupAttributesMap["object_class"].(string),
		ObjectIdentifier:     groupAttributesMap["unique_identifier"].(string),
		GroupName:            groupAttributesMap["name"].(string),
		Membership:           groupAttributesMap["membership"].(string),
		MembershipIdentifier: groupAttributesMap["group_membership_identifier"].(string),
		BackLinkIdentifier:   groupAttributesMap["group_back_link_identifier"].(string),
	}

	if uiLabel, ok := customSettingsMap["custom_ui_button_label"]; ok && uiLabel != "" {
		settings.CustomOrgLdapSettings.CustomUiButtonLabel = addrOf(uiLabel.(string))
	}

	return &settings, nil
}

// resourceVcfaOrgLdapImport is responsible for importing the resource.
// The d.ID() field as being passed from `terraform import _resource_name_ _the_id_string_ requires
// a name based dot-formatted path to the object to lookup the object and sets the id of object.
// `terraform import` automatically performs `refresh` operation which loads up all other fields.
// For this resource, the import path is just the org name.
//
// Example import path (id): orgName
func resourceVcfaOrgLdapImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	orgName := d.Id()

	tmClient := meta.(ClientContainer).tmClient
	tmOrg, err := tmClient.GetTmOrgByName(orgName)
	if err != nil {
		return nil, err
	}

	dSet(d, "org_id", tmOrg.TmOrg.ID)

	d.SetId(tmOrg.TmOrg.ID)
	return []*schema.ResourceData{d}, nil
}
