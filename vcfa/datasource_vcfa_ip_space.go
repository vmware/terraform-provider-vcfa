package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

var dsIpSpaceInternalScopeSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("ID of internal scope within %s", labelVcfaIpSpace),
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("Name of internal scope within %s", labelVcfaIpSpace),
		},
		"cidr": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("The CIDR that represents this IP block within %s", labelVcfaIpSpace),
		},
	},
}

func datasourceVcfaIpSpace() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaIpSpaceRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaIpSpace),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Region ID for this %s", labelVcfaIpSpace),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaIpSpace),
			},
			"external_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "External scope in CIDR format",
			},
			"default_quota_max_subnet_size": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Maximum subnet size represented as a prefix length (e.g. 24, 28) in %s", labelVcfaIpSpace),
			},
			"default_quota_max_cidr_count": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Maximum number of subnets that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelVcfaIpSpace),
			},
			"default_quota_max_ip_count": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Maximum number of single floating IP addresses that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelVcfaIpSpace),
			},
			"internal_scope": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("Internal scope of %s", labelVcfaIpSpace),
				Elem:        dsIpSpaceInternalScopeSchema,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaIpSpace),
			},
		},
	}
}

func datasourceVcfaIpSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(MetaContainer).VcfaClient

	getIpSpaceByName := func(name string) (*govcd.TmIpSpace, error) {
		return vcfaClient.GetTmIpSpaceByNameAndRegionId(name, d.Get("region_id").(string))
	}

	c := dsReadConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:    labelVcfaIpSpace,
		getEntityFunc:  getIpSpaceByName,
		stateStoreFunc: setIpSpaceData,
	}
	return readDatasource(ctx, d, meta, c)
}
