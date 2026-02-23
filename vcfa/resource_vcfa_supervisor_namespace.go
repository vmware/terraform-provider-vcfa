// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const labelSupervisorNamespace = "Supervisor Namespace"

var supervisorNamespaceConditionsSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"message": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Human-readable message with details about the condition",
		},
		"reason": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Machine-readable CamelCase reason code",
		},
		"severity": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Severity level: `Info`, `Warning`, `Error`",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Condition status: `True`, `False`, `Unknown`)",
		},
		"type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Condition type identifier (e.g., `Ready`, `Realized`, ...)",
		},
	},
}

var supervisorNamespaceStatusContentLibrariesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: " Name of the content library",
		},
		"type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Type of content source",
		},
	},
}
var supervisorNamespaceContentSourcesClassConfigOverridesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // Update not supported
			Description: "Name of the content library",
		},
		"type": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // Update not supported
			Description: "Type of content source",
		},
	},
}

var supervisorNamespaceInfraPoliciesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"mandatory": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Infra policy is auto enforced if mandatory",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the Infra Policy",
		},
	},
}

var supervisorNamespaceStorageClassesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"limit": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the Storage Class",
		},
	},
}

var supervisorNamespaceStorageClassesClassConfigOverridesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"limit": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // Update not supported
			Description: "Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // Update not supported
			Description: "Name of the Storage Class",
		},
	},
}

var supervisorNamespaceVMClassesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the VM Class",
		},
	},
}

var supervisorNamespaceVMClassesClassConfigOverridesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the VM Class",
		},
	},
}

var supervisorNamespaceZonesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"cpu_limit": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "CPU limit (format: `<number><unit>`, where `<unit>` can be `M` or `G`)",
		},
		"cpu_reservation": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "CPU reservation (format: `<number><unit>`, where `<unit>` can be `M` or `G`)",
		},
		"marked_for_removal": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Indicates if this zone is scheduled for removal during a scale-down operation",
		},
		"memory_limit": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Memory limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)",
		},
		"memory_reservation": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Memory reservation (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the Zone",
		},
	},
}

var supervisorNamespaceZonesClassConfigOverridesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"cpu_limit": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // Update not supported
			Description: "CPU limit (format: `<number><unit>`, where `<unit>` can be `M` or `G`)",
		},
		"cpu_reservation": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // Update not supported
			Description: "CPU reservation (format: `<number><unit>`, where `<unit>` can be `M` or `G`)",
		},
		"memory_limit": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // Update not supported
			Description: "Memory limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)",
		},
		"memory_reservation": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // Update not supported
			Description: "Memory reservation (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true, // Update not supported
			Description: "Name of the Zone",
		},
	},
}

func resourceVcfaSupervisorNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcfaSupervisorNamespaceCreate,
		ReadContext:   resourceVcfaSupervisorNamespaceRead,
		UpdateContext: resourceVcfaSupervisorNamespaceUpdate,
		DeleteContext: resourceVcfaSupervisorNamespaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcfaSupervisorNamespaceImport,
		},

		Schema: map[string]*schema.Schema{
			"name_prefix": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Supervisor Namespaces names cannot be changed
				Description: fmt.Sprintf("Prefix for the %s name", labelSupervisorNamespace),
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringMatch(rfc1123LabelNameRegex, "Name must match RFC 1123 Label name (lower case alphabet, 0-9 and hyphen -)"),
				),
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Name of the %s", labelSupervisorNamespace),
			},
			"project_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Update not supported
				Description: fmt.Sprintf("The name of the Project the %s belongs to", labelSupervisorNamespace),
			},
			"class_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Update not supported
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
				Description: fmt.Sprintf("Content libraries currently available in the %s", labelSupervisorNamespace),
				Elem:        supervisorNamespaceStatusContentLibrariesSchema,
			},
			"content_sources_class_config_overrides": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Class Config Overrides for Content Sources",
				Elem:        supervisorNamespaceContentSourcesClassConfigOverridesSchema,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
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
				Optional:    true,
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
				Required:    true,
				ForceNew:    true, // Update not supported
				Description: fmt.Sprintf("Name of the %s", labelVcfaRegion),
			},
			"seg_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: fmt.Sprintf("Service Engine Group associated with the %s", labelSupervisorNamespace),
			},
			"shared_subnet_names": {
				Type:        schema.TypeSet,
				Optional:    true,
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
				Type:         schema.TypeSet,
				Optional:     true,
				ForceNew:     true, // Update not supported
				MinItems:     1,
				Computed:     true,
				Description:  "Class Config Overrides for Storage Classes",
				Elem:         supervisorNamespaceStorageClassesClassConfigOverridesSchema,
				ExactlyOneOf: []string{"storage_classes_class_config_overrides", "storage_classes_initial_class_config_overrides"},
			},
			"storage_classes_initial_class_config_overrides": {
				Type:         schema.TypeSet,
				Optional:     true,
				ForceNew:     true, // Update not supported
				MinItems:     1,
				Computed:     true,
				Deprecated:   "Please use `storage_classes_class_config_overrides` instead",
				Description:  "Initial Class Config Overrides for Storage Classes",
				Elem:         supervisorNamespaceStorageClassesClassConfigOverridesSchema,
				ExactlyOneOf: []string{"storage_classes_class_config_overrides", "storage_classes_initial_class_config_overrides"},
			},
			"vm_classes": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("%s VM Classes", labelSupervisorNamespace),
				Elem:        supervisorNamespaceVMClassesSchema,
			},
			"vm_classes_class_config_overrides": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: fmt.Sprintf("%s VM Classes", labelSupervisorNamespace),
				Elem:        supervisorNamespaceVMClassesClassConfigOverridesSchema,
			},
			"vpc_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Update not supported
				Description: "Name of the VPC",
			},
			"zones": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("%s Zones", labelSupervisorNamespace),
				Elem:        supervisorNamespaceZonesSchema,
			},
			"zones_class_config_overrides": {
				Type:         schema.TypeSet,
				Optional:     true,
				ForceNew:     true, // Update not supported
				MinItems:     1,
				Computed:     true,
				Description:  "Class Config Overrides for Zones",
				Elem:         supervisorNamespaceZonesClassConfigOverridesSchema,
				ExactlyOneOf: []string{"zones_class_config_overrides", "zones_initial_class_config_overrides"},
			},
			"zones_initial_class_config_overrides": {
				Type:         schema.TypeSet,
				Optional:     true,
				ForceNew:     true, // Update not supported
				MinItems:     1,
				Computed:     true,
				Deprecated:   "Please use `zones_class_config_overrides` instead",
				Description:  "Initial Class Config Overrides for Zones",
				Elem:         supervisorNamespaceZonesClassConfigOverridesSchema,
				ExactlyOneOf: []string{"zones_class_config_overrides", "zones_initial_class_config_overrides"},
			},
		},
	}
}

func resourceVcfaSupervisorNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	namePrefix, oknamePrefix := d.GetOk("name_prefix")
	if !oknamePrefix {
		return diag.Errorf("name_prefix not specified")
	}
	projectName, okProjectName := d.GetOk("project_name")
	if !okProjectName {
		return diag.Errorf("project_name not specified")
	}

	supervisorNamespace := ccitypes.SupervisorNamespace{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.SupervisorNamespaceKind,
			APIVersion: ccitypes.SupervisorNamespaceAPI + "/" + ccitypes.SupervisorNamespaceVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			GenerateName: namePrefix.(string),
			Namespace:    projectName.(string),
		},
		Spec: ccitypes.SupervisorNamespaceSpec{
			ClassName:            d.Get("class_name").(string),
			Description:          d.Get("description").(string),
			ClassConfigOverrides: ccitypes.SupervisorNamespaceSpecClassConfigOverrides{},
			InfraPolicyNames:     convertSchemaSetToSliceOfStrings(d.Get("infra_policy_names").(*schema.Set)),
			RegionName:           d.Get("region_name").(string),
			SegName:              d.Get("seg_name").(string),
			SharedSubnetNames:    convertSchemaSetToSliceOfStrings(d.Get("shared_subnet_names").(*schema.Set)),
			VpcName:              d.Get("vpc_name").(string),
		},
	}

	contentSourcesClassConfigOverridesList := d.Get("content_sources_class_config_overrides").(*schema.Set).List()
	if len(contentSourcesClassConfigOverridesList) > 0 {
		contentSourcesClassConfigOverrides := make([]ccitypes.SupervisorNamespaceSpecClassConfigOverridesContentSources, len(contentSourcesClassConfigOverridesList))
		for i, k := range contentSourcesClassConfigOverridesList {
			storageClass := k.(map[string]interface{})
			contentSourcesClassConfigOverrides[i] = ccitypes.SupervisorNamespaceSpecClassConfigOverridesContentSources{
				Name: storageClass["name"].(string),
				Type: storageClass["type"].(string),
			}
		}
		supervisorNamespace.Spec.ClassConfigOverrides.ContentSources = contentSourcesClassConfigOverrides
	}

	// Deprecation compatibility: If `storage_classes_class_config_overrides` is not set, fallback to deprecated one.
	var storageClassesClassConfigOverridesList []any
	if storageClassesClassConfigOverrides, ok := d.GetOk("storage_classes_class_config_overrides"); ok {
		storageClassesClassConfigOverridesList = storageClassesClassConfigOverrides.(*schema.Set).List()
	} else {
		storageClassesClassConfigOverridesList = d.Get("storage_classes_initial_class_config_overrides").(*schema.Set).List()
	}
	if len(storageClassesClassConfigOverridesList) > 0 {
		storageClassesClassConfigOverrides := make([]ccitypes.SupervisorNamespaceSpecClassConfigOverridesStorageClass, len(storageClassesClassConfigOverridesList))
		for i, k := range storageClassesClassConfigOverridesList {
			storageClass := k.(map[string]interface{})
			storageClassesClassConfigOverrides[i] = ccitypes.SupervisorNamespaceSpecClassConfigOverridesStorageClass{
				Limit: storageClass["limit"].(string),
				Name:  storageClass["name"].(string),
			}
		}
		supervisorNamespace.Spec.ClassConfigOverrides.StorageClasses = storageClassesClassConfigOverrides
	}

	vmClassesClassConfigOverridesList := d.Get("vm_classes_class_config_overrides").(*schema.Set).List()
	if len(vmClassesClassConfigOverridesList) > 0 {
		vmClassesClassConfigOverrides := make([]ccitypes.SupervisorNamespaceSpecClassConfigOverridesVmClass, len(vmClassesClassConfigOverridesList))
		for i, k := range vmClassesClassConfigOverridesList {
			storageClass := k.(map[string]interface{})
			vmClassesClassConfigOverrides[i] = ccitypes.SupervisorNamespaceSpecClassConfigOverridesVmClass{
				Name: storageClass["name"].(string),
			}
		}
		supervisorNamespace.Spec.ClassConfigOverrides.VmClasses = vmClassesClassConfigOverrides
	}

	// Deprecation compatibility: If `zones_class_config_overrides` is not set, fallback to deprecated one.
	var zonesClassConfigOverridesList []any
	if zonesClassConfigOverrides, ok := d.GetOk("zones_class_config_overrides"); ok {
		zonesClassConfigOverridesList = zonesClassConfigOverrides.(*schema.Set).List()
	} else {
		zonesClassConfigOverridesList = d.Get("zones_initial_class_config_overrides").(*schema.Set).List()
	}
	if len(zonesClassConfigOverridesList) > 0 {
		zonesClassConfigOverrides := make([]ccitypes.SupervisorNamespaceSpecClassConfigOverridesZone, len(zonesClassConfigOverridesList))
		for i, k := range zonesClassConfigOverridesList {
			zone := k.(map[string]interface{})
			zonesClassConfigOverrides[i] = ccitypes.SupervisorNamespaceSpecClassConfigOverridesZone{
				CpuLimit:          zone["cpu_limit"].(string),
				CpuReservation:    zone["cpu_reservation"].(string),
				MemoryLimit:       zone["memory_limit"].(string),
				MemoryReservation: zone["memory_reservation"].(string),
				Name:              zone["name"].(string),
			}
		}
		supervisorNamespace.Spec.ClassConfigOverrides.Zones = zonesClassConfigOverrides
	}

	supervisorNamespaceOut, err := createSupervisorNamespace(tmClient, projectName.(string), supervisorNamespace)
	if err != nil {
		return diag.Errorf("error creating %s: %s", labelSupervisorNamespace, err)
	}

	stateChangeFunc := retry.StateChangeConf{
		Pending: []string{"CREATING", "WAITING"},
		Target:  []string{"CREATED"},
		Refresh: func() (any, string, error) {
			supervisorNamespace, err := readSupervisorNamespace(tmClient, projectName.(string), supervisorNamespaceOut.GetName())
			if err != nil {
				return nil, "", err
			}

			log.Printf("[DEBUG] %s %s current phase is %s", labelSupervisorNamespace, supervisorNamespaceOut.GetName(), supervisorNamespace.Status.Phase)
			if strings.ToUpper(supervisorNamespace.Status.Phase) == "ERROR" {
				return nil, "", fmt.Errorf("%s %s is in an ERROR state", labelSupervisorNamespace, supervisorNamespaceOut.GetName())
			}

			return supervisorNamespace, strings.ToUpper(supervisorNamespace.Status.Phase), nil
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err = stateChangeFunc.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for %s %s in Project %s to be created: %s", labelSupervisorNamespace, supervisorNamespaceOut.GetName(), projectName, err)
	}

	d.SetId(buildResourceId(projectName.(string), supervisorNamespaceOut.GetName()))

	return resourceVcfaSupervisorNamespaceRead(ctx, d, meta)
}

func resourceVcfaSupervisorNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("%s updates are not supported", labelSupervisorNamespace)
}

func resourceVcfaSupervisorNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	projectName, name, err := parseResourceId(d.Id())
	if err != nil {
		return diag.Errorf("error parsing %s resource id %s: %s", labelSupervisorNamespace, d.Id(), err)
	}

	supervisorNamespace, err := readSupervisorNamespace(tmClient, projectName, name)
	if err != nil {
		return diag.Errorf("error reading %s: %s", labelSupervisorNamespace, err)
	}

	if err := setSupervisorNamespaceData(tmClient, d, projectName, name, supervisorNamespace); err != nil {
		return diag.Errorf("error setting %s data: %s", labelSupervisorNamespace, err)
	}

	return nil
}

func resourceVcfaSupervisorNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tmClient := meta.(ClientContainer).tmClient
	projectName, name, err := parseResourceId(d.Id())
	if err != nil {
		return diag.Errorf("error parsing %s resource id %s: %s", labelSupervisorNamespace, d.Id(), err)
	}

	if err := deleteSupervisorNamespace(tmClient, projectName, name); err != nil {
		return diag.Errorf("error deleting %s: %s", labelSupervisorNamespace, err)
	}

	stateChangeFunc := retry.StateChangeConf{
		Pending: []string{"DELETING", "WAITING"},
		Target:  []string{"DELETED"},
		Refresh: func() (any, string, error) {
			supervisorNamespace, err := readSupervisorNamespace(tmClient, projectName, name)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return "", "DELETED", nil
				}
				return nil, "", err
			}

			log.Printf("[DEBUG] %s %s current phase is %s", labelSupervisorNamespace, name, supervisorNamespace.Status.Phase)
			if strings.ToUpper(supervisorNamespace.Status.Phase) == "ERROR" {
				return nil, "", fmt.Errorf("%s %s is in an ERROR state", labelSupervisorNamespace, name)
			}

			return supervisorNamespace, strings.ToUpper(supervisorNamespace.Status.Phase), nil
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err = stateChangeFunc.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for %s %s in Project %s to be deleted: %s", labelSupervisorNamespace, name, projectName, err)
	}

	d.SetId("")

	return nil
}

func resourceVcfaSupervisorNamespaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tmClient := meta.(ClientContainer).tmClient
	idSlice := strings.Split(d.Id(), ImportSeparator)
	if len(idSlice) != 2 {
		return nil, fmt.Errorf("expected import ID to be <project_name>%s<supervisor_namespace_name>", ImportSeparator)
	}
	projectName := idSlice[0]
	name := idSlice[1]
	if _, err := readSupervisorNamespace(tmClient, projectName, name); err != nil {
		return nil, fmt.Errorf("error reading %s: %s", labelSupervisorNamespace, err)
	}

	d.SetId(buildResourceId(projectName, name))

	return []*schema.ResourceData{d}, nil
}

func createSupervisorNamespace(tmClient *VCDClient, projectName string, supervisorNamespace ccitypes.SupervisorNamespace) (ccitypes.SupervisorNamespace, error) {
	var supervisorNamespaceOut ccitypes.SupervisorNamespace
	supervisorNamespaceURL, err := buildSupervisorNamespaceURL(tmClient, projectName, "")
	if err != nil {
		return supervisorNamespace, fmt.Errorf("error building %s URL: %s", labelSupervisorNamespace, err)
	}
	if err := tmClient.VCDClient.Client.PostEntity(supervisorNamespaceURL, nil, &supervisorNamespace, &supervisorNamespaceOut, nil); err != nil {
		return supervisorNamespace, fmt.Errorf("error creating %s in Project %s: %s", labelSupervisorNamespace, projectName, err)
	}
	return supervisorNamespaceOut, nil
}

func readSupervisorNamespace(tmClient *VCDClient, projectName string, supervisorNamespaceName string) (ccitypes.SupervisorNamespace, error) {
	var supervisorNamespace ccitypes.SupervisorNamespace
	supervisorNamespaceURL, err := buildSupervisorNamespaceURL(tmClient, projectName, supervisorNamespaceName)
	if err != nil {
		return supervisorNamespace, fmt.Errorf("error building %s URL: %s", labelSupervisorNamespace, err)
	}
	if err := tmClient.VCDClient.Client.GetEntity(supervisorNamespaceURL, nil, &supervisorNamespace, nil); err != nil {
		return supervisorNamespace, fmt.Errorf("error reading %s %s in Project %s: %s", labelSupervisorNamespace, supervisorNamespaceName, projectName, err)
	}
	return supervisorNamespace, nil
}

func deleteSupervisorNamespace(tmClient *VCDClient, projectName string, supervisorNamespaceName string) error {
	supervisorNamespaceURL, err := buildSupervisorNamespaceURL(tmClient, projectName, supervisorNamespaceName)
	if err != nil {
		return fmt.Errorf("error building %s URL: %s", labelSupervisorNamespace, err)
	}
	if err := tmClient.Client.DeleteEntity(supervisorNamespaceURL, nil, nil); err != nil {
		return fmt.Errorf("error deleting %s %s in Project %s: %s", labelSupervisorNamespace, supervisorNamespaceName, projectName, err)
	}
	return nil
}

