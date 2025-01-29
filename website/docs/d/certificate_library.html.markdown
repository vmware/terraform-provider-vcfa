---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_library_certificate"
sidebar_current: "docs-vcfa-data-source-certificate-library"
description: |-
  Provides a data source to read certificates in System or Org library.
---

# vcfa\_certificate\_library

Provides a data source to read certificate in System or Org library and reference in other resources.

~> Only `System Administrator` can access System certificates using this data source.

## Example Usage

```hcl
data "vcfa_library_certificate" "certificate1" {
  org   = "myOrg"
  alias = "SAML Encryption"
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Optional)  - alias (name) of certificate
* `id` - (Optional)  - ID of certificate

`alias` or `id` is required field.

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_library_certificate`](/providers/vmware/vcfa/latest/docs/resources/certificate_library) resource are available.