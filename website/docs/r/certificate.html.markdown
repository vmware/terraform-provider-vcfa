---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_certificate"
sidebar_current: "docs-vcfa-resource-certificate"
description: |-
  Provides a VMware Cloud Foundation Automation Certificate resource. This can be used to manage the certificates that
  VCF Automation provides to others. They can be used when creating services that must be secured.
---

# vcfa\_certificate

Provides a VMware Cloud Foundation Automation Certificate resource. This can be used to manage the certificates of
servers that VCF Automation has trusted communication with. These certificates are used for verification of the
credentials of other servers.

~> Only `System Administrator` can create this resource.

## Example Usage

```hcl
data "vcfa_org" "org1" {
  name = "myOrg"
}

resource "vcfa_certificate" "new-certificate" {
  org_id                 = data.vcfa_org.org1.id
  alias                  = "SAML certificate"
  description            = "Created by Terraform VCFA Provider"
  certificate            = file("/home/user/cert.pem")
  private_key            = file("/home/user/key.pem")
  private_key_passphrase = "passphrase"
}
```

Creating a Certificate in System (Provider) context:

```hcl
data "vcfa_org" "system" {
  name = "System"
}

resource "vcfa_certificate" "new-certificate-for-system" {
  org_id                 = data.vcfa_org.system.id
  alias                  = "provider certificate"
  description            = "Created by Terraform VCFA Provider"
  certificate            = file("/home/user/provider-cert.pem")
  private_key            = file("/home/user/provider-key.pem")
  private_key_passphrase = "passphrase"
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) ID of the Organization that owns the Certificate
* `alias` - (Required) Alias (name) of the Certificate
* `description` - (Optional) Certificate description
* `certificate` - (Required) Content of the Certificate. **Note:** Do not use trailing
  newlines in the Certificate, as VCFA trims them and `plan/apply` reports a difference in such case
* `private_key` - (Optional) Content of the private key
* `private_key_passphrase` - (Optional) Private key pass phrase 

## Attribute Reference

The following attributes are exported on this resource:

* `id` - The added to Certificate library ID

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Certificate from library can be [imported][docs-import] into this resource
via supplying the full dot separated path Certificate in library. `System` org should be used to import system
certificates. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcfa_certificate.imported my-org.my-certificate-alias
```

The above would import the Certificate named `my-certificate-alias` which is configured in organization named `my-org`.
