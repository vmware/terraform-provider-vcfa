---
page_title: "VMware Cloud Foundation Automation: vcfa_org_oidc"
description: |-
  Provides a data source to read the OpenID Connect (OIDC) configuration of an Organization in VMware Cloud Foundation Automation.
---

# vcfa_org_oidc

Provides a data source to read the OpenID Connect (OIDC) configuration of an [Organization][vcfa_org-ds] in VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_org" "my_org" {
  name = "my-org"
}

data "vcfa_org_oidc" "oidc_settings" {
  org_id = data.vcfa_org.my_org.id
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) - ID of the [Organization][vcfa_org-ds] containing the OIDC settings

## Attribute Reference

All the arguments and attributes from [the `vcfa_org_oidc` resource](/providers/vmware/vcfa/latest/docs/resources/org_oidc) are available as read-only.

[vcfa_org-ds]: /providers/vmware/vcfa/latest/docs/data-sources/org