---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_nsxt_manager"
sidebar_current: "docs-vcfa-data-source-nsxt-manager"
description: |-
  Provides a data source for reading available NSX-T Managers attached to VMware Cloud Foundation Automation Tenant Manager.
---

# vcfa\_nsxt\_manager

Provides a data source for reading available NSX-T Managers attached to VMware Cloud Foundation Automation Tenant Manager.

## Example Usage 

```hcl
data "vcfa_nsxt_manager" "main" {
  name = "nsxt-manager-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) NSX-T manager name

## Attribute reference

* `id` - ID of the manager
* `href` - Full URL of the manager

All attributes defined in
[`vcfa_nsxt_manager`](/providers/vmware/vcfa/latest/docs/resources/tm_nsxt_manager#attribute-reference)
are supported.
