---
page_title: "VMware Cloud Foundation Automation: vcfa_region_vm_class"
subcategory: ""
description: |-
  Provides a data source to read Region Virtual Machine Classes in VMware Cloud Foundation Automation. These are useful
  when configuring an Organization Region Quota "region_vm_class_ids" argument.
---

# Data Source: vcfa_region_vm_class

Provides a data source to read Region Virtual Machine Classes in VMware Cloud Foundation Automation. These are useful
when configuring an [Organization Region Quota](/providers/vmware/vcfa/latest/docs/resources/org_region_quota) `region_vm_class_ids` argument.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_region" "region1" {
  name = "my-region"
}

data "vcfa_region_vm_class" "vm_class" {
  name      = "best-effort-4xlarge"
  region_id = data.vcfa_region.region1.id
}

data "vcfa_region_vm_class" "vm_class2" {
  name      = "best-effort-8xlarge"
  region_id = data.vcfa_region.region1.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the Region VM Class
- `region_id` - (Required) An ID for the parent [Region](/providers/vmware/vcfa/latest/docs/data-sources/region)

## Attribute Reference

- `cpu_reservation_mhz` - CPU that a Virtual Machine reserves when this Region VM Class is applied
- `memory_reservation_mib` - Memory in MiB that a Virtual Machine reserves when this Region VM Class is applied
- `cpu_count` - Number of CPUs that a Virtual Machine gets when this Region VM Class is applied
- `memory_mib` - Memory in MiB that a Virtual Machine gets when this Region VM Class is applied
- `reserved` - Whether this Region VM Class can be used to reserve number of its instances within a namespace
