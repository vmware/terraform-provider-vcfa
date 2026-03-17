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

const labelVcfaDistributedVlanConnection = "Distributed Vlan Connection"

func resourceVcfaDistributedVlanConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaDistributedVlanConnectionCreate,
		ReadContext:   resourceVcfaDistributedVlanConnectionRead,
		UpdateContext: resourceVcfaDistributedVlanConnectionUpdate,
		DeleteContext: resourceVcfaDistributedVlanConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaDistributedVlanConnectionImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaDistributedVlanConnection),
			},
			"backing_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("ID for the matching %s in NSX", labelVcfaDistributedVlanConnection),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaDistributedVlanConnection),
			},
			"gateway_cidr": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The gateway CIDR for the %s", labelVcfaDistributedVlanConnection),
			},
			"ip_space_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: fmt.Sprintf("Reference to an IP Block that is used for the external connection for this %s", labelVcfaDistributedVlanConnection),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Region ID for this %s", labelVcfaDistributedVlanConnection),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaDistributedVlanConnection),
			},
			"subnet_exclusive": {
				Type:        schema.TypeBool,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Whether this %s is exclusively for the gateway CIDR only", labelVcfaDistributedVlanConnection),
			},
			"vlan_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The VLAN ID for the external traffic",
			},
			"zone_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: fmt.Sprintf("The supervisor zones this %s spans", labelVcfaDistributedVlanConnection),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcfaDistributedVlanConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	unlock := tmClient.lockById(d.Get("region_id").(string))
	defer unlock()

	c := crudConfig[*govcd.TmDistributedVlanConnection, types.TmDistributedVlanConnection]{
		entityLabel:      labelVcfaDistributedVlanConnection,
		getTypeFunc:      getDistributedVlanConnectionType,
		stateStoreFunc:   setDistributedVlanConnectionData,
		createAsyncFunc:  tmClient.CreateTmDistributedVlanConnectionAsync,
		getEntityFunc:    tmClient.GetTmDistributedVlanConnectionById,
		resourceReadFunc: resourceVcfaDistributedVlanConnectionRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaDistributedVlanConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	unlock := tmClient.lockById(d.Get("region_id").(string))
	defer unlock()

	c := crudConfig[*govcd.TmDistributedVlanConnection, types.TmDistributedVlanConnection]{
		entityLabel:      labelVcfaDistributedVlanConnection,
		getTypeFunc:      getDistributedVlanConnectionType,
		getEntityFunc:    tmClient.GetTmDistributedVlanConnectionById,
		resourceReadFunc: resourceVcfaDistributedVlanConnectionRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaDistributedVlanConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmDistributedVlanConnection, types.TmDistributedVlanConnection]{
		entityLabel:    labelVcfaDistributedVlanConnection,
		getEntityFunc:  tmClient.GetTmDistributedVlanConnectionById,
		stateStoreFunc: setDistributedVlanConnectionData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaDistributedVlanConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	unlock := tmClient.lockById(d.Get("region_id").(string))
	defer unlock()

	c := crudConfig[*govcd.TmDistributedVlanConnection, types.TmDistributedVlanConnection]{
		entityLabel:   labelVcfaDistributedVlanConnection,
		getEntityFunc: tmClient.GetTmDistributedVlanConnectionById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaDistributedVlanConnectionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as region-name.distributed-vlan-connection-name")
	}
	regionName, distributedVlanConnectionName := resourceURI[0], resourceURI[1]

	tmClient := meta.(ClientContainer).tmClient
	region, err := tmClient.GetRegionByName(regionName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name '%s': %s", labelVcfaRegion, regionName, err)
	}

	distributedVlanConnection, err := tmClient.GetTmDistributedVlanConnectionByNameAndRegionId(distributedVlanConnectionName, region.Region.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by given name '%s': %s", labelVcfaDistributedVlanConnection, distributedVlanConnectionName, err)
	}

	dSet(d, "region_id", region.Region.ID)
	d.SetId(distributedVlanConnection.TmDistributedVlanConnection.ID)
	return []*schema.ResourceData{d}, nil
}

func getDistributedVlanConnectionType(tmClient *VCDClient, d *schema.ResourceData) (*types.TmDistributedVlanConnection, error) {
	t := &types.TmDistributedVlanConnection{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		GatewayCidr:     d.Get("gateway_cidr").(string),
		RegionRef:       types.OpenApiReference{ID: d.Get("region_id").(string)},
		SubnetExclusive: d.Get("subnet_exclusive").(bool),
		VlanId:          d.Get("vlan_id").(int),
		ZoneRefs:        convertSliceOfStringsToOpenApiReferenceIds(convertTypeListToSliceOfStrings(d.Get("zone_ids").(*schema.Set).List())),
	}

	if ipSpaceId, ok := d.GetOk("ip_space_id"); ok {
		t.IpSpaceRef = &types.OpenApiReference{ID: ipSpaceId.(string)}
	}

	return t, nil
}

func setDistributedVlanConnectionData(_ *VCDClient, d *schema.ResourceData, dvc *govcd.TmDistributedVlanConnection) error {
	if dvc == nil || dvc.TmDistributedVlanConnection == nil {
		return fmt.Errorf("nil %s received", labelVcfaDistributedVlanConnection)
	}

	d.SetId(dvc.TmDistributedVlanConnection.ID)
	dSet(d, "name", dvc.TmDistributedVlanConnection.Name)
	dSet(d, "backing_id", dvc.TmDistributedVlanConnection.BackingId)
	dSet(d, "description", dvc.TmDistributedVlanConnection.Description)
	dSet(d, "gateway_cidr", dvc.TmDistributedVlanConnection.GatewayCidr)
	dSet(d, "region_id", dvc.TmDistributedVlanConnection.RegionRef.ID)
	dSet(d, "status", dvc.TmDistributedVlanConnection.Status)
	dSet(d, "subnet_exclusive", dvc.TmDistributedVlanConnection.SubnetExclusive)
	dSet(d, "vlan_id", dvc.TmDistributedVlanConnection.VlanId)

	if dvc.TmDistributedVlanConnection.IpSpaceRef != nil {
		dSet(d, "ip_space_id", dvc.TmDistributedVlanConnection.IpSpaceRef.ID)
	}

	zoneIds := extractIdsFromOpenApiReferences(dvc.TmDistributedVlanConnection.ZoneRefs)
	if err := d.Set("zone_ids", zoneIds); err != nil {
		return fmt.Errorf("error storing 'zone_ids': %s", err)
	}

	return nil
}
