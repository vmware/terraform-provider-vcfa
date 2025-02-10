package vcfa

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const labelSupervisorNamespace = "Supervisor Namespace"

var supervisorNamespaceStorageClassesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"limit_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Limit in MiB",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the Storage Class",
		},
	},
}

var supervisorNamespaceStorageClassesInitialClassConfigOverridesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"limit_mib": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Limit in MiB",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
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

var supervisorNamespaceZonesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"cpu_limit_mhz": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "CPU limit in MHz",
		},
		"cpu_reservation_mhz": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "CPU reservation in MHz",
		},
		"memory_limit_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Memory limit in MiB",
		},
		"memory_reservation_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Memory reservation in MiB",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the Zone",
		},
	},
}

var supervisorNamespaceZonesInitialClassConfigOverridesSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"cpu_limit_mhz": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "CPU limit in MHz",
		},
		"cpu_reservation_mhz": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "CPU reservation in MHz",
		},
		"memory_limit_mib": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Memory limit in MiB",
		},
		"memory_reservation_mib": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Memory reservation in MiB",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
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
				ValidateDiagFunc: validation.AllDiag(
					validation.ToDiagFunc(
						validation.StringMatch(rfc1123LabelNameRegex, "Name must match RFC 1123 Label name (lower case alphabet, 0-9 and hyphen -)"),
					),
					validation.ToDiagFunc(
						validation.StringIsNotEmpty,
					),
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
				Description: fmt.Sprintf("The name of the Project the %s belongs to", labelSupervisorNamespace),
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringIsNotEmpty,
				),
			},
			"class_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Supervisor Namespace Class",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
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
				Required:    true,
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
				Required:    true,
				MinItems:    1,
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
				Required:    true,
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
				Required:    true,
				MinItems:    1,
				Description: "Initial Class Config Overrides for Zones",
				Elem:        supervisorNamespaceZonesInitialClassConfigOverridesSchema,
			},
		},
	}
}

func resourceVcfaSupervisorNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cciClient := meta.(ClientContainer).cciClient

	projectName := d.Get("project_name").(string)

	supervisorNamespace := getsupervisorNamespaceType(d)
	createdSupervisorNamespace, err := cciClient.CreateSupervisorNamespace(projectName, supervisorNamespace)
	if err != nil {
		return diag.Errorf("error creating %s: %s", labelSupervisorNamespace, err)
	}

	d.SetId(buildSupervisorNamespaceResourceId(d.Get("project_name").(string), createdSupervisorNamespace.SupervisorNamespace.GetName()))

	return resourceVcfaSupervisorNamespaceRead(ctx, d, meta)
}

func resourceVcfaSupervisorNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("%s updates are not supported", labelSupervisorNamespace)
}

func resourceVcfaSupervisorNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cciClient := meta.(ClientContainer).cciClient

	projectName, name, err := parseSupervisorNamespaceResourceId(d.Id())
	if err != nil {
		return diag.Errorf("error parsing %s resource id %s: %s", labelSupervisorNamespace, d.Id(), err)
	}

	supervisorNamespace, err := cciClient.GetSupervisorNamespaceByName(projectName, name)
	if err != nil {
		return diag.Errorf("error retrieving %s '%s' in Project '%s': %s", labelSupervisorNamespace, name, projectName, err)
	}

	if err := setsupervisorNamespaceData(d, projectName, name, supervisorNamespace.SupervisorNamespace); err != nil {
		return diag.Errorf("error setting %s data: %s", labelSupervisorNamespace, err)
	}

	return nil
}

func resourceVcfaSupervisorNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	cciClient := meta.(ClientContainer).cciClient

	projectName, name, err := parseSupervisorNamespaceResourceId(d.Id())
	if err != nil {
		return diag.Errorf("error parsing %s resource id %s: %s", labelSupervisorNamespace, d.Id(), err)
	}

	supervisorNamespace, err := cciClient.GetSupervisorNamespaceByName(projectName, name)
	if err != nil {
		return diag.Errorf("error retrieving %s '%s' in Project '%s': %s", labelSupervisorNamespace, name, projectName, err)
	}

	err = supervisorNamespace.Delete()
	if err != nil {
		return diag.Errorf("error removing %s '%s' from Project '%s': %s", labelSupervisorNamespace, name, projectName, err)
	}

	d.SetId("")

	return nil
}

func resourceVcfaSupervisorNamespaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	cciClient := meta.(ClientContainer).cciClient
	idSlice := strings.Split(d.Id(), ImportSeparator)
	if len(idSlice) != 2 {
		return nil, fmt.Errorf("expected import ID to be <project_name>%s<supervisor_namespace_name>", ImportSeparator)
	}
	projectName := idSlice[0]
	name := idSlice[1]

	if _, err := cciClient.GetSupervisorNamespaceByName(projectName, name); err != nil {
		return nil, fmt.Errorf("error reading %s: %s", labelSupervisorNamespace, err)
	}

	d.SetId(buildSupervisorNamespaceResourceId(projectName, name))

	return []*schema.ResourceData{d}, nil
}

func buildSupervisorNamespaceResourceId(projectName string, supervisorNamespaceName string) string {
	return fmt.Sprintf("%s:%s", projectName, supervisorNamespaceName)
}

func parseSupervisorNamespaceResourceId(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("id %s does not contain two parts", id)
	}
	return idParts[0], idParts[1], nil
}

func getsupervisorNamespaceType(d *schema.ResourceData) *ccitypes.SupervisorNamespace {
	supervisorNamespace := &ccitypes.SupervisorNamespace{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.SupervisorNamespaceKind,
			APIVersion: ccitypes.SupervisorNamespaceAPI + "/" + ccitypes.SupervisorNamespaceVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			GenerateName: d.Get("name_prefix").(string),
			Namespace:    d.Get("project_name").(string),
		},
		Spec: ccitypes.SupervisorNamespaceSpec{
			ClassName:                   d.Get("class_name").(string),
			Description:                 d.Get("description").(string),
			InitialClassConfigOverrides: ccitypes.SupervisorNamespaceSpecInitialClassConfigOverrides{},
			RegionName:                  d.Get("region_name").(string),
			VpcName:                     d.Get("vpc_name").(string),
		},
	}

	storageClassesInitialClassConfigOverridesList := d.Get("storage_classes_initial_class_config_overrides").(*schema.Set).List()
	if len(storageClassesInitialClassConfigOverridesList) > 0 {
		storageClassesInitialClassConfigOverrides := make([]ccitypes.SupervisorNamespaceSpecInitialClassConfigOverridesStorageClass, len(storageClassesInitialClassConfigOverridesList))
		for i, k := range storageClassesInitialClassConfigOverridesList {
			storageClass := k.(map[string]interface{})
			storageClassesInitialClassConfigOverrides[i] = ccitypes.SupervisorNamespaceSpecInitialClassConfigOverridesStorageClass{
				LimitMiB: int64(storageClass["limit_mib"].(int)),
				Name:     storageClass["name"].(string),
			}
		}
		supervisorNamespace.Spec.InitialClassConfigOverrides.StorageClasses = storageClassesInitialClassConfigOverrides
	}

	zonesInitialClassConfigOverridesList := d.Get("zones_initial_class_config_overrides").(*schema.Set).List()
	if len(zonesInitialClassConfigOverridesList) > 0 {
		zonesInitialClassConfigOverrides := make([]ccitypes.SupervisorNamespaceSpecInitialClassConfigOverridesZone, len(zonesInitialClassConfigOverridesList))
		for i, k := range zonesInitialClassConfigOverridesList {
			zone := k.(map[string]interface{})
			zonesInitialClassConfigOverrides[i] = ccitypes.SupervisorNamespaceSpecInitialClassConfigOverridesZone{
				CpuLimitMHz:          int64(zone["cpu_limit_mhz"].(int)),
				CpuReservationMHz:    int64(zone["cpu_reservation_mhz"].(int)),
				MemoryLimitMiB:       int64(zone["memory_limit_mib"].(int)),
				MemoryReservationMiB: int64(zone["memory_reservation_mib"].(int)),
				Name:                 zone["name"].(string),
			}
		}
		supervisorNamespace.Spec.InitialClassConfigOverrides.Zones = zonesInitialClassConfigOverrides
	}

	return supervisorNamespace
}

