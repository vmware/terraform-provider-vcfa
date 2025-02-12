package vcfa

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const labelProject = "Project"

func resourceVcfaProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaProjectCreate,
		ReadContext:   resourceVcfaProjectRead,
		DeleteContext: resourceVcfaProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaProjectImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: fmt.Sprintf("Name of the %s", labelProject),
			},
			"description": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: fmt.Sprintf("Description of the %s", labelProject),
			},
		},
	}
}

func resourceVcfaProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cciClient := meta.(ClientContainer).cciClient
	if cciClient.VCDClient.Client.IsSysAdmin {
		return diag.Errorf("this resource requires Org user")
	}
	project := getprojectType(d)
	createdProject, err := cciClient.CreateProject(project)
	if err != nil {
		return diag.Errorf("error creating %s: %s", labelProject, err)
	}

	d.SetId(createdProject.Project.GetName())

	return resourceVcfaProjectRead(ctx, d, meta)
}

func resourceVcfaProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cciClient := meta.(ClientContainer).cciClient
	if cciClient.VCDClient.Client.IsSysAdmin {
		return diag.Errorf("this resource requires Org user")
	}
	project, err := cciClient.GetProjectByName(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			// entity is no more found - removing from state
			d.SetId("")
			return nil
		}
		return diag.Errorf("error retrieving %s '%s': %s", labelProject, d.Id(), err)
	}

	if err := setprojectData(d, project.Project); err != nil {
		return diag.Errorf("error setting %s data: %s", labelProject, err)
	}

	return nil
}

func resourceVcfaProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cciClient := meta.(ClientContainer).cciClient
	if cciClient.VCDClient.Client.IsSysAdmin {
		return diag.Errorf("this resource requires Org user")
	}
	project, err := cciClient.GetProjectByName(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving %s '%s' : %s", labelProject, d.Id(), err)
	}

	err = project.Delete()
	if err != nil {
		return diag.Errorf("error removing %s '%s' : %s", labelProject, d.Id(), err)
	}

	d.SetId("")

	return nil
}

func resourceVcfaProjectImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	cciClient := meta.(ClientContainer).cciClient
	idSlice := strings.Split(d.Id(), ImportSeparator)
	if len(idSlice) != 1 {
		return nil, fmt.Errorf("expected import ID to be <project_name>")
	}
	projectName := idSlice[0]

	project, err := cciClient.GetProjectByName(projectName)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", labelProject, err)
	}

	d.SetId(project.Project.Name)

	return []*schema.ResourceData{d}, nil
}

func getprojectType(d *schema.ResourceData) *ccitypes.Project {
	project := &ccitypes.Project{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.ProjectKind,
			APIVersion: ccitypes.ProjectCciAPI + "/" + ccitypes.ApiVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: d.Get("name").(string),
		},
		Spec: ccitypes.ProjectSpec{
			Description: d.Get("description").(string),
		},
	}

	return project
}

func setprojectData(d *schema.ResourceData, project *ccitypes.Project) error {
	d.SetId(project.Name)
	dSet(d, "name", project.Name)
	dSet(d, "description", project.Spec.Description)

	return nil
}
