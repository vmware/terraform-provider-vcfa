package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaProviderGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaProviderGatewayRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaProviderGateway),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelVcfaRegion, labelVcfaProviderGateway),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaProviderGateway),
			},
			"nsxt_tier0_gateway_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelVcfaTier0Gateway, labelVcfaProviderGateway),
			},
			"ip_space_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("A set of supervisor IDs used in this %s", labelVcfaRegion),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaProviderGateway),
			},
		},
	}
}

func datasourceVcfaProviderGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	getProviderGateway := func(name string) (*govcd.TmProviderGateway, error) {
		return vcdClient.GetTmProviderGatewayByNameAndRegionId(name, d.Get("region_id").(string))
	}
	c := dsReadConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:    labelVcfaProviderGateway,
		getEntityFunc:  getProviderGateway,
		stateStoreFunc: setProviderGatewayData,
	}
	return readDatasource(ctx, d, meta, c)
}
