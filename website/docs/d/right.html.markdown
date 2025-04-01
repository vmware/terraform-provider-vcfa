---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_right"
sidebar_current: "docs-vcfa-data-source-right"
description: |-
  Provides a data source to read available Rights in VMware Cloud Foundation Automation.
---

# vcfa\_right

Provides a data source to read available Rights in VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example usage

```hcl
data "vcfa_right" "some-right" {
  name = "Organization: Edit Limits"
}

output "some-right" {
  value = data.vcfa_right.some-right
}
```

Sample output:
```
some-right = {
  "bundle_key" = "RIGHT_ORG_OPERATIONS_LIMIT_EDIT"
  "category_id" = "urn:vcloud:rightsCategory:d6b25879-2ff0-3f82-933c-74eeb8aef591"
  "description" = "Organization: Edit Limits"
  "id" = "urn:vcloud:right:23272fe2-b7e3-3a82-8561-2dd7fda260e4"
  "implied_rights" = toset([
    {
      "id" = "urn:vcloud:right:30a64c60-c5cc-3b4f-a321-5e6f2bca02c2"
      "name" = "Organization: View"
    },
  ])
  "name" = "Organization: Edit Limits"
  "right_type" = "MODIFY"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Right.

## Attribute reference

* `description` - A description of the Right
* `category_id` - The ID of the category for this Right
* `bundle_key` - Key used for internationalization
* `right type` - Type of the Right (VIEW or MODIFY)
* `implied_rights` - List of Rights that are implied with this one

## More information

See [Roles management](/providers/vmware/vcfa/latest/docs/guides/roles_management) for a broader description of how roles and
rights work together.