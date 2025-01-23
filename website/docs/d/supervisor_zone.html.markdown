---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor_zone"
sidebar_current: "docs-vcfa-data-source-supervisor-zone"
description: |-
  Provides a data source to read Supervisor Zones in VMware Cloud Foundation Automation.
---

# vcfa\_supervisor\_zone

Provides a data source to read Supervisor Zones in VMware Cloud Foundation Automation.

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

* `supervisor_id` - (Required) ID of parent Supervisor
* `name` - (Required) The name of Supervisor Zone

## Attribute Reference

* `vcenter_id` - vCenter server ID that contains this Supervisor
* `region_id` - Region ID that consumes this Supervisor
* `cpu_capacity_mhz` - The CPU capacity (in MHz) in this zone. Total CPU consumption in this zone
  cannot cross this limit.
* `cpu_used_mhz` - Total CPU used (in MHz) in this zone.
* `memory_capacity_mib` - The memory capacity (in mebibytes) in this zone. Total memory consumption
  in this zone cannot cross this limit.
* `memory_used_mib` - Total memory used (in mebibytes) in this zone.
