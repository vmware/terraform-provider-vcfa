---
page_title: "VMware Cloud Foundation Automation: vcfa_org_oidc"
subcategory: ""
description: |-
  Provides a resource to configure or remove OpenID Connect (OIDC) for an Organization in VMware Cloud Foundation Automation.
---

# vcfa\_org\_oidc

Provides a resource to configure or remove OpenID Connect (OIDC) for an [Organization][vcfa_org] in VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

## Example Usage with Well-known Configuration Endpoint

The well-known configuration endpoint retrieves all the OpenID Connect settings values:

```hcl
data "vcfa_org" "my_org" {
  name = "my-org"
}

resource "vcfa_org_oidc" "oidc" {
  org_id                 = data.vcfa_org.my_org.id
  enabled                = true
  prefer_id_token        = false
  client_id              = "clientId"
  client_secret          = "clientSecret"
  max_clock_skew_seconds = 60
  wellknown_endpoint     = "https://my-idp.company.com/oidc/.well-known/openid-configuration"
}
```

Users can override any value retrieved by the well-known configuration:

```hcl
data "vcfa_org" "my_org" {
  name = "my-org"
}

resource "vcfa_org_oidc" "oidc" {
  org_id                 = data.vcfa_org.my_org.id
  enabled                = true
  prefer_id_token        = false
  client_id              = "clientId"
  client_secret          = "clientSecret"
  max_clock_skew_seconds = 60
  wellknown_endpoint     = "https://my-idp.company.com/oidc/.well-known/openid-configuration"

  # Overrides:
  access_token_endpoint = "https://my-other-idp.company.com/oidc/token"
  userinfo_endpoint     = "https://my-other-idp.company.com/oidc/userinfo"
}
```

Once the OIDC settings are created, if users want to restore an overridden value to the original one given by the
well-known configuration endpoint, they must perform an update in code to set the previous value explicitly.

In other words, removing the argument or setting it to `""` **won't** make the original value from the well-known configuration endpoint
to be restored during updates.

```hcl
data "vcfa_org" "my_org" {
  name = "my-org"
}

resource "vcfa_org_oidc" "oidc" {
  org_id                 = data.vcfa_org.my_org.id
  enabled                = true
  prefer_id_token        = false
  client_id              = "clientId"
  client_secret          = "clientSecret"
  max_clock_skew_seconds = 60
  wellknown_endpoint     = "https://my-idp.company.com/oidc/.well-known/openid-configuration"

  access_token_endpoint = "https://my-idp.company.com/oidc/token"          # Restores the previous value
  userinfo_endpoint     = "https://my-other-idp.company.com/oidc/userinfo" # Still overridden
}
```

## Example Usage without Well-known Configuration Endpoint

