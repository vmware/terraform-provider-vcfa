---
page_title: "VMware Cloud Foundation Automation: vcfa_distributed_vlan_connection"
subcategory: ""
description: |-
  Provides a data source to read a Distributed VLAN Connection in VMware Cloud Foundation Automation.
---

# vcfa_distributed_vlan_connection

Provides a data source to read a Distributed VLAN Connection in VMware Cloud Foundation Automation.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_region" "demo" {
  name = "demo-region"
}

data "vcfa_distributed_vlan_connection" "demo" {
  name      = "demo-distributed-vlan-connection"
  region_id = data.vcfa_region.demo.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the Distributed VLAN Connection
- `region_id` - (Required) The ID of the Region that has this Distributed VLAN Connection. Can be looked up using
  [`vcfa_region`](/providers/vmware/vcfa/latest/docs/data-sources/region)

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_distributed_vlan_connection`](/providers/vmware/vcfa/latest/docs/resources/distributed_vlan_connection) resource are available.
