// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

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

var ipSpaceIpBlockSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("ID of the IP Block within %s", labelVcfaIpSpace),
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: fmt.Sprintf("Name of the IP Block within %s", labelVcfaIpSpace),
		},
		"cidr": {
			Type:        schema.TypeString,
			Required:    true,
			Description: fmt.Sprintf("The CIDR that represents this IP Block within %s", labelVcfaIpSpace),
		},
	},
}

var ipSpaceIpRangeSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("ID of IP Range within %s", labelVcfaIpSpace),
		},
		"start_ip_address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Starting IP address in the range",
		},
		"end_ip_address": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Ending IP address in the range",
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
			"backing_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID for the matching IP Block in NSX",
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
				Deprecated:  "Use 'inbound_remote_networks' in 'vcfa_provider_gateway' resource instead",
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
			"cidr_blocks": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Description:   fmt.Sprintf("CIDR blocks of %s. Along with 'ip_address_ranges' typically defines the span of IP addresses used within a Data Center", labelVcfaIpSpace),
				AtLeastOneOf:  []string{"cidr_blocks", "internal_scope", "ip_address_ranges"},
				ConflictsWith: []string{"internal_scope"},
				Elem:          ipSpaceIpBlockSchema,
				MaxItems:      30,
			},
			"internal_scope": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Deprecated:    "Use 'cidr_blocks' instead",
				Description:   fmt.Sprintf("Internal scope of %s. Along with 'ip_address_ranges' typically defines the span of IP addresses used within a Data Center", labelVcfaIpSpace),
				AtLeastOneOf:  []string{"cidr_blocks", "internal_scope", "ip_address_ranges"},
				ConflictsWith: []string{"cidr_blocks"},
				Elem:          ipSpaceIpBlockSchema,
				MaxItems:      30,
			},
			"ip_address_ranges": {
				Type:         schema.TypeSet,
				Optional:     true,
				Description:  fmt.Sprintf("IP address ranges of %s. Along with 'internal_scope' typically defines the span of IP addresses used within a Data Center", labelVcfaIpSpace),
				AtLeastOneOf: []string{"cidr_blocks", "internal_scope", "ip_address_ranges"},
				Elem:         ipSpaceIpRangeSchema,
				MaxItems:     30,
			},
			"is_imported_ip_block": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if the IP Block is imported from an existing NSX IP Block",
			},
			"provider_visibility_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: fmt.Sprintf("If set to true, the %s details will be hidden from organizations", labelVcfaIpSpace),
			},
			"reserved_ip_address_ranges": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: fmt.Sprintf("IP addresses that will not be considered for IP allocation. Reserved IPs have to be part of one of the CIDRs or IP Ranges of %s", labelVcfaIpSpace),
				Elem:        ipSpaceIpRangeSchema,
				MaxItems:    128,
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaIpSpace),
			},
			"subnet_exclusive": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether this IP Block is exclusively for a single CIDR",
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
		Name:                   d.Get("name").(string),
		Description:            d.Get("description").(string),
		RegionRef:              types.OpenApiReference{ID: d.Get("region_id").(string)},
		ExternalScopeCidr:      d.Get("external_scope").(string),
		ProviderVisibilityOnly: d.Get("provider_visibility_only").(bool),
	}

	// error is ignored because validation is enforced in schema fields
	maxCidrCountInt, _ := strconv.Atoi(d.Get("default_quota_max_cidr_count").(string))
	maxIPCountInt, _ := strconv.Atoi(d.Get("default_quota_max_ip_count").(string))
	maxSubnetSizeInt, _ := strconv.Atoi(d.Get("default_quota_max_subnet_size").(string))
	t.DefaultQuota = types.TmIpSpaceQuota{
		MaxCidrCount:  maxCidrCountInt,
		MaxIPCount:    maxIPCountInt,
		MaxSubnetSize: maxSubnetSizeInt,
	}

	if _, ok := d.GetOk("cidr_blocks"); ok {
		t.InternalScopeCidrBlocks = getIpBlocksFromSchema(d, "cidr_blocks")
	} else if _, ok := d.GetOk("internal_scope"); ok {
		t.InternalScopeCidrBlocks = getIpBlocksFromSchema(d, "internal_scope")
	}

	t.IpAddressRanges = getIpRangesFromSchema(d, "ip_address_ranges")
	t.ReservedIpAddressRanges = getIpRangesFromSchema(d, "reserved_ip_address_ranges")

	return t, nil
}

