---
page_title: "VMware Cloud Foundation Automation: vcfa_org_ldap"
subcategory: ""
description: |-
  Provides a data source to read LDAP configuration from an Organization.
---

# Data Source: vcfa_org_ldap

Provides a data source to read LDAP configuration from an [Organization][vcfa_org-ds].

_Used by: **Provider**, **Tenant**_

-> To read LDAP of the Provider (System) organization, please use [`vcfa_provider_ldap` data source](/providers/vmware/vcfa/latest/docs/data-sources/provider_ldap) instead

## Example Usage

```hcl
data "vcfa_org" "my-org" {
  name = "my-org"
}

data "vcfa_org_ldap" "first" {
  org_id = data.vcfa_org.my-org.id
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required)  - ID of the [Organization][vcfa_org-ds] containing the LDAP settings

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_ldap`](/providers/vmware/vcfa/latest/docs/resources/org_ldap) resource are available, except `password`.

[vcfa_org-ds]: /providers/vmware/vcfa/latest/docs/data-sources/org
