---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_api_token"
sidebar_current: "docs-vcfa-resource-api-token"
description: |-
  Provides a resource to manage API Tokens. API Tokens are an easy way to authenticate to VCFA. 
  They are user-based and have the same role as the user.
---

# vcfa\_api\_token 

Provides a resource to manage API Tokens. API Tokens are an easy way to authenticate to VCFA. 
They are user-based and have the same role as the user.

## Example usage

```hcl
resource "vcfa_api_token" "example_token" {
  name             = "example_token"
  file_name        = "example_token.json"
  allow_token_file = true
}
```

## Argument reference

The following arguments are supported:

* `name` - (Required) The unique name of the API Token for a specific user.
* `file_name` - (Required) The name of the file which will be created containing the API Token
* `allow_token_file` - (Required) An additional check that the user is aware that the file contains
  SENSITIVE information. Must be set to `true` or it will return a validation error.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. However, an experimental feature in Terraform 1.5+ allows also code generation.
See [Importing resources][importing-resources] for more information.

An existing API Token can be [imported][docs-import] into this resource via supplying
the full dot separated path. An example is below:

```
terraform import vcfa_api_token.example_token example_token
```

[docs-import]: https://www.terraform.io/docs/import/
[provider-api-token-file]: /providers/vmware/vcfa/latest/docs#api_token_file
