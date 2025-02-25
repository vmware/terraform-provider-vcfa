package vcfa

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaContentLibraryItem = "Content Library Item"

func resourceVcfaContentLibraryItem() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaContentLibraryItemCreate,
		ReadContext:   resourceVcfaContentLibraryItemRead,
		UpdateContext: resourceVcfaContentLibraryItemUpdate,
		DeleteContext: resourceVcfaContentLibraryItemDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaContentLibraryItemImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaContentLibraryItem),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("The description of the %s", labelVcfaContentLibraryItem),
			},
			"content_library_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("ID of the %s that this %s belongs to", labelVcfaContentLibrary, labelVcfaContentLibraryItem),
			},
			"file_paths": {
				Type:        schema.TypeSet,
				Optional:    true, // Not needed when Importing
				ForceNew:    true,
				Description: fmt.Sprintf("A single path to an OVA/ISO, or multiple paths for an OVF and its referenced files, to create the %s", labelVcfaContentLibraryItem),
			},
			"upload_piece_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: fmt.Sprintf("When uploading the %s, this argument defines the size of the file chunks in which it is split on every upload request. It can possibly impact upload performance. Default 1 MB", labelVcfaContentLibraryItem),
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The ISO-8601 timestamp representing when this %s was created", labelVcfaContentLibraryItem),
			},
			"item_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The type of %s", labelVcfaContentLibraryItem),
			},
			"image_identifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Virtual Machine Identifier (VMI) of the %s. This is a read only field", labelVcfaContentLibraryItem),
			},
			"is_published": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is published", labelVcfaContentLibraryItem),
			},
			"is_subscribed": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is subscribed", labelVcfaContentLibraryItem),
			},
			"last_successful_sync": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The ISO-8601 timestamp representing when this %s was last synced if subscribed", labelVcfaContentLibraryItem),
			},
			"owner_org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The reference to the %s that the %s belongs to", labelVcfaOrg, labelVcfaContentLibraryItem),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of this %s", labelVcfaContentLibraryItem),
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: fmt.Sprintf("The version of this %s. For a subscribed library, this version is same as in publisher library", labelVcfaContentLibraryItem),
			},
		},
	}
}

func resourceVcfaContentLibraryItemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	clId := d.Get("content_library_id").(string)
	cl, err := tmClient.GetContentLibraryById(clId, nil)
	if err != nil {
		return diag.Errorf("could not retrieve %s with ID '%s': %s", labelVcfaContentLibrary, clId, err)
	}

	if _, ok := d.GetOk("file_paths"); !ok {
		return diag.Errorf("the argument 'file_paths' is required during creation")
	}

	uploadArgs := govcd.ContentLibraryItemUploadArguments{
		UploadPieceSize: int64(d.Get("upload_piece_size").(int)) * 1024 * 1024,
	}

	filePaths := d.Get("file_paths").(*schema.Set).List()
	if len(filePaths) == 1 {
		// ISO/OVA
		uploadArgs.FilePath = filePaths[0].(string)
	} else {
		// OVF
		for _, p := range filePaths {
			if filepath.Ext(p.(string)) == ".ovf" {
				uploadArgs.FilePath = p.(string)
			} else {
				uploadArgs.OvfFilesPaths = append(uploadArgs.OvfFilesPaths, p.(string))
			}
		}
	}

	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
		entityLabel:    labelVcfaContentLibraryItem,
		getTypeFunc:    getContentLibraryItemType,
		stateStoreFunc: setContentLibraryItemData,
		createFunc: func(config *types.ContentLibraryItem) (*govcd.ContentLibraryItem, error) {
			return cl.CreateContentLibraryItem(config, uploadArgs)
		},
		resourceReadFunc: resourceVcfaContentLibraryItemRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaContentLibraryItemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	clId := d.Get("content_library_id").(string)
	cl, err := tmClient.GetContentLibraryById(clId, nil)
	if err != nil {
		return diag.Errorf("could not retrieve Content Library with ID '%s': %s", clId, err)
	}

	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
		entityLabel:      labelVcfaContentLibraryItem,
		getTypeFunc:      getContentLibraryItemType,
		getEntityFunc:    cl.GetContentLibraryItemById,
		resourceReadFunc: resourceVcfaContentLibraryItemRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaContentLibraryItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	clId := d.Get("content_library_id").(string)
	cl, err := tmClient.GetContentLibraryById(clId, nil)
	if err != nil {
		return diag.Errorf("could not retrieve %s with ID '%s': %s", labelVcfaContentLibrary, clId, err)
	}

	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
		entityLabel:    labelVcfaContentLibraryItem,
		getEntityFunc:  cl.GetContentLibraryItemById,
		stateStoreFunc: setContentLibraryItemData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaContentLibraryItemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	clId := d.Get("content_library_id").(string)
	cl, err := tmClient.GetContentLibraryById(clId, nil)
	if err != nil {
		return diag.Errorf("could not retrieve %s with ID '%s': %s", labelVcfaContentLibrary, clId, err)
	}

	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
		entityLabel:   labelVcfaContentLibraryItem,
		getEntityFunc: cl.GetContentLibraryItemById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaContentLibraryItemImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	id := strings.Split(d.Id(), ImportSeparator)
	var tenantContext *govcd.TenantContext
	clName, cliName := "", ""
	switch len(id) {
	case 3:
		org, err := tmClient.GetTmOrgByName(id[0])
		if err != nil {
			return nil, fmt.Errorf("error getting %s with name '%s' for import: %s", labelVcfaOrg, id[0], err)
		}
		tenantContext = &govcd.TenantContext{
			OrgId:   org.TmOrg.ID,
			OrgName: org.TmOrg.Name,
		}
		clName = id[1]
		cliName = id[2]
	case 2:
		clName = id[0]
		cliName = id[1]
	default:
		return nil, fmt.Errorf("ID syntax should be <%s name>%s<%s name>%s<%s name> or <%s name>%s<%s name>", labelVcfaOrg, ImportSeparator, labelVcfaContentLibrary, ImportSeparator, labelVcfaContentLibraryItem, labelVcfaContentLibrary, ImportSeparator, labelVcfaContentLibraryItem)
	}

	cl, err := tmClient.GetContentLibraryByName(clName, tenantContext)
	if err != nil {
		return nil, fmt.Errorf("error getting %s with name '%s' for import: %s", labelVcfaContentLibrary, clName, err)
	}

	cli, err := cl.GetContentLibraryItemByName(cliName)
	if err != nil {
		return nil, fmt.Errorf("error getting %s with name '%s': %s", labelVcfaContentLibraryItem, cliName, err)
	}

	d.SetId(cli.ContentLibraryItem.ID)
	dSet(d, "content_library_id", cl.ContentLibrary.ID)
	return []*schema.ResourceData{d}, nil
}

func getContentLibraryItemType(_ *VCDClient, d *schema.ResourceData) (*types.ContentLibraryItem, error) {
	t := &types.ContentLibraryItem{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	// This happens during updates. We need to send this value, otherwise the operation fails
	if itemType := d.Get("item_type"); itemType != nil {
		t.ItemType = itemType.(string)
	}

	return t, nil
}

func setContentLibraryItemData(_ *VCDClient, d *schema.ResourceData, cli *govcd.ContentLibraryItem) error {
	if cli == nil || cli.ContentLibraryItem == nil {
		return fmt.Errorf("cannot save state for nil %s", labelVcfaContentLibraryItem)
	}

	dSet(d, "content_library_id", cli.ContentLibraryItem.ContentLibrary.ID)
	dSet(d, "name", cli.ContentLibraryItem.Name)
	dSet(d, "description", cli.ContentLibraryItem.Description)
	dSet(d, "item_type", cli.ContentLibraryItem.ItemType)
	dSet(d, "creation_date", cli.ContentLibraryItem.CreationDate)
	dSet(d, "image_identifier", cli.ContentLibraryItem.ImageIdentifier)
	dSet(d, "is_published", cli.ContentLibraryItem.IsPublished)
	dSet(d, "is_subscribed", cli.ContentLibraryItem.IsSubscribed)
	dSet(d, "last_successful_sync", cli.ContentLibraryItem.LastSuccessfulSync)
	if cli.ContentLibraryItem.Org != nil {
		dSet(d, "owner_org_id", cli.ContentLibraryItem.Org.ID)
	}
	dSet(d, "status", cli.ContentLibraryItem.Status)
	dSet(d, "version", cli.ContentLibraryItem.Version)
	d.SetId(cli.ContentLibraryItem.ID)

	return nil
}
