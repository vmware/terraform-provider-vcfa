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
- `description` - Description
- `phase` - Phase of the Supervisor Namespace
- `ready` - Whether the Supervisor Namespace is in a ready status or not
- `region_name` - Name of the Region
- `storage_classes` - A set of Supervisor Namespace Storage Classes. See [Storage Classes](#storage-classes) section for details
- `storage_classes_initial_class_config_overrides` - A set of Supervisor Namespace Storage Classes Initial Class Config Overrides. See [Storage Classes Initial Class Config Overridess](#storage-classes-initial-class-config-overrides) section for details
- `vpc_name` - Name of the VPC
- `vm_classes` - A set of Supervisor Namespace VM Classes. See [VM Classes](#vm-classes) section for details
- `zones` - A set of Supervisor Namespace Zones. See [Zones](#zones) section for details
- `zones_initial_class_config_overrides` - A set of Supervisor Namespace Zones Initial Class Config Overrides. See [Zones Initial Class Config Overrides](#zones-initial-class-config-overrides) section for details

<a id="storage-classes"></a>
## Storage Classes

The `storage_classes` is a set of entries that have the following structure:

- `limi` - Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the Storage Class

<a id="storage-classes-initial-class-config-overrides"></a>
## Storage Classes Initial Class Config Overrides

The `storage_classes_initial_class_config_overrides` is a set of entries that have the following structure:

- `limit` - Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the Storage Class

<a id="vm-classes"></a>
## VM Classes

The `vm_classes` is a set of entries that have the following structure:

- `name` - Name of the VM Class

<a id="zones"></a>
## Zones

The `zones` is a set of entries that have the following structure:

- `cpu_limit` - CPU limit (format: `<number><unit>`, where `<unit>` can be `M` or `G`)
- `cpu_reservation` - CPU reservation (format: `<number><unit>`, where `<unit>` can be `M` or `G`)
- `memory_limit` - Memory limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `memory_reservation` - Memory reservation (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the Zone

<a id="zones-initial-class-config-overrides"></a>
## Zones Initial Class Config Overrides

The `zones_initial_class_config_overrides` is a set of entries that have the following structure:

- `cpu_limit` - CPU limit (format: `<number><unit>`, where `<unit>` can be `M` or `G`)
- `cpu_reservation` - CPU reservation (format: `<number><unit>`, where `<unit>` can be `M` or `G`)
- `memory_limit` - Memory limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `memory_reservation` - Memory reservation (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the Zone
