---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_vdc"
sidebar_current: "docs-vcfa-resource-org-vdc"
description: |-
  Provides a resource to manage VMware Cloud Foundation Automation Organization VDCs (Region Quotas).
---

# vcfa\_org\_vdc

Provides a resource to manage VMware Cloud Foundation Automation Organization VDCs (Region Quotas).

## Example Usage

```hcl
data "vcfa_vcenter" "vc" {
  name = "my-vcenter"
}

data "vcfa_supervisor" "supervisor" {
  name       = "my-supervisor"
  vcenter_id = vcfa_vcenter.vc.id
}

data "vcfa_region" "one" {
  name = "region-one"
}

data "vcfa_region_zone" "one" {
  region_id = data.vcfa_region.one.id
  name      = "my-zone"
}

data "vcfa_region_vm_class" "vm_class1" {
  name      = "best-effort-4xlarge"
  region_id = data.vcfa_region.region1.id
}

data "vcfa_region_vm_class" "vm_class2" {
  name      = "best-effort-8xlarge"
  region_id = data.vcfa_region.region1.id
}

resource "vcfa_org_vdc" "first" {
  org_id         = vcfa_org.test.id
  region_id      = data.vcfa_region.one.id
  supervisor_ids = [data.vcfa_supervisor.test.id]
  zone_resource_allocations {
    region_zone_id         = data.vcfa_region_zone.one.id
    cpu_limit_mhz          = 2000
    cpu_reservation_mhz    = 100
    memory_limit_mib       = 1024
    memory_reservation_mib = 512
  }
  region_vm_class_ids = [
    data.vcfa_region_vm_class.vm_class1.id,
    data.vcfa_region_vm_class.vm_class2.id,
  ]
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) An Org ID for this Organization VDC (Region Quota) to be assigned to
- `region_id` - (Required) A Region ID that this Organization VDC (Region Quota) should be backed by
- `supervisor_ids` - (Required) A set of Supervisor IDs that back this Organization VDC (Region Quota). Can be looked up
  using [`vcfa_supervisor`](/providers/vmware/vcfa/latest/docs/data-sources/supervisor) data source
- `zone_resource_allocations` - (Required) A set of Zone Resource Allocation definitions. See [Zone Resource Allocations](#zone-resource-allocations-block)
- `region_vm_class_ids` - (Required) A set of Region VM Class IDs. These can be fetched with [`vcfa_region_vm_class` data source](/providers/vmware/vcfa/latest/docs/data-sources/region_vm_class)

<a id="zone-resource-allocations-block"></a>
## Zone Resource Allocations

- `region_zone_id` - (Required) Can be looked up using
  [`vcfa_region_zone`](/providers/vmware/vcfa/latest/docs/data-sources/region_zone) data source
- `cpu_limit_mhz` - (Required) Maximum CPU consumption limit in MHz
- `cpu_reservation_mhz` - (Required) Defines reserved CPU capacity in MHz
- `memory_limit_mib` - (Required) Maximum memory consumption limit in MiB
- `memory_reservation_mib` - (Required) Defines reserved memory capacity in MiB

A computed attribute `region_zone_name` will be set in each `zone_resource_allocations` block.


## Attribute Reference

The following attributes are exported on this resource:

- `name` - The name of the Organization VDC (Region Quota), it's assigned on creation and can't be changed
- `status` - The creation status of the Organization VDC (Region Quota). Possible values are `READY`, `NOT_READY`, `ERROR`,
  `FAILED`

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Organization VDC (Region Quota) configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcfa_org_vdc.imported my-org-name.my-region-name
```

The above would import the Organization VDC (Region Quota) that belongs to `my-org-name` Organization and `my-region-name` Region.