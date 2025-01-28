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

const labelVcfaRightsBundle = "Rights Bundle"

func resourceVcfaRightsBundle() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaRightsBundleCreate,
		ReadContext:   resourceVcfaRightsBundleRead,
		UpdateContext: resourceVcfaRightsBundleUpdate,
		DeleteContext: resourceVcfaRightsBundleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaVcfaRightsBundleImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaRightsBundle),
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s description", labelVcfaRightsBundle),
			},
			"bundle_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is read-only", labelVcfaRightsBundle),
			},
			"rights": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: fmt.Sprintf("Set of %ss assigned to this %s", labelVcfaRight, labelVcfaRightsBundle),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"publish_to_all_orgs": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: fmt.Sprintf("When true, publishes the %s to all %ss", labelVcfaRightsBundle, labelVcfaOrg),
			},
			"org_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: fmt.Sprintf("Set of %ss IDs to which this %s is published", labelVcfaOrg, labelVcfaRightsBundle),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}
func resourceVcfaRightsBundleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	rightsBundleName := d.Get("name").(string)
	publishToAllTenants := d.Get("publish_to_all_orgs").(bool)

	inputRights, err := getRights(vcdClient, nil, fmt.Sprintf("%s create", labelVcfaRightsBundle), d)
	if err != nil {
		return diag.FromErr(err)
	}
	rightsBundle, err := vcdClient.Client.CreateRightsBundle(&types.RightsBundle{
		Name:        rightsBundleName,
		Description: d.Get("description").(string),
		BundleKey:   types.VcloudUndefinedKey,
		PublishAll:  &publishToAllTenants,
	})
	if err != nil {
		return diag.Errorf("[%s create] error creating role %s: %s", labelVcfaRightsBundle, rightsBundleName, err)
	}
	if len(inputRights) > 0 {
		err = rightsBundle.AddRights(inputRights)
		if err != nil {
			return diag.Errorf("[%s create] error adding rights to %s %s: %s", labelVcfaRightsBundle, labelVcfaRightsBundle, rightsBundleName, err)
		}
	}

	inputTenants, err := getOrganizations(vcdClient, fmt.Sprintf("%s create", labelVcfaRightsBundle), d)
	if err != nil {
		return diag.FromErr(err)
	}
	if publishToAllTenants {
		err = rightsBundle.PublishAllTenants()
		if err != nil {
			return diag.Errorf("[%s create] error publishing to all %ss - %s %s: %s", labelVcfaRightsBundle, labelVcfaOrg, labelVcfaRightsBundle, rightsBundleName, err)
		}
	}
	if len(inputTenants) > 0 {
		err = rightsBundle.PublishTenants(inputTenants)
		if err != nil {
			return diag.Errorf("[%s create] error publishing to %s - %s %s: %s", labelVcfaRightsBundle, labelVcfaOrg, labelVcfaRightsBundle, rightsBundleName, err)
		}
	}
	d.SetId(rightsBundle.RightsBundle.Id)
	return genericVcfaRightsBundleRead(ctx, d, meta, "resource", "create")
}

func resourceVcfaRightsBundleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaRightsBundleRead(ctx, d, meta, "resource", "read")
}

func genericVcfaRightsBundleRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin, operation string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	var rightsBundle *govcd.RightsBundle
	var err error
	rightsBundleName := d.Get("name").(string)
	identifier := d.Id()

	if identifier == "" {
		rightsBundle, err = vcdClient.Client.GetRightsBundleByName(rightsBundleName)
	} else {
		rightsBundle, err = vcdClient.Client.GetRightsBundleById(identifier)
	}
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[%s read-%s] error retrieving %s %s: %s", labelVcfaRightsBundle, operation, labelVcfaRightsBundle, rightsBundleName, err)
	}

	d.SetId(rightsBundle.RightsBundle.Id)
	dSet(d, "description", rightsBundle.RightsBundle.Description)
	dSet(d, "bundle_key", rightsBundle.RightsBundle.BundleKey)
	dSet(d, "read_only", rightsBundle.RightsBundle.ReadOnly)

	rights, err := rightsBundle.GetRights(nil)
	if err != nil {
		return diag.Errorf("[%s read-%s] error while querying %s rights: %s", labelVcfaRightsBundle, operation, labelVcfaRightsBundle, err)
	}
	var assignedRights []interface{}

	for _, right := range rights {
		assignedRights = append(assignedRights, right.Name)
	}
	if len(assignedRights) > 0 {
		err = d.Set("rights", assignedRights)
		if err != nil {
			return diag.Errorf("[%s read-%s] error setting rights for %s %s: %s", labelVcfaRightsBundle, operation, labelVcfaRightsBundle, rightsBundleName, err)
		}
	}

	orgs, err := rightsBundle.GetTenants(nil)
	if err != nil {
		return diag.Errorf("[%s read-%s] error while querying %s %ss: %s", labelVcfaRightsBundle, operation, labelVcfaRightsBundle, labelVcfaOrg, err)
	}
	var registeredTenants []interface{}

	publishAll := false
	if rightsBundle.RightsBundle.PublishAll != nil {
		publishAll = *rightsBundle.RightsBundle.PublishAll
	}
	dSet(d, "publish_to_all_orgs", publishAll)
	for _, org := range orgs {
		registeredTenants = append(registeredTenants, org.ID)
	}
	if !publishAll {
		if len(registeredTenants) > 0 {
			err = d.Set("org_ids", registeredTenants)
			if err != nil {
				return diag.Errorf("[%s read-%s] error setting %s for %s %s: %s", labelVcfaRightsBundle, operation, labelVcfaOrg, labelVcfaRightsBundle, rightsBundleName, err)
			}
		}
	}

	return nil
}

func resourceVcfaRightsBundleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rightsBundleName := d.Get("name").(string)
	publishToAllTenants := d.Get("publish_to_all_orgs").(bool)

	rightsBundle, err := vcdClient.Client.GetRightsBundleById(d.Id())
	if err != nil {
		return diag.Errorf("[%s update] error retrieving %s %s: %s", labelVcfaRightsBundle, labelVcfaRightsBundle, rightsBundleName, err)
	}

	var inputRights []types.OpenApiReference
	var inputTenants []types.OpenApiReference
	var changedRights = d.HasChange("rights")
	var changedTenants = d.HasChange("org_ids") || d.HasChange("publish_to_all_orgs")
	if changedRights {
		inputRights, err = getRights(vcdClient, nil, fmt.Sprintf("%s update", labelVcfaRightsBundle), d)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("publish_to_all_orgs") {
		rightsBundle.RightsBundle.Name = rightsBundleName
		rightsBundle.RightsBundle.Description = d.Get("description").(string)
		rightsBundle.RightsBundle.PublishAll = &publishToAllTenants
		_, err = rightsBundle.Update()
		if err != nil {
			return diag.Errorf("[%s update] error updating %s %s: %s", labelVcfaRightsBundle, labelVcfaRightsBundle, rightsBundleName, err)
		}
	}

	if changedRights {
		if len(inputRights) > 0 {
			err = rightsBundle.UpdateRights(inputRights)
			if err != nil {
				return diag.Errorf("[%s update] error updating %s %s rights: %s", labelVcfaRightsBundle, labelVcfaRightsBundle, rightsBundleName, err)
			}
		} else {
			currentRights, err := rightsBundle.GetRights(nil)
			if err != nil {
				return diag.Errorf("[%s update] error retrieving %s %s rights: %s", labelVcfaRightsBundle, labelVcfaRightsBundle, rightsBundleName, err)
			}
			if len(currentRights) > 0 {
				err = rightsBundle.RemoveAllRights()
				if err != nil {
					return diag.Errorf("[%s update] error removing %s %s rights: %s", labelVcfaRightsBundle, labelVcfaRightsBundle, rightsBundleName, err)
				}
			}
		}
	}
	if changedTenants {
		inputTenants, err = getOrganizations(vcdClient, fmt.Sprintf("%s create", labelVcfaRightsBundle), d)
		if err != nil {
			return diag.FromErr(err)
		}
		if publishToAllTenants {
			err = rightsBundle.PublishAllTenants()
			if err != nil {
				return diag.Errorf("[%s update] error publishing to all %ss - %s %s: %s", labelVcfaRightsBundle, labelVcfaOrg, labelVcfaRightsBundle, rightsBundleName, err)
			}
		} else {
			if len(inputTenants) > 0 {
				err = rightsBundle.ReplacePublishedTenants(inputTenants)
				if err != nil {
					return diag.Errorf("[%s update] error publishing to %ss - %s %s: %s", labelVcfaRightsBundle, labelVcfaOrg, labelVcfaRightsBundle, rightsBundleName, err)
				}
			} else {
				err = rightsBundle.UnpublishAllTenants()
				if err != nil {
					return diag.Errorf("[%s update] error unpublishing from all %ss - %s %s: %s", labelVcfaRightsBundle, labelVcfaOrg, labelVcfaRightsBundle, rightsBundleName, err)
				}
			}
		}
	}

	return genericVcfaRightsBundleRead(ctx, d, meta, "resource", "update")
}

func resourceVcfaRightsBundleDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	rightsBundleName := d.Get("name").(string)

	var rightsBundle *govcd.RightsBundle
	var err error
	identifier := d.Id()
	if identifier == "" {
		rightsBundle, err = vcdClient.Client.GetRightsBundleByName(rightsBundleName)
	} else {
		rightsBundle, err = vcdClient.Client.GetRightsBundleById(identifier)
	}

	if err != nil {
		return diag.Errorf("[%s delete] error retrieving %s %s: %s", labelVcfaRightsBundle, labelVcfaRightsBundle, rightsBundleName, err)
	}

	err = rightsBundle.Delete()
	if err != nil {
		return diag.Errorf("[%s delete] error deleting %s %s: %s", labelVcfaRightsBundle, labelVcfaRightsBundle, rightsBundleName, err)
	}
	return nil
}

func resourceVcfaVcfaRightsBundleImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceVcfaURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceVcfaURI) != 1 {
		return nil, fmt.Errorf("resource name must be specified as rightsBundle-name")
	}
	rightsBundleName := resourceVcfaURI[0]

	vcdClient := meta.(*VCDClient)

	rightsBundle, err := vcdClient.Client.GetRightsBundleByName(rightsBundleName)
	if err != nil {
		return nil, fmt.Errorf("[%s import] error retrieving %s %s: %s", labelVcfaRightsBundle, labelVcfaRightsBundle, rightsBundleName, err)
	}
	dSet(d, "name", rightsBundleName)
	dSet(d, "description", rightsBundle.RightsBundle.Description)
	dSet(d, "bundle_key", rightsBundle.RightsBundle.BundleKey)

	publishAll := false
	if rightsBundle.RightsBundle.PublishAll != nil {
		publishAll = *rightsBundle.RightsBundle.PublishAll
	}
	dSet(d, "publish_to_all_orgs", publishAll)
	d.SetId(rightsBundle.RightsBundle.Id)
	return []*schema.ResourceData{d}, nil
}

// getOrganizations returns a list of Organizations for provider level rights containers (global role, rights bundle)
func getOrganizations(client *VCDClient, label string, d *schema.ResourceData) ([]types.OpenApiReference, error) {
	var inputTenants []types.OpenApiReference

	orgIds := d.Get("org_ids").(*schema.Set).List()

	for _, oi := range orgIds {
		tenantId := oi.(string)
		org, err := client.GetTmOrgById(tenantId)
		if err != nil {
			return nil, fmt.Errorf("[%s] error retrieving tenant %s: %s", label, oi, err)
		}
		inputTenants = append(inputTenants, types.OpenApiReference{Name: org.TmOrg.Name, ID: tenantId})
	}
	return inputTenants, nil
}

// getRights will collect the list of rights of a rights collection (role, global role, rights bundle)
// and check whether the necessary implied rights are included.
// Calling resources should provide a client and optionally an Org (role)
// The "label" identifies the calling resource and operation and it is used to form error messages
func getRights(client *VCDClient, org *govcd.AdminOrg, label string, d *schema.ResourceData) ([]types.OpenApiReference, error) {
	var inputRights []types.OpenApiReference

	if client == nil {
		return nil, fmt.Errorf("[getRights - %s] client was empty", label)
	}
	rights := d.Get("rights").(*schema.Set).List()

	var right *types.Right
	var err error

	for _, r := range rights {
		rn := r.(string)
		if org != nil {
			right, err = org.GetRightByName(rn)
		} else {
			right, err = client.Client.GetRightByName(rn)
		}
		if err != nil {
			return nil, fmt.Errorf("[%s] error retrieving %s %s: %s", label, labelVcfaRight, rn, err)
		}
		inputRights = append(inputRights, types.OpenApiReference{Name: rn, ID: right.ID})
	}

	missingImpliedRights, err := govcd.FindMissingImpliedRights(&client.Client, inputRights)
	if err != nil {
		return nil, fmt.Errorf("[%s] error inspecting implied %ss: %s", label, labelVcfaRight, err)
	}

	if len(missingImpliedRights) > 0 {
		message := "The rights set requires the following implied rights to be added:"
		rightsList := ""
		for _, right := range missingImpliedRights {
			rightsList += fmt.Sprintf("\"%s\",\n", right.Name)
		}
		return nil, fmt.Errorf("%s\n%s", message, rightsList)
	}
	return inputRights, nil
}
