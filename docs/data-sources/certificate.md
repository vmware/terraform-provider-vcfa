---
page_title: "VMware Cloud Foundation Automation: vcfa_certificate"
subcategory: ""
description: |-
  Provides a data source to read a Certificate in VMware Cloud Foundation Automation. This can be used to read the certificates that
  VCF Automation provides to others. They can be used when creating services that must be secured.
---

# Data Source: vcfa_certificate

Provides a data source to read a Certificate in VMware Cloud Foundation Automation. This can be used to read the certificates that
VCF Automation provides to others. They can be used when creating services that must be secured.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
data "vcfa_org" "system" {
  name = "System"
}

# This certificate if read from the Provider ("System" org)
data "vcfa_certificate" "system_certificate" {
  org_id = data.vcfa_org.system.id
  alias  = "SAML Encryption"
}

data "vcfa_org" "tenant" {
  name = "my-tenant1"
}

# This certificate if read from the tenant
data "vcfa_certificate" "tenant_certificate" {
  org_id = data.vcfa_org.tenant.id
  alias  = "Example certificate"
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) - ID of the Organization that owns the Certificate
- `alias` - (Optional)  - Alias (name) of the Certificate. Either `alias` or `id` are required
- `id` - (Optional)  - ID of the Certificate. Either `alias` or `id` are required

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_certificate`](/providers/vmware/vcfa/latest/docs/resources/certificate) resource are available.
