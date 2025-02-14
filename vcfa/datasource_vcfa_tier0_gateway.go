package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaTier0Gateway = "Tier 0 Gateway"

func datasourceVcfaTier0Gateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaTier0GatewayRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Display Name of %s", labelVcfaTier0Gateway),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s ID", labelVcfaRegion),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaTier0Gateway),
			},
			"parent_tier_0_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelVcfaTier0Gateway, labelVcfaTier0Gateway),
			},
			"already_imported": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines if the T0 is already imported of %s", labelVcfaTier0Gateway),
			},
		},
	}
}

func datasourceVcfaTier0GatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	// Fetching the region to conform to standard of returning 'ErrorEntityNotFound', because the API behind
	// GetTmTier0GatewayWithContextByName does not handle it well
	region, err := tmClient.GetRegionById(d.Get("region_id").(string))
	if err != nil {
		return diag.Errorf("no region with ID '%s' found: %s", d.Get("region_id").(string), err)
	}

	getT0ByName := func(name string) (*govcd.TmTier0Gateway, error) {
		return tmClient.GetTmTier0GatewayWithContextByName(name, region.Region.ID, true)
	}

	c := dsReadConfig[*govcd.TmTier0Gateway, types.TmTier0Gateway]{
		entityLabel:    labelVcfaTier0Gateway,
		getEntityFunc:  getT0ByName,
		stateStoreFunc: setTier0GatewayData,
	}
	return readDatasource(ctx, d, meta, c)
}

func setTier0GatewayData(_ *VCDClient, d *schema.ResourceData, t *govcd.TmTier0Gateway) error {
	d.SetId(t.TmTier0Gateway.ID) // So far the API returns plain UUID (not URN)
	dSet(d, "name", t.TmTier0Gateway.DisplayName)
	dSet(d, "description", t.TmTier0Gateway.Description)
	dSet(d, "parent_tier_0_id", t.TmTier0Gateway.ParentTier0ID)
	dSet(d, "already_imported", t.TmTier0Gateway.AlreadyImported)

	return nil
}
