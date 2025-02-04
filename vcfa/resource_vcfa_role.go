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

const labelVcfaRole = "Role"

func resourceVcfaRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaRoleCreate,
		ReadContext:   resourceVcfaRoleRead,
		UpdateContext: resourceVcfaRoleUpdate,
		DeleteContext: resourceVcfaRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaRoleImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaRole),
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("The ID of the %s of the %s", labelVcfaOrg, labelVcfaRole),
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s description", labelVcfaRole),
			},
			"bundle_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"read_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether this %s is read-only", labelVcfaRole),
			},
			"rights": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: fmt.Sprintf("Set of %ss assigned to this %s", labelVcfaRight, labelVcfaRole),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVcfaRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	roleName := d.Get("name").(string)
	orgId := d.Get("org_id").(string)

	// TODO: TM: Change to tmClient.GetTmOrgById(orgId), requires implementing Role support for that type
	org, err := tmClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[%s create] error retrieving %s '%s': %s", labelVcfaRole, labelVcfaOrg, orgId, err)
	}

	// Check rights early, so that we can show a friendly error message when there are missing implied rights
	inputRights, err := getRights(tmClient, org, fmt.Sprintf("%s create", labelVcfaRole), d)
	if err != nil {
		return diag.FromErr(err)
	}

	role, err := org.CreateRole(&types.Role{
		Name:        roleName,
		Description: d.Get("description").(string),
		BundleKey:   types.VcloudUndefinedKey,
	})
	if err != nil {
		return diag.Errorf("[%s create] error creating %s '%s': %s", labelVcfaRole, labelVcfaRole, roleName, err)
	}
	if len(inputRights) > 0 {
		err = role.AddRights(inputRights)
		if err != nil {
			return diag.Errorf("[%s create] error adding %ss to %s '%s': %s", labelVcfaRole, labelVcfaRight, labelVcfaRole, roleName, err)
		}
	}

	d.SetId(role.Role.ID)
	return genericVcfaRoleRead(ctx, d, meta, "resource", "create")
}

func resourceVcfaRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcfaRoleRead(ctx, d, meta, "resource", "read")
}

func genericVcfaRoleRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin, operation string) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	roleName := d.Get("name").(string)
	orgId := d.Get("org_id").(string)
	identifier := d.Id()

	var role *govcd.Role
	var err error

	// TODO: TM: Change to tmClient.GetTmOrgById(orgId), requires implementing Role support for that type
	org, err := tmClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[%s %s-%s] error retrieving %s '%s': %s", labelVcfaRole, operation, origin, labelVcfaOrg, orgId, err)
	}

	if identifier == "" {
		role, err = org.GetRoleByName(roleName)
	} else {
		role, err = org.GetRoleById(identifier)
	}
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(role.Role.ID)
	dSet(d, "description", role.Role.Description)
	dSet(d, "bundle_key", role.Role.BundleKey)
	dSet(d, "read_only", role.Role.ReadOnly)

	rights, err := role.GetRights(nil)
	if err != nil {
		return diag.Errorf("[%s %s-%s] error while querying %s %ss: %s", labelVcfaRole, operation, origin, labelVcfaRole, labelVcfaRight, err)

	}
	var assignedRights []interface{}

	for _, right := range rights {
		assignedRights = append(assignedRights, right.Name)
	}
	if len(assignedRights) > 0 {
		err = d.Set("rights", assignedRights)
		if err != nil {
			return diag.Errorf("[%s %s-%s] error setting %ss for %s '%s': %s", labelVcfaRole, operation, origin, labelVcfaRight, labelVcfaRole, roleName, err)
		}
	}
	return nil
}

func resourceVcfaRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	roleName := d.Get("name").(string)
	orgId := d.Get("org_id").(string)

	// TODO: TM: Change to tmClient.GetTmOrgById(orgId), requires implementing Role support for that type
	org, err := tmClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[%s update] error retrieving %s '%s': %s", labelVcfaRole, labelVcfaOrg, orgId, err)
	}

	role, err := org.GetRoleById(d.Id())
	if err != nil {
		return diag.Errorf("[%s update] error retrieving %s '%s': %s", labelVcfaRole, labelVcfaRole, roleName, err)
	}

	if d.HasChange("name") || d.HasChange("description") {
		role.Role.Name = roleName
		role.Role.Description = d.Get("description").(string)
		_, err = role.Update()
		if err != nil {
			return diag.Errorf("[%s update] error updating %s '%s': %s", labelVcfaRole, labelVcfaRole, roleName, err)
		}
	}

	inputRights, err := getRights(tmClient, org, fmt.Sprintf("%s update", labelVcfaRole), d)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(inputRights) > 0 {
		err = role.UpdateRights(inputRights)
		if err != nil {
			return diag.Errorf("[%s update] error updating %s '%s' %ss: %s", labelVcfaRole, labelVcfaRole, roleName, labelVcfaRight, err)
		}
	} else {
		currentRights, err := role.GetRights(nil)
		if err != nil {
			return diag.Errorf("[%s update] error retrieving %s '%s' %ss: %s", labelVcfaRole, labelVcfaRole, roleName, labelVcfaRight, err)
		}
		if len(currentRights) > 0 {
			err = role.RemoveAllRights()
			if err != nil {
				return diag.Errorf("[%s update] error removing %s '%s' %ss: %s", labelVcfaRole, labelVcfaRole, roleName, labelVcfaRight, err)
			}
		}
	}
	return genericVcfaRoleRead(ctx, d, meta, "resource", "update")
}

func resourceVcfaRoleDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	roleName := d.Get("name").(string)
	orgId := d.Get("org_id").(string)

	// TODO: TM: Change to tmClient.GetTmOrgById(orgId), requires implementing Role support for that type
	org, err := tmClient.GetAdminOrgById(orgId)
	if err != nil {
		return diag.Errorf("[%s delete] error retrieving %s '%s': %s", labelVcfaRole, labelVcfaOrg, orgId, err)
	}
	var role *govcd.Role
	identifier := d.Id()
	if identifier != "" {
		role, err = org.GetRoleById(identifier)
	} else {
		role, err = org.GetRoleByName(roleName)
	}
	if err != nil {
		return diag.Errorf("[%s delete] error retrieving %s '%s': %s", labelVcfaRole, labelVcfaRole, roleName, err)
	}
	err = role.Delete()
	if err != nil {
		return diag.Errorf("[%s delete] error deleting %s '%s': %s", labelVcfaRole, labelVcfaRole, roleName, err)
	}
	return nil
}

func resourceVcfaRoleImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 2 {
		return nil, fmt.Errorf("resource name must be specified as org-name%srole-name", ImportSeparator)
	}
	orgName, roleName := resourceURI[0], resourceURI[1]

	tmClient := meta.(ClientContainer).tmClient

	// TODO: TM: Change to tmClient.GetTmOrgByName(orgName), requires implementing Role support for that type
	org, err := tmClient.GetAdminOrgByName(orgName)
	if err != nil {
		return nil, fmt.Errorf("[%s import] error retrieving %s '%s': %s", labelVcfaRole, labelVcfaOrg, orgName, err)
	}

	role, err := org.GetRoleByName(roleName)
	if err != nil {
		return nil, fmt.Errorf("[%s import] error retrieving %s '%s': %s", labelVcfaRole, labelVcfaRole, roleName, err)
	}
	dSet(d, "org_id", org.AdminOrg.ID)
	dSet(d, "name", roleName)
	dSet(d, "description", role.Role.Description)
	dSet(d, "bundle_key", role.Role.BundleKey)
	d.SetId(role.Role.ID)
	return []*schema.ResourceData{d}, nil
}
