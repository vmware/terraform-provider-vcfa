---
page_title: "VMware Cloud Foundation Automation: vcfa_org_external_group"
subcategory: ""
description: |-
  Provides a resource to manage external Groups from an Organization in VMware Cloud Foundation Automation. External Groups are sourced from an external identity provider such as LDAP, OAUTH or SAML.
---

# vcfa_org_external_group

Provides a resource to manage external Groups from an [Organization][vcfa_org] in VMware Cloud Foundation Automation. External Groups are sourced from an external identity provider such as LDAP, OAUTH or SAML.

_Used by: **Provider**_

## Example Usage

```hcl
resource "vcfa_org" "demo" {
  name         = "terraform-org"
  display_name = "terraform-org"
  description  = "Terraform demo"
  is_enabled   = true
}

data "vcfa_role" "org-user" {
  org_id = vcfa_org.demo.id
  name   = "Organization User"
}

resource "vcfa_org_ldap" "demo" {
  org_id = vcfa_org.demo.id
  # ... LDAP configuration ...
}

resource "vcfa_org_external_group" "demo" {
  org_id        = vcfa_org.demo.id
  provider_type = "LDAP"
  role_ids      = [data.vcfa_role.org-user.id]
  name          = "demo-external-group"

  depends_on = [vcfa_org_ldap.demo]
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) An [Organization][vcfa_org] ID for this External Group to be imported in.
  This field cannot be updated after creation
- `name` - (Required) The name that identifies this group in the external identity provider.
  This field cannot be updated after creation
- `provider_type` - (Required) The external identity provider type. Must be one of `LDAP`, `OAUTH` or `SAML`.
  This field cannot be updated after creation
- `role_ids` - (Required) A set of [Role][vcfa_global_role] IDs to assign to this External Group

## Attribute Reference

- `id` - The ID of the External Group
- `description` - The description of the External Group, as set by the identity provider

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Organization External Group configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

```shell
terraform import vcfa_org_external_group.imported my-org-name.my-group-name
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-group-name` external group from `my-org-name` Organization.

After that, you can expand the configuration file and either update or delete the Organization External Group as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Organization External Group's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org
[vcfa_global_role]: /providers/vmware/vcfa/latest/docs/resources/global_role
