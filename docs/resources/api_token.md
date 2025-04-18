---
page_title: "VMware Cloud Foundation Automation: vcfa_api_token"
subcategory: ""
description: |-
  Provides a resource to manage API Tokens. API Tokens are an easy way to authenticate to VMware Cloud Foundation Automation. 
  They are user-based and have the same Role as the user.
---

# vcfa_api_token

Provides a resource to manage API Tokens. API Tokens are an easy way to authenticate to VMware Cloud Foundation Automation.
They are user-based and have the same [Role](/providers/vmware/vcfa/latest/docs/resources/role) as the user.

_Used by: **Provider**, **Tenant**_

## Example usage

```hcl
# The vcfa_api_token below generates an API Token for the "bob" user configured in the provider block.
provider "vcfa" {
  user     = "bob"
  password = var.my_password
  org      = "tenant1"
  # Omitted arguments...
}

resource "vcfa_api_token" "example_token" {
  name             = "example_token"
  file_name        = "example_token.json"
  allow_token_file = true
}

# Creating an API Token as the System Administrator.
provider "vcfa" {
  user     = "serviceadministrator"
  password = var.system_password
  org      = "System"
  # Omitted arguments...
}

resource "vcfa_api_token" "system_token" {
  name             = "system_token"
  file_name        = "system_token.json"
  allow_token_file = true
}
```

## Argument reference

The following arguments are supported:

- `name` - (Required) The unique name of the API Token for a specific user.
- `file_name` - (Required) The name of the file which will be created containing the API Token. The file will have the following
JSON contents:

```json
{
  "token_type": "API Token",
  "refresh_token": "24JVMAQvaayIDuA7wayPPfa376mrfraB",
  "updated_by": "terraform-provider-vcfa/ (darwin/amd64; isProvider:true)",
  "updated_on": "2025-01-29T10:00:43+01:00"
 }
```

- `allow_token_file` - (Required) An additional check that the user is aware that the file contains
  **SENSITIVE** information. Must be set to `true` or it will return a validation error.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. However, an experimental feature in Terraform 1.5+ allows also code generation.
See [Importing resources][importing-resources] for more information.

An existing API Token can be [imported][docs-import] into this resource via supplying
the full dot separated path. An example is below:

```shell
terraform import vcfa_api_token.example_token example_token
```

[docs-import]: https://www.terraform.io/docs/import/
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
