// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaContentLibrary = "Content Library"

func resourceVcfaContentLibrary() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaContentLibraryCreate,
		ReadContext:   resourceVcfaContentLibraryRead,
		UpdateContext: resourceVcfaContentLibraryUpdate,
		DeleteContext: resourceVcfaContentLibraryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaContentLibraryImport,
		},
		CustomizeDiff: resourceVcfaContentLibraryCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The name of the %s", labelVcfaContentLibrary),
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Can't be changed after created
				Description: fmt.Sprintf("The reference to the %s that the %s belongs to", labelVcfaOrg, labelVcfaContentLibrary),
			},
			"delete_force": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("On deletion, forcefully deletes the %s and its %ss. Only for PROVIDER Content Libraries", labelVcfaContentLibrary, labelVcfaContentLibraryItem),
			},
			"delete_recursive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("On deletion, deletes the %s, including its %ss, in a single operation", labelVcfaContentLibrary, labelVcfaContentLibraryItem),
			},
			"storage_class_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: fmt.Sprintf("A set of %s IDs used by this %s", labelVcfaStorageClass, labelVcfaContentLibrary),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"auto_attach": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true, // Cannot be updated
				Description: fmt.Sprintf("For Tenant Content Libraries this field represents whether this %s should be "+
					"automatically attached to all current and future namespaces in the %s. If no value is "+
					"supplied during creation then this field will default to true. If a value of false is supplied, "+
					"then this Tenant %s will only be attached to namespaces that explicitly request it. "+
					"For Provider Content Libraries this field is not needed for creation and will always be returned as true. "+
					"This field cannot be updated after creation", labelVcfaContentLibrary, labelVcfaOrg, labelVcfaContentLibrary),
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The ISO-8601 timestamp representing when this %s was created", labelVcfaContentLibrary),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, // Subscribed libraries inherit publisher's description
				Description: fmt.Sprintf("The description of the %s", labelVcfaContentLibrary),
			},
			"is_shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is shared with other %ss", labelVcfaContentLibrary, labelVcfaOrg),
			},
			"is_subscribed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is subscribed from an external published library", labelVcfaContentLibrary),
			},
			"library_type": {
				Type:     schema.TypeString,
				Computed: true,
				Description: fmt.Sprintf("The type of %s, can be either PROVIDER (%s that is scoped to a "+
					"provider) or TENANT (%s that is scoped to a tenant organization)", labelVcfaContentLibrary, labelVcfaContentLibrary, labelVcfaContentLibrary),
			},
			"subscription_config": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: fmt.Sprintf("A block representing subscription settings of a %s", labelVcfaContentLibrary),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subscription_url": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true, // Can't change subscription url
							Description: fmt.Sprintf("Subscription url of this %s", labelVcfaContentLibrary),
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "Password to use to authenticate with the publisher",
						},
					},
				},
			},
			"version_number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Version number of this %s", labelVcfaContentLibrary),
			},
			"is_project_scoped": {
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Whether this %s is scoped to specific projects in the %s", labelVcfaContentLibrary, labelVcfaOrg),
			},
			"all_projects_permission": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf("Permissions to apply to all projects in the %s for this %s. Can be 'READ_ONLY' or 'READ_WRITE'. "+
					"Only applicable when 'is_project_scoped' is set to 'true'", labelVcfaOrg, labelVcfaContentLibrary),
				ValidateFunc: validation.StringInSlice([]string{"READ_ONLY", "READ_WRITE"}, false),
			},
			"project_permissions": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: fmt.Sprintf("A set of project permissions for this %s. Only applicable when 'is_project_scoped' is set to 'true'", labelVcfaContentLibrary),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permissions": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The type of project permission ('READ_ONLY' or 'READ_WRITE')",
							ValidateFunc: validation.StringInSlice([]string{"READ_ONLY", "READ_WRITE"}, false),
						},
						"project_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the project that this permission applies to",
						},
						"project_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the project that this permission applies to",
						},
					},
				},
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of this %s. Can be 'READY', 'NOT_READY', 'FAILED' or 'PARTIALLY_READY'", labelVcfaContentLibrary),
			},
		},
	}
}

func resourceVcfaContentLibraryCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if d.Get("is_project_scoped").(bool) {
		hasAllProjectsPermission := false
		if v, ok := d.GetOk("all_projects_permission"); ok {
			if s, ok := v.(string); ok && s != "" {
				hasAllProjectsPermission = true
			}
		}
		hasProjectPermissions := false
		if v, ok := d.GetOk("project_permissions"); ok {
			if set, ok := v.(*schema.Set); ok && set != nil && set.Len() > 0 {
				hasProjectPermissions = true
			}
		}
		if !hasAllProjectsPermission && !hasProjectPermissions {
			return fmt.Errorf("when %q is true, either %q or %q must be specified", "is_project_scoped", "all_projects_permission", "project_permissions")
		}
		return nil
	}
	if _, ok := d.GetOk("all_projects_permission"); ok {
		return fmt.Errorf("%q may only be set when %q is true", "all_projects_permission", "is_project_scoped")
	}
	if v, ok := d.GetOk("project_permissions"); ok {
		if set, ok := v.(*schema.Set); ok && set != nil && set.Len() > 0 {
			return fmt.Errorf("%q may only be set when %q is true", "project_permissions", "is_project_scoped")
		}
	}
	return nil
}

func resourceVcfaContentLibraryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	cl, err := tmClient.CreateContentLibrary(getContentLibraryType(d), tenantContext)
	if err != nil {
		return diag.FromErr(err)
	}
	err = setContentLibraryData(tmClient, d, cl, "resource")
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceVcfaContentLibraryRead(ctx, d, meta)
}

func resourceVcfaContentLibraryRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var cl *govcd.ContentLibrary
	idOrName := d.Id()
	if idOrName != "" {
		cl, err = tmClient.GetContentLibraryById(idOrName, tenantContext)
	} else {
		idOrName = d.Get("name").(string)
		cl, err = tmClient.GetContentLibraryByName(idOrName, tenantContext)
	}
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		log.Printf("[DEBUG] %s no longer exists. Removing from tfstate", labelVcfaContentLibrary)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	err = setContentLibraryData(tmClient, d, cl, "resource")
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcfaContentLibraryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	cl, err := tmClient.GetContentLibraryById(d.Id(), tenantContext)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = cl.Update(getContentLibraryType(d))
	if err != nil {
		return diag.FromErr(err)
	}
	return resourceVcfaContentLibraryRead(ctx, d, meta)
}

func resourceVcfaContentLibraryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	cl, err := tmClient.GetContentLibraryById(d.Id(), tenantContext)
	if err != nil {
		return diag.FromErr(err)
	}

	deleteForce := d.Get("delete_force").(bool)
	if cl.ContentLibrary.LibraryType != "PROVIDER" {
		deleteForce = false // Forcefully deletion is not available for non-PROVIDER Content Libraries
	}
	err = cl.Delete(deleteForce, d.Get("delete_recursive").(bool))
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcfaContentLibraryImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	idSplit := strings.Split(d.Id(), ImportSeparator)
	if len(idSplit) != 2 {
		return nil, fmt.Errorf("invalid import identifier '%s', should be <%s name>%s<%s name> for Tenant Content Libraries or System%s<%s name> for Provider Content Libraries", d.Id(), labelVcfaOrg, ImportSeparator, labelVcfaContentLibrary, ImportSeparator, labelVcfaContentLibrary)
	}
	var cl *govcd.ContentLibrary
	var org *govcd.TmOrg
	var err error
	if strings.EqualFold(idSplit[0], "system") {
		// Provider Content Library
		cl, err = tmClient.GetContentLibraryByName(idSplit[1], nil)
	} else {
		org, err = tmClient.GetTmOrgByName(idSplit[0])
		if err != nil {
			return nil, err
		}
		cl, err = tmClient.GetContentLibraryByName(idSplit[1], &govcd.TenantContext{
			OrgId:   org.TmOrg.ID,
			OrgName: org.TmOrg.Name,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s with identifier '%s': %s", labelVcfaContentLibrary, d.Id(), err)
	}

	d.SetId(cl.ContentLibrary.ID)
	dSet(d, "name", cl.ContentLibrary.Name)
	if cl.ContentLibrary.Org != nil {
		dSet(d, "org_id", cl.ContentLibrary.Org.ID)
	}
	return []*schema.ResourceData{d}, nil
}

func getContentLibraryType(d *schema.ResourceData) *types.ContentLibrary {
	t := &types.ContentLibrary{
		Name:                  d.Get("name").(string),
		Description:           d.Get("description").(string),
		AutoAttach:            d.Get("auto_attach").(bool),
		StorageClasses:        convertSliceOfStringsToOpenApiReferenceIds(convertTypeListToSliceOfStrings(d.Get("storage_class_ids").(*schema.Set).List())),
		IsProjectScoped:       d.Get("is_project_scoped").(bool),
		AllProjectsPermission: d.Get("all_projects_permission").(string),
	}
	if v, ok := d.GetOk("subscription_config"); ok {
		subsConfig := v.([]interface{})[0].(map[string]interface{})
		t.SubscriptionConfig = &types.ContentLibrarySubscriptionConfig{
			SubscriptionUrl: subsConfig["subscription_url"].(string),
			Password:        subsConfig["password"].(string),
		}
	}
	if v, ok := d.GetOk("project_permissions"); ok {
		ppSet := v.(*schema.Set).List()
		projectPermissions := make([]types.ContentLibraryProjectPermission, len(ppSet))
		for i, pp := range ppSet {
			ppMap := pp.(map[string]interface{})
			projectPermissions[i] = types.ContentLibraryProjectPermission{
				Permissions:       ppMap["permissions"].(string),
				ProjectAssignment: types.OpenApiReference{ID: ppMap["project_id"].(string)},
			}
		}
		t.ProjectPermissions = projectPermissions
	}
	return t
}

func setContentLibraryData(_ *VCDClient, d *schema.ResourceData, cl *govcd.ContentLibrary, origin string) error {
	if cl == nil || cl.ContentLibrary == nil {
		return fmt.Errorf("provided %s is nil", labelVcfaContentLibrary)
	}

	dSet(d, "name", cl.ContentLibrary.Name)
	dSet(d, "auto_attach", cl.ContentLibrary.AutoAttach)
	dSet(d, "creation_date", cl.ContentLibrary.CreationDate)
	dSet(d, "description", cl.ContentLibrary.Description)
	dSet(d, "is_shared", cl.ContentLibrary.IsShared)
	dSet(d, "is_subscribed", cl.ContentLibrary.IsSubscribed)
	dSet(d, "library_type", cl.ContentLibrary.LibraryType)
	dSet(d, "version_number", cl.ContentLibrary.VersionNumber)
	dSet(d, "is_project_scoped", cl.ContentLibrary.IsProjectScoped)
	dSet(d, "all_projects_permission", cl.ContentLibrary.AllProjectsPermission)
	dSet(d, "status", cl.ContentLibrary.Status)
	if cl.ContentLibrary.Org != nil {
		dSet(d, "org_id", cl.ContentLibrary.Org.ID)
	}

	projectPermissions := make([]map[string]interface{}, len(cl.ContentLibrary.ProjectPermissions))
	for i, pp := range cl.ContentLibrary.ProjectPermissions {
		projectPermissions[i] = map[string]interface{}{
			"permissions":  pp.Permissions,
			"project_id":   pp.ProjectAssignment.ID,
			"project_name": pp.ProjectAssignment.Name,
		}
	}
	if err := d.Set("project_permissions", projectPermissions); err != nil {
		return err
	}

	scs := make([]string, len(cl.ContentLibrary.StorageClasses))
	for i, sc := range cl.ContentLibrary.StorageClasses {
		scs[i] = sc.ID
	}
	err := d.Set("storage_class_ids", scs)
	if err != nil {
		return err
	}

	subscriptionConfig := make([]interface{}, 0)
	if cl.ContentLibrary.SubscriptionConfig != nil {
		subscriptionConfig = []interface{}{
			map[string]interface{}{
				"subscription_url": cl.ContentLibrary.SubscriptionConfig.SubscriptionUrl,
			},
		}
		// Password is only available in resource
		if origin == "resource" {
			// Password is never returned by backend. We save what we have currently
			if p := d.Get("subscription_config.0.password"); p != "" {
				subscriptionConfig[0].(map[string]interface{})["password"] = p
			}
		}
	}

	err = d.Set("subscription_config", subscriptionConfig)
	if err != nil {
		return err
	}

	d.SetId(cl.ContentLibrary.ID)
	return nil
}
