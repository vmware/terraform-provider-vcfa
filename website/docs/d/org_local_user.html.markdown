---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_local_user"
sidebar_current: "docs-vcfa-data-source-org-local-user"
description: |-
  Provides a data source to read local users in VMware Cloud Foundation Automation Organizations.
---

# vcfa\_org\_local\_user

Provides a data source to read local users in VMware Cloud Foundation Automation Organizations.

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

- `org_id` - (Required) A parent Org ID for looking up this user
- `username` - (Required) The name of existing user

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_local_user`](/providers/vmware/vcfa/latest/docs/resources/org_local_user) resource are
available.
