package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaNsxManager() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaNsxManagerRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaNsxManager),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaNsxManager),
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Username for authenticating to %s", labelVcfaNsxManager),
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("URL of %s", labelVcfaNsxManager),
			},
			"network_provider_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Network Provider Scope for %s", labelVcfaNsxManager),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaNsxManager),
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("HREF of %s", labelVcfaNsxManager),
			},
		},
	}
}

func datasourceVcfaNsxManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).VcfaClient
	c := dsReadConfig[*govcd.NsxtManagerOpenApi, types.NsxtManagerOpenApi]{
		entityLabel:    labelVcfaNsxManager,
		getEntityFunc:  vcfaClient.GetNsxtManagerOpenApiByName,
		stateStoreFunc: setNsxManagerData,
	}
	return readDatasource(ctx, d, meta, c)
}
