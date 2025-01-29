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

const labelVcfaRegionalNetworkingSetting = "Org Regional Networking"

func resourceVcfaOrgRegionalNetworking() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgRegionalNetworkingCreate,
		ReadContext:   resourceVcfaOrgRegionalNetworkingRead,
		UpdateContext: resourceVcfaOrgRegionalNetworkingUpdate,
		DeleteContext: resourceVcfaOrgRegionalNetworkingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgRegionalNetworkingImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaRegionalNetworkingSetting),
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaOrg, labelVcfaRegionalNetworkingSetting),
			},
			"provider_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaProviderGateway, labelVcfaRegionalNetworkingSetting),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaRegion, labelVcfaRegionalNetworkingSetting),
			},
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
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

func resourceVcfaOrgRegionalNetworkingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmRegionalNetworkingSetting, types.TmRegionalNetworkingSetting]{
		entityLabel:      labelVcfaRegionalNetworkingSetting,
		getTypeFunc:      getTmRegionalNetworkingSettingType,
		stateStoreFunc:   setTmRegionalNetworkingSettingData,
		createAsyncFunc:  vcfaClient.CreateTmRegionalNetworkingSettingAsync,
		getEntityFunc:    vcfaClient.GetTmRegionalNetworkingSettingById,
		resourceReadFunc: resourceVcfaOrgRegionalNetworkingRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionalNetworkingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmRegionalNetworkingSetting, types.TmRegionalNetworkingSetting]{
		entityLabel:      labelVcfaRegionalNetworkingSetting,
		getTypeFunc:      getTmRegionalNetworkingSettingType,
		getEntityFunc:    vcfaClient.GetTmRegionalNetworkingSettingById,
		resourceReadFunc: resourceVcfaOrgRegionalNetworkingRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionalNetworkingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmRegionalNetworkingSetting, types.TmRegionalNetworkingSetting]{
		entityLabel:    labelVcfaRegionalNetworkingSetting,
		getEntityFunc:  vcfaClient.GetTmRegionalNetworkingSettingById,
		stateStoreFunc: setTmRegionalNetworkingSettingData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionalNetworkingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmRegionalNetworkingSetting, types.TmRegionalNetworkingSetting]{
		entityLabel:   labelVcfaRegionalNetworkingSetting,
		getEntityFunc: vcfaClient.GetTmRegionalNetworkingSettingById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionalNetworkingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcfaClient := meta.(*VCDClient)

	id := strings.Split(d.Id(), ImportSeparator)
	if len(id) != 2 {
		return nil, fmt.Errorf("ID syntax should be <%s name>%s<%s name>", labelVcfaOrg, ImportSeparator, labelVcfaRegionalNetworkingSetting)
	}

	org, err := vcfaClient.GetTmOrgByName(id[0])
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s': %s", labelVcfaOrg, id[0], err)
	}

	rns, err := vcfaClient.GetTmRegionalNetworkingSettingByNameAndOrgId(id[1], org.TmOrg.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s' within %s '%s': %s",
			labelVcfaRegionalNetworkingSetting, id[1], labelVcfaOrg, id[0], err)
	}

	d.SetId(rns.TmRegionalNetworkingSetting.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmRegionalNetworkingSettingType(vcfaClient *VCDClient, d *schema.ResourceData) (*types.TmRegionalNetworkingSetting, error) {
	t := &types.TmRegionalNetworkingSetting{
		Name:               d.Get("name").(string),
		OrgRef:             types.OpenApiReference{ID: d.Get("org_id").(string)},
		RegionRef:          types.OpenApiReference{ID: d.Get("region_id").(string)},
		ProviderGatewayRef: types.OpenApiReference{ID: d.Get("provider_gateway_id").(string)},
	}

	edgeClusterId := d.Get("edge_cluster_id").(string)
	if edgeClusterId != "" { // Edge cluster will be picked automatically if not specified
		t.ServiceEdgeClusterRef = &types.OpenApiReference{ID: edgeClusterId}
	}

	return t, nil
}

func setTmRegionalNetworkingSettingData(_ *VCDClient, d *schema.ResourceData, cfg *govcd.TmRegionalNetworkingSetting) error {
	if cfg == nil || cfg.TmRegionalNetworkingSetting == nil {
		return fmt.Errorf("nil configuration received")
	}

	d.SetId(cfg.TmRegionalNetworkingSetting.ID)
	dSet(d, "name", cfg.TmRegionalNetworkingSetting.Name)
	dSet(d, "org_id", cfg.TmRegionalNetworkingSetting.OrgRef.ID)
	dSet(d, "region_id", cfg.TmRegionalNetworkingSetting.RegionRef.ID)
	dSet(d, "provider_gateway_id", cfg.TmRegionalNetworkingSetting.ProviderGatewayRef.ID)
	if cfg.TmRegionalNetworkingSetting.ServiceEdgeClusterRef != nil {
		dSet(d, "edge_cluster_id", cfg.TmRegionalNetworkingSetting.ServiceEdgeClusterRef.ID)
	} else {
		dSet(d, "edge_cluster_id", "")
	}
	dSet(d, "status", cfg.TmRegionalNetworkingSetting.Status)

	return nil
}
