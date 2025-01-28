---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_role"
sidebar_current: "docs-vcfa-data-source-role"
description: |-
 Provides a VMware Cloud Foundation Automation Role. This can be used to read Roles.
---

# vcfa\_role

Provides a VMware Cloud Foundation Automation Role data source. This can be used to read Roles.

## Example Usage

```hcl
data "vcfa_org" "system" {
  name = "System"
}

data "vcfa_role" "sysadmin" {
  org_id = data.vcfa_org.system.id
  name   = "System Administrator"
}
```

Sample output:
```
sysadmin-out = {
  "bundle_key" = "ROLE_SYSTEM_ADMINISTRATOR"
  "description" = "Built-in rights for administering this installation"
  "id" = "urn:vcloud:role:67e119b7-083b-349e-8dfd-6cf0c19b83cf"
  "name" = "System Administrator"
  "org_id" = "urn:vcloud:org:a93c9db9-7471-3192-8d09-a8f7eeda85f9"
  "read_only" = true
  "rights" = toset([
    "AMQP Settings: Manage",
    "AMQP Settings: View",
    "API Explorer: View",
    "API Tokens: Manage",
    "API Tokens: Manage All",
    "Access Any Namespace with Elevated Privileges",
    "Access Any Namespace: Edit",
    "Access Any Namespace: View",
    "Access Control List: Manage",
    "Access Control List: View",
    "Access Metrics Endpoint",
    "Advisory Definitions: Create and Delete",
    "Advisory Definitions: Read",
    "Allowed Origins: Manage",
    "Allowed Origins: View",
    "Alternate Admin Entity: View",
    "Approvals: Manage",
    "Approvals: View",
    "Assembler Administrator",
    "Assembler User",
    "Assembler Viewer",
    "Audit: Manage",
    "Billing: View",
    "Blueprint Request: Manage",
    "Blueprint Request: View",
    "Blueprint: Edit",
    "Blueprint: Manage",
    "Blueprint: Publish",
    "Blueprint: View",
    "Catalog Instance: Manage",
    "Catalog: Manage",
    ...
  ])
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `name` - (Required) The name of the Role

## Attribute Reference

* `read_only` - Whether this Role is read-only
* `description` - A description of the Role
* `bundle_key` - Key used for internationalization
* `rights` - Set of rights assigned to this Role
