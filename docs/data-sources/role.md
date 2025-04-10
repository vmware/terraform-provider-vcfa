---
page_title: "VMware Cloud Foundation Automation: vcfa_role"
subcategory: ""
description: |-
  Provides a data source to read a Role from VMware Cloud Foundation Automation.
---

# vcfa_role

Provides a data source to read a Role from VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
# Reading a Role from System (Provider) organization
data "vcfa_org" "system" {
  name = "System"
}

data "vcfa_role" "sysadmin" {
  org_id = data.vcfa_org.system.id
  name   = "System Administrator"
}

# Reading a Role from a regular tenant
data "vcfa_org" "org1" {
  name = "org1"
}

data "vcfa_role" "sysadmin" {
  org_id = data.vcfa_org.org1.id
  name   = "Organization Administrator"
}
```

Sample output:

```shell
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

- `org_id` - (Required) The ID of Organization where the Role belongs. Can be fetched with [`vcfa_org` data source](/providers/vmware/vcfa/latest/docs/data-sources/org)
- `name` - (Required) The name of the Role

## Attribute Reference

- `read_only` - Whether this Role is read-only
- `description` - A description of the Role
- `bundle_key` - Key used for internationalization
- `rights` - Set of rights assigned to this Role

## More information

See [Roles management](/providers/vmware/vcfa/latest/docs/guides/roles_management) for a broader description of how roles and
rights work together.
