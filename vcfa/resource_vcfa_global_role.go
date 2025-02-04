package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelVcfaGlobalRole = "Global Role"

func resourceVcfaGlobalRole() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcfaGlobalRoleRead,
		CreateContext: resourceVcfaGlobalRoleCreate,
		UpdateContext: resourceVcfaGlobalRoleUpdate,
		DeleteContext: resourceVcfaGlobalRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaGlobalRoleImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaGlobalRole),
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s description", labelVcfaGlobalRole),
			},
			"bundle_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is read-only", labelVcfaGlobalRole),
			},
			"rights": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: fmt.Sprintf("List of %ss assigned to this %s", labelVcfaRight, labelVcfaGlobalRole),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"publish_to_all_orgs": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: fmt.Sprintf("When true, publishes the %s to all %ss", labelVcfaGlobalRole, labelVcfaOrg),
			},
			"org_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: fmt.Sprintf("List of IDs of %ss to which this %s is published", labelVcfaOrg, labelVcfaGlobalRole),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVcfaGlobalRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).VcfaClient

	globalRoleName := d.Get("name").(string)
	publishToAllOrgs := d.Get("publish_to_all_orgs").(bool)

	inputRights, err := getRights(vcfaClient, nil, fmt.Sprintf("%s create", labelVcfaGlobalRole), d)
	if err != nil {
		return diag.FromErr(err)
	}
	globalRole, err := vcfaClient.Client.CreateGlobalRole(&types.GlobalRole{
		Name:        globalRoleName,
		Description: d.Get("description").(string),
		BundleKey:   types.VcloudUndefinedKey,
		PublishAll:  &publishToAllOrgs,
	})
	if err != nil {
		return diag.Errorf("[%s create] error creating %s '%s': %s", labelVcfaGlobalRole, labelVcfaGlobalRole, globalRoleName, err)
	}
	if len(inputRights) > 0 {
		err = globalRole.AddRights(inputRights)
		if err != nil {
			return diag.Errorf("[%s create] error adding %ss to %s '%s': %s", labelVcfaGlobalRole, labelVcfaRight, labelVcfaGlobalRole, globalRoleName, err)
		}
	}

	inputTenants, err := getOrganizations(vcfaClient, fmt.Sprintf("%s create", labelVcfaGlobalRole), d)
	if err != nil {
		return diag.FromErr(err)
	}
	if publishToAllOrgs {
		err = globalRole.PublishAllTenants()
		if err != nil {
			return diag.Errorf("[%s create] error publishing to all %ss - %s '%s': %s", labelVcfaGlobalRole, labelVcfaOrg, labelVcfaGlobalRole, globalRoleName, err)
		}
	}
	if len(inputTenants) > 0 {
		err = globalRole.PublishTenants(inputTenants)
		if err != nil {
			return diag.Errorf("[%s create] error publishing to %ss - %s '%s': %s", labelVcfaGlobalRole, labelVcfaOrg, labelVcfaGlobalRole, globalRoleName, err)
		}
	}
	d.SetId(globalRole.GlobalRole.Id)
	return genericGlobalRoleRead(ctx, d, meta, "resource", "create")
}

func resourceVcfaGlobalRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericGlobalRoleRead(ctx, d, meta, "resource", "read")
}

func genericGlobalRoleRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin, operation string) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).VcfaClient

	var globalRole *govcd.GlobalRole
	var err error
	globalRoleName := d.Get("name").(string)
	identifier := d.Id()
	if identifier == "" {
		globalRole, err = vcfaClient.Client.GetGlobalRoleByName(globalRoleName)
	} else {
		globalRole, err = vcfaClient.Client.GetGlobalRoleById(identifier)
	}

	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[%s read-%s] error retrieving %s '%s': %s", labelVcfaGlobalRole, operation, labelVcfaGlobalRole, globalRoleName, err)
	}

	publishAll := false
	if globalRole.GlobalRole.PublishAll != nil {
		publishAll = *globalRole.GlobalRole.PublishAll
	}
	d.SetId(globalRole.GlobalRole.Id)
	dSet(d, "description", globalRole.GlobalRole.Description)
	dSet(d, "bundle_key", globalRole.GlobalRole.BundleKey)
	dSet(d, "read_only", globalRole.GlobalRole.ReadOnly)
	err = d.Set("publish_to_all_orgs", publishAll)
	if err != nil {
		return diag.Errorf("[%s read-%s] error setting publish_to_all_orgs: %s", labelVcfaGlobalRole, operation, err)
	}

	rights, err := globalRole.GetRights(nil)
	if err != nil {
		return diag.Errorf("[%s read-%s] error while querying %s %ss: %s", labelVcfaGlobalRole, operation, labelVcfaGlobalRole, labelVcfaRight, err)
	}
	var assignedRights []interface{}

	for _, right := range rights {
		assignedRights = append(assignedRights, right.Name)
	}
	if len(assignedRights) > 0 {
		err = d.Set("rights", assignedRights)
		if err != nil {
			return diag.Errorf("[%s read-%s] error setting %ss for %s '%s': %s", labelVcfaGlobalRole, operation, labelVcfaRight, labelVcfaGlobalRole, globalRoleName, err)
		}
	}

	orgs, err := globalRole.GetTenants(nil)
	if err != nil {
		return diag.Errorf("[%s read-%s] error while querying %s %ss: %s", labelVcfaGlobalRole, operation, labelVcfaGlobalRole, labelVcfaOrg, err)
	}
	var registeredTenants []interface{}

	for _, org := range orgs {
		registeredTenants = append(registeredTenants, org.ID)
	}
	if len(registeredTenants) > 0 {
		err = d.Set("org_ids", registeredTenants)
		if err != nil {
			return diag.Errorf("[%s read-%s] error setting %ss for %s '%s': %s", labelVcfaGlobalRole, operation, labelVcfaOrg, labelVcfaGlobalRole, globalRoleName, err)
		}
	}

	return nil
}

func resourceVcfaGlobalRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).VcfaClient

	globalRoleName := d.Get("name").(string)

	publishToAllTenants := d.Get("publish_to_all_orgs").(bool)

	globalRole, err := vcfaClient.Client.GetGlobalRoleById(d.Id())
	if err != nil {
		return diag.Errorf("[%s update] error retrieving %s '%s': %s", labelVcfaGlobalRole, labelVcfaGlobalRole, globalRoleName, err)
	}

	var inputRights []types.OpenApiReference
	var inputTenants []types.OpenApiReference
	var changedRights = d.HasChange("rights")
	var changedTenants = d.HasChange("org_ids") || d.HasChange("publish_to_all_orgs")

	if changedRights {
		inputRights, err = getRights(vcfaClient, nil, fmt.Sprintf("%s update", labelVcfaGlobalRole), d)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name") || d.HasChange("description") || d.HasChange("publish_to_all_orgs") {
		globalRole.GlobalRole.Name = globalRoleName
		globalRole.GlobalRole.Description = d.Get("description").(string)
		globalRole.GlobalRole.PublishAll = &publishToAllTenants
		_, err = globalRole.Update()
		if err != nil {
			return diag.Errorf("[%s update] error updating %s '%s': %s", labelVcfaGlobalRole, labelVcfaGlobalRole, globalRoleName, err)
		}
	}

	if changedRights {
		if len(inputRights) > 0 {
			err = globalRole.UpdateRights(inputRights)
			if err != nil {
				return diag.Errorf("[%s update] error updating %s '%s' %ss: %s", labelVcfaGlobalRole, labelVcfaGlobalRole, globalRoleName, labelVcfaRight, err)
			}
		} else {
			currentRights, err := globalRole.GetRights(nil)
			if err != nil {
				return diag.Errorf("[%s update] error retrieving %s '%s' %ss: %s", labelVcfaGlobalRole, labelVcfaGlobalRole, globalRoleName, labelVcfaRight, err)
			}
			if len(currentRights) > 0 {
				err = globalRole.RemoveAllRights()
				if err != nil {
					return diag.Errorf("[%s update] error removing %s '%s' %ss: %s", labelVcfaGlobalRole, labelVcfaGlobalRole, globalRoleName, labelVcfaRight, err)
				}
			}
		}
	}
	if changedTenants {
		inputTenants, err = getOrganizations(vcfaClient, fmt.Sprintf("%s create", labelVcfaGlobalRole), d)
		if err != nil {
			return diag.FromErr(err)
		}
		if publishToAllTenants {
			err = globalRole.PublishAllTenants()
			if err != nil {
				return diag.Errorf("[%s update] error publishing to all %ss - %s '%s': %s", labelVcfaGlobalRole, labelVcfaOrg, labelVcfaGlobalRole, globalRoleName, err)
			}
		} else {
			if len(inputTenants) > 0 {
				err = globalRole.ReplacePublishedTenants(inputTenants)
				if err != nil {
					return diag.Errorf("[%s update] error publishing to %ss - %s '%s': %s", labelVcfaGlobalRole, labelVcfaOrg, labelVcfaGlobalRole, globalRoleName, err)
				}
			} else {
				if !publishToAllTenants {
					err = globalRole.UnpublishAllTenants()
					if err != nil {
						return diag.Errorf("[%s update] error unpublishing from all %ss - %s '%s': %s", labelVcfaGlobalRole, labelVcfaOrg, labelVcfaGlobalRole, globalRoleName, err)
					}
				}
			}
		}
	}

	return genericGlobalRoleRead(ctx, d, meta, "resource", "update")
}

func resourceVcfaGlobalRoleDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).VcfaClient

	globalRoleName := d.Get("name").(string)

	var globalRole *govcd.GlobalRole
	var err error
	identifier := d.Id()
	if identifier == "" {
		globalRole, err = vcfaClient.Client.GetGlobalRoleByName(globalRoleName)
	} else {
		globalRole, err = vcfaClient.Client.GetGlobalRoleById(identifier)
	}

	if err != nil {
		return diag.Errorf("[%s delete] error retrieving %s '%s': %s", labelVcfaGlobalRole, labelVcfaGlobalRole, globalRoleName, err)
	}

	err = globalRole.Delete()
	if err != nil {
		return diag.Errorf("[%s delete] error deleting %s '%s': %s", labelVcfaGlobalRole, labelVcfaGlobalRole, globalRoleName, err)
	}
	return nil
}

func resourceVcfaGlobalRoleImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcfaClient := meta.(ClientContainer).VcfaClient
	globalRole, err := vcfaClient.Client.GetGlobalRoleByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("[%s import] error retrieving %s '%s': %s", labelVcfaGlobalRole, labelVcfaGlobalRole, d.Id(), err)
	}
	dSet(d, "name", globalRole.GlobalRole.Name)
	dSet(d, "description", globalRole.GlobalRole.Description)
	dSet(d, "bundle_key", globalRole.GlobalRole.BundleKey)
	publishAll := false
	if globalRole.GlobalRole.PublishAll != nil {
		publishAll = *globalRole.GlobalRole.PublishAll
	}
	dSet(d, "publish_to_all_orgs", publishAll)
	d.SetId(globalRole.GlobalRole.Id)
	return []*schema.ResourceData{d}, nil
}
