package vcfa

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The name of the %s", labelVcfaContentLibrary),
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, // If not configured, Organization ID is retrieved and saved
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
					"automatically attached to all current and future namespaces in the tenant organization. If no value is "+
					"supplied during creation then this field will default to true. If a value of false is supplied, "+
					"then this Tenant %s will only be attached to namespaces that explicitly request it. "+
					"For Provider Content Libraries this field is not needed for creation and will always be returned as true. "+
					"This field cannot be updated after %s creation", labelVcfaContentLibrary, labelVcfaContentLibrary, labelVcfaContentLibrary),
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
				ForceNew:    true, // Can't change subscription settings
				Description: fmt.Sprintf("A block representing subscription settings of a %s", labelVcfaContentLibrary),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subscription_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: fmt.Sprintf("Subscription url of this %s", labelVcfaContentLibrary),
						},
						"password": {
							Type:        schema.TypeString,
							Optional:    true, // Required at Runtime as cannot be Required + Computed in schema. (It is computed as password cannot be recovered)
							Computed:    true,
							Description: "Password to use to authenticate with the publisher",
						},
						"need_local_copy": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to eagerly download content from publisher and store it locally",
						},
					},
				},
			},
			"version_number": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("Version number of this %s", labelVcfaContentLibrary),
			},
		},
	}
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
	err = setContentLibraryData(tmClient, d, cl)
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

	err = setContentLibraryData(tmClient, d, cl)
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
	if len(idSplit) > 2 {
		return nil, fmt.Errorf("invalid import identifier '%s', should be either <%s name>, or <%s name>%s<%s name>", labelVcfaContentLibrary, labelVcfaOrg, labelVcfaContentLibrary, d.Id(), ImportSeparator)
	}
	var cl *govcd.ContentLibrary
	var org *govcd.TmOrg
	var err error
	if len(idSplit) == 1 {
		// Nor Organization specified, meaning that is a PROVIDER Content Library
		cl, err = tmClient.GetContentLibraryByName(idSplit[0], nil)
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
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		AutoAttach:     d.Get("auto_attach").(bool),
		StorageClasses: convertSliceOfStringsToOpenApiReferenceIds(convertTypeListToSliceOfStrings(d.Get("storage_class_ids").(*schema.Set).List())),
	}
	if v, ok := d.GetOk("subscription_config"); ok {
		subsConfig := v.([]interface{})[0].(map[string]interface{})
		t.SubscriptionConfig = &types.ContentLibrarySubscriptionConfig{
			SubscriptionUrl: subsConfig["subscription_url"].(string),
			NeedLocalCopy:   subsConfig["need_local_copy"].(bool),
			Password:        subsConfig["password"].(string),
		}
	}
	return t
}

func setContentLibraryData(_ *VCDClient, d *schema.ResourceData, cl *govcd.ContentLibrary) error {
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
	if cl.ContentLibrary.Org != nil {
		dSet(d, "org_id", cl.ContentLibrary.Org.ID)
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
				"password":         cl.ContentLibrary.SubscriptionConfig.Password,
				"need_local_copy":  cl.ContentLibrary.SubscriptionConfig.NeedLocalCopy,
			},
		}
	}
	err = d.Set("subscription_config", subscriptionConfig)
	if err != nil {
		return err
	}

	d.SetId(cl.ContentLibrary.ID)
	return nil
}
