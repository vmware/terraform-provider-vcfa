---
page_title: "VMware Cloud Foundation Automation: vcfa_org_local_user"
subcategory: ""
description: |-
  Provides a data source to read a local User from an Organization in VMware Cloud Foundation Automation.
---

# Data Source: vcfa_org_local_user

Provides a data source to read a local User from an [Organization][vcfa_org-ds] in VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "demo-org"
}

data "vcfa_org_local_user" "demo" {
  org_id   = data.vcfa_org.demo.id
  username = "demo-local-user"
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) A parent [Organization][vcfa_org-ds] ID for looking up this user
- `username` - (Required) The name of existing user

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_local_user`](/providers/vmware/vcfa/latest/docs/resources/org_local_user) resource are
available.

[vcfa_org-ds]: /providers/vmware/vcfa/latest/docs/data-sources/org
