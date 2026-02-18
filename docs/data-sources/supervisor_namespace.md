---
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor_namespace"
subcategory: ""
description: |-
  Provides a data source to read a Supervisor Namespace from VMware Cloud Foundation Automation.
---

# vcfa_supervisor_namespace

Provides a data source to read a Supervisor Namespace from VMware Cloud Foundation Automation.

_Used by: **Tenant**_

-> This data source may use the [Kubernetes provider](https://registry.terraform.io/providers/hashicorp/kubernetes),
to see how to obtain the Kubeconfig, please check the [`vcfa_kubeconfig`](/providers/vmware/vcfa/latest/docs/data-sources/kubeconfig) data source.

## Example Usage

```hcl
# A project data source read with the Kubernetes provider. This project already exists
data "kubernetes_resource" "project" {
  api_version = "project.cci.vmware.com/v1alpha2"
  kind        = "Project"
  metadata {
    name = "tf-tenant-demo-project"
  }
}

data "vcfa_supervisor_namespace" "supervisor_namespace" {
  name         = "tf-supervisor-namespace"
  project_name = data.kubernetes_resource.project.object["metadata"]["name"]
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the Supervisor Namespace
- `project_name` - (Required) The name of the Project where the Supervisor Namespace belongs to. Can be fetched
  with the Kubernetes provider [`kubernetes_resource`](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/data-sources/resource)
  data source for existing Projects

## Attribute Reference

- `class_name` - The name of the Supervisor Namespace Class
- `conditions` - Detailed conditions tracking Supervisor Namespace health and lifecycle events. See [Conditions](#conditions)
- `content_libraries` - Content libraries currently available in the Supervisor Namespace. See [Content Libraries](#content-libraries)
- `content_sources_class_config_overrides` - Class Config Overrides for Content Sources. See [Content Sources Class Config Overrides](#content-sources-class-config-overrides)
- `description` - Description
- `infra_policies` - List of Infra Policies associated with the Supervisor Namespace. See [Infra Policies](#infra-policies)
- `infra_policy_names` - List of non-mandatory Infra Policy names
- `phase` - Phase of the Supervisor Namespace
- `ready` - Whether the Supervisor Namespace is in a ready status or not
- `region_name` - Name of the Region
- `seg_name` - Service Engine Group associated with the Supervisor Namespace
- `shared_subnet_names` - Shared subnets associated with the Supervisor Namespace
- `storage_classes` - A set of Supervisor Namespace Storage Classes. See [Storage Classes](#storage-classes)
- `storage_classes_class_config_overrides` - Class Config Overrides for Storage Classes. See [Storage Classes Class Config Overrides](#storage-classes-class-config-overrides)
- `storage_classes_initial_class_config_overrides` - (**Deprecated**) Use `storage_classes_class_config_overrides` instead. See [Storage Classes Class Config Overrides](#storage-classes-class-config-overrides)
- `vpc_name` - Name of the VPC
- `vm_classes` - A set of Supervisor Namespace VM Classes. See [VM Classes](#vm-classes)
- `vm_classes_class_config_overrides` - Class Config Overrides for VM Classes. See [VM Classes Class Config Overrides](#vm-classes-class-config-overrides)
- `zones` - A set of Supervisor Namespace Zones. See [Zones](#zones)
- `zones_class_config_overrides` - Class Config Overrides for Zones. See [Zones Class Config Overrides](#zones-class-config-overrides)
- `zones_initial_class_config_overrides` - (**Deprecated**) Use `zones_class_config_overrides` instead. See [Zones Class Config Overrides](#zones-class-config-overrides)

## Conditions

The `conditions` attribute is a set of entries with the following structure:

- `message` - Human-readable message with details about the condition
- `reason` - Machine-readable CamelCase reason code
- `severity` - Severity level: `Info`, `Warning`, `Error`
- `status` - Condition status: `True`, `False`, `Unknown`
- `type` - Condition type identifier (e.g. `Ready`, `Realized`)

## Content Libraries

The `content_libraries` attribute is a set of entries with the following structure:

- `name` - Name of the content library
- `type` - Type of content source

## Content Sources Class Config Overrides

The `content_sources_class_config_overrides` is a set of entries that have the following structure:

- `name` - Name of the content library
- `type` - Type of content source

## Infra Policies

The `infra_policies` attribute is a set of entries with the following structure:

- `name` - Name of the Infra Policy
- `mandatory` - Whether the Infra Policy is auto-enforced when mandatory

## Storage Classes

The `storage_classes` is a set of entries that have the following structure:

- `limit` - Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the Storage Class

## Storage Classes Class Config Overrides

The `storage_classes_class_config_overrides` is a set of entries that have the following structure:

- `limit` - Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the Storage Class

## VM Classes

The `vm_classes` is a set of entries that have the following structure:

- `name` - Name of the VM Class

## VM Classes Class Config Overrides

The `vm_classes_class_config_overrides` is a set of entries that have the following structure:

- `name` - Name of the VM Class

## Zones

The `zones` is a set of entries that have the following structure:

- `cpu_limit` - CPU limit (format: `<number><unit>`, where `<unit>` can be `M` or `G`)
- `cpu_reservation` - CPU reservation (format: `<number><unit>`, where `<unit>` can be `M` or `G`)
- `marked_for_removal` - Indicates if this zone is scheduled for removal during a scale-down operation
- `memory_limit` - Memory limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `memory_reservation` - Memory reservation (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the Zone

## Zones Class Config Overrides

The `zones_class_config_overrides` is a set of entries that have the following structure:

- `cpu_limit` - CPU limit (format: `<number><unit>`, where `<unit>` can be `M` or `G`)
- `cpu_reservation` - CPU reservation (format: `<number><unit>`, where `<unit>` can be `M` or `G`)
- `memory_limit` - Memory limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `memory_reservation` - Memory reservation (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the Zone
