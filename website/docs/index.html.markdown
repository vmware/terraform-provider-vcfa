---
layout: "vcfa"
page_title: "Provider: VMware Cloud Foundation Automation"
sidebar_current: "docs-vcfa-index"
description: |-
  The VMware Cloud Foundation Automation provider is used to interact with the resources supported by VMware Cloud Foundation Automation. The provider needs to be configured with the proper credentials before it can be used.
---

# VMware Cloud Foundation Automation Provider 0.1

The VMware Cloud Foundation Automation provider is used to interact with the resources supported by VMware Cloud Foundation Automation. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources. Please refer to
[CHANGELOG.md](https://github.com/vmware/terraform-provider-vcfa/blob/main/CHANGELOG.md)
to track feature additions.

## Supported VCFA Versions

The following VCFA versions are supported by this provider:

- 9.0

## Connecting as System Administrator

When you want to manage resources across different organizations from a single configuration.

```hcl
terraform {
  required_providers {
    vcfa = {
      source  = "vmware/vcfa"
      version = "= 0.1.0"
    }
  }
}

# Configure the VMware Cloud Foundation Automation Provider
provider "vcfa" {
  user                 = "serviceadministrator"
  password             = var.vcfa_pass
  auth_type            = "integrated"
  org                  = "System"
  url                  = var.vcfa_url
  allow_unverified_ssl = var.vcfa_allow_unverified_ssl
  logging              = true # Enables logging
  logging_file         = "vcfa.log"
}

# Fetch the Tenant Manager version
data "vcfa_tm_version" "version" {
  condition         = ">= 10.7.0"
  fail_if_not_match = false
}

```

## Argument Reference

The following arguments are used to configure the VMware Cloud Foundation Automation Provider:

- `user` - (Required) This is the username for VCFA API operations. Can also be specified
  with the `VCFA_USER` environment variable. `user` may be "serviceadministrator" (set `org` or
  `sysorg` to "System" in this case).

- `password` - (Required) This is the password for VCFA API operations. Can
  also be specified with the `VCFA_PASSWORD` environment variable.

- `auth_type` - (Optional) `integrated`, `token`, `api_token` or `service_account_token_file`. 
  Default is `integrated`. Can also be set with `VCFA_AUTH_TYPE` environment variable. 
    - `integrated` - VCFA local users and LDAP users (provided LDAP is configured for Organization).
  - `token` allows to specify token in `token` field.
  - `api_token` allows to specify an API token.
  - `api_token_file` allows to specify a file containing an API token.
  - `service_account_token_file` allows to specify a file containing a service account's token.
  
- `token` - (Optional) This is the bearer token that can be used instead of username
   and password (in combination with field `auth_type=token`). When this is set, username and
   password will be ignored, but should be left in configuration either empty or with any custom
   values. A token can be specified with the `VCFA_TOKEN` environment variable.
   Both a (deprecated) authorization token or a bearer token can be used in this field.

- `api_token` - (Optional) This is the API token that a System or organization administrator can create and 
   distribute to users. It is used instead of username and password (in combination with `auth_type=api_token`). When
   this field is filled, username and password are ignored. An API token can also be specified with the `VCFA_API_TOKEN`
   environment variable. This token requires at least VCFA 9.0. There are restrictions to its use, as defined in the documentation

- `api_token_file` - (Optional) Same as `api_token`, only provided 
   as a JSON file. Can also be specified with the `VCFA_API_TOKEN_FILE` environment variable.
 
- `service_account_token_file` - (Optional) This is the file that contains a Service Account API token. The
   path to the file could be provided as absolute or relative to the working directory. It is used instead of username
   and password (in combination with `auth_type=service_account_token_file`. The file can also be specified with the 
   `VCFA_SA_TOKEN_FILE` environment variable. There are restrictions to its use, as defined in 
   the documentation

- `allow_service_account_token_file` - (Optional) When using `auth_type=service_account_token_file`,
  if set to `true`, will suppress a warning to the user about the service account token file containing *sensitive information*.
  Can also be set with `VCFA_ALLOW_SA_TOKEN_FILE`.

- `org` - (Required) This is the Tenant Manager Org on which to run API
  operations. Can also be specified with the `VCFA_ORG` environment
  variable.  
  `org` may be set to "System" when connection as Sys Admin is desired
  (set `user` to "administrator" in this case).  
  Note: `org` value is case sensitive.
  
- `sysorg` - (Optional) - Organization for user authentication. Can also be
   specified with the `VCFA_SYS_ORG` environment variable. Set `sysorg` to "System" and
   `user` to "administrator" to free up `org` argument for setting a default organization
   for resources to use.
   
- `url` - (Required) This is the URL for the Tenant Manager API endpoint. e.g.
  https://server.domain.com/tm/api. Can also be specified with the `VCFA_URL` environment variable.
  
- `vdc` - (Optional) This is the virtual datacenter within Tenant Manager to run
  API operations against. If not set the plugin will select the first virtual
  datacenter available to your Org. Can also be specified with the `VCFA_VDC` environment
  variable.

- `allow_unverified_ssl` - (Optional) Boolean that can be set to true to
  disable SSL certificate verification. This should be used with care as it
  could allow an attacker to intercept your auth token. If omitted, default
  value is false. Can also be specified with the
  `VCFA_ALLOW_UNVERIFIED_SSL` environment variable.

- `logging` - (Optional) Boolean that enables API calls logging from upstream library `go-vcloud-director`. 
   The logging file will record all API requests and responses, plus some debug information that is part of this 
   provider. Logging can also be activated using the `VCFA_API_LOGGING` environment variable.

- `logging_file` - (Optional) The name of the log file (when `logging` is enabled). By default is 
  `go-vcloud-director` and it can also be changed using the `VCFA_API_LOGGING_FILE` environment variable.
  
- `import_separator` - (Optional) The string to be used as separator with `terraform import`. By default
  it is a dot (`.`).

## Connection Cache

VCFA connection calls can be expensive, and if a definition file contains several resources, it may trigger 
multiple connections. There is a cache engine, disabled by default, which can be activated by the `VCFA_CACHE` 
environment variable. When enabled, the provider will not reconnect, but reuse an active connection for up to 20 
minutes, and then connect again.

[service-account]: /providers/vmware/vcfa/latest/docs/resources/service_account
[service-account-script]: https://github.com/vmware/terraform-provider-vcfa/blob/main/scripts/create_service_account.sh
[api-token]: /providers/vmware/vcfa/latest/docs/resource/api_token
