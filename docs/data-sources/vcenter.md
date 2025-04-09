---
page_title: "VMware Cloud Foundation Automation: vcfa_vcenter"
subcategory: ""
description: |-
  Provides a data source to read a vCenter server attached to VMware Cloud Foundation Automation.
---

# vcfa_vcenter

Provides a data source to read a vCenter server attached to VMware Cloud Foundation Automation.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_vcenter" "vc" {
  name = "vcenter-one"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) vCenter name

## Attribute reference

All attributes defined in
[`vcfa_vcenter`](/providers/vmware/vcfa/latest/docs/resources/vcenter#attribute-reference) are
supported.
