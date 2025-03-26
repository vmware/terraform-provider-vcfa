---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_nsx_manager"
sidebar_current: "docs-vcfa-data-source-nsx-manager"
description: |-
  Provides a data source for reading available NSX Managers attached to VMware Cloud Foundation Automation.
---

# vcfa\_nsx\_manager

Provides a data source for reading available NSX Managers attached to VMware Cloud Foundation Automation.

~> This data source can only be used by **System Administrators**

## Example Usage 

```hcl
data "vcfa_nsx_manager" "main" {
  name = "nsx-manager-one"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) NSX manager name

## Attribute reference

* `id` - ID of the manager
* `href` - Full URL of the manager

All attributes defined in
[`vcfa_nsx_manager`](/providers/vmware/vcfa/latest/docs/resources/nsx_manager#attribute-reference)
are supported.
