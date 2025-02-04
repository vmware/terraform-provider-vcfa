package vcfa

import (
	"context"
	"fmt"
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
		// TODO: TM: Update not supported yet
		// UpdateContext: resourceVcfaContentLibraryItemUpdate,
		DeleteContext: resourceVcfaContentLibraryItemDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaContentLibraryItemImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // TODO: TM: Update not supported yet
				Description: fmt.Sprintf("Name of the %s", labelVcfaContentLibraryItem),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true, // TODO: TM: Update not supported yet
				Description: fmt.Sprintf("The description of the %s", labelVcfaContentLibraryItem),
			},
			"content_library_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("ID of the %s that this %s belongs to", labelVcfaContentLibrary, labelVcfaContentLibraryItem),
			},
			"file_path": {
				Type:        schema.TypeString,
				Optional:    true, // Not needed when Importing
				ForceNew:    true, // TODO: TM: Update not supported yet
				Description: fmt.Sprintf("Path to the OVA/ISO to create the %s", labelVcfaContentLibraryItem),
			},
			"upload_piece_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true, // TODO: TM: Update not supported yet
				Default:     1,
				Description: fmt.Sprintf("When uploading the %s, this argument defines the size of the file chunks in which it is split on every upload request. It can possibly impact upload performance. Default 1 MB", labelVcfaContentLibraryItem),
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The ISO-8601 timestamp representing when this %s was created", labelVcfaContentLibraryItem),
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
	vcfaClient := meta.(ClientContainer).VcfaClient

	clId := d.Get("content_library_id").(string)
	// TODO: TM: Tenant Context should not be nil and depend on the configured owner_org_id
	cl, err := vcfaClient.GetContentLibraryById(clId, nil)
	if err != nil {
		return diag.Errorf("could not retrieve %s with ID '%s': %s", labelVcfaContentLibrary, clId, err)
	}

	filePath := d.Get("file_path").(string)
	uploadPieceSize := d.Get("upload_piece_size").(int)

	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
		entityLabel:    labelVcfaContentLibraryItem,
		getTypeFunc:    getContentLibraryItemType,
		stateStoreFunc: setContentLibraryItemData,
		createFunc: func(config *types.ContentLibraryItem) (*govcd.ContentLibraryItem, error) {
			return cl.CreateContentLibraryItem(config, govcd.ContentLibraryItemUploadArguments{
				FilePath:        filePath,
				UploadPieceSize: int64(uploadPieceSize) * 1024 * 1024,
			})
		},
		resourceReadFunc: resourceVcfaContentLibraryItemRead,
	}
	return createResource(ctx, d, meta, c)
}

//func resourceVcfaContentLibraryItemUpdate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
//	// TODO: TM: Update is not supported yet
//	return diag.Errorf("update not supported")
//	vcfaClient := meta.(MetaContainer).VcfaClient
//
//	clId := d.Get("content_library_id").(string)
//	cl, err := vcfaClient.GetContentLibraryById(clId)
//	if err != nil {
//		return diag.Errorf("could not retrieve Content Library with ID '%s': %s", clId, err)
//	}
//
//	c := crudConfig[*govcd.ContentLibraryItem, types.ContentLibraryItem]{
//		entityLabel:      labelVcfaContentLibraryItem,
//		getTypeFunc:      getContentLibraryItemType,
//		getEntityFunc:    cl.GetContentLibraryItemById,
//		resourceReadFunc: resourceVcfaContentLibraryItemRead,
//	}
//
//	return updateResource(ctx, d, meta, c)
//}

func resourceVcfaContentLibraryItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).VcfaClient

	clId := d.Get("content_library_id").(string)
	// TODO: TM: Tenant Context should not be nil and depend on the configured owner_org_id
	cl, err := vcfaClient.GetContentLibraryById(clId, nil)
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
	vcfaClient := meta.(ClientContainer).VcfaClient

	clId := d.Get("content_library_id").(string)
	// TODO: TM: Tenant Context should not be nil and depend on the configured owner_org_id
	cl, err := vcfaClient.GetContentLibraryById(clId, nil)
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
	vcfaClient := meta.(ClientContainer).VcfaClient

	id := strings.Split(d.Id(), ImportSeparator)
	if len(id) != 2 {
		return nil, fmt.Errorf("ID syntax should be <%s name>%s<%s name>", labelVcfaContentLibrary, ImportSeparator, labelVcfaContentLibraryItem)
	}

	// TODO: TM: Tenant Context should not be nil and depend on the configured owner_org_id
	cl, err := vcfaClient.GetContentLibraryByName(id[0], nil)
	if err != nil {
		return nil, fmt.Errorf("error getting %s with name '%s' for import: %s", labelVcfaContentLibrary, id[0], err)
	}

	cli, err := cl.GetContentLibraryItemByName(id[1])
	if err != nil {
		return nil, fmt.Errorf("error getting %s with name '%s': %s", labelVcfaContentLibraryItem, id[1], err)
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

	return t, nil
}

func setContentLibraryItemData(_ *VCDClient, d *schema.ResourceData, cli *govcd.ContentLibraryItem) error {
	if cli == nil || cli.ContentLibraryItem == nil {
		return fmt.Errorf("cannot save state for nil %s", labelVcfaContentLibraryItem)
	}

	dSet(d, "content_library_id", cli.ContentLibraryItem.ContentLibrary.ID)
	dSet(d, "name", cli.ContentLibraryItem.Name)
	dSet(d, "description", cli.ContentLibraryItem.Description)
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
