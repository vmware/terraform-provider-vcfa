package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaOrgRegionalNetworking() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaOrgRegionalNetworkingRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaRegionalNetworkingSetting),
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaOrg, labelVcfaRegionalNetworkingSetting),
			},
			"provider_gateway_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaProviderGateway, labelVcfaRegionalNetworkingSetting),
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaRegion, labelVcfaRegionalNetworkingSetting),
			},
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Backing %s ID for %s. Will be autoselected if not specified.", labelVcfaEdgeCluster, labelVcfaRegionalNetworkingSetting),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaRegionalNetworkingSetting),
			},
		},
	}
}

func datasourceVcfaOrgRegionalNetworkingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	getTmRegionalNetworkingSettingByName := func(name string) (*govcd.TmRegionalNetworkingSetting, error) {
		return tmClient.GetTmRegionalNetworkingSettingByNameAndOrgId(name, d.Get("org_id").(string))
	}

	c := dsReadConfig[*govcd.TmRegionalNetworkingSetting, types.TmRegionalNetworkingSetting]{
		entityLabel:    labelVcfaRegionalNetworkingSetting,
		getEntityFunc:  getTmRegionalNetworkingSettingByName,
		stateStoreFunc: setTmRegionalNetworkingSettingData,
	}
	return readDatasource(ctx, d, meta, c)
}
