package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcfaVcenter() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaVcenterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaVirtualCenter),
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("URL of %s", labelVcfaVirtualCenter),
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Username of %s", labelVcfaVirtualCenter),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Should the %s be enabled", labelVcfaVirtualCenter),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaVirtualCenter),
			},
			"has_proxy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s has proxy defined", labelVcfaVirtualCenter),
			},
			"is_connected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s is connected", labelVcfaVirtualCenter),
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelVcfaVirtualCenter),
			},
			"connection_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Listener state of %s", labelVcfaVirtualCenter),
			},
			"cluster_health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelVcfaVirtualCenter),
			},
			"vcenter_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Version of %s", labelVcfaVirtualCenter),
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s UUID", labelVcfaVirtualCenter),
			},
			"vcenter_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s hostname", labelVcfaVirtualCenter),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter status",
			},
		},
	}
}

func datasourceVcfaVcenterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	c := dsReadConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelVcfaVirtualCenter,
		getEntityFunc:  tmClient.GetVCenterByName,
		stateStoreFunc: setVcenterData,
	}
	return readDatasource(ctx, d, meta, c)
}
