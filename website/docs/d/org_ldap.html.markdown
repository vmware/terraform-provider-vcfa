---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_ldap"
sidebar_current: "docs-vcfa-data-source-org-ldap"
description: |-
  Provides a data source to read LDAP configuration for an Organization.
---

# vcfa\_org\_ldap

Provides a data source to read LDAP configuration for an Organization.

-> To read LDAP of the Provider (System) organization, please use [`vcfa_ldap` data source](/providers/vmware/vcfa/latest/docs/data-sources/ldap) instead

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

* `org_id` - (Required)  - ID of the organization containing the LDAP settings

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_ldap`](/providers/vmware/vcfa/latest/docs/resources/org_ldap) resource are available.
