---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_certificate"
sidebar_current: "docs-vcfa-data-source-certificate"
description: |-
  Provides a VMware Cloud Foundation Automation Certificate data source. This can be used to read the certificates that
  VCF Automation provides to others. They can be used when creating services that must be secured.
---

# vcfa\_certificate

Provides a VMware Cloud Foundation Automation Certificate data source. This can be used to read the certificates that
VCF Automation provides to others. They can be used when creating services that must be secured.

~> Only `System Administrator` can access System certificates using this data source.

## Example Usage

```hcl
data "vcfa_org" "system" {
  name = "System"
}

data "vcfa_certificate" "certificate1" {
  org_id = data.vcfa_org.system.id
  alias  = "SAML Encryption"
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) - ID of the Organization that owns the Certificate
* `alias` - (Optional)  - alias (name) of the Certificate. Either `alias` or `id` are required
* `id` - (Optional)  - ID of the Certificate. Either `alias` or `id` are required

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_certificate`](/providers/vmware/vcfa/latest/docs/resources/certificate) resource are available.