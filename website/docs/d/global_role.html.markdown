---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_global_role"
sidebar_current: "docs-vcfa-data-source-global-role"
description: |-
 Provides a VMware Cloud Foundation Automation Global Role data source . This can be used to read Global Roles.
---

# vcfa\_global\_role

Provides a VMware Cloud Foundation Automation Global Role data source. This can be used to read Global Roles.

## Example Usage

```hcl
data "vcfa_global_role" "vapp-author" {
  name = "vApp Author"
}
```

Sample output:
```

```


## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Global Role.

## Attribute Reference

* `description` - A description of the Global Role
* `bundle_key` - Key used for internationalization
* `rights` - List of rights assigned to this role
* `publish_to_all_orgs` - When true, publishes the Global Role to all Organizations
* `org_ids` - List of IDs of Organizations to which this Global Role gets published. Ignored if `publish_to_all_orgs` is true
* `read_only` - Whether this Global Role is read-only
