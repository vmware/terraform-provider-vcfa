---
page_title: "VMware Cloud Foundation Automation: vcfa_org"
subcategory: ""
description: |-
  Provides a data source to read an Organization in VMware Cloud Foundation Automation.
---

# Data Source: vcfa_org

Provides a data source to read an Organization in VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_org" "existing" {
  name = "my-org-name"
}

# Reads the System (Provider) organization. This can only be done by System administrators (Providers)
data "vcfa_org" "system" {
  name = "System"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of organization. Setting `name="System"`" will fetch the Provider organization,
  this can only be done by System administrators

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org`](/providers/vmware/vcfa/latest/docs/resources/org) resource are available.
