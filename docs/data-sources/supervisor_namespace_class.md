---
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor_namespace_class"
subcategory: ""
description: |-
  Provides a data source to read Supervisor Namespace Classes from VMware Cloud Foundation Automation.
---

# vcfa_supervisor_namespace_class

Provides a data source to read a Supervisor Namespace Class from VMware Cloud Foundation Automation. To manage one, use the [`vcfa_supervisor_namespace_class`](/providers/vmware/vcfa/latest/docs/resources/supervisor_namespace_class) resource.

_Used by: **Tenant**_

## Example Usage

```hcl
data "vcfa_supervisor_namespace_class" "small" {
  name = "small"
}

output "small_vm_classes" {
  value = data.vcfa_supervisor_namespace_class.small.vm_classes
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the Supervisor Namespace Class.

## Attribute Reference

The following attributes are exported:

- `id` - Internal identifier, which is the name of the Supervisor Namespace Class.
- `description` - Description of the Supervisor Namespace Class.
- `storage_classes` - Storage classes, each with a `name` and a default `limit`.
- `vm_classes` - Names of the available VM classes.
- `content_sources` - Content sources, each with a `name` and a `type`.
- `default_limits` - Default per-zone limits: `cpu_limit`, `memory_limit`, `cpu_reservation` and `memory_reservation`.
- `zones` - Zones with their own limits: `name`, `cpu_limit`, `cpu_reservation`, `memory_limit`, `memory_reservation` and `vm_class_reservations` (each with a `vm_class_name` and a `count`).
