---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_networking"
sidebar_current: "docs-vcfa-data-source-org-networking"
description: |-
  Provides a data source to read VMware Cloud Foundation Automation Organization Networking Settings.
---

# vcfa\_org\_networking

Provides a data source to read VMware Cloud Foundation Automation Organization Networking Settings.

-> For general Organization settings, see [`vcfa_org_settings`](/providers/vmware/vcfa/latest/docs/data-sources/org_settings) data source

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "my-org-name"
}

data "vcfa_org_networking" "demo" {
  org_id = data.vcfa_org.demo.id
}
```

## Argument Reference

The following arguments are supported:

- `org_id` - (Required) The ID of Organization.

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_networking`](/providers/vmware/vcfa/latest/docs/resources/org_networking) resource are
available.