func getIpBlocksFromSchema(d *schema.ResourceData, fieldName string) []types.TmIpAddressSpaceIpBlock {
	ipBlocks := d.Get(fieldName).(*schema.Set)
	ipBlocksSlice := ipBlocks.List()
	if len(ipBlocksSlice) == 0 {
		return nil
	}

	result := make([]types.TmIpAddressSpaceIpBlock, len(ipBlocksSlice))
	for i, r := range ipBlocksSlice {
		ipBlockMap := convertToStringMap(r.(map[string]interface{}))
		result[i] = types.TmIpAddressSpaceIpBlock{
			Name: ipBlockMap["name"],
			Cidr: ipBlockMap["cidr"],
		}

		// ID of IP Block is important for updates
		// Terraform TypeSet cannot natively identify the ID between previous and new states
		// To work around this, an attempt to retrieve ID from state and correlate it with new payload is done
		// An important fact is that `cidr` field is not updatable, therefore one can be sure
		// that ID from state can be looked up based on CIDR.
		// If there was no such cidr in previous state - it means that this is a new IP Block
		// and it doesn't need an ID
		result[i].ID = getIpBlockIdFromFromPreviousState(d, fieldName, ipBlockMap["name"], ipBlockMap["cidr"])
	}
	return result
}

func getIpRangesFromSchema(d *schema.ResourceData, fieldName string) []types.TmIpAddressSpaceRange {
	ipRanges := d.Get(fieldName).(*schema.Set)
	ipRangesSlice := ipRanges.List()
	if len(ipRangesSlice) == 0 {
		return nil
	}

	result := make([]types.TmIpAddressSpaceRange, len(ipRangesSlice))
	for i, r := range ipRangesSlice {
		rangeMap := convertToStringMap(r.(map[string]interface{}))
		result[i] = types.TmIpAddressSpaceRange{
			StartIpAddress: rangeMap["start_ip_address"],
			EndIpAddress:   rangeMap["end_ip_address"],
		}
	}
	return result
}

func setIpSpaceData(_ *VCDClient, d *schema.ResourceData, i *govcd.TmIpSpace) error {
	if i == nil || i.TmIpSpace == nil {
		return fmt.Errorf("nil %s received", labelVcfaIpSpace)
	}

	d.SetId(i.TmIpSpace.ID)
	dSet(d, "name", i.TmIpSpace.Name)
	dSet(d, "backing_id", i.TmIpSpace.BackingId)
	dSet(d, "description", i.TmIpSpace.Description)
	dSet(d, "external_scope", i.TmIpSpace.ExternalScopeCidr)
	dSet(d, "is_imported_ip_block", i.TmIpSpace.IsImportedIpBlock)
	dSet(d, "provider_visibility_only", i.TmIpSpace.ProviderVisibilityOnly)
	dSet(d, "region_id", i.TmIpSpace.RegionRef.ID)
	dSet(d, "subnet_exclusive", i.TmIpSpace.SubnetExclusive)
	dSet(d, "status", i.TmIpSpace.Status)

	dSet(d, "default_quota_max_subnet_size", strconv.Itoa(i.TmIpSpace.DefaultQuota.MaxSubnetSize))
	dSet(d, "default_quota_max_cidr_count", strconv.Itoa(i.TmIpSpace.DefaultQuota.MaxCidrCount))
	dSet(d, "default_quota_max_ip_count", strconv.Itoa(i.TmIpSpace.DefaultQuota.MaxIPCount))

	cidrBlocks := setIpAddressSpaceIpBlocksToState(i.TmIpSpace.InternalScopeCidrBlocks)
	if err := d.Set("cidr_blocks", cidrBlocks); err != nil {
		return fmt.Errorf("error storing 'cidr_blocks': %s", err)
	}

	if err := d.Set("internal_scope", cidrBlocks); err != nil {
		return fmt.Errorf("error storing 'internal_scope': %s", err)
	}

	if err := d.Set("ip_address_ranges", setIpAddressSpaceRangesToState(i.TmIpSpace.IpAddressRanges)); err != nil {
		return fmt.Errorf("error storing 'ip_address_ranges': %s", err)
	}

	if err := d.Set("reserved_ip_address_ranges", setIpAddressSpaceRangesToState(i.TmIpSpace.ReservedIpAddressRanges)); err != nil {
		return fmt.Errorf("error storing 'reserved_ip_address_ranges': %s", err)
	}

	return nil
}

