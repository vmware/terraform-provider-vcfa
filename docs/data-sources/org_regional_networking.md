---
page_title: "VMware Cloud Foundation Automation: vcfa_org_regional_networking"
description: |-
  Provides a data source to read the Regional Networking Settings from an Organization in VMware Cloud Foundation Automation.
---

# vcfa_org_regional_networking

Provides a data source to read the Regional Networking Settings from an [Organization][vcfa_org-ds] in VMware Cloud Foundation Automation.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "my-org-name"
}

data "vcfa_org_regional_networking" "demo" {
  name   = "my-name"
  org_id = vcfa_org.demo.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of Organization Regional Networking configuration
- `org_id` - (Required) The ID of [Organization][vcfa_org-ds]

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_regional_networking`](/providers/vmware/vcfa/latest/docs/resources/org_regional_networking)
resource are available.

[vcfa_org-ds]: /providers/vmware/vcfa/latest/docs/data-sources/org
