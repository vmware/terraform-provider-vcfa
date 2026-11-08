// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

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
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"provider_gateway_id", "distributed_vlan_connection_id", "distributed_vxlan_connection_id"},
				Description:  fmt.Sprintf("Parent %s ID for %s. Mutually exclusive with 'distributed_vlan_connection_id' and 'distributed_vxlan_connection_id'", labelVcfaProviderGateway, labelVcfaRegionalNetworkingSetting),
			},
			"distributed_vlan_connection_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"provider_gateway_id", "distributed_vlan_connection_id", "distributed_vxlan_connection_id"},
				Description:  fmt.Sprintf("Distributed VLAN Connection ID used for egress by the %s, as an alternative to a %s", labelVcfaRegionalNetworkingSetting, labelVcfaProviderGateway),
			},
			"distributed_vxlan_connection_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"provider_gateway_id", "distributed_vlan_connection_id", "distributed_vxlan_connection_id"},
				Description:  fmt.Sprintf("Distributed VXLAN Connection ID used for egress by the %s, as an alternative to a %s", labelVcfaRegionalNetworkingSetting, labelVcfaProviderGateway),
			},
			"virtual_network_appliance_cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Virtual Network Appliance Cluster ID used with a distributed connection in the %s", labelVcfaRegionalNetworkingSetting),
			},
			"default_vpc_private_subnet_cidr_override": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Optional IPv4 CIDR override for the default VPC's private subnet. It can only be set on creation",
			},
			"project_tgw_private_subnet_cidr_override": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Optional IPv4 CIDR override for the project transit gateway's private subnet. It can only be set on creation",
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
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmRegionalNetworkingSetting, types.TmRegionalNetworkingSetting]{
		entityLabel:      labelVcfaRegionalNetworkingSetting,
		getTypeFunc:      getTmRegionalNetworkingSettingType,
		stateStoreFunc:   setTmRegionalNetworkingSettingData,
		createAsyncFunc:  tmClient.CreateTmRegionalNetworkingSettingAsync,
		getEntityFunc:    tmClient.GetTmRegionalNetworkingSettingById,
		resourceReadFunc: resourceVcfaOrgRegionalNetworkingRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionalNetworkingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmRegionalNetworkingSetting, types.TmRegionalNetworkingSetting]{
		entityLabel:      labelVcfaRegionalNetworkingSetting,
		getTypeFunc:      getTmRegionalNetworkingSettingType,
		getEntityFunc:    tmClient.GetTmRegionalNetworkingSettingById,
		resourceReadFunc: resourceVcfaOrgRegionalNetworkingRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionalNetworkingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmRegionalNetworkingSetting, types.TmRegionalNetworkingSetting]{
		entityLabel:    labelVcfaRegionalNetworkingSetting,
		getEntityFunc:  tmClient.GetTmRegionalNetworkingSettingById,
		stateStoreFunc: setTmRegionalNetworkingSettingData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionalNetworkingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	c := crudConfig[*govcd.TmRegionalNetworkingSetting, types.TmRegionalNetworkingSetting]{
		entityLabel:   labelVcfaRegionalNetworkingSetting,
		getEntityFunc: tmClient.GetTmRegionalNetworkingSettingById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionalNetworkingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	id := strings.Split(d.Id(), ImportSeparator)
	if len(id) != 2 {
		return nil, fmt.Errorf("ID syntax should be <%s name>%s<%s name>", labelVcfaOrg, ImportSeparator, labelVcfaRegionalNetworkingSetting)
	}

	org, err := tmClient.GetTmOrgByName(id[0])
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s': %s", labelVcfaOrg, id[0], err)
	}

	rns, err := tmClient.GetTmRegionalNetworkingSettingByNameAndOrgId(id[1], org.TmOrg.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s '%s' within %s '%s': %s",
			labelVcfaRegionalNetworkingSetting, id[1], labelVcfaOrg, id[0], err)
	}

	d.SetId(rns.TmRegionalNetworkingSetting.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmRegionalNetworkingSettingType(tmClient *VCDClient, d *schema.ResourceData) (*types.TmRegionalNetworkingSetting, error) {
	t := &types.TmRegionalNetworkingSetting{
		Name:                                d.Get("name").(string),
		OrgRef:                              types.OpenApiReference{ID: d.Get("org_id").(string)},
		RegionRef:                           types.OpenApiReference{ID: d.Get("region_id").(string)},
		DefaultVpcPrivateSubnetCidrOverride: d.Get("default_vpc_private_subnet_cidr_override").(string),
		ProjectTgwPrivateSubnetCidrOverride: d.Get("project_tgw_private_subnet_cidr_override").(string),
	}
	if v := d.Get("provider_gateway_id").(string); v != "" {
		t.ProviderGatewayRef = &types.OpenApiReference{ID: v}
	}
	if v := d.Get("distributed_vlan_connection_id").(string); v != "" {
		t.DistributedVlanConnectionRef = &types.OpenApiReference{ID: v}
	}
	if v := d.Get("distributed_vxlan_connection_id").(string); v != "" {
		t.DistributedVxlanConnectionRef = &types.OpenApiReference{ID: v}
	}
	if v := d.Get("virtual_network_appliance_cluster_id").(string); v != "" {
		t.VirtualNetworkApplianceClusterRef = &types.OpenApiReference{ID: v}
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
	dSet(d, "provider_gateway_id", refId(cfg.TmRegionalNetworkingSetting.ProviderGatewayRef))
	dSet(d, "distributed_vlan_connection_id", refId(cfg.TmRegionalNetworkingSetting.DistributedVlanConnectionRef))
	dSet(d, "distributed_vxlan_connection_id", refId(cfg.TmRegionalNetworkingSetting.DistributedVxlanConnectionRef))
	dSet(d, "virtual_network_appliance_cluster_id", refId(cfg.TmRegionalNetworkingSetting.VirtualNetworkApplianceClusterRef))
	if cfg.TmRegionalNetworkingSetting.ServiceEdgeClusterRef != nil {
		dSet(d, "edge_cluster_id", cfg.TmRegionalNetworkingSetting.ServiceEdgeClusterRef.ID)
	} else {
		dSet(d, "edge_cluster_id", "")
	}
	dSet(d, "status", cfg.TmRegionalNetworkingSetting.Status)

	return nil
}
