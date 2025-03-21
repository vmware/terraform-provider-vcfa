---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_global_role"
sidebar_current: "docs-vcfa-data-source-global-role"
description: |-
 Provides a VMware Cloud Foundation Automation Global Role data source . This can be used to read Global Roles.
---

# vcfa\_global\_role

Provides a data source to read a Global Role in VMware Cloud Foundation Automation, it can be used to retrieve details
of an existing Global Role, like the [Organizations][vcfa_org] in which it is published.

## Example Usage

```hcl
data "vcfa_global_role" "org_user" {
  name = "Organization User"
}

output "org_user_out" {
  value = data.vcfa_global_role.org_user
}
```

Sample output:
```
org_user_out = {
  "bundle_key" = "ROLE_ORGANIZATION_USER"
  "description" = "Rights given to an organization user"
  "id" = "urn:vcloud:globalRole:b49c5a15-73fd-4390-9e87-1e1d47e69c39"
  "name" = "Organization User"
  "org_ids" = toset([
    "urn:vcloud:org:9361eddf-cfe2-410f-8400-4b1b25b26cea",
  ])
  "publish_to_all_orgs" = true
  "read_only" = true
  "rights" = toset([
    "API Tokens: Manage",
    "Metrics: View",
    "Namespace Usage: Manage",
    "Namespace Usage: View",
  ])
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Global Role.

## Attribute Reference

* `description` - A description of the Global Role
* `bundle_key` - Key used for internationalization
* `rights` - List of rights assigned to this Global Role
* `publish_to_all_orgs` - When `true`, publishes the Global Role to all [Organizations][vcfa_org]
* `org_ids` - List of IDs of [Organizations][vcfa_org] to which this Global Role gets published. Ignored if `publish_to_all_orgs` is `true`
* `read_only` - Whether this Global Role is read-only

[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org