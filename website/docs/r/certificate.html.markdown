---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_certificate"
sidebar_current: "docs-vcfa-resource-certificate"
description: |-
  Provides a VMware Cloud Foundation Automation Certificate resource. It can be used to manage the certificates of
  servers that VCF Automation has trusted communication with. These certificates are used for verification of the
  credentials of other servers.
---

# vcfa\_certificate

Provides a VMware Cloud Foundation Automation Certificate resource. It can be used to manage the certificates of
servers that VCF Automation has trusted communication with. These certificates are used for verification of the
credentials of other servers.

_Used by: **Provider**, **Tenant**_

## Example Usage

```hcl
# Creating a Certificate in a tenant
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

# Creating a Certificate in System (Provider) organization
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

* `org_id` - (Required) ID of the [Organization](/providers/vmware/vcfa/latest/docs/resources/org) that owns the Certificate
* `alias` - (Required) Alias (name) of the Certificate
* `description` - (Optional) Certificate description
* `certificate` - (Required) Content of the Certificate. **Note:** Do not use trailing
  newlines in the Certificate, as VCFA trims them and `plan/apply` reports a difference in such case
* `private_key` - (Optional) Content of the private key
* `private_key_passphrase` - (Optional) Private key passphrase 

## Attribute Reference

The following attributes are exported on this resource:

* `id` - The ID of the Certificate added to the Certificates Library

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Certificate can be [imported][docs-import] into this resource
via supplying the full dot separated path Certificate. To import certificates in the System (Provider) Organization,
one can use `System`. An example is below:

```
terraform import vcfa_certificate.imported System.my-system-certificate-alias
```

```
terraform import vcfa_certificate.imported my-org.my-certificate-alias
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the Certificate named `my-system-certificate-alias` that exists in the System (Provider) Organization and
`my-certificate-alias` which is configured in Organization named `my-org`.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources