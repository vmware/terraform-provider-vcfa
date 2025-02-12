package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func datasourceVcfaSupervisorNamespace() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaSupervisorNamespaceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelSupervisorNamespace),
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringIsNotEmpty,
				),
			},
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The name of the Project the %s belongs to", labelSupervisorNamespace),
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringIsNotEmpty,
				),
			},
			"class_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Supervisor Namespace Class",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description",
			},
			"phase": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Phase of the %s", labelSupervisorNamespace),
			},
			"ready": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Whether the %s is in a ready status or not", labelSupervisorNamespace),
			},
			"region_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Name of the %s", labelVcfaRegion),
			},
			"storage_classes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("%s Storage Classes", labelSupervisorNamespace),
				Elem:        supervisorNamespaceStorageClassesSchema,
			},
			"storage_classes_initial_class_config_overrides": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Initial Class Config Overrides for Storage Classes",
				Elem:        supervisorNamespaceStorageClassesInitialClassConfigOverridesSchema,
			},
			"vm_classes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("%s VM Classes", labelSupervisorNamespace),
				Elem:        supervisorNamespaceVMClassesSchema,
			},
			"vpc_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the VPC",
			},
			"zones": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("%s Zones", labelSupervisorNamespace),
				Elem:        supervisorNamespaceZonesSchema,
			},
			"zones_initial_class_config_overrides": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Initial Class Config Overrides for Zones",
				Elem:        supervisorNamespaceZonesInitialClassConfigOverridesSchema,
			},
		},
	}
}

func datasourceVcfaSupervisorNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cciClient := meta.(ClientContainer).cciClient
	if cciClient.VCDClient.Client.IsSysAdmin {
		return diag.Errorf("this resource requires Org user")
	}
	projectName := d.Get("project_name").(string)
	name := d.Get("name").(string)

	supervisorNamespace, err := cciClient.GetSupervisorNamespaceByName(projectName, name)
	if err != nil {
		return diag.Errorf("error retrieving %s '%s' in Project '%s': %s", labelSupervisorNamespace, name, projectName, err)
	}

	if err := setsupervisorNamespaceData(d, projectName, name, supervisorNamespace.SupervisorNamespace); err != nil {
		return diag.Errorf("error setting %s data: %s", labelSupervisorNamespace, err)
	}

	return nil
}
