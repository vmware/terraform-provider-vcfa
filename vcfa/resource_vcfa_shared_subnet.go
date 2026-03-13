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

const labelVcfaSharedSubnet = "Shared Subnet"

func resourceVcfaSharedSubnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaSharedSubnetCreate,
		ReadContext:   resourceVcfaSharedSubnetRead,
		UpdateContext: resourceVcfaSharedSubnetUpdate,
		DeleteContext: resourceVcfaSharedSubnetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaSharedSubnetImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaSharedSubnet),
			},
			"backing_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID for the matching Subnet in NSX",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaSharedSubnet),
			},
			"gateway_cidr": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("The gateway CIDR for the %s. This cannot be updated", labelVcfaSharedSubnet),
			},
			"ip_space_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("The IP Block that is automatically created for this %s", labelVcfaSharedSubnet),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Region ID for this %s", labelVcfaSharedSubnet),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaSharedSubnet),
			},
			"subnet_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Type of %s", labelVcfaSharedSubnet),
			},
			"vlan_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The VLAN ID if type is VLAN",
			},
		},
	}
}

func resourceVcfaSharedSubnetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	unlock := tmClient.lockById(d.Get("region_id").(string))
	defer unlock()

	c := crudConfig[*govcd.TmSharedSubnet, types.TmSharedSubnet]{
		entityLabel:      labelVcfaSharedSubnet,
		getTypeFunc:      getSharedSubnetType,
		stateStoreFunc:   setSharedSubnetData,
		createAsyncFunc:  tmClient.CreateTmSharedSubnetAsync,
		getEntityFunc:    tmClient.GetTmSharedSubnetById,
		resourceReadFunc: resourceVcfaSharedSubnetRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaSharedSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	unlock := tmClient.lockById(d.Get("region_id").(string))
	defer unlock()

	c := crudConfig[*govcd.TmSharedSubnet, types.TmSharedSubnet]{
		entityLabel:      labelVcfaSharedSubnet,
		getTypeFunc:      getSharedSubnetType,
		getEntityFunc:    tmClient.GetTmSharedSubnetById,
		resourceReadFunc: resourceVcfaSharedSubnetRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaSharedSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmSharedSubnet, types.TmSharedSubnet]{
		entityLabel:    labelVcfaSharedSubnet,
		getEntityFunc:  tmClient.GetTmSharedSubnetById,
		stateStoreFunc: setSharedSubnetData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaSharedSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	unlock := tmClient.lockById(d.Get("region_id").(string))
	defer unlock()

	c := crudConfig[*govcd.TmSharedSubnet, types.TmSharedSubnet]{
		entityLabel:   labelVcfaSharedSubnet,
		getEntityFunc: tmClient.GetTmSharedSubnetById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaSharedSubnetImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as region-name.shared-subnet-name")
	}
	regionName, sharedSubnetName := resourceURI[0], resourceURI[1]

	tmClient := meta.(ClientContainer).tmClient
	region, err := tmClient.GetRegionByName(regionName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name '%s': %s", labelVcfaRegion, regionName, err)
	}

	sharedSubnet, err := tmClient.GetTmSharedSubnetByNameAndRegionId(sharedSubnetName, region.Region.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by given name '%s': %s", labelVcfaSharedSubnet, sharedSubnetName, err)
	}

	dSet(d, "region_id", region.Region.ID)
	d.SetId(sharedSubnet.TmSharedSubnet.ID)
	return []*schema.ResourceData{d}, nil
}

func getSharedSubnetType(tmClient *VCDClient, d *schema.ResourceData) (*types.TmSharedSubnet, error) {
	t := &types.TmSharedSubnet{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		GatewayCidr: d.Get("gateway_cidr").(string),
		RegionRef:   types.OpenApiReference{ID: d.Get("region_id").(string)},
		SubnetType:  d.Get("subnet_type").(string),
		VlanId:      d.Get("vlan_id").(int),
	}

	return t, nil
}

func setSharedSubnetData(_ *VCDClient, d *schema.ResourceData, i *govcd.TmSharedSubnet) error {
	if i == nil || i.TmSharedSubnet == nil {
		return fmt.Errorf("nil %s received", labelVcfaSharedSubnet)
	}

	d.SetId(i.TmSharedSubnet.ID)
	dSet(d, "name", i.TmSharedSubnet.Name)
	dSet(d, "backing_id", i.TmSharedSubnet.BackingId)
	dSet(d, "description", i.TmSharedSubnet.Description)
	dSet(d, "gateway_cidr", i.TmSharedSubnet.GatewayCidr)
	dSet(d, "ip_space_id", i.TmSharedSubnet.IPSpaceRef.ID)
	dSet(d, "region_id", i.TmSharedSubnet.RegionRef.ID)
	dSet(d, "status", i.TmSharedSubnet.Status)
	dSet(d, "subnet_type", i.TmSharedSubnet.SubnetType)
	dSet(d, "vlan_id", i.TmSharedSubnet.VlanId)

	return nil
}
