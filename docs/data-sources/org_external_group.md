---
page_title: "VMware Cloud Foundation Automation: vcfa_org_external_group"
subcategory: ""
description: |-
  Provides a data source to read an external Group from an Organization in VMware Cloud Foundation Automation.
---

# vcfa_org_external_group

Provides a data source to read an external Group from an [Organization][vcfa_org-ds] in VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "demo-org"
}

data "vcfa_org_external_group" "demo" {
  org_id = data.vcfa_org.demo.id
  name   = "demo-external-group"
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) A parent [Organization][vcfa_org-ds] ID for looking up this group
- `name` - (Required) The name of the existing external group

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_external_group`](/providers/vmware/vcfa/latest/docs/resources/org_external_group) resource are
available.

[vcfa_org-ds]: /providers/vmware/vcfa/latest/docs/data-sources/org
