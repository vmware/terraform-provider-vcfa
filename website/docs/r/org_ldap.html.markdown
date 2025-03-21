---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_ldap"
sidebar_current: "docs-vcfa-resource-org-ldap"
description: |-
  Provides a VMware Cloud Foundation Automation LDAP resource for Organizations. This can be used to create, delete, and 
  update LDAP configuration for an organization.
---

# vcfa\_org\_ldap

Provides a VMware Cloud Foundation Automation LDAP resource for Organizations.
This can be used to create, update and delete LDAP configuration for an Organization.

-> To configure LDAP for the Provider (System) organization, please use [`vcfa_provider_ldap` resource](/providers/vmware/vcfa/latest/docs/resources/provider_ldap) instead

## Example Usage 1 - Custom configuration

```hcl
data "vcfa_org" "my-org" {
  name = "my-org"
}

resource "vcfa_org_ldap" "my-org-ldap" {
  org_id                 = data.vcfa_org.my-org.id
  ldap_mode              = "CUSTOM"
  auto_trust_certificate = false # Because is_ssl = false
  custom_settings {
    server                  = "192.168.1.172"
    port                    = 389
    is_ssl                  = false
    username                = "cn=admin,dc=foo,dc=com"
    password                = "StrongPassword"
    base_distinguished_name = "dc=foo,dc=com"
    connector_type          = "OPEN_LDAP"
    user_attributes {
      object_class                = "inetOrgPerson"
      unique_identifier           = "uid"
      display_name                = "cn"
      username                    = "uid"
      given_name                  = "givenName"
      surname                     = "sn"
      telephone                   = "telephoneNumber"
      group_membership_identifier = "dn"
      email                       = "mail"
    }
    group_attributes {
      name                        = "cn"
      object_class                = "group"
      membership                  = "member"
      unique_identifier           = "cn"
      group_membership_identifier = "dn"
    }
  }
}
```

## Example Usage 2 - Using system configuration

```hcl
data "vcfa_org" "my-org" {
  name = "my-org"
}

resource "vcfa_org_ldap" "my-org-ldap" {
  org_id                 = data.vcfa_org.my-org.id
  ldap_mode              = "SYSTEM"
  auto_trust_certificate = false
  custom_user_ou         = "ou=Foo,dc=domain,dc=local base DN"
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) Org ID: there is only one LDAP configuration available for an organization. Thus, the resource can be identified by the Org.
* `ldap_mode` - (Required) One of `NONE`, `CUSTOM`, `SYSTEM`. Note that using `NONE` has the effect of removing the LDAP settings
* `auto_trust_certificate` - (Required) Defines if the LDAP certificate should automatically be trusted, only makes sense if `custom_settings.0.is_ssl=true` (see [Custom Settings](#custom-settings))
  The certificate is not removed by Terraform when this resource is destroyed
* `custom_user_ou` - (Optional) If `ldap_mode` is `SYSTEM`, specifies an LDAP `attribute=value` pair to use for OU (organizational unit)
* `custom_settings` - (Optional) LDAP server configuration. Becomes mandatory if `ldap_mode` is set to `CUSTOM`. See [Custom Settings](#custom-settings) below for details

<a id="custom-settings"></a>
## Custom Settings

The `custom_settings` section contains the configuration for the LDAP server

* `server` - (Required) The IP address or host name of the server providing the LDAP service
* `port` - (Required) Port number of the LDAP server (usually 389 for LDAP, 636 for LDAPS)
* `connector_type` - (Required) Type of connector: one of `OPEN_LDAP`, `ACTIVE_DIRECTORY`
* `base_distinguished_name` - (Required) LDAP search base
* `is_ssl` - (Optional) True if the LDAP service requires an SSL connection. If the certificate is not trusted already, `auto_trust_certificate=true` is needed.
* `username` - (Optional) _Username_ to use when logging in to LDAP, specified using LDAP attribute=value pairs 
  (for example: cn="ldap-admin", c="example", dc="com")
* `password` - (Optional) _Password_ for the user identified by `username`. This value is never returned on reads
* `user_attributes` - (Required) User settings when `ldap_mode` is `CUSTOM` See [User Attributes](#user-attributes) below for details
* `group_attributes` - (Required) Group settings when `ldap_mode` is `CUSTOM` See [Group Attributes](#group-attributes) below for details

<a id="user-attributes"></a>
### User Attributes

* `object_class` - (Required)  LDAP _objectClass_ of which imported users are members. For example, _"user"_ or _"person"_ attributes
* `unique_identifier` - (Required) LDAP attribute to use as the unique identifier for a user. For example, the _"objectGuid"_ attribute
* `username` - (Required) LDAP attribute to use when looking up a username to import. For example, _"userPrincipalName"_ or _"samAccountName"_ attributes
* `email` - (Required) LDAP attribute to use for the user's email address. For example, the _"mail"_ attribute
* `display_name` - (Required) LDAP attribute to use for the user's full name. For example, the _"displayName"_ attribute
* `given_name` - (Required) LDAP attribute to use for the user's given name. For example, the _"givenName"_ attribute
* `surname` - (Required) LDAP attribute to use for the user's surname. For example, the _"sn"_ attribute
* `telephone` - (Required) LDAP attribute to use for the user's telephone number. For example, the _"telephoneNumber"_ attribute
* `group_membership_identifier` - (Required) LDAP attribute that identifies a user as a member of a group. For example, the _"dn"_ attribute
* `group_back_link_identifier` - (Optional) LDAP attribute that returns the identifiers of all the groups of which the user is a member

<a id="group-attributes"></a>
### Group Attributes

* `object_class` - (Required) LDAP _objectClass_ of which imported groups are members. For example, the _"group"_ attribute
* `unique_identifier` - (Required) LDAP attribute to use as the unique identifier for a group. For example, the _"objectGuid"_ attribute
* `name` - (Required) LDAP attribute to use for the group name. For example, the _"cn"_ attribute
* `membership` - (Required) LDAP attribute to use when getting the members of a group. For example, the _"member"_ attribute
* `group_membership_identifier` - (Required) LDAP attribute that identifies a group as a member of another group. For example, the _"dn"_ attribute
* `group_back_link_identifier` - (Optional) LDAP group attribute used to identify a group member

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing LDAP configuration for an Organization can be [imported][docs-import] into this resource via supplying the path for an Organization name.
For example, using this structure, representing an existing LDAP configuration that was **not** created using Terraform:

```hcl
data "vcfa_org" "my-org" {
  name = "my-org"
}

resource "vcfa_org_ldap" "my-org-ldap" {
  org_id = data.vcfa_org.my-org.id
}
```

You can import such LDAP configuration into terraform state using one of the following commands

```
terraform import vcfa_org_ldap.my-org-ldap organization_name
```

After that, you must expand the configuration file before you can either update or delete the LDAP configuration. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the stored properties.

[docs-import]:https://www.terraform.io/docs/import/
