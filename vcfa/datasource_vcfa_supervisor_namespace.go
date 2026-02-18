// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcfaSupervisorNamespace() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcfaSupervisorNamespaceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelSupervisorNamespace),
			},
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("The name of the Project the %s belongs to", labelSupervisorNamespace),
			},
			"class_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Supervisor Namespace Class",
			},
			"conditions": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("Detailed conditions tracking %s health and lifecycle events", labelSupervisorNamespace),
				Elem:        supervisorNamespaceConditionsSchema,
			},
			"content_libraries": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("Content libraries currently available in the  %s", labelSupervisorNamespace),
				Elem:        supervisorNamespaceStatusContentLibrariesSchema,
			},
			"content_sources_class_config_overrides": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Class Config Overrides for Content Sources",
				Elem:        supervisorNamespaceContentSourcesClassConfigOverridesSchema,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description",
			},
			"infra_policies": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("List of Infra Policies associated with the %s", labelSupervisorNamespace),
				Elem:        supervisorNamespaceInfraPoliciesSchema,
			},
			"infra_policy_names": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("List of Non-mandatory Infra Policies to be associated with the %s", labelSupervisorNamespace),
				Elem:        &schema.Schema{Type: schema.TypeString},
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
			"seg_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Service Engine Group associated with the %s", labelSupervisorNamespace),
			},
			"shared_subnet_names": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("Shared subnets associated with the %s", labelSupervisorNamespace),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"storage_classes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("%s Storage Classes", labelSupervisorNamespace),
				Elem:        supervisorNamespaceStorageClassesSchema,
			},
			"storage_classes_class_config_overrides": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Class Config Overrides for Storage Classes",
				Elem:        supervisorNamespaceStorageClassesClassConfigOverridesSchema,
			},
			"storage_classes_initial_class_config_overrides": {
				Type:        schema.TypeSet,
				Computed:    true,
				Deprecated:  "Please use `storage_classes_class_config_overrides` instead",
				Description: "Initial Class Config Overrides for Storage Classes",
				Elem:        supervisorNamespaceStorageClassesClassConfigOverridesSchema,
			},
			"vm_classes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("%s VM Classes", labelSupervisorNamespace),
				Elem:        supervisorNamespaceVMClassesSchema,
			},
			"vm_classes_class_config_overrides": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Class Config Overrides for VM Classes",
				Elem:        supervisorNamespaceVMClassesClassConfigOverridesSchema,
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
			"zones_class_config_overrides": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Class Config Overrides for Zones",
				Elem:        supervisorNamespaceZonesClassConfigOverridesSchema,
			},
			"zones_initial_class_config_overrides": {
				Type:        schema.TypeSet,
				Computed:    true,
				Deprecated:  "Please use `zones_class_config_overrides` instead",
				Description: "Initial Class Config Overrides for Zones",
				Elem:        supervisorNamespaceZonesClassConfigOverridesSchema,
			},
		},
	}
}

func datasourceVcfaSupervisorNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	name, okName := d.GetOk("name")
	if !okName {
		return diag.Errorf("name not specified")
	}
	projectName, okProjectName := d.GetOk("project_name")
	if !okProjectName {
		return diag.Errorf("project_name not specified")
	}

	supervisorNamespace, err := readSupervisorNamespace(tmClient, projectName.(string), name.(string))
	if err != nil {
		return diag.Errorf("error reading %s: %s", labelSupervisorNamespace, err)
	}
	if err := setSupervisorNamespaceData(tmClient, d, projectName.(string), name.(string), supervisorNamespace); err != nil {
		return diag.Errorf("error setting %s data: %s", labelSupervisorNamespace, err)
	}

	return nil
}
