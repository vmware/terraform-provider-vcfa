---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_settings"
sidebar_current: "docs-vcfa-data-source-org-settings"
description: |-
  Provides a data source to read the General Settings from an Organization in VMware Cloud Foundation Automation.
---

# vcfa\_org\_settings

Provides a data source to read the General Settings from an [Organization][vcfa_org-ds] in VMware Cloud Foundation Automation.

_Used by: **Provider**, **Tenant**_

-> For Networking Settings of an Organization, see [`vcfa_org_networking`](/providers/vmware/vcfa/latest/docs/data-sources/org_networking) data source 

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

- `org_id` - (Required) The ID of the [Organization][vcfa_org-ds].

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_settings`](/providers/vmware/vcfa/latest/docs/resources/org_settings) resource are
available.

[vcfa_org-ds]: /providers/vmware/vcfa/latest/docs/data-sources/org