func setsupervisorNamespaceData(d *schema.ResourceData, projectName string, supervisorNamespaceName string, supervisorNamespace *ccitypes.SupervisorNamespace) error {
	if supervisorNamespace == nil {
		return fmt.Errorf("error - provided Supervisor Namespace")
	}

	d.SetId(buildSupervisorNamespaceResourceId(projectName, supervisorNamespaceName))
	dSet(d, "name", supervisorNamespaceName)
	dSet(d, "project_name", projectName)
	dSet(d, "class_name", supervisorNamespace.Spec.ClassName)
	dSet(d, "description", supervisorNamespace.Spec.Description)
	dSet(d, "region_name", supervisorNamespace.Spec.RegionName)
	dSet(d, "phase", supervisorNamespace.Status.Phase)
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

	storageClasses := make([]interface{}, 0, len(supervisorNamespace.Status.StorageClasses))
	for _, storageClass := range supervisorNamespace.Status.StorageClasses {
		sc := map[string]interface{}{
			"limit_mib": storageClass.LimitMiB,
			"name":      storageClass.Name,
		}

		storageClasses = append(storageClasses, sc)
	}
	d.Set("storage_classes", storageClasses)

	storageClassesInitialClassConfigOverrides := make([]interface{}, 0, len(supervisorNamespace.Spec.InitialClassConfigOverrides.StorageClasses))
	for _, storageClass := range supervisorNamespace.Spec.InitialClassConfigOverrides.StorageClasses {
		storageClassInitialClassConfigOverride := map[string]interface{}{
			"limit_mib": storageClass.LimitMiB,
			"name":      storageClass.Name,
		}

		storageClassesInitialClassConfigOverrides = append(storageClassesInitialClassConfigOverrides, storageClassInitialClassConfigOverride)
	}
	d.Set("storage_classes_initial_class_config_overrides", storageClassesInitialClassConfigOverrides)

	vmClasses := make([]interface{}, 0, len(supervisorNamespace.Status.VMClasses))
	for _, vmClass := range supervisorNamespace.Status.VMClasses {
		vc := map[string]interface{}{
			"name": vmClass.Name,
		}

		vmClasses = append(vmClasses, vc)
	}
	d.Set("vm_classes", vmClasses)

	zones := make([]interface{}, 0, len(supervisorNamespace.Status.Zones))
	for _, zone := range supervisorNamespace.Status.Zones {
		z := map[string]interface{}{
			"cpu_limit_mhz":          zone.CpuLimitMHz,
			"cpu_reservation_mhz":    zone.CpuReservationMHz,
			"memory_limit_mib":       zone.MemoryLimitMiB,
			"memory_reservation_mib": zone.MemoryReservationMiB,
			"name":                   zone.Name,
		}

		zones = append(zones, z)
	}
	d.Set("zones", zones)

	zonesInitialClassConfigOverrides := make([]interface{}, 0, len(supervisorNamespace.Spec.InitialClassConfigOverrides.Zones))
	for _, zone := range supervisorNamespace.Spec.InitialClassConfigOverrides.Zones {
		zoneInitialClassConfigOverride := map[string]interface{}{
			"cpu_limit_mhz":          zone.CpuLimitMHz,
			"cpu_reservation_mhz":    zone.CpuReservationMHz,
			"memory_limit_mib":       zone.MemoryLimitMiB,
			"memory_reservation_mib": zone.MemoryReservationMiB,
			"name":                   zone.Name,
		}

		zonesInitialClassConfigOverrides = append(zonesInitialClassConfigOverrides, zoneInitialClassConfigOverride)
	}
	d.Set("zones_initial_class_config_overrides", zonesInitialClassConfigOverrides)

	return nil
}
