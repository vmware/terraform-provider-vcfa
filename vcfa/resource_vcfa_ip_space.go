package vcfa

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
)

const labelVcfaIpSpace = "IP Space"

var ipSpaceInternalScopeSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("ID of internal scope within %s", labelVcfaIpSpace),
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: fmt.Sprintf("Name of internal scope within %s", labelVcfaIpSpace),
		},
		"cidr": {
			Type:        schema.TypeString,
			Required:    true,
			Description: fmt.Sprintf("The CIDR that represents this IP block within %s", labelVcfaIpSpace),
		},
	},
}

func resourceVcfaIpSpace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaIpSpaceCreate,
		ReadContext:   resourceVcfaIpSpaceRead,
		UpdateContext: resourceVcfaIpSpaceUpdate,
		DeleteContext: resourceVcfaIpSpaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaIpSpaceImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaIpSpace),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaIpSpace),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Region ID for this %s", labelVcfaIpSpace),
			},
			"external_scope": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "External scope in CIDR format",
			},
			"default_quota_max_subnet_size": {
				Type:             schema.TypeString, // Values are 'ints', TypeString + validation is used to handle 0
				Required:         true,
				Description:      fmt.Sprintf("Maximum subnet size represented as a prefix length (e.g. 24, 28) in %s", labelVcfaIpSpace),
				ValidateDiagFunc: IsIntAndAtLeast(-1),
			},
			"default_quota_max_cidr_count": {
				Type:             schema.TypeString, // Values are 'ints', TypeString + validation is used to handle 0
				Required:         true,
				Description:      fmt.Sprintf("Maximum number of subnets that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelVcfaIpSpace),
				ValidateDiagFunc: IsIntAndAtLeast(-1),
			},
			"default_quota_max_ip_count": {
				Type:             schema.TypeString, // Values are 'ints', TypeString + validation is used to handle 0
				Required:         true,
				Description:      fmt.Sprintf("Maximum number of single floating IP addresses that can be allocated from internal scope in this %s. ('-1' for unlimited)", labelVcfaIpSpace),
				ValidateDiagFunc: IsIntAndAtLeast(-1),
			},
			"internal_scope": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: fmt.Sprintf("Internal scope of %s", labelVcfaIpSpace),
				Elem:        ipSpaceInternalScopeSchema,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaIpSpace),
			},
		},
	}
}

func resourceVcfaIpSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	unlock := tmClient.lockById(d.Get("region_id").(string))
	defer unlock()

	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:      labelVcfaIpSpace,
		getTypeFunc:      getIpSpaceType,
		stateStoreFunc:   setIpSpaceData,
		createAsyncFunc:  tmClient.CreateTmIpSpaceAsync,
		getEntityFunc:    tmClient.GetTmIpSpaceById,
		resourceReadFunc: resourceVcfaIpSpaceRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaIpSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	unlock := tmClient.lockById(d.Get("region_id").(string))
	defer unlock()

	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:      labelVcfaIpSpace,
		getTypeFunc:      getIpSpaceType,
		getEntityFunc:    tmClient.GetTmIpSpaceById,
		resourceReadFunc: resourceVcfaIpSpaceRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaIpSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:    labelVcfaIpSpace,
		getEntityFunc:  tmClient.GetTmIpSpaceById,
		stateStoreFunc: setIpSpaceData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaIpSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	unlock := tmClient.lockById(d.Get("region_id").(string))
	defer unlock()

	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:   labelVcfaIpSpace,
		getEntityFunc: tmClient.GetTmIpSpaceById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaIpSpaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as region-name.ip-space-name")
	}
	regionName, ipSpaceName := resourceURI[0], resourceURI[1]

	tmClient := meta.(ClientContainer).tmClient
	region, err := tmClient.GetRegionByName(regionName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name '%s': %s", labelVcfaRegion, regionName, err)
	}

	ipSpace, err := tmClient.GetTmIpSpaceByNameAndRegionId(ipSpaceName, region.Region.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by given name '%s': %s", labelVcfaIpSpace, d.Id(), err)
	}

	dSet(d, "region_id", region.Region.ID)
	d.SetId(ipSpace.TmIpSpace.ID)
	return []*schema.ResourceData{d}, nil
}

