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

const labelVcfaOrgRegionQuota = "Org Region Quota"

func resourceVcfaOrgRegionQuota() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgRegionQuotaCreate,
		ReadContext:   resourceVcfaOrgRegionQuotaRead,
		UpdateContext: resourceVcfaOrgRegionQuotaUpdate,
		DeleteContext: resourceVcfaOrgRegionQuotaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgRegionQuotaImport,
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s ID", labelVcfaOrg),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s ID", labelVcfaRegion),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of the %s", labelVcfaOrgRegionQuota),
			},
			"supervisor_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("A set of Supervisor IDs that back this %s", labelVcfaOrgRegionQuota),
			},
			"zone_resource_allocations": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        orgRegionQuotaZoneResourceAllocation,
				Description: fmt.Sprintf("A set of %ss and their resource allocations", labelVcfaRegionZone),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s status", labelVcfaOrgRegionQuota),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaOrgRegionQuota),
			},
			"region_vm_class_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("A set of %s IDs to assign to this %s", labelVcfaRegionVmClass, labelVcfaOrgRegionQuota),
			},
			"region_storage_policy": {
				Type:        schema.TypeSet,
				MinItems:    1,
				Required:    true,
				Elem:        orgRegionQuotaRegionStoragePolicy,
				Description: fmt.Sprintf("A set of %s to assign to this %s", labelVcfaRegionStoragePolicy, labelVcfaOrgRegionQuota),
			},
		},
	}
}

var orgRegionQuotaZoneResourceAllocation = &schema.Resource{
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

var orgRegionQuotaRegionStoragePolicy = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"region_storage_policy_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: fmt.Sprintf("The ID of the %s for this %s", labelVcfaRegionStoragePolicy, labelVcfaOrgRegionQuota),
		},
		"storage_limit_mib": {
			Type:             schema.TypeInt,
			Required:         true,
			Description:      "Maximum allowed storage allocation in mebibytes. Minimum value: 0",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		},
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("The ID of the %s Storage Policy", labelVcfaOrgRegionQuota),
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The name of the Storage Policy. It follows RFC 1123 Label Names to conform with Kubernetes standards",
		},
		"storage_used_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Amount of storage used in mebibytes",
		},
	},
}

func assignVmClassesToRegionQuota(d *schema.ResourceData, tmClient *VCDClient) error {
	vmClassIds := convertSchemaSetToSliceOfStrings(d.Get("region_vm_class_ids").(*schema.Set))
	err := tmClient.AssignVmClassesToRegionQuota(d.Id(), &types.RegionVirtualMachineClasses{Values: convertSliceOfStringsToOpenApiReferenceIds(vmClassIds)})
	if err != nil {
		return err
	}
	return nil
}

func createStoragePoliciesInRegionQuota(d *schema.ResourceData, tmClient *VCDClient) error {
	regionQuotaId := d.Id()
	rsps := d.Get("region_storage_policy").(*schema.Set).List()
	payload := make([]types.VirtualDatacenterStoragePolicy, len(rsps))
	for i, rsp := range rsps {
		regionStoragePolicy := rsp.(map[string]interface{})
		payload[i] = types.VirtualDatacenterStoragePolicy{
			RegionStoragePolicy: types.OpenApiReference{
				ID: regionStoragePolicy["region_storage_policy_id"].(string),
			},
			StorageLimitMiB: int64(regionStoragePolicy["storage_limit_mib"].(int)),
			VirtualDatacenter: types.OpenApiReference{
				ID: regionQuotaId,
			},
		}
	}
	_, err := tmClient.CreateRegionQuotaStoragePolicies(regionQuotaId, &types.VirtualDatacenterStoragePolicies{Values: payload})
	if err != nil {
		return err
	}
	return nil
}

