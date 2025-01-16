---
layout: "vcfa"
page_title: "Provider: VMware Cloud Foundation Automation"
sidebar_current: "docs-vcfa-index"
description: |-
  The VMware Cloud Foundation Automation provider is used to interact with the resources supported by VMware Cloud Foundation Automation. The provider needs to be configured with the proper credentials before it can be used.
---

# VMware Cloud Foundation Automation Provider 1.0

The VMware Cloud Foundation Automation provider is used to interact with the resources supported by VMware Cloud Foundation Automation. The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources. Please refer to
[CHANGELOG.md](https://github.com/vmware/terraform-provider-vcfa/blob/main/CHANGELOG.md)
to track feature additions.

## Supported VCFA Versions

The following Cloud Director versions are supported by this provider:

* 9.0

## Connecting as System Administrator

When you want to manage resources across different organizations from a single configuration.

```hcl
# Configure the VMware Cloud Foundation Automation Provider
provider "vcfa" {
  user                 = "administrator"
  password             = var.vcfa_pass
  auth_type            = "integrated"
  org                  = "System"
  url                  = var.vcfa_url
  allow_unverified_ssl = var.vcfa_allow_unverified_ssl
}

# Create a new organization
resource "vcfa_org" "org" {
  name = "Org1"
  # ...
}

```

## Argument Reference

The following arguments are used to configure the VMware Cloud Foundation Automation Provider:

* `user` - (Required) This is the username for VCFA API operations. Can also be specified
  with the `VCFA_USER` environment variable. *v1.0+* `user` may be "serviceadministrator" (set `org` or
  `sysorg` to "System" in this case).
  
* `password` - (Required) This is the password for VCFA API operations. Can
  also be specified with the `VCFA_PASSWORD` environment variable.

* `auth_type` - (Optional) `integrated`, `token`, `api_token`, `service_account_token_file` or `saml_adfs`. 
  Default is `integrated`. Can also be set with `VCFA_AUTH_TYPE` environment variable. 
  * `integrated` - VCFA local users and LDAP users (provided LDAP is configured for Organization).
  * `saml_adfs` allows to use SAML login flow with Active Directory Federation
  Services (ADFS) using "/adfs/services/trust/13/usernamemixed" endpoint. Please note that
  credentials for ADFS should be formatted as `user@contoso.com` or `contoso.com\user`. 
  `saml_adfs_rpt_id` can be used to specify a different RPT ID.
  * `token` allows to specify token in `token` field.
  * `api_token` allows to specify an API token.
  * `api_token_file` allows to specify a file containing an API token.
  * `service_account_token_file` allows to specify a file containing a service account's token.
  
* `token` - (Optional; *v2.6+*) This is the bearer token that can be used instead of username
   and password (in combination with field `auth_type=token`). When this is set, username and
   password will be ignored, but should be left in configuration either empty or with any custom
   values. A token can be specified with the `VCFA_TOKEN` environment variable.
   Both a (deprecated) authorization token or a bearer token (*v3.1+*) can be used in this field.

* `api_token` - (Optional; *v3.5+*) This is the API token that a System or organization administrator can create and 
   distribute to users. It is used instead of username and password (in combination with `auth_type=api_token`). When
   this field is filled, username and password are ignored. An API token can also be specified with the `VCFA_API_TOKEN`
   environment variable. This token requires at least VCFA 9.0. There are restrictions to its use, as defined in
   [the documentation](https://docs.vmware.com/en/VMware-Cloud-Director/10.3/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-A1B3B2FA-7B2C-4EE1-9D1B-188BE703EEDE.html)

* `api_token_file` - (Optional; *v3.10+*)) Same as `api_token`, only provided 
   as a JSON file. Can also be specified with the `VCFA_API_TOKEN_FILE` environment variable.
 
* `service_account_token_file` - (Optional; *v3.9+, VCFA 9.0+*) This is the file that contains a Service Account API token. The
   path to the file could be provided as absolute or relative to the working directory. It is used instead of username
   and password (in combination with `auth_type=service_account_token_file`. The file can also be specified with the 
   `VCFA_SA_TOKEN_FILE` environment variable. There are restrictions to its use, as defined in 
   [the documentation](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-8CD3C8BE-3187-4769-B960-3E3315492C16.html)

* `allow_service_account_token_file` - (Optional; *v3.9+, VCFA 9.0+*) When using `auth_type=service_account_token_file`,
  if set to `true`, will suppress a warning to the user about the service account token file containing *sensitive information*.
  Can also be set with `VCFA_ALLOW_SA_TOKEN_FILE`.

* `saml_adfs_rpt_id` - (Optional) When using `auth_type=saml_adfs` VCFA SAML entity ID will be used
  as Relaying Party Trust Identifier (RPT ID) by default. If a different RPT ID is needed - one can
  set it using this field. It can also be set with `VCFA_SAML_ADFS_RPT_ID` environment variable.

* `saml_adfs_cookie` - (Optional; *v3.14+*) An additional cookie that can be injected when looking
up ADFS server from VCFA. Example `sso-preferred=yes; sso_redirect_org={{.Org}}`. `{{.Org}}` will be
replaced with actual Org during runtime.

* `org` - (Required) This is the Cloud Director Org on which to run API
  operations. Can also be specified with the `VCFA_ORG` environment
  variable.  
  *v2.0+* `org` may be set to "System" when connection as Sys Admin is desired
  (set `user` to "administrator" in this case).  
  Note: `org` value is case sensitive.
  
* `sysorg` - (Optional; *v2.0+*) - Organization for user authentication. Can also be
   specified with the `VCFA_SYS_ORG` environment variable. Set `sysorg` to "System" and
   `user` to "administrator" to free up `org` argument for setting a default organization
   for resources to use.
   
* `url` - (Required) This is the URL for the Cloud Director API endpoint. e.g.
  https://server.domain.com/api. Can also be specified with the `VCFA_URL` environment variable.
  
* `vdc` - (Optional) This is the virtual datacenter within Cloud Director to run
  API operations against. If not set the plugin will select the first virtual
  datacenter available to your Org. Can also be specified with the `VCFA_VDC` environment
  variable.

* `allow_unverified_ssl` - (Optional) Boolean that can be set to true to
  disable SSL certificate verification. This should be used with care as it
  could allow an attacker to intercept your auth token. If omitted, default
  value is false. Can also be specified with the
  `VCFA_ALLOW_UNVERIFIED_SSL` environment variable.

* `logging` - (Optional; *v2.0+*) Boolean that enables API calls logging from upstream library `go-vcloud-director`. 
   The logging file will record all API requests and responses, plus some debug information that is part of this 
   provider. Logging can also be activated using the `VCFA_API_LOGGING` environment variable.

* `logging_file` - (Optional; *v2.0+*) The name of the log file (when `logging` is enabled). By default is 
  `go-vcloud-director` and it can also be changed using the `VCFA_API_LOGGING_FILE` environment variable.
  
* `import_separator` - (Optional; *v2.5+*) The string to be used as separator with `terraform import`. By default
  it is a dot (`.`).

## Connection Cache (*1.0+*)

Cloud Director connection calls can be expensive, and if a definition file contains several resources, it may trigger 
multiple connections. There is a cache engine, disabled by default, which can be activated by the `VCFA_CACHE` 
environment variable. When enabled, the provider will not reconnect, but reuse an active connection for up to 20 
minutes, and then connect again.

[service-account]: /providers/vmware/vcfa/latest/docs/resources/service_account
[service-account-script]: https://github.com/vmware/terraform-provider-vcfa/blob/main/scripts/create_service_account.sh
[api-token]: /providers/vmware/vcfa/latest/docs/resource/api_token