func getIpSpaceType(tmClient *VCDClient, d *schema.ResourceData) (*types.TmIpSpace, error) {
	t := &types.TmIpSpace{
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		RegionRef:         types.OpenApiReference{ID: d.Get("region_id").(string)},
		ExternalScopeCidr: d.Get("external_scope").(string),
	}

	// error is ignored because validation is enforced in schema fields
	maxCidrCountInt, _ := strconv.Atoi(d.Get("default_quota_max_cidr_count").(string))
	maxIPCountInt, _ := strconv.Atoi(d.Get("default_quota_max_ip_count").(string))
	maxSubnetSizeInt, _ := strconv.Atoi(d.Get("default_quota_max_subnet_size").(string))
	t.DefaultQuota = types.TmIpSpaceDefaultQuota{
		MaxCidrCount:  maxCidrCountInt,
		MaxIPCount:    maxIPCountInt,
		MaxSubnetSize: maxSubnetSizeInt,
	}

	// internal_scope
	internalScope := d.Get("internal_scope").(*schema.Set)
	internalScopeSlice := internalScope.List()
	if len(internalScopeSlice) > 0 {
		isSlice := make([]types.TmIpSpaceInternalScopeCidrBlocks, len(internalScopeSlice))
		for internalScopeIndex := range internalScopeSlice {
			internalScopeBlockStrings := convertToStringMap(internalScopeSlice[internalScopeIndex].(map[string]interface{}))

			isSlice[internalScopeIndex].Name = internalScopeBlockStrings["name"]
			isSlice[internalScopeIndex].Cidr = internalScopeBlockStrings["cidr"]

			// ID of internal_scope is important for updates
			// Terraform TypeSet cannot natively identify the ID between previous and new states
			// To work around this, an attempt to retrieve ID from state and correlate it with new payload is done
			// An important fact is that `cidr` field is not updatable, therefore one can be sure
			// that ID from state can be looked up based on CIDR.
			// If there was no such cidr in previous state - it means that this is a new 'internal_scope' block
			// and it doesn't need an ID
			isSlice[internalScopeIndex].ID = getInternalScopeIdFromFromPreviousState(d, internalScopeBlockStrings["name"], internalScopeBlockStrings["cidr"])

		}
		t.InternalScopeCidrBlocks = isSlice
	}

	return t, nil
}

func setIpSpaceData(_ *VCDClient, d *schema.ResourceData, i *govcd.TmIpSpace) error {
	if i == nil || i.TmIpSpace == nil {
		return fmt.Errorf("nil %s received", labelVcfaIpSpace)
	}

	d.SetId(i.TmIpSpace.ID)
	dSet(d, "name", i.TmIpSpace.Name)
	dSet(d, "description", i.TmIpSpace.Description)
	dSet(d, "region_id", i.TmIpSpace.RegionRef.ID)
	dSet(d, "external_scope", i.TmIpSpace.ExternalScopeCidr)
	dSet(d, "status", i.TmIpSpace.Status)

	dSet(d, "default_quota_max_subnet_size", strconv.Itoa(i.TmIpSpace.DefaultQuota.MaxSubnetSize))
	dSet(d, "default_quota_max_cidr_count", strconv.Itoa(i.TmIpSpace.DefaultQuota.MaxCidrCount))
	dSet(d, "default_quota_max_ip_count", strconv.Itoa(i.TmIpSpace.DefaultQuota.MaxIPCount))

	// internal_scope
	internalScopeInterface := make([]interface{}, len(i.TmIpSpace.InternalScopeCidrBlocks))
	for i, val := range i.TmIpSpace.InternalScopeCidrBlocks {
		singleScope := make(map[string]interface{})

		singleScope["id"] = val.ID
		singleScope["name"] = val.Name
		singleScope["cidr"] = val.Cidr

		internalScopeInterface[i] = singleScope
	}
	err := d.Set("internal_scope", internalScopeInterface)
	if err != nil {
		return fmt.Errorf("error storing 'internal_scope': %s", err)
	}

	return nil
}

func getInternalScopeIdFromFromPreviousState(d *schema.ResourceData, desiredName, desiredCidr string) string {
	internalScopeOld, _ := d.GetChange("internal_scope")
	internalScopeOldSchema := internalScopeOld.(*schema.Set)
	internalScopeOldSlice := internalScopeOldSchema.List()

	util.Logger.Printf("[TRACE] Looking for ID of 'internal_scope' with name '%s', cidr '%s'\n", desiredName, desiredCidr)
	var foundPartialId string
	for internalScopeIndex := range internalScopeOldSlice {
		singleScopeOld := internalScopeOldSlice[internalScopeIndex]
		singleScopeOldMap := convertToStringMap(singleScopeOld.(map[string]interface{}))

		// exact match
		if singleScopeOldMap["cidr"] == desiredCidr && singleScopeOldMap["name"] == desiredName {
			util.Logger.Printf("[TRACE] Found exact match for ID '%s' of 'internal_scope' with name '%s', cidr '%s' \n", singleScopeOldMap["id"], desiredName, desiredCidr)
			return singleScopeOldMap["id"]
		}

		// partial match based on cidr
		if singleScopeOldMap["cidr"] == desiredCidr {
			util.Logger.Printf("[TRACE] Found partial match for ID '%s' of 'internal_scope' with cidr '%s'. 'name' is ignored'\n", singleScopeOldMap["id"], desiredCidr)
			foundPartialId = singleScopeOldMap["id"]
		}
	}

	if foundPartialId != "" {
		util.Logger.Printf("[TRACE] Returning partial match for ID '%s' of 'internal_scope' with cidr '%s'. 'name' are ignored'\n", desiredCidr, desiredName)
		return foundPartialId
	}

	util.Logger.Printf("[TRACE] Not found 'internal_scope' ID with name '%s', cidr '%s'\n", desiredName, desiredCidr)
	// No ID was found at all
	return ""
}
