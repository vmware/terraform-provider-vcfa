package vcfa

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaRegion = "Region"

// rfc1123LabelNameRegex matches strings with at most 31 characters, composed only by lowercase
// alphanumeric characters or '-', that must start with an alphabetic character, and end with an
// alphanumeric.
var rfc1123LabelNameRegex = regexp.MustCompile(`^[a-z](?:[a-z0-9-]{0,29}[a-z0-9])?$`)

func resourceVcfaRegion() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaRegionCreate,
		ReadContext:   resourceVcfaRegionRead,
		UpdateContext: resourceVcfaRegionUpdate,
		DeleteContext: resourceVcfaRegionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaRegionImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Region names cannot be changed
				Description: fmt.Sprintf("%s name", labelVcfaRegion),
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringMatch(rfc1123LabelNameRegex, "Name must match RFC 1123 Label name (lower case alphabet, 0-9 and hyphen -)"),
				),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("%s description", labelVcfaRegion),
			},
			"nsx_manager_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX Manager ID",
			},
			"supervisor_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: fmt.Sprintf("A set of supervisor IDs used in this %s", labelVcfaRegion),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"storage_policy_names": { // TODO: TM: check if the API accepts IDs and if it should use
				Type:        schema.TypeSet,
				Required:    true,
				Description: "A set of storage policy names",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cpu_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU Capacity in MHz",
			},
			"cpu_reservation_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU reservation in MHz",
			},
			"memory_capacity_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory capacity in MiB",
			},
			"memory_reservation_capacity_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory reservation in MiB",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of the %s", labelVcfaRegion),
			},
		},
	}
}

func resourceVcfaRegionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	c := crudConfig[*govcd.Region, types.Region]{
		entityLabel:      labelVcfaRegion,
		getTypeFunc:      getRegionType,
		stateStoreFunc:   setRegionData,
		createFunc:       vcdClient.CreateRegion,
		resourceReadFunc: resourceVcfaRegionRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaRegionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	c := crudConfig[*govcd.Region, types.Region]{
		entityLabel:      labelVcfaRegion,
		getTypeFunc:      getRegionType,
		getEntityFunc:    vcdClient.GetRegionById,
		resourceReadFunc: resourceVcfaRegionRead,
	}
	return updateResource(ctx, d, meta, c)
}

func resourceVcfaRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient
	c := crudConfig[*govcd.Region, types.Region]{
		entityLabel:    labelVcfaRegion,
		getEntityFunc:  vcdClient.GetRegionById,
		stateStoreFunc: setRegionData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaRegionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(MetaContainer).VcfaClient

	c := crudConfig[*govcd.Region, types.Region]{
		entityLabel:   labelVcfaRegion,
		getEntityFunc: vcdClient.GetRegionById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaRegionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(MetaContainer).VcfaClient
	region, err := vcdClient.GetRegionByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving Region: %s", err)
	}

	d.SetId(region.Region.ID)

	return []*schema.ResourceData{d}, nil
}

func getRegionType(vcdClient *VCDClient, d *schema.ResourceData) (*types.Region, error) {
	t := &types.Region{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		NsxManager:  &types.OpenApiReference{ID: d.Get("nsx_manager_id").(string)},
	}

	// API requires Names to be sent with IDs, but Terraform native approach is to use IDs only
	// therefore names need to be looked up for IDs
	supervisorIds := convertSchemaSetToSliceOfStrings(d.Get("supervisor_ids").(*schema.Set))
	superVisorReferences := make([]types.OpenApiReference, 0)
	for _, singleSupervisorId := range supervisorIds {
		supervisor, err := vcdClient.GetSupervisorById(singleSupervisorId)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Supervisor with ID %s: %s", singleSupervisorId, err)
		}

		superVisorReferences = append(superVisorReferences, types.OpenApiReference{
			ID:   supervisor.Supervisor.SupervisorID,
			Name: supervisor.Supervisor.Name,
		})
	}
	t.Supervisors = superVisorReferences

	storagePolicyNames := convertSchemaSetToSliceOfStrings(d.Get("storage_policy_names").(*schema.Set))
	t.StoragePolicies = storagePolicyNames

	return t, nil
}

func setRegionData(_ *VCDClient, d *schema.ResourceData, r *govcd.Region) error {
	if r == nil || r.Region == nil {
		return fmt.Errorf("nil Region entity")
	}

	d.SetId(r.Region.ID)
	dSet(d, "name", r.Region.Name)
	dSet(d, "description", r.Region.Description)
	dSet(d, "nsx_manager_id", r.Region.NsxManager.ID)

	dSet(d, "cpu_capacity_mhz", r.Region.CPUCapacityMHz)
	dSet(d, "cpu_reservation_capacity_mhz", r.Region.CPUReservationCapacityMHz)
	dSet(d, "memory_capacity_mib", r.Region.MemoryCapacityMiB)
	dSet(d, "memory_reservation_capacity_mib", r.Region.MemoryReservationCapacityMiB)
	dSet(d, "status", r.Region.Status)

	err := d.Set("supervisor_ids", extractIdsFromOpenApiReferences(r.Region.Supervisors))
	if err != nil {
		return fmt.Errorf("error storing 'supervisors': %s", err)
	}

	err = d.Set("storage_policy_names", r.Region.StoragePolicies)
	if err != nil {
		return fmt.Errorf("error storing 'storage_policy_names': %s", err)
	}

	return nil
}