// updateStoragePoliciesInRegionQuota updates the set of Region Quota Storage Policies.
func updateStoragePoliciesInRegionQuota(d *schema.ResourceData, tmClient *VCDClient) error {
	regionQuotaId := d.Id()
	oldRsps, newRsps := d.GetChange("region_storage_policy")

	deleteFunc := func(oldSp map[string]interface{}) error {
		return tmClient.DeleteRegionQuotaStoragePolicy(oldSp["id"].(string))
	}

	updateFunc := func(newSp, oldSp map[string]interface{}) error {
		_, err := tmClient.UpdateRegionQuotaStoragePolicy(oldSp["id"].(string), &types.VirtualDatacenterStoragePolicy{
			ID: oldSp["id"].(string),
			RegionStoragePolicy: types.OpenApiReference{
				ID: oldSp["region_storage_policy_id"].(string),
			},
			Name:            oldSp["name"].(string),
			StorageLimitMiB: int64(newSp["storage_limit_mib"].(int)),
			VirtualDatacenter: types.OpenApiReference{
				ID: regionQuotaId,
			},
		})
		if err != nil {
			return err
		}
		return nil
	}

	var policiesToCreate []types.VirtualDatacenterStoragePolicy
	createFunc := func(newSp map[string]interface{}) error {
		policiesToCreate = append(policiesToCreate, types.VirtualDatacenterStoragePolicy{
			RegionStoragePolicy: types.OpenApiReference{
				ID: newSp["region_storage_policy_id"].(string),
			},
			StorageLimitMiB: int64(newSp["storage_limit_mib"].(int)),
			VirtualDatacenter: types.OpenApiReference{
				ID: regionQuotaId,
			},
		})
		return nil
	}

	oldRspsSet := oldRsps.(*schema.Set)
	newRspsSet := newRsps.(*schema.Set)

	err := searchSetAndApply(newRspsSet, oldRspsSet, "region_storage_policy_id", updateFunc, createFunc)
	if err != nil {
		return fmt.Errorf("could not update Storage Policies of Region Quota '%s': %s", regionQuotaId, err)
	}
	if len(policiesToCreate) > 0 {
		_, err := tmClient.CreateRegionQuotaStoragePolicies(regionQuotaId, &types.VirtualDatacenterStoragePolicies{Values: policiesToCreate})
		if err != nil {
			return fmt.Errorf("could not create Storage Policies during Region Quota '%s' update: %s", regionQuotaId, err)
		}
	}

	err = searchSetAndApply(oldRspsSet, newRspsSet, "region_storage_policy_id", nil, deleteFunc)
	if err != nil {
		return fmt.Errorf("could not delete old Storage Policies during Region Quota '%s' update: %s", regionQuotaId, err)
	}

	return nil
}

func saveVmClassesInState(tmClient *VCDClient, d *schema.ResourceData, rqId string) error {
	vmClasses, err := tmClient.GetVmClassesFromRegionQuota(rqId)
	if err != nil {
		return fmt.Errorf("could not fetch VM Classes from Region Quota '%s': %s", rqId, err)
	}
	if vmClasses != nil {
		vmcIds := make([]interface{}, len(vmClasses.Values))
		for i, vmc := range vmClasses.Values {
			vmcIds[i] = vmc.ID
		}
		err = d.Set("region_vm_class_ids", vmcIds)
		if err != nil {
			return fmt.Errorf("error setting 'region_vm_class_ids' after read: %s", err)
		}
	}
	return nil
}

func saveRegionStoragePoliciesInState(d *schema.ResourceData, regionQuota *govcd.RegionQuota, origin string) error {
	storagePolicies, err := regionQuota.GetAllStoragePolicies(nil)
	if err != nil {
		return fmt.Errorf("could not fetch Storage Policies from Region Quota '%s': %s", regionQuota.TmVdc.ID, err)
	}

	spsAttr := make([]interface{}, len(storagePolicies))
	for i, sp := range storagePolicies {
		spAttr := make(map[string]interface{})
		spAttr["id"] = sp.VirtualDatacenterStoragePolicy.ID
		spAttr["region_storage_policy_id"] = sp.VirtualDatacenterStoragePolicy.RegionStoragePolicy.ID // Can't be nil
		spAttr["storage_limit_mib"] = int(sp.VirtualDatacenterStoragePolicy.StorageLimitMiB)
		spAttr["name"] = sp.VirtualDatacenterStoragePolicy.Name
		spAttr["storage_used_mib"] = int(sp.VirtualDatacenterStoragePolicy.StorageUsedMiB)
		spsAttr[i] = spAttr
	}

	var attr *schema.Set
	if origin == "resource" {
		attr = schema.NewSet(schema.HashResource(orgRegionQuotaRegionStoragePolicy), spsAttr)
	} else {
		attr = schema.NewSet(schema.HashResource(orgRegionQuotaDsRegionStoragePolicy), spsAttr)
	}

	err = d.Set("region_storage_policy", attr)
	if err != nil {
		return fmt.Errorf("error setting 'region_storage_policy' after read: %s", err)
	}
	return nil
}

func resourceVcfaOrgRegionQuotaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.RegionQuota, types.TmVdc]{
		entityLabel:      labelVcfaOrgRegionQuota,
		getTypeFunc:      getOrgRegionQuotaType,
		stateStoreFunc:   setOrgRegionQuotaData,
		createFunc:       tmClient.CreateRegionQuota,
		resourceReadFunc: nil, // We don't use generic Read, as we didn't finish creation yet
	}
	diags := createResource(ctx, d, meta, c)
	if diags != nil {
		return diags
	}

	err := assignVmClassesToRegionQuota(d, tmClient)
	if err != nil {
		return diag.FromErr(err)
	}
	err = createStoragePoliciesInRegionQuota(d, tmClient)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVcfaOrgRegionQuotaRead(ctx, d, meta)
}

func resourceVcfaOrgRegionQuotaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.RegionQuota, types.TmVdc]{
		entityLabel:      labelVcfaOrgRegionQuota,
		getTypeFunc:      getOrgRegionQuotaType,
		getEntityFunc:    tmClient.GetRegionQuotaById,
		resourceReadFunc: nil, // We don't use generic Read, as we didn't finish update yet
	}

	diags := updateResource(ctx, d, meta, c)
	if diags != nil {
		return diags
	}

	err := assignVmClassesToRegionQuota(d, tmClient)
	if err != nil {
		return diag.FromErr(err)
	}
	err = updateStoragePoliciesInRegionQuota(d, tmClient)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVcfaOrgRegionQuotaRead(ctx, d, meta)
}

func resourceVcfaOrgRegionQuotaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.RegionQuota, types.TmVdc]{
		entityLabel:   labelVcfaOrgRegionQuota,
		getEntityFunc: tmClient.GetRegionQuotaById,
		stateStoreFunc: func(tmClient *VCDClient, d *schema.ResourceData, outerType *govcd.RegionQuota) error {
			err := setOrgRegionQuotaData(tmClient, d, outerType)
			if err != nil {
				return err
			}
			err = saveVmClassesInState(tmClient, d, outerType.TmVdc.ID)
			if err != nil {
				return err
			}
			return saveRegionStoragePoliciesInState(d, outerType, "resource")
		},
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionQuotaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	c := crudConfig[*govcd.RegionQuota, types.TmVdc]{
		entityLabel:   labelVcfaOrgRegionQuota,
		getEntityFunc: tmClient.GetRegionQuotaById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaOrgRegionQuotaImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	idSlice := strings.Split(d.Id(), ImportSeparator)
	if len(idSlice) != 2 {
		return nil, fmt.Errorf("expected import ID to be <org name>%s<region name>", ImportSeparator)
	}

	rq, err := tmClient.GetRegionQuotaByName(fmt.Sprintf("%s_%s", idSlice[0], idSlice[1]))
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s: %s", labelVcfaOrgRegionQuota, err)
	}

	d.SetId(rq.TmVdc.ID)
	return []*schema.ResourceData{d}, nil
}

func getOrgRegionQuotaType(tmClient *VCDClient, d *schema.ResourceData) (*types.TmVdc, error) {
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

func setOrgRegionQuotaData(_ *VCDClient, d *schema.ResourceData, rq *govcd.RegionQuota) error {
	if rq == nil || rq.TmVdc == nil {
		return fmt.Errorf("provided %s is nil", labelVcfaOrgRegionQuota)
	}

	d.SetId(rq.TmVdc.ID)
	dSet(d, "name", rq.TmVdc.Name)
	dSet(d, "description", rq.TmVdc.Description)
	dSet(d, "status", rq.TmVdc.Status)

	orgId := ""
	if rq.TmVdc.Org != nil {
		orgId = rq.TmVdc.Org.ID
	}
	dSet(d, "org_id", orgId)

	regionId := ""
	if rq.TmVdc.Region != nil {
		regionId = rq.TmVdc.Region.ID
	}
	dSet(d, "region_id", regionId)

	supervisors := extractIdsFromOpenApiReferences(rq.TmVdc.Supervisors)
	err := d.Set("supervisor_ids", supervisors)
	if err != nil {
		return fmt.Errorf("error storing 'supervisor_ids': %s", err)
	}

	zoneCompute := make([]interface{}, len(rq.TmVdc.ZoneResourceAllocation))
	for zoneIndex, zone := range rq.TmVdc.ZoneResourceAllocation {
		oneZone := make(map[string]interface{})
		oneZone["region_zone_name"] = zone.Zone.Name
		oneZone["region_zone_id"] = zone.Zone.ID
		oneZone["memory_limit_mib"] = zone.ResourceAllocation.MemoryLimitMiB
		oneZone["memory_reservation_mib"] = zone.ResourceAllocation.MemoryReservationMiB
		oneZone["cpu_limit_mhz"] = zone.ResourceAllocation.CPULimitMHz
		oneZone["cpu_reservation_mhz"] = zone.ResourceAllocation.CPUReservationMHz

		zoneCompute[zoneIndex] = oneZone
	}

	autoAllocatedSubnetSet := schema.NewSet(schema.HashResource(orgRegionQuotaZoneResourceAllocation), zoneCompute)
	err = d.Set("zone_resource_allocations", autoAllocatedSubnetSet)
	if err != nil {
		return fmt.Errorf("error setting 'zone_resource_allocations' after read: %s", err)
	}
	return nil
}
