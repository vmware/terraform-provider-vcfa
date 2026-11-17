---
page_title: "VMware Cloud Foundation Automation: vcfa_project"
subcategory: ""
description: |-
  Provides a resource to manage Projects, and the users and groups that have access to them, in VMware Cloud Foundation Automation.
---

# vcfa_project

Provides a resource to manage Projects in VMware Cloud Foundation Automation, together with the users and groups that have access to them.

A Project groups resources, such as [Supervisor Namespaces](/providers/vmware/vcfa/latest/docs/resources/supervisor_namespace), and the users and groups that can work with them. Each member is assigned a role in the Project.

_Used by: **Tenant**_

## Example Usage

```hcl
resource "vcfa_project" "team_a" {
  name        = "team-a"
  description = "Project of Team A"

  members = [
    { kind = "User", name = "alice", role = "admin" },
    { kind = "User", name = "bob@vsphere.local", role = "edit" },
    { kind = "Group", name = "team-a-developers", role = "edit_adv" },
    { kind = "Group", name = "auditors", role = "view" },
  ]
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the Project. It must be a lowercase DNS label: at most 63 characters, consisting of lowercase alphanumeric characters or `-`, and starting and ending with an alphanumeric character. Changing it forces a new resource to be created.
- `description` - (Optional) Description of the Project, of at most 2000 characters.
- `members` - (Optional) Users and groups that have access to the Project, and their role. See [Members](#members).

## Attribute Reference

In addition to the arguments above, the following computed attributes are exported:

- `id` - Internal identifier, which is the name of the Project.
- `uid` - Unique identifier of the Project.

## Members

Each element of `members` assigns a role in the Project to a user or a group:

- `kind` - (Required) Kind of member: `User` or `Group`.
- `name` - (Required) Name of the user or group, as known to the identity provider of the Organization. A user of a specific domain can be given as `<name>@<domain>`. It must not contain `:`.
- `role` - (Required) Role of the member in the Project:
  - `admin` - Project Administrator.
  - `edit` - Project User.
  - `edit_adv` - Project Advanced User.
  - `view` - Project Auditor.

Each user or group can only have a single role in a Project, so a member can appear at most once in `members`. Changing the `role` of a member updates its assignment in place. Adding or removing an element assigns or unassigns the user or group.

`members` is **authoritative** when set: on every apply, users and groups that were assigned to the Project by other means, such as the UI, are unassigned, and assignments removed outside of Terraform are restored. Setting `members = []` unassigns every user and group. When `members` is not set, the members of the Project are not managed at all, and removing `members` from the configuration stops managing them without unassigning anyone.

~> **Note:** A Project can only be deleted once it holds no resources, such as Supervisor Namespaces. Deleting the Project unassigns all its members.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows also code generation. See [Importing resources][importing-resources] for more information.

An existing Project can be [imported][docs-import] into this resource via its name.
For example, using this structure, representing an existing Project that was **not** created using Terraform:

```hcl
resource "vcfa_project" "existing" {
  name        = "team-a"
  description = "Project of Team A"

  members = [
    { kind = "User", name = "alice", role = "admin" },
  ]
}
```

You can import such Project into terraform state using this command:

```shell
terraform import vcfa_project.existing "team-a"
```

The import reads the current members of the Project into `members`. If the configuration does not set `members`, the next apply stops managing them, without unassigning anyone.

[docs-import]: https://developer.hashicorp.com/terraform/cli/import
[importing-resources]: /providers/vmware/vcfa/latest/docs/guides/importing_resources
