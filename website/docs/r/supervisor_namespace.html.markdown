---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor_namespace"
sidebar_current: "docs-vcfa-resource-supervisor-namespace"
description: |-
 Provides a VMware Cloud Foundation Automation Supervisor Namespace. This can be used to create, modify, and delete Supervisor namespaces.
---

# vcfa\_supervisor\_namespace

Provides a VMware Cloud Foundation Automation Supervisor Namespace. This can be used to create, modify, and delete Supervisor namespaces.

## Example Usage

```hcl
resource "vcfa_supervisor_namespace" "supervisor_namespace" {
  name_prefix  = "terraform-demo"
  project_name = "default-project"
  class_name   = "small"
  description  = "Supervisor Namespace created by Terraform"
  region_name  = "default-region"
  vpc_name     = "default-vpc"

  storage_classes_initial_class_config_overrides {
    limit_mib = 10000
    name      = "vSAN Default Storage Policy"
  }

  zones_initial_class_config_overrides {
    cpu_limit_mhz          = 1000
    cpu_reservation_mhz    = 0
    memory_limit_mib       = 1000
    memory_reservation_mib = 0
    name                   = "default-zone"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name_prefix` (Required) Prefix for the Supervisor Namespace Name. It must match RFC 1123 Label name (lower-case alphabet,
  numbers between 0 and 9 and hyphen `-`)
* `project_name` - (Required) The name of the Project where the Supervisor Namespace belongs to
* `class_name` - (Required) The name of the Supervisor Namespace Class
* `description` - (Optional) Description
* `storage_classes_initial_class_config_overrides` - (Required) A set of Supervisor Namespace Storage Classes Initial Class Config Overrides. At least one is required. See [Storage Classes Initial Class Config Overridess](#storage-classes-initial-class-config-overrides) section for details
* `region_name` - (Required)Name of the Region
* `vpc_name` - (Required) Name of the VPC
* `zones_initial_class_config_overrides` - (Required) A set of Supervisor Namespace Zones Initial Class Config Overrides. At least one is required. See [Zones Initial Class Config Overrides](#zones-initial-class-config-overrides) section for details

## Attribute Reference

* `name` - The name of the Supervisor Namespace
- `phase` - Phase of the Supervisor Namespace
- `ready` - Whether the Supervisor Namespace is in a ready status or not
- `storage_classes` - A set of Supervisor Namespace Storage Classes. See [Storage Classes](#storage-classes) section for details
- `vm_classes` - A set of Supervisor Namespace VM Classes. See [VM Classes](#vm-classes) section for details
- `zones` - A set of Supervisor Namespace Zones. See [Zones](#zones) section for details

<a id="storage-classes"></a>
## Storage Classes

The `storage_classes` is a set of metadata entries that have the following structure:

* `limit_mib` - Limit in MiB
* `name` - Name of the Storage Class

<a id="storage-classes-initial-class-config-overrides"></a>
## Storage Classes Initial Class Config Overrides

The `storage_classes_initial_class_config_overrides` is a set of metadata entries that have the following structure:

* `limit_mib` - Limit in MiB
* `name` - Name of the Storage Class

<a id="vm-classes"></a>
## VM Classes

The `vm_classes` is a set of metadata entries that have the following structure:

* `name` - Name of the VM Class

<a id="zones"></a>
## Zones

The `zones` is a set of metadata entries that have the following structure:

* `cpu_limit_mhz` - CPU limit in MHz
* `cpu_reservation_mhz` - CPU reservation in MHz
* `memory_limit_mib` - Memory limit in MiB
* `memory_reservation_mib` - Memory reservation in MiB
* `name` - Name of the Zone

<a id="zones-initial-class-config-overrides"></a>
## Zones Initial Class Config Overrides

The `zones_initial_class_config_overrides` is a set of metadata entries that have the following structure:

* `cpu_limit_mhz` - CPU limit in MHz
* `cpu_reservation_mhz` - CPU reservation in MHz
* `memory_limit_mib` - Memory limit in MiB
* `memory_reservation_mib` - Memory reservation in MiB
* `name` - Name of the Zone

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
    limit_mib = 10000
    name      = "vSAN Default Storage Policy"
  }

  zones_initial_class_config_overrides {
    cpu_limit_mhz          = 1000
    cpu_reservation_mhz    = 0
    memory_limit_mib       = 1000
    memory_reservation_mib = 0
    name                   = "default-zone"
  }
}
```

You can import such Supervisor Namespace into terraform state using this command

```
terraform import vcfa_supervisor_namespace.existing_supervisor_namespace "project_name.supervisor_namespace_name"
```

Where `project_name` is the name of the Project and `supervisor_namespace_name` is the name of the Supervisor Namespace.

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCFA_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the Supervisor Namespace as needed. Running `terraform plan` at this stage will show the difference between the minimal configuration file and the Supervisor Namespace's stored properties.
