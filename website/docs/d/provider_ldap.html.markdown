---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_provider_ldap"
sidebar_current: "docs-vcfa-data-source-ldap"
description: |-
  Provides a data source to read the LDAP configuration of the Provider (System).
---

# vcfa\_provider\_ldap

Provides a data source to read the LDAP configuration of the **Provider (System)**.

_Used by: **Provider**_

-> To read LDAP for a regular Organization (tenant), please use [`vcfa_org_ldap` data source](/providers/vmware/vcfa/latest/docs/data-sources/org_ldap) instead

## Example Usage

```hcl
data "vcfa_provider_ldap" "system-ldap" {}
```

## Argument Reference

No arguments are required as the System LDAP configuration is unique.

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_provider_ldap`](/providers/vmware/vcfa/latest/docs/resources/provider_ldap) resource are available, except `password`.
