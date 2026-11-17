---
page_title: "VMware Cloud Foundation Automation: vcfa_project"
subcategory: ""
description: |-
  Provides a data source to read Projects, and the users and groups that have access to them, from VMware Cloud Foundation Automation.
---

# vcfa_project

Provides a data source to read a Project from VMware Cloud Foundation Automation, together with the users and groups that have access to it. To manage a Project, use the [`vcfa_project`](/providers/vmware/vcfa/latest/docs/resources/project) resource.

_Used by: **Tenant**_

## Example Usage

```hcl
data "vcfa_project" "team_a" {
  name = "team-a"
}

output "team_a_admins" {
  value = [for m in data.vcfa_project.team_a.members : m.name if m.role == "admin"]
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the Project.

## Attribute Reference

The following attributes are exported:

- `id` - Internal identifier, which is the name of the Project.
- `description` - Description of the Project.
- `uid` - Unique identifier of the Project.
- `members` - Users and groups that have access to the Project, and their role:
  - `kind` - Kind of member: `User` or `Group`.
  - `name` - Name of the user or group.
  - `role` - Role of the member in the Project: `admin` (Project Administrator), `edit` (Project User), `edit_adv` (Project Advanced User) or `view` (Project Auditor).
