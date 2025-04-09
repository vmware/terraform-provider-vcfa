---
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor_zone"
subcategory: ""
description: |-
  Provides a data source to read a Supervisor Zone in VMware Cloud Foundation Automation. These are useful
  when configuring an Organization Region Quota "zone_resource_allocations" argument.
---

# vcfa_supervisor_zone

Provides a data source to read a Supervisor Zone in VMware Cloud Foundation Automation. These are useful
when configuring an [Organization Region Quota](/providers/vmware/vcfa/latest/docs/resources/org_region_quota) `zone_resource_allocations` argument.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_vcenter" "one" {
  name = "vcenter-one"
}

data "vcfa_supervisor" "one" {
  name       = "my-supervisor-name"
  vcenter_id = data.vcfa_vcenter.one.id
}

data "vcfa_supervisor_zone" "one" {
  supervisor_id = data.vcfa_supervisor.one.id
  name          = "domain-c8"
}
```

## Argument Reference

The following arguments are supported:

- `supervisor_id` - (Required) ID of parent [Supervisor](/providers/vmware/vcfa/latest/docs/data-sources/supervisor)
- `name` - (Required) The name of Supervisor Zone

## Attribute Reference

- `vcenter_id` - vCenter server ID that contains this Supervisor
- `region_id` - Region ID that consumes this Supervisor
- `cpu_capacity_mhz` - The CPU capacity (in MHz) in this zone. Total CPU consumption in this zone
  cannot cross this limit.
- `cpu_used_mhz` - Total CPU used (in MHz) in this zone.
- `memory_capacity_mib` - The memory capacity (in mebibytes) in this zone. Total memory consumption
  in this zone cannot cross this limit.
- `memory_used_mib` - Total memory used (in mebibytes) in this zone.
