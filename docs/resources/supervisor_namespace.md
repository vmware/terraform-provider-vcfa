---
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor_namespace"
subcategory: ""
description: |-
  Provides a resource to manage Supervisor Namespaces in VMware Cloud Foundation Automation.
---

# vcfa_supervisor_namespace

Provides a resource to manage Supervisor Namespaces in VMware Cloud Foundation Automation.

_Used by: **Tenant**_

-> This resource may use the [Kubernetes provider](https://registry.terraform.io/providers/hashicorp/kubernetes),
to see how to obtain the Kubeconfig, please check the [`vcfa_kubeconfig`](/providers/vmware/vcfa/latest/docs/data-sources/kubeconfig) data source.

## Example Usage

```hcl
# Retrieve the Kubeconfig, to be able to use the Kubernetes provider.
data "vcfa_kubeconfig" "kube_config" {}

provider "kubernetes" {
  host     = data.vcfa_kubeconfig.kube_config.host
  insecure = data.vcfa_kubeconfig.kube_config.insecure_skip_tls_verify
  token    = data.vcfa_kubeconfig.kube_config.token
}

# With the Kubernetes provider, fetch the Project for the Supervisor Namespace
data "kubernetes_resource" "project" {
  api_version = kubernetes_manifest.project.manifest["apiVersion"]
  kind        = kubernetes_manifest.project.manifest["kind"]
  metadata {
    name = kubernetes_manifest.project.manifest["metadata"]["name"]
  }
}

data "vcfa_region" "demo" {
  name = "default-region"
}

resource "vcfa_supervisor_namespace" "supervisor_namespace" {
  name_prefix  = "terraform-demo"
  project_name = data.kubernetes_resource.project.object["metadata"]["name"]
  class_name   = "small"
  description  = "Supervisor Namespace created by Terraform"
  region_name  = data.vcfa_region.demo.name
  vpc_name     = "default-vpc"

  storage_classes_class_config_overrides {
    limit = "10000Mi"
    name  = "vSAN Default Storage Policy"
  }

  zones_class_config_overrides {
    cpu_limit          = "1000M"
    cpu_reservation    = "0M"
    memory_limit       = "1000Mi"
    memory_reservation = "0Mi"
    name               = "default-zone"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name_prefix` - (Required) Prefix for the Supervisor Namespace name. It must match RFC 1123 Label name (lower-case alphabet,
  numbers between 0 and 9 and hyphen `-`)
- `project_name` - (Required) The name of the Project where the Supervisor Namespace belongs to. Can be fetched
  with the Kubernetes provider [`kubernetes_resource`](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/data-sources/resource) data source
  for existing Projects, or with a reference to the [`kubernetes_manifest`](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest)
  if the Project is managed in the same Terraform configuration
- `class_name` - (Required) The name of the Supervisor Namespace Class
- `description` - (Optional) Description
- `region_name` - (Required) Name of the [Region](/providers/vmware/vcfa/latest/docs/data-sources/region)
- `vpc_name` - (Required) Name of the VPC
- `content_sources_class_config_overrides` - (Optional) Class Config Overrides for Content Sources. Each entry has `name` and `type` (e.g. `ContentLibrary`). Update not supported. See [Content Sources Class Config Overrides](#content-sources-class-config-overrides)
- `infra_policy_names` - (Optional) List of non-mandatory Infra Policies to associate with the Supervisor Namespace
- `seg_name` - (Optional) Service Engine Group associated with the Supervisor Namespace
- `shared_subnet_names` - (Optional) List of shared subnets associated with the Supervisor Namespace
- `storage_classes_class_config_overrides` - (Optional) Class Config Overrides for Storage Classes. At least one of this or `storage_classes_initial_class_config_overrides` is required. See [Storage Classes Class Config Overrides](#storage-classes-class-config-overrides)
- `storage_classes_initial_class_config_overrides` - (Optional, **Deprecated**) Use `storage_classes_class_config_overrides` instead. Exactly one of this or `storage_classes_class_config_overrides` must be set. See [Storage Classes Class Config Overrides](#storage-classes-class-config-overrides)
- `vm_classes_class_config_overrides` - (Optional) Class Config Overrides for VM Classes. See [VM Classes Class Config Overrides](#vm-classes-class-config-overrides)
- `zones_class_config_overrides` - (Optional) Class Config Overrides for Zones. At least one of this or `zones_initial_class_config_overrides` is required. See [Zones Class Config Overrides](#zones-class-config-overrides)
- `zones_initial_class_config_overrides` - (Optional, **Deprecated**) Use `zones_class_config_overrides` instead. Exactly one of this or `zones_class_config_overrides` must be set. See [Zones Class Config Overrides](#zones-class-config-overrides)

## Attribute Reference

- `name` - The name of the Supervisor Namespace
- `phase` - Phase of the Supervisor Namespace
- `ready` - Whether the Supervisor Namespace is in a ready status or not
- `conditions` - Detailed conditions tracking Supervisor Namespace health and lifecycle events. See [Conditions](#conditions)
- `content_libraries` - Content libraries currently available in the Supervisor Namespace. See [Content Libraries](#content-libraries)
- `infra_policies` - List of Infra Policies associated with the Supervisor Namespace. See [Infra Policies](#infra-policies)
- `storage_classes` - A set of Supervisor Namespace Storage Classes. See [Storage Classes](#storage-classes)
- `vm_classes` - A set of Supervisor Namespace VM Classes. See [VM Classes](#vm-classes)
- `zones` - A set of Supervisor Namespace Zones. See [Zones](#zones)

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

- `name` - (Required) Name of the content library
- `type` - (Required) Type of content source (e.g. `ContentLibrary`). Update not supported.

## Infra Policies

The `infra_policies` attribute is a set of entries with the following structure:

- `name` - Name of the Infra Policy
- `mandatory` - Whether the Infra Policy is auto-enforced when mandatory

## Storage Classes

The `storage_classes` attribute is a set of entries that have the following structure:

- `limit` - Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the [Storage Class](/providers/vmware/vcfa/latest/docs/data-sources/storage_class)

## Storage Classes Class Config Overrides

The `storage_classes_class_config_overrides` is a set of entries that have the following structure:

- `limit` - Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the [Storage Class](/providers/vmware/vcfa/latest/docs/data-sources/storage_class)

## VM Classes

The `vm_classes` attribute is a set of entries that have the following structure:

- `name` - Name of the [VM Class](/providers/vmware/vcfa/latest/docs/data-sources/region_vm_class)

## VM Classes Class Config Overrides

The `vm_classes_class_config_overrides` is a set of entries that have the following structure:

- `name` - (Required) Name of the VM Class

## Zones

The `zones` attribute is a set of entries that have the following structure:

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

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Supervisor Namespace can be [imported][docs-import] into this resource via supplying the full dot separated path for a Supervisor Namespace.
For example, using this structure, representing an existing Supervisor Namespace that was **not** created using Terraform:

```hcl
resource "vcfa_supervisor_namespace" "existing_supervisor_namespace" {
  name_prefix  = "terraform-demo"
  project_name = "default-project"
  class_name   = "small"
  description  = "Supervisor Namespace created by Terraform"
  region_name  = "default-region"
  vpc_name     = "default-vpc"

  storage_classes_class_config_overrides {
    limit = "10000Mi"
    name  = "vSAN Default Storage Policy"
  }

  zones_class_config_overrides {
    cpu_limit          = "1000M"
    cpu_reservation    = "0M"
    memory_limit       = "1000Mi"
    memory_reservation = "0Mi"
    name               = "default-zone"
  }
}
```

You can import such Supervisor Namespace into terraform state using this command

```shell
terraform import vcfa_supervisor_namespace.existing_supervisor_namespace "project_name.supervisor_namespace_name"
```

Where `project_name` is the name of the Project and `supervisor_namespace_name` is the name of the Supervisor Namespace.

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

After that, you can expand the configuration file and either update or delete the Supervisor Namespace as needed.
Running `terraform plan` at this stage will show the difference between the minimal configuration file and the Supervisor Namespace's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
