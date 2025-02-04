package vcfa

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaOrgVdc = "Org VDC"

func resourceVcfaOrgVdc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrgVdcCreate,
		ReadContext:   resourceOrgVdcRead,
		UpdateContext: resourceOrgVdcUpdate,
		DeleteContext: resourceOrgVdcDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOrgVdcImport,
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent Organization ID",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent Region ID",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of the %s", labelVcfaOrgVdc),
			},
			"supervisor_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("A set of Supervisor IDs that back this %s", labelVcfaOrgVdc),
			},
			"zone_resource_allocations": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        orgVdcZoneResourceAllocation,
				Description: "A set of Region Zones and their resource allocations",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s status", labelVcfaOrgVdc),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaOrgVdc),
			},
		},
	}
}

var orgVdcZoneResourceAllocation = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"region_zone_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("%s Name", labelVcfaRegionZone),
		},
		"region_zone_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: fmt.Sprintf("%s ID", labelVcfaRegionZone),
		},
		"memory_limit_mib": {
			Type:             schema.TypeInt,
			Required:         true,
			Description:      "Memory limit in MiB",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"memory_reservation_mib": {
			Type:             schema.TypeInt,
			Required:         true,
			Description:      "Memory reservation in MiB",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"cpu_limit_mhz": {
			Type:             schema.TypeInt,
			Required:         true,
			Description:      "CPU limit in MHz",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"cpu_reservation_mhz": {
			Type:             schema.TypeInt,
			Required:         true,
			Description:      "CPU reservation in MHz",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
	},
}

func resourceOrgVdcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:      labelVcfaOrgVdc,
		getTypeFunc:      getTmVdcType,
		stateStoreFunc:   setTmVdcData,
		createFunc:       tmClient.CreateTmVdc,
		resourceReadFunc: resourceOrgVdcRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceOrgVdcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:      labelVcfaOrgVdc,
		getTypeFunc:      getTmVdcType,
		getEntityFunc:    tmClient.GetTmVdcById,
		resourceReadFunc: resourceOrgVdcRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceOrgVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:    labelVcfaOrgVdc,
		getEntityFunc:  tmClient.GetTmVdcById,
		stateStoreFunc: setTmVdcData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceOrgVdcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:   labelVcfaOrgVdc,
		getEntityFunc: tmClient.GetTmVdcById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceOrgVdcImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	idSlice := strings.Split(d.Id(), ImportSeparator)
	if len(idSlice) != 2 {
		return nil, fmt.Errorf("expected import ID to be <org name>%s<region name>", ImportSeparator)
	}

	vdc, err := tmClient.GetTmVdcByName(fmt.Sprintf("%s_%s", idSlice[0], idSlice[1]))
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s: %s", labelVcfaOrgVdc, err)
	}

	d.SetId(vdc.TmVdc.ID)
	return []*schema.ResourceData{d}, nil
}

func getTmVdcType(tmClient *VCDClient, d *schema.ResourceData) (*types.TmVdc, error) {
	name := d.Get("name").(string)
	if name == "" {
		org, err := tmClient.GetOrgById(d.Get("org_id").(string))
		if err != nil {
			return nil, err
		}
		region, err := tmClient.GetRegionById(d.Get("region_id").(string))
		if err != nil {
			return nil, err
		}
		name = fmt.Sprintf("%s_%s", org.Org.Name, region.Region.Name)
	}
	t := &types.TmVdc{
		Name:        name,
		Description: d.Get("description").(string),
		Org:         &types.OpenApiReference{ID: d.Get("org_id").(string)},
		Region:      &types.OpenApiReference{ID: d.Get("region_id").(string)},
	}

	supervisorIds := convertSchemaSetToSliceOfStrings(d.Get("supervisor_ids").(*schema.Set))
	t.Supervisors = convertSliceOfStringsToOpenApiReferenceIds(supervisorIds)

	zra := d.Get("zone_resource_allocations").(*schema.Set)
	r := make([]*types.TmVdcZoneResourceAllocation, zra.Len())
	for zoneIndex, singleZone := range zra.List() {
		singleZoneMap := singleZone.(map[string]interface{})
		singleZoneType := &types.TmVdcZoneResourceAllocation{
			Zone: &types.OpenApiReference{
				ID: singleZoneMap["region_zone_id"].(string),
			},
			ResourceAllocation: types.TmVdcResourceAllocation{
				CPULimitMHz:          singleZoneMap["cpu_limit_mhz"].(int),
				CPUReservationMHz:    singleZoneMap["cpu_reservation_mhz"].(int),
				MemoryLimitMiB:       singleZoneMap["memory_limit_mib"].(int),
				MemoryReservationMiB: singleZoneMap["memory_reservation_mib"].(int),
			},
		}
		r[zoneIndex] = singleZoneType
	}
	t.ZoneResourceAllocation = r

	return t, nil
}

func setTmVdcData(_ *VCDClient, d *schema.ResourceData, vdc *govcd.TmVdc) error {
	if vdc == nil {
		return fmt.Errorf("provided VDC is nil")
	}

	d.SetId(vdc.TmVdc.ID)
	dSet(d, "name", vdc.TmVdc.Name)
	dSet(d, "description", vdc.TmVdc.Description)
	dSet(d, "status", vdc.TmVdc.Status)

	orgId := ""
	if vdc.TmVdc.Org != nil {
		orgId = vdc.TmVdc.Org.ID
	}
	dSet(d, "org_id", orgId)

	regionId := ""
	if vdc.TmVdc.Region != nil {
		regionId = vdc.TmVdc.Region.ID
	}
	dSet(d, "region_id", regionId)

	supervisors := extractIdsFromOpenApiReferences(vdc.TmVdc.Supervisors)
	err := d.Set("supervisor_ids", supervisors)
	if err != nil {
		return fmt.Errorf("error storing 'supervisor_ids': %s", err)
	}

	zoneCompute := make([]interface{}, len(vdc.TmVdc.ZoneResourceAllocation))
	for zoneIndex, zone := range vdc.TmVdc.ZoneResourceAllocation {
		oneZone := make(map[string]interface{})
		oneZone["region_zone_name"] = zone.Zone.Name
		oneZone["region_zone_id"] = zone.Zone.ID
		oneZone["memory_limit_mib"] = zone.ResourceAllocation.MemoryLimitMiB
		oneZone["memory_reservation_mib"] = zone.ResourceAllocation.MemoryReservationMiB
		oneZone["cpu_limit_mhz"] = zone.ResourceAllocation.CPULimitMHz
		oneZone["cpu_reservation_mhz"] = zone.ResourceAllocation.CPUReservationMHz

		zoneCompute[zoneIndex] = oneZone
	}

	autoAllocatedSubnetSet := schema.NewSet(schema.HashResource(orgVdcZoneResourceAllocation), zoneCompute)
	err = d.Set("zone_resource_allocations", autoAllocatedSubnetSet)
	if err != nil {
		return fmt.Errorf("error setting 'zone_resource_allocations' after read: %s", err)
	}

	return nil
}
