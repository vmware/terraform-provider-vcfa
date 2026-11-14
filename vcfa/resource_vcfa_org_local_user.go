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

const labelOrgLocalUser = "Org Local User"

func resourceVcfaOrgLocalUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaOrgLocalUserCreate,
		ReadContext:   resourceVcfaOrgLocalUserRead,
		UpdateContext: resourceVcfaOrgLocalUserUpdate,
		DeleteContext: resourceVcfaOrgLocalUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaOrgLocalUserImport,
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaOrg, labelOrgLocalUser),
			},
			"role_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("%ss to use for %s", labelVcfaRole, labelOrgLocalUser),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s username", labelOrgLocalUser),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Password for %s", labelOrgLocalUser),
			},
		},
	}
}

func resourceVcfaOrgLocalUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	createFunc := func(config *types.OpenApiUser) (*govcd.OpenApiUser, error) {
		return tmClient.CreateUser(config, tenantContext)
	}

	c := crudConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:      labelOrgLocalUser,
		getTypeFunc:      getOrgLocalUserType,
		stateStoreFunc:   setOrgLocalUserData,
		createFunc:       createFunc,
		resourceReadFunc: resourceVcfaOrgLocalUserRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaOrgLocalUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	getByIdFunc := func(id string) (*govcd.OpenApiUser, error) {
		return tmClient.GetUserById(id, tenantContext)
	}

	c := crudConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:      labelOrgLocalUser,
		getTypeFunc:      getOrgLocalUserType,
		getEntityFunc:    getByIdFunc,
		resourceReadFunc: resourceVcfaOrgLocalUserRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaOrgLocalUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) { // Org no longer exists, therefore user is also gone
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	getByIdFunc := func(id string) (*govcd.OpenApiUser, error) {
		return tmClient.GetUserById(id, tenantContext)
	}

	c := crudConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:    labelOrgLocalUser,
		getEntityFunc:  getByIdFunc,
		stateStoreFunc: setOrgLocalUserData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaOrgLocalUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient

	tenantContext, err := getTenantContextFromOrgId(tmClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	getByIdFunc := func(id string) (*govcd.OpenApiUser, error) {
		return tmClient.GetUserById(id, tenantContext)
	}

	c := crudConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:   labelOrgLocalUser,
		getEntityFunc: getByIdFunc,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaOrgLocalUserImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient

	idSlice := strings.Split(d.Id(), ImportSeparator)
	if len(idSlice) != 2 {
		return nil, fmt.Errorf("expected import ID to be <org name>%s<user name>", ImportSeparator)
	}

	org, err := tmClient.GetTmOrgByName(idSlice[0])
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %s", labelVcfaOrg, err)
	}

	tenantContext := &govcd.TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	}

	user, err := tmClient.GetUserByName(idSlice[1], tenantContext)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s: %s", labelOrgLocalUser, err)
	}

	d.SetId(user.User.ID)
	dSet(d, "org_id", org.TmOrg.ID)

	return []*schema.ResourceData{d}, nil
}

func getOrgLocalUserType(tmClient *VCDClient, d *schema.ResourceData) (*types.OpenApiUser, error) {
	org, err := tmClient.GetTmOrgById(d.Get("org_id").(string))
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %s", labelVcfaOrg, err)
	}

	roleSet := convertSchemaSetToSliceOfStrings(d.Get("role_ids").(*schema.Set))
	t := &types.OpenApiUser{
		OrgEntityRef:   &types.OpenApiReference{ID: d.Get("org_id").(string), Name: org.TmOrg.Name},
		Username:       d.Get("username").(string),
		Password:       d.Get("password").(string),
		ProviderType:   "LOCAL",
		RoleEntityRefs: convertSliceOfStringsToOpenApiReferenceIds(roleSet),
		Locked:         addrOf(false),
	}

	// Update requires ID being present
	if d.Id() != "" { // update operation
		t.ID = d.Id()
		user, err := tmClient.GetUserById(d.Id(), nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving %s by ID %s: %s", labelOrgLocalUser, d.Id(), err)
		}
		// Name In Source must be set to "previous" username when performing update
		t.NameInSource = user.User.Username

		// if password has not changed - send exactly '******' to prevent updating password just like UI
		if !d.HasChange("password") {
			t.Password = "******"
		}
	}

	return t, nil
}

func setOrgLocalUserData(_ *VCDClient, d *schema.ResourceData, user *govcd.OpenApiUser) error {
	if user == nil || user.User == nil {
		return fmt.Errorf("nil user structure")
	}
	d.SetId(user.User.ID)
	dSet(d, "username", user.User.Username)

	roleIds := extractIdsFromOpenApiReferences(user.User.RoleEntityRefs)
	err := d.Set("role_ids", roleIds)
	if err != nil {
		return fmt.Errorf("error storing 'role_ids': %s", err)
	}

	dSet(d, "org_id", "")
	if user.User.OrgEntityRef != nil {
		dSet(d, "org_id", user.User.OrgEntityRef.ID)
	}
	// dSet(d, "password", user.User.Password) // password is never returned on read

	return nil
}
