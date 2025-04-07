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

const labelVcfaProviderGateway = "Provider Gateway"
const labelVcfaProviderGatewayIpSpaceAssociations = "IP Space Associations"

// Provider Gateway has an asymmetric API - it requires are least one IP Space reference when
// creating a Provider Gateway, but it will not return Associated IP Spaces afterwards. Instead,
// to update associated IP Spaces one must use separate API endpoint and structure
// (`TmIpSpaceAssociation`) to manage associations after initial Provider Gateway creation

func resourceVcfaProviderGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaProviderGatewayCreate,
		ReadContext:   resourceVcfaProviderGatewayRead,
		UpdateContext: resourceVcfaProviderGatewayUpdate,
		DeleteContext: resourceVcfaProviderGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaProviderGatewayImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVcfaProviderGateway),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Description of %s", labelVcfaProviderGateway),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelVcfaRegion, labelVcfaProviderGateway),
			},
			"tier0_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s of %s", labelVcfaTier0Gateway, labelVcfaProviderGateway),
			},
			"ip_space_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: fmt.Sprintf("A set of supervisor IDs used in this %s", labelVcfaRegion),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Status of %s", labelVcfaProviderGateway),
			},
		},
	}
}

func resourceVcfaProviderGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:      labelVcfaProviderGateway,
		getTypeFunc:      getProviderGatewayType,
		stateStoreFunc:   setProviderGatewayData,
		createAsyncFunc:  tmClient.CreateTmProviderGatewayAsync,
		getEntityFunc:    tmClient.GetTmProviderGatewayById,
		resourceReadFunc: resourceVcfaProviderGatewayRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaProviderGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	// Update IP Space associations using separate endpoint (more details at the top of file)
	if d.HasChange("ip_space_ids") {
		previous, new := d.GetChange("ip_space_ids")
		previousSet := previous.(*schema.Set)
		newSet := new.(*schema.Set)

		toRemoveSet := previousSet.Difference(newSet)
		toAddSet := newSet.Difference(previousSet)

		// Adding new ones first, because it can happen that all previous IP Spaces are removed and
		// new ones added, however API prohibits removal of all IP Space associations for Provider
		// Gateway (at least one IP Space must always be associated)
		err := addIpSpaceAssociations(tmClient, d.Id(), convertSchemaSetToSliceOfStrings(toAddSet))
		if err != nil {
			return diag.FromErr(err)
		}

		// Remove associations that are no more in configuration
		err = removeIpSpaceAssociations(tmClient, d.Id(), convertSchemaSetToSliceOfStrings(toRemoveSet))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// This is the default entity update path - other fields can be updated, by updating IP Space itself
	if d.HasChangeExcept("ip_space_ids") {
		c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
			entityLabel:      labelVcfaProviderGateway,
			getTypeFunc:      getProviderGatewayType,
			getEntityFunc:    tmClient.GetTmProviderGatewayById,
			resourceReadFunc: resourceVcfaProviderGatewayRead,
		}

		return updateResource(ctx, d, meta, c)
	}

	return nil
}

func resourceVcfaProviderGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:    labelVcfaProviderGateway,
		getEntityFunc:  tmClient.GetTmProviderGatewayById,
		stateStoreFunc: setProviderGatewayData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaProviderGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:   labelVcfaProviderGateway,
		getEntityFunc: tmClient.GetTmProviderGatewayById,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaProviderGatewayImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as region-name.provider-gateway-name")
	}
	regionName, providerGatewayName := resourceURI[0], resourceURI[1]

	tmClient := meta.(ClientContainer).tmClient
	region, err := tmClient.GetRegionByName(regionName)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s by name '%s': %s", labelVcfaRegion, regionName, err)
	}

	providerGateway, err := tmClient.GetTmProviderGatewayByNameAndRegionId(providerGatewayName, region.Region.ID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving Provider Gateway: %s", err)
	}

	d.SetId(providerGateway.TmProviderGateway.ID)
	return []*schema.ResourceData{d}, nil
}

