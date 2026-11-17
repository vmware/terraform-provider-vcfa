---
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor_namespace_class"
subcategory: ""
description: |-
  Provides a resource to manage Supervisor Namespace Classes in VMware Cloud Foundation Automation.
---

# vcfa_supervisor_namespace_class

Provides a resource to manage Supervisor Namespace Classes in VMware Cloud Foundation Automation.

A Supervisor Namespace Class is a template for Supervisor Namespaces: it defines the storage classes, VM classes, content sources and resource limits that the Supervisor Namespaces created from it get.

_Used by: **Tenant**_

## Example Usage

```hcl
resource "vcfa_supervisor_namespace_class" "small" {
  name        = "small"
  description = "Small Supervisor Namespaces for development"

  storage_classes = [
    { name = "vsan-default-storage-policy", limit = "100Gi" },
  ]

  vm_classes = ["*"]

  content_sources = [
    { name = "my-content-library" },
  ]

  zones = [
    {
      name               = "zone-1"
      cpu_limit          = "20"
      memory_limit       = "128Gi"
      cpu_reservation    = "8"
      memory_reservation = "64Gi"

      vm_class_reservations = [
        { vm_class_name = "best-effort-small", count = 2 },
      ]
    },
  ]
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the Supervisor Namespace Class. It must be a lowercase DNS label of at most 63 characters. Changing it forces a new resource to be created.
- `description` - (Optional) Description of the Supervisor Namespace Class, of at most 2000 characters.
- `storage_classes` - (Required) Storage classes available in the Supervisor Namespaces created from the class. At least one is required:
  - `name` - (Required) Name of the storage class, or `*` for all of them.
  - `limit` - (Required) Default storage limit, e.g. `100Gi`.
- `vm_classes` - (Required) Names of the VM classes available in the Supervisor Namespaces created from the class, or `*` for all of them. At least one is required.
- `content_sources` - (Optional) Content sources available in the Supervisor Namespaces created from the class:
  - `name` - (Required) Name of the content source.
  - `type` - (Optional) Type of the content source. Only `ContentLibrary`, the default, is supported.
- `default_limits` - (Optional) Default resource limits per zone. Exactly one of `default_limits` and `zones` is required:
  - `cpu_limit`, `cpu_reservation` - (Required) CPU quantities, e.g. `10` or `5000m`.
  - `memory_limit`, `memory_reservation` - (Required) Memory quantities, e.g. `64Gi`.
- `zones` - (Optional) Zones with their own limits. Exactly one of `default_limits` and `zones` is required:
  - `name` - (Required) Name of the zone, or `*` for all of them.
  - `cpu_limit`, `cpu_reservation`, `memory_limit`, `memory_reservation` - (Required) Limits and reservations of the zone, in the same format as in `default_limits`.
  - `vm_class_reservations` - (Optional) VM classes to reserve in the zone by default:
    - `vm_class_name` - (Required) Name of the VM class.
    - `count` - (Required) Number of instances to reserve, at least 1.

## Attribute Reference

In addition to the arguments above, the following computed attributes are exported:

- `id` - Internal identifier, which is the name of the Supervisor Namespace Class.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows also code generation. See [Importing resources][importing-resources] for more information.

An existing Supervisor Namespace Class can be [imported][docs-import] into this resource via its name:

```shell
terraform import vcfa_supervisor_namespace_class.small "small"
```

[docs-import]: https://developer.hashicorp.com/terraform/cli/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
