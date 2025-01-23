package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcfaContentLibraryItem() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaContentLibraryItemRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaContentLibraryItem),
			},
			"content_library_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("ID of the Content Library that this %s belongs to", labelVcfaContentLibraryItem),
			},
			"creation_date": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The ISO-8601 timestamp representing when this %s was created", labelVcfaContentLibraryItem),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The description of the %s", labelVcfaContentLibraryItem),
			},
			"image_identifier": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Virtual Machine Identifier (VMI) of the %s. This is a ReadOnly field", labelVcfaContentLibraryItem),
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
				Type: schema.TypeString,
				// TODO: TM: This should be optional: Either Provider or Tenant can create CLs
				Computed:    true,
				Description: fmt.Sprintf("The reference to the organization that the %s belongs to", labelVcfaContentLibraryItem),
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

func datasourceVcfaContentLibraryItemRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// TODO: TM: Tenant Context should not be nil and depend on the configured owner_org_id
	cl, err := vcdClient.GetContentLibraryById(d.Get("content_library_id").(string), nil)
	if err != nil {
		return diag.Errorf("error retrieving Content Library: %s", err)
	}

	cli, err := cl.GetContentLibraryItemByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Content Library Item: %s", err)
	}

	err = setContentLibraryItemData(vcdClient, d, cli)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
