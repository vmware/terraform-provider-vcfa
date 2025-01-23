---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_vcenter"
sidebar_current: "docs-vcfa-data-source-vcenter"
description: |-
  Provides a data source for reading vCenters attached to VMware Cloud Foundation Automation.
---

# vcfa\_vcenter

Provides a data source for reading vCenters attached to VMware Cloud Foundation Automation.

## Example Usage

```hcl
data "vcfa_vcenter" "vc" {
  name = "vcenter-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) vCenter name

## Attribute reference

All attributes defined in
[`vcfa_vcenter`](/providers/vmware/vcfa/latest/docs/resources/vcenter#attribute-reference) are
supported.
