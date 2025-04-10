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

  storage_classes_initial_class_config_overrides {
    limit = "10000Mi"
    name  = "vSAN Default Storage Policy"
  }

  zones_initial_class_config_overrides {
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

- `name_prefix` (Required) Prefix for the Supervisor Namespace name. It must match RFC 1123 Label name (lower-case alphabet,
  numbers between 0 and 9 and hyphen `-`)
- `project_name` - (Required) The name of the Project where the Supervisor Namespace belongs to. Can be fetched
  with the Kubernetes provider [`kubernetes_resource`](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/data-sources/resource) data source
  for existing Projects, or with a reference to the [`kubernetes_manifest`](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest)
  if the Project is managed in the same Terraform configuration
- `class_name` - (Required) The name of the Supervisor Namespace Class
- `description` - (Optional) Description
- `storage_classes_initial_class_config_overrides` - (Required) A set of Supervisor Namespace Storage Classes Initial Class Config Overrides. At least one is required. See [Storage Classes Initial Class Config Overrides](#storage-classes-initial-class-config-overrides) section for details
- `region_name` - (Required) Name of the [Region](/providers/vmware/vcfa/latest/docs/data-sources/region)
- `vpc_name` - (Required) Name of the VPC
- `zones_initial_class_config_overrides` - (Required) A set of Supervisor Namespace Zones Initial Class Config Overrides. At least one is required. See [Zones Initial Class Config Overrides](#zones-initial-class-config-overrides) section for details

## Attribute Reference

- `name` - The name of the Supervisor Namespace
- `phase` - Phase of the Supervisor Namespace
- `ready` - Whether the Supervisor Namespace is in a ready status or not
- `storage_classes` - A set of Supervisor Namespace Storage Classes. See [Storage Classes](#storage-classes) section for details
- `vm_classes` - A set of Supervisor Namespace VM Classes. See [VM Classes](#vm-classes) section for details
- `zones` - A set of Supervisor Namespace Zones. See [Zones](#zones) section for details

<a id="storage-classes"></a>

## Storage Classes

The `storage_classes` is a set of entries that have the following structure:

- `limit` - Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the [Storage Class](/providers/vmware/vcfa/latest/docs/data-sources/storage_class)

<a id="storage-classes-initial-class-config-overrides"></a>

## Storage Classes Initial Class Config Overrides

The `storage_classes_initial_class_config_overrides` is a set of entries that have the following structure:

- `limit` - Limit (format: `<number><unit>`, where `<unit>` can be `Mi`, `Gi`, or `Ti`)
- `name` - Name of the [Storage Class](/providers/vmware/vcfa/latest/docs/data-sources/storage_class)

<a id="vm-classes"></a>

## VM Classes

The `vm_classes` is a set of entries that have the following structure:

- `name` - Name of the [VM Class](/providers/vmware/vcfa/latest/docs/data-sources/region_vm_class)

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

  storage_classes_initial_class_config_overrides {
    limit = "10000Mi"
    name  = "vSAN Default Storage Policy"
  }

  zones_initial_class_config_overrides {
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
