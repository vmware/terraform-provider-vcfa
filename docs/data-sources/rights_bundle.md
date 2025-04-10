---
page_title: "VMware Cloud Foundation Automation: vcfa_rights_bundle"
subcategory: ""
description: |-
  Provides a data source to read a Rights Bundle from VMware Cloud Foundation Automation.
---

# vcfa_rights_bundle

Provides a data source to read a Rights Bundle from VMware Cloud Foundation Automation.

_Used by: **Provider**_

## Example Usage

```hcl
data "vcfa_rights_bundle" "default-set" {
  name = "Default Tenant Rights Bundle"
}

output "default-rb" {
  value = data.vcfa_rights_bundle.default-set
}
```

Sample output:

```shell
default-rb = {
  "bundle_key" = "DEFAULT_TENANT_RIGHTS_BUNDLE"
  "description" = "Default set of tenant rights"
  "id" = "urn:vcloud:rightsBundle:bce22806-be59-4e89-b48c-9354c0e18f78"
  "name" = "Default Tenant Rights Bundle"
  "org_ids" = toset([])
  "publish_to_all_orgs" = true
  "read_only" = true
  "rights" = toset([
    "API Explorer: View",
    "API Tokens: Manage",
    "API Tokens: Manage All",
    "Access Control List: Manage",
    "Access Control List: View",
    "Advisory Definitions: Create and Delete",
    "Advisory Definitions: Read",
    "Approvals: Manage",
    "Approvals: View",
    "Audit: Manage",
    "Billing: View",
    "Blueprint Request: Manage",
    "Blueprint Request: View",
    "Blueprint: Edit",
    "Blueprint: Manage",
    "Blueprint: Publish",
    "Blueprint: View",
    "Catalog Instance: Manage",
    "Catalog: Manage",
    "Catalog: Request",
    "Catalog: View",
    "Certificate Library: Manage",
    "Certificate Library: View",
    "Connectivity Profile: Manage",
    "Connectivity Profile: View",
    "Content Library Item: Manage",
    "Content Library Item: View",
    "Content Library: Manage",
    "Content Library: View",
    "Custom entity: View all custom entity instances in org",
    "Custom entity: View custom entity definitions",
    "Custom entity: View custom entity instance",
    "General: Administrator Control",
    "General: Administrator View",
    "General: View Error Details",
    "Group / User: Manage",
    "Group / User: View",
    "IP Blocks: Manage",
    "IP Blocks: View",
    "Integrations: Manage",
    "Integrations: View",
    "Inventory: View",
    "LDAP Settings: Manage",
    "LDAP Settings: View",
    "Metadata File Entry: Create/Modify",
    "Metrics: View",
    "Namespace Class: Manage",
    "Namespace Class: View",
    "Namespace Lifecycle: View",
    "Namespace: Manage",
    "Namespace: View",
    "Notifications: Manage",
    "Notifications: View",
    "Organization vDC: Edit",
    "Organization vDC: View",
    "Organization: Edit Federation Settings",
    "Organization: Edit LDAP Settings",
    "Organization: Edit Limits",
    "Organization: Edit Name",
    "Organization: Edit OAuth Settings",
    "Organization: Edit Password Policy",
    "Organization: Edit Properties",
    "Organization: Edit Quotas Policy",
    "Organization: Edit SMTP Settings",
    "Organization: View",
    "Organization: View OAuth Settings",
    "Organization: View Other Orgs",
    "Policy: Manage",
    "Policy: View",
    "Projects: Manage",
    "Projects: View",
    "Property Group: Manage",
    "Property Group: View",
    "Region: Simple View",
    "Regions: View",
    "Right: View",
    "Role: Create, Edit, Delete, or Copy",
    "SSL: Test Connection",
    "Secrets: Manage",
    "Secrets: View",
    "Service Account: Manage",
    "Service Account: Simple View",
    "Service Account: View",
    "Settings: Manage",
    "Settings: View",
    "Storage Classes: View",
    "Task: Update",
    "Task: View Tasks",
    "Toggles: Manage",
    "Token: Manage",
    "Token: Manage All",
    "Transit Gateway: View",
    "Truststore: Manage",
    "Truststore: View",
    "UI Plugins: View",
    "VPC: Manage",
    "VPC: View",
    "vApp: Use Console",
  ])
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the Rights Bundle

## Attribute Reference

- `description` - A description of the Rights Bundle
- `bundle_key` - Key used for internationalization
- `rights` - Set of rights assigned to this role
- `publish_to_all_orgs` - When `true`, publishes the Rights Bundle to all [Organizations](/providers/vmware/vcfa/latest/docs/resources/org)
- `org_ids` - Set of IDs of the Organizations to which this Rights Bundle gets published. Ignored if `publish_to_all_orgs` is `true`
- `read_only` - Whether this Rights Bundle is read-only

## More information

See [Roles management](/providers/vmware/vcfa/latest/docs/guides/roles_management) for a broader description of how roles and
rights work together.
