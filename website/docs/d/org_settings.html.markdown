---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_settings"
sidebar_current: "docs-vcfa-data-source-org-settings"
description: |-
  Provides a data source to read VMware Cloud Foundation Automation Organization general Settings.
---

# vcfa\_org\_settings

Provides a data source to read VMware Cloud Foundation Automation Organization general Settings.

~> This data source can only be used by **System Administrators**

-> For Organization Networking settings, see [`vcfa_org_networking`](/providers/vmware/vcfa/latest/docs/data-sources/org_networking) data source 

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "my-org-name"
}

data "vcfa_org_settings" "demo" {
  org_id = data.vcfa_org.demo.id
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) The ID of Organization.

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_settings`](/providers/vmware/vcfa/latest/docs/resources/org_settings) resource are
available.