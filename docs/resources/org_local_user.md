---
page_title: "VMware Cloud Foundation Automation: vcfa_org_local_user"
description: |-
  Provides a resource to manage local Users from an Organization in VMware Cloud Foundation Automation.
---

# vcfa_org_local_user

Provides a resource to manage Local Users from an [Organization][vcfa_org] in VMware Cloud Foundation Automation.

_Used by: **Provider**_

## Example Usage

```hcl
resource "vcfa_org" "demo" {
  name         = "terraform-org"
  display_name = "terraform-org"
  description  = "Terraform demo"
  is_enabled   = true
}

data "vcfa_role" "org-admin" {
  org_id = vcfa_org.demo.id
  name   = "Organization Administrator"
}

data "vcfa_role" "org-user" {
  org_id = vcfa_org.demo.id
  name   = "Organization User"
}

resource "vcfa_org_local_user" "demo" {
  org_id   = vcfa_org.demo.id
  role_ids = [data.vcfa_role.org-admin.id, data.vcfa_role.org-user.id]
  username = "demo-local-user"
  password = "CHANGE-ME"
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) An [Organization][vcfa_org] ID for this Local User to be created in 
- `role_ids` - (Required) A set of [Role][vcfa_global_role] IDs to assign to this Local User
- `username` - (Required) Username for this Local User
- `password` - (Required) A password for the Local User

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing Organization Local User configuration can be [imported][docs-import] into this resource via
supplying path for it. An example is below:

```
terraform import vcfa_org_local_user.imported my-org-name.my-user-name
```

_NOTE_: The default separator `.` can be changed using provider's `import_separator` argument or environment variable `VCFA_IMPORT_SEPARATOR`

The above would import the `my-user-name` local user from  `my-org-name` Organization.

After that, you can expand the configuration file and either update or delete the Organization Local User as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Organization Local User's stored properties.

[docs-import]: https://www.terraform.io/docs/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
[vcfa_org]: /providers/vmware/vcfa/latest/docs/resources/org
[vcfa_global_role]: /providers/vmware/vcfa/latest/docs/resources/global_role