```hcl
data "vcfa_org" "my_org" {
  name = "my-org"
}

resource "vcfa_org_oidc" "oidc" {
  org_id                      = data.vcfa_org.my_org.id
  enabled                     = true
  prefer_id_token             = false
  client_id                   = "clientId"
  client_secret               = "clientSecret"
  max_clock_skew_seconds      = 60
  issuer_id                   = "https://my-idp.company.com/oidc"
  user_authorization_endpoint = "https://my-idp.company.com/oidc/authorize"
  access_token_endpoint       = "https://my-idp.company.com/oidc/token"
  userinfo_endpoint           = "https://my-idp.company.com/oidc/userinfo"
  scopes                      = ["openid", "profile", "email", "address", "phone", "offline_access"]
  claims_mapping {
    email      = "email"
    subject    = "sub"
    last_name  = "family_name"
    first_name = "given_name"
    full_name  = "name"
  }
  key {
    id              = "rsa1"
    algorithm       = "RSA"
    certificate     = file("certificate.pem")
    expiration_date = "2077-05-13"
  }
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) ID of the [Organization][vcfa_org] that will have the OpenID Connect settings configured. There must be only one
  resource `vcfa_org_oidc` per `org_id`, as there is only one OpenID configuration per Organization
- `client_id` - (Required) Client ID to use with the OIDC provider
- `client_secret` - (Required) Client Secret to use with the OIDC provider
- `enabled` - (Required) Either `true` or `false`, specifies whether the OIDC authentication is enabled for the given organization
- `wellknown_endpoint` - (Optional) This endpoint retrieves the OIDC provider configuration and automatically sets
  the following arguments, without setting them explicitly: `issuer_id`, `user_authorization_endpoint`, `access_token_endpoint`, 
  `userinfo_endpoint`, the `claims_mapping` block, the `key` blocks, and `scopes`. These mentioned attributes will be computed, and
  can be overridden by setting them explicitly in HCL code
- `issuer_id` - (Optional) The issuer ID for the OIDC provider.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**.
  This allows users to override the configuration given by `wellknown_endpoint`
- `user_authorization_endpoint` - (Optional) The endpoint to use for authorization.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**.
  This allows users to override the configuration given by `wellknown_endpoint`
- `access_token_endpoint` - (Optional) The endpoint to use for access tokens.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**.
  This allows users to override the configuration given by `wellknown_endpoint`
- `userinfo_endpoint` - (Optional) The endpoint to use for User Info.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**.
  This allows users to override the configuration given by `wellknown_endpoint`
- `prefer_id_token` - (Optional) If you want to combine claims from `userinfo_endpoint` and the ID Token, set this to `true`.
  The identity providers do not provide all the required claims set in `userinfo_endpoint`. By setting this argument to `true`,
  VMware Cloud Foundation Automation can fetch and consume claims from both sources
- `max_clock_skew_seconds` - (Optional) The maximum clock skew is the maximum allowable time difference between the client and server.
  This time compensates for any small-time differences in the timestamps when verifying tokens. The **default** value is `60` seconds
- `scopes` - (Optional) A set of scopes to use with the OIDC provider. They are used to authorize access to user details,
  by defining the permissions that the access tokens have to access user information.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**. This allows users
  to override the scopes given by `wellknown_endpoint`. Setting `scopes = []` will make Terraform to set the scopes provided originally
  by the `wellknown_endpoint`
- `claims_mapping` - (Optional) A single configuration block that specifies the claim mappings to use with the OIDC provider.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**. This allows users
  to override the claims given by `wellknown_endpoint`. The supported claims are:
  - `email` - Required if `wellknown_endpoint` doesn't give info about it
  - `subject` - Required if `wellknown_endpoint` doesn't give info about it
  - `last_name` - Required if `wellknown_endpoint` doesn't give info about it
  - `first_name` - Required if `wellknown_endpoint` doesn't give info about it
  - `full_name` - Required if `wellknown_endpoint` doesn't give info about it
  - `groups` - Optional
  - `roles` - Optional
- `key` - (Optional) One or more configuration blocks that specify the keys to use with the OIDC provider.
  If `wellknown_endpoint` is **not** set, then this argument is **required**. Otherwise, it is **optional**. This allows users
  to override the keys given by `wellknown_endpoint`. Each key requires the following:
  - `id` - Identifier of the key
  - `algorithm` - Algorithm used by the key. Can be `RSA` or `EC`
  - `certificate` - The contents of a PEM file to create/update the key
  - `expiration_date` - Expiration date for the key. The accepted format is `YYYY-MM-DD`, like `2077-12-31`
- `key_refresh_endpoint` - (Optional) Endpoint used to refresh the keys. If set, `key_refresh_period_hours` and `key_refresh_strategy` will be required.
  If `wellknown_endpoint` is set, then this argument will override the obtained endpoint
- `key_refresh_period_hours` - (Optional) Required if `key_refresh_endpoint` is set. Defines the frequency of key refresh. Maximum value is `720` (30 days)
- `key_refresh_strategy` - (Optional) Required if `key_refresh_endpoint` is set. Defines the strategy of key refresh. One of `ADD`, `REPLACE`, `EXPIRE_AFTER`
- `key_expire_duration_hours` - (Optional) Required if `key_refresh_endpoint` is set and `key_refresh_strategy=EXPIRE_AFTER`. Defines the expiration period of the key. Maximum value is `24`
- `ui_button_label` - (Optional) Customizes the label of the UI button of the login screen

## Attribute Reference

- `redirect_uri` - The client configuration redirect URI used to create a client application registration with an identity provider
  that complies with the OpenID Connect standard

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing OIDC configuration for an Organization can be [imported][docs-import] into this resource via supplying the path for an Organization.
For example, using this structure, representing an existing OIDC configuration that was **not** created using Terraform:

```hcl
data "vcfa_org" "my_org" {
  name = "my-org"
}

resource "vcfa_org_oidc" "my_org_oidc" {
  org_id = data.vcfa_org.my_org.id
}
```

You can import such OIDC configuration into terraform state using one of the following commands

```
terraform import vcfa_org_oidc.my_org_oidc organization_name
# OR
terraform import vcfa_org_oidc.my_org_oidc organization_id
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

After that, you must expand the configuration file before you can either update or delete the OIDC configuration. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org