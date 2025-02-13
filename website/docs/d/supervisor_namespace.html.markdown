---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor_namespace"
sidebar_current: "docs-data-source-vcfa-supervisor-namespace"
description: |-
  Provides a VMware Cloud Foundation Automation Supervisor Namespace data source.
---

# vcfa\_supervisor\_namespace

Provides a VMware Cloud Foundation Automation Supervisor Namespace data source.

## Example Usage

```hcl
data "vcfa_supervisor_namespace" "supervisor_namespace" {
  name         = "demo-supervisor-namespace"
  project_name = "default-project"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Supervisor Namespace
* `project_name` - (Required) The name of the Project where the Supervisor Namespace belongs to

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