func buildSupervisorNamespaceURL(tmClient *VCDClient, projectName string, supervisorNamespaceName string) (*url.URL, error) {
	supervisorNamespaceRawURL := fmt.Sprintf(ccitypes.SupervisorNamespacesURL, projectName)
	if supervisorNamespaceName != "" {
		supervisorNamespaceRawURL = supervisorNamespaceRawURL + "/" + supervisorNamespaceName
	}

	return tmClient.Client.GetEntityUrl(supervisorNamespaceRawURL)
}

func buildResourceId(projectName string, supervisorNamespaceName string) string {
	return fmt.Sprintf("%s:%s", projectName, supervisorNamespaceName)
}

func parseResourceId(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("id %s does not contain two parts", id)
	}
	return idParts[0], idParts[1], nil
}

func setSupervisorNamespaceData(_ *VCDClient, d *schema.ResourceData, projectName string, supervisorNamespaceName string, supervisorNamespace ccitypes.SupervisorNamespace) error {
	d.SetId(buildResourceId(projectName, supervisorNamespaceName))
	dSet(d, "name", supervisorNamespaceName)
	dSet(d, "project_name", projectName)
	dSet(d, "class_name", supervisorNamespace.Spec.ClassName)
	dSet(d, "description", supervisorNamespace.Spec.Description)
	dSet(d, "phase", supervisorNamespace.Status.Phase)
	dSet(d, "region_name", supervisorNamespace.Spec.RegionName)
	dSet(d, "seg_name", supervisorNamespace.Spec.SegName)
	dSet(d, "vpc_name", supervisorNamespace.Spec.VpcName)

	d.Set("ready", false)
	for _, condition := range supervisorNamespace.Status.Conditions {
		if strings.ToLower(condition.Type) == "ready" {
			if strings.ToLower(condition.Status) == "true" {
				d.Set("ready", true)
			}
			break
		}
	}

	conditions := make([]interface{}, 0, len(supervisorNamespace.Status.Conditions))
	for _, condition := range supervisorNamespace.Status.Conditions {
		c := map[string]interface{}{
			"message":  condition.Message,
			"reason":   condition.Reason,
			"severity": condition.Severity,
			"status":   condition.Status,
			"type":     condition.Type,
		}

		conditions = append(conditions, c)
	}
	d.Set("conditions", conditions)

	contentLibraries := make([]interface{}, 0, len(supervisorNamespace.Status.ContentLibraries))
	for _, contentLibrary := range supervisorNamespace.Status.ContentLibraries {
		cl := map[string]interface{}{
			"name": contentLibrary.Name,
			"type": contentLibrary.Type,
		}

		contentLibraries = append(contentLibraries, cl)
	}
	d.Set("content_libraries", contentLibraries)

	contentSourcesClassConfigOverrides := make([]interface{}, 0, len(supervisorNamespace.Spec.ClassConfigOverrides.ContentSources))
	for _, contentSource := range supervisorNamespace.Spec.ClassConfigOverrides.ContentSources {
		cs := map[string]interface{}{
			"name": contentSource.Name,
			"type": contentSource.Type,
		}

		contentSourcesClassConfigOverrides = append(contentSourcesClassConfigOverrides, cs)
	}
	d.Set("content_sources_class_config_overrides", contentSourcesClassConfigOverrides)

	infraPolicies := make([]interface{}, 0, len(supervisorNamespace.Status.InfraPolicies))
	for _, infraPolicy := range supervisorNamespace.Status.InfraPolicies {
		ip := map[string]interface{}{
			"mandatory": infraPolicy.Mandatory,
			"name":      infraPolicy.Name,
		}

		infraPolicies = append(infraPolicies, ip)
	}
	d.Set("infra_policies", infraPolicies)

	infraPolicyNames := make([]interface{}, 0, len(supervisorNamespace.Spec.InfraPolicyNames))
	for _, infraPolicyName := range supervisorNamespace.Spec.InfraPolicyNames {
		infraPolicyNames = append(infraPolicyNames, infraPolicyName)
	}
	d.Set("infra_policy_names", infraPolicyNames)

	sharedSubnetNames := make([]interface{}, 0, len(supervisorNamespace.Spec.SharedSubnetNames))
	for _, sharedSubnetName := range supervisorNamespace.Spec.SharedSubnetNames {
		sharedSubnetNames = append(sharedSubnetNames, sharedSubnetName)
	}
	d.Set("shared_subnet_names", sharedSubnetNames)

	storageClasses := make([]interface{}, 0, len(supervisorNamespace.Status.StorageClasses))
	for _, storageClass := range supervisorNamespace.Status.StorageClasses {
		sc := map[string]interface{}{
			"limit": storageClass.Limit,
			"name":  storageClass.Name,
		}

		storageClasses = append(storageClasses, sc)
	}
	d.Set("storage_classes", storageClasses)

	storageClassesClassConfigOverrides := make([]interface{}, 0, len(supervisorNamespace.Spec.ClassConfigOverrides.StorageClasses))
	for _, storageClass := range supervisorNamespace.Spec.ClassConfigOverrides.StorageClasses {
		storageClassClassConfigOverride := map[string]interface{}{
			"limit": storageClass.Limit,
			"name":  storageClass.Name,
		}

		storageClassesClassConfigOverrides = append(storageClassesClassConfigOverrides, storageClassClassConfigOverride)
	}
	d.Set("storage_classes_class_config_overrides", storageClassesClassConfigOverrides)
	d.Set("storage_classes_initial_class_config_overrides", storageClassesClassConfigOverrides)

	vmClasses := make([]interface{}, 0, len(supervisorNamespace.Status.VMClasses))
	for _, vmClass := range supervisorNamespace.Status.VMClasses {
		vc := map[string]interface{}{
			"name": vmClass.Name,
		}

		vmClasses = append(vmClasses, vc)
	}
	d.Set("vm_classes", vmClasses)

	vmClassessClassConfigOverrides := make([]interface{}, 0, len(supervisorNamespace.Spec.ClassConfigOverrides.VmClasses))
	for _, vmClass := range supervisorNamespace.Spec.ClassConfigOverrides.VmClasses {
		vmClassClassConfigOverride := map[string]interface{}{
			"name": vmClass.Name,
		}

		vmClassessClassConfigOverrides = append(vmClassessClassConfigOverrides, vmClassClassConfigOverride)
	}
	d.Set("vm_classes_class_config_overrides", vmClassessClassConfigOverrides)

	zones := make([]interface{}, 0, len(supervisorNamespace.Status.Zones))
	for _, zone := range supervisorNamespace.Status.Zones {
		z := map[string]interface{}{
			"cpu_limit":          zone.CpuLimit,
			"cpu_reservation":    zone.CpuReservation,
			"marked_for_removal": zone.MarkedForRemoval,
			"memory_limit":       zone.MemoryLimit,
			"memory_reservation": zone.MemoryReservation,
			"name":               zone.Name,
		}

		zones = append(zones, z)
	}
	d.Set("zones", zones)

	zonesClassConfigOverrides := make([]interface{}, 0, len(supervisorNamespace.Spec.ClassConfigOverrides.Zones))
	for _, zone := range supervisorNamespace.Spec.ClassConfigOverrides.Zones {
		zoneClassConfigOverride := map[string]interface{}{
			"cpu_limit":          zone.CpuLimit,
			"cpu_reservation":    zone.CpuReservation,
			"memory_limit":       zone.MemoryLimit,
			"memory_reservation": zone.MemoryReservation,
			"name":               zone.Name,
		}

		zonesClassConfigOverrides = append(zonesClassConfigOverrides, zoneClassConfigOverride)
	}
	d.Set("zones_class_config_overrides", zonesClassConfigOverrides)
	d.Set("zones_initial_class_config_overrides", zonesClassConfigOverrides)

	return nil
}
