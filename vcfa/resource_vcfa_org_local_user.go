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

const labelLocalUser = "Org Local User"

func resourceVcfaLocalUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaLocalUserCreate,
		ReadContext:   resourceVcfaLocalUserRead,
		UpdateContext: resourceVcfaLocalUserUpdate,
		DeleteContext: resourceVcfaLocalUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaLocalUserImport,
		},

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Parent %s ID for %s", labelVcfaOrg, labelLocalUser),
			},
			"role_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("%ss to use for %s", labelVcfaRole, labelLocalUser),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s username", labelLocalUser),
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: fmt.Sprintf("Password for %s", labelLocalUser),
			},
		},
	}
}

func resourceVcfaLocalUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(vcfaClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	createFunc := func(config *types.OpenApiUser) (*govcd.OpenApiUser, error) {
		return vcfaClient.CreateUser(config, tenantContext)
	}

	c := crudConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:      labelLocalUser,
		getTypeFunc:      getLocalUserType,
		stateStoreFunc:   setLocalUserData,
		createFunc:       createFunc,
		resourceReadFunc: resourceVcfaLocalUserRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcfaLocalUserUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(vcfaClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	getByIdFunc := func(id string) (*govcd.OpenApiUser, error) {
		return vcfaClient.GetUserById(id, tenantContext)
	}

	c := crudConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:      labelLocalUser,
		getTypeFunc:      getLocalUserType,
		getEntityFunc:    getByIdFunc,
		resourceReadFunc: resourceVcfaLocalUserRead,
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcfaLocalUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).tmClient
	tenantContext, err := getTenantContextFromOrgId(vcfaClient, d.Get("org_id").(string))
	if err != nil {
		if govcd.ContainsNotFound(err) { // Org no longer exists, therefore user is also gone
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	getByIdFunc := func(id string) (*govcd.OpenApiUser, error) {
		return vcfaClient.GetUserById(id, tenantContext)
	}

	c := crudConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:    labelLocalUser,
		getEntityFunc:  getByIdFunc,
		stateStoreFunc: setLocalUserData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcfaLocalUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcfaClient := meta.(ClientContainer).tmClient

	tenantContext, err := getTenantContextFromOrgId(vcfaClient, d.Get("org_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	getByIdFunc := func(id string) (*govcd.OpenApiUser, error) {
		return vcfaClient.GetUserById(id, tenantContext)
	}

	c := crudConfig[*govcd.OpenApiUser, types.OpenApiUser]{
		entityLabel:   labelLocalUser,
		getEntityFunc: getByIdFunc,
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcfaLocalUserImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcfaClient := meta.(ClientContainer).tmClient

	idSlice := strings.Split(d.Id(), ImportSeparator)
	if len(idSlice) != 2 {
		return nil, fmt.Errorf("expected import ID to be <org name>%s<user name>", ImportSeparator)
	}

	org, err := vcfaClient.GetTmOrgByName(idSlice[0])
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %s", labelVcfaOrg, err)
	}

	tenantContext := &govcd.TenantContext{
		OrgId:   org.TmOrg.ID,
		OrgName: org.TmOrg.Name,
	}

	user, err := vcfaClient.GetUserByName(idSlice[1], tenantContext)
	if err != nil {
		return nil, fmt.Errorf("error retrieving %s: %s", labelLocalUser, err)
	}

	d.SetId(user.User.ID)
	dSet(d, "org_id", org.TmOrg.ID)

	return []*schema.ResourceData{d}, nil
}

func getLocalUserType(vcfaClient *VCDClient, d *schema.ResourceData) (*types.OpenApiUser, error) {
	org, err := vcfaClient.GetTmOrgById(d.Get("org_id").(string))
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
		user, err := vcfaClient.GetUserById(d.Id(), nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving %s by ID %s: %s", labelLocalUser, d.Id(), err)
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

func setLocalUserData(_ *VCDClient, d *schema.ResourceData, user *govcd.OpenApiUser) error {
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