func setIpAddressSpaceIpBlocksToState(cidrBlocks []types.TmIpAddressSpaceIpBlock) []interface{} {
	cidrBlocksInterface := make([]interface{}, len(cidrBlocks))
	for i, cidrBlock := range cidrBlocks {
		cidrBlocksInterface[i] = map[string]interface{}{
			"id":   cidrBlock.ID,
			"name": cidrBlock.Name,
			"cidr": cidrBlock.Cidr,
		}
	}
	return cidrBlocksInterface
}

func setIpAddressSpaceRangesToState(ranges []types.TmIpAddressSpaceRange) []interface{} {
	rangesInterface := make([]interface{}, len(ranges))
	for i, r := range ranges {
		rangesInterface[i] = map[string]interface{}{
			"id":               r.ID,
			"start_ip_address": r.StartIpAddress,
			"end_ip_address":   r.EndIpAddress,
		}
	}
	return rangesInterface
}

func getIpBlockIdFromFromPreviousState(d *schema.ResourceData, fieldName string, desiredName, desiredCidr string) string {
	ipBlocksOld, _ := d.GetChange(fieldName)
	ipBlocksOldSchema := ipBlocksOld.(*schema.Set)
	ipBlocksOldSlice := ipBlocksOldSchema.List()

	util.Logger.Printf("[TRACE] Looking for ID of %s' with name '%s', cidr '%s'\n", fieldName, desiredName, desiredCidr)
	var foundPartialId string
	for ipBlockIndex := range ipBlocksOldSlice {
		ipBlockOld := ipBlocksOldSlice[ipBlockIndex]
		ipBlockOldMap := convertToStringMap(ipBlockOld.(map[string]interface{}))

		// exact match
		if ipBlockOldMap["cidr"] == desiredCidr && ipBlockOldMap["name"] == desiredName {
			util.Logger.Printf("[TRACE] Found exact match for ID '%s' of '%s' with name '%s', cidr '%s' \n", ipBlockOldMap["id"], fieldName, desiredName, desiredCidr)
			return ipBlockOldMap["id"]
		}

		// partial match based on cidr
		if ipBlockOldMap["cidr"] == desiredCidr {
			util.Logger.Printf("[TRACE] Found partial match for ID '%s' of '%s' with cidr '%s'. 'name' is ignored'\n", ipBlockOldMap["id"], fieldName, desiredCidr)
			foundPartialId = ipBlockOldMap["id"]
		}
	}

	if foundPartialId != "" {
		util.Logger.Printf("[TRACE] Returning partial match for ID '%s' of '%s' with cidr '%s'. 'name' are ignored'\n", desiredCidr, fieldName, desiredName)
		return foundPartialId
	}

	util.Logger.Printf("[TRACE] Not found '%s' ID with name '%s', cidr '%s'\n", fieldName, desiredName, desiredCidr)
	// No ID was found at all
	return ""
}
