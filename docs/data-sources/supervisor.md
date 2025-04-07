---
page_title: "VMware Cloud Foundation Automation: vcfa_supervisor"
description: |-
  Provides a data source to read a Supervisor in VMware Cloud Foundation Automation.
---

# vcfa_supervisor

Provides a data source to read a Supervisor in VMware Cloud Foundation Automation.

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
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of Supervisor
- `vcenter_id` - (Required) ID of the [vCenter server](/providers/vmware/vcfa/latest/docs/data-sources/vcenter) that contains this Supervisor

## Attribute Reference

- `region_id` - Region ID that consumes this Supervisor