func getProviderGatewayType(tmClient *VCDClient, d *schema.ResourceData) (*types.TmProviderGateway, error) {
	t := &types.TmProviderGateway{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		RegionRef:   types.OpenApiReference{ID: d.Get("region_id").(string)},
		BackingRef:  types.OpenApiReference{ID: d.Get("tier0_gateway_id").(string)},
	}

	ipSpaceIds := convertSchemaSetToSliceOfStrings(d.Get("ip_space_ids").(*schema.Set))
	t.IPSpaceRefs = convertSliceOfStringsToOpenApiReferenceIds(ipSpaceIds)

	// Update operation fails if the ID is not set for update
	if d.Id() != "" {
		t.ID = d.Id()
	}

	// IP Spaces associations are populated on create only. Updates are done using separate endpoint
	// (more details at the top of file)
	if d.Id() != "" {
		t.IPSpaceRefs = []types.OpenApiReference{}
	}

	return t, nil
}

func setProviderGatewayData(tmClient *VCDClient, d *schema.ResourceData, p *govcd.TmProviderGateway) error {
	if p == nil || p.TmProviderGateway == nil {
		return fmt.Errorf("nil entity received")
	}

	d.SetId(p.TmProviderGateway.ID)
	dSet(d, "name", p.TmProviderGateway.Name)
	dSet(d, "description", p.TmProviderGateway.Description)
	dSet(d, "region_id", p.TmProviderGateway.RegionRef.ID)
	dSet(d, "tier0_gateway_id", p.TmProviderGateway.BackingRef.ID)
	dSet(d, "status", p.TmProviderGateway.Status)

	// IP Space Associations have to be read separatelly after creation (more details at the top of file)
	associations, err := tmClient.GetAllTmIpSpaceAssociationsByProviderGatewayId(p.TmProviderGateway.ID)
	if err != nil {
		return fmt.Errorf("error retrieving %s for %s", labelVcfaProviderGatewayIpSpaceAssociations, labelVcfaProviderGateway)
	}
	associationIds := make([]string, len(associations))
	for index, singleAssociation := range associations {
		associationIds[index] = singleAssociation.TmIpSpaceAssociation.IPSpaceRef.ID
	}

	err = d.Set("ip_space_ids", associationIds)
	if err != nil {
		return fmt.Errorf("error storing 'ip_space_ids': %s", err)
	}

	return nil
}

func addIpSpaceAssociations(tmClient *VCDClient, providerGatewayId string, addIpSpaceIds []string) error {
	for _, addIpSpaceId := range addIpSpaceIds {
		at := &types.TmIpSpaceAssociation{
			IPSpaceRef:         &types.OpenApiReference{ID: addIpSpaceId},
			ProviderGatewayRef: &types.OpenApiReference{ID: providerGatewayId},
		}
		_, err := tmClient.CreateTmIpSpaceAssociation(at)
		if err != nil {
			return fmt.Errorf("error adding new %s for %s with ID '%s': %s",
				labelVcfaProviderGatewayIpSpaceAssociations, labelVcfaIpSpace, addIpSpaceId, err)
		}
	}

	return nil
}

func removeIpSpaceAssociations(tmClient *VCDClient, providerGatewayId string, removeIpSpaceIds []string) error {
	existingIpSpaceAssociations, err := tmClient.GetAllTmIpSpaceAssociationsByProviderGatewayId(providerGatewayId)
	if err != nil {
		return fmt.Errorf("error reading %s for update: %s", labelVcfaProviderGatewayIpSpaceAssociations, err)
	}

	for _, singleIpSpaceId := range removeIpSpaceIds {
		for _, singleAssociation := range existingIpSpaceAssociations {
			if singleAssociation.TmIpSpaceAssociation.IPSpaceRef.ID == singleIpSpaceId {
				err = singleAssociation.Delete()
				if err != nil {
					return fmt.Errorf("error removing %s '%s' for %s '%s': %s",
						labelVcfaProviderGatewayIpSpaceAssociations, singleAssociation.TmIpSpaceAssociation.ID, labelVcfaIpSpace, singleIpSpaceId, err)
				}
			}
		}
	}

	return nil
}
