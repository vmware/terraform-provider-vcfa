---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_org_regional_networking"
sidebar_current: "docs-vcfa-data-source-org-regional-networking"
description: |-
  Provides a data source to read VMware Cloud Foundation Automation Organization Regional Networking Settings.
---

# vcfa\_org\_regional\_networking

Provides a data source to read VMware Cloud Foundation Automation Organization Regional Networking Settings.

## Example Usage

```hcl
data "vcfa_org" "demo" {
  name = "my-org-name"
}

data "vcfa_org_regional_networking" "demo" {
  name   = "
  org_id = vcfa_org.demo.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of Organization Regional Networking configuration
- `org_id` - (Required) The ID of Organization

## Attribute Reference

All the arguments and attributes defined in
[`vcfa_org_regional_networking`](/providers/vmware/vcfa/latest/docs/resources/org_regional_networking)
resource are available.
