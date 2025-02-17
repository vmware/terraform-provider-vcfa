---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_ldap"
sidebar_current: "docs-vcfa-data-source-ldap"
description: |-
  Provides a data source to read LDAP configuration of the Provider (System).
---

# vcfa\_ldap

Provides a data source to read LDAP configuration of the Provider (System).

-> To read LDAP for a regular organization (tenant), please use [`vcfa_org_ldap` data source](/providers/vmware/vcfa/latest/docs/data-sources/org_ldap) instead

## Example Usage

```hcl
data "vcfa_ldap" "system-ldap" {
}
```

## Argument Reference

No arguments are required as the System LDAP configuration is unique per VCFA.

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_ldap`](/providers/vmware/vcfa/latest/docs/resources/ldap) resource are available, except `password`.
