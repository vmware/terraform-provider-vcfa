---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_version"
sidebar_current: "docs-data-source-vcfa-version"
description: |-
  Provides a data source to fetch the version details from VMware Cloud Foundation Automation,
  its maximum supported API version and perform some optional checks with version constraints.
---

# vcfa\_version

Provides a data source to fetch the version details from VMware Cloud Foundation Automation, its maximum supported API version and
perform some optional checks with version constraints.

_Used by: **Provider**_

## Example Usage

```hcl
# This data source will assert that the VCFA version is exactly 9.0.0, otherwise it will fail
data "vcfa_version" "eq_9" {
  condition         = "= 9.0.0"
  fail_if_not_match = true
}

# This data source will assert that the VCFA version is greater than or equal to 9.1, but it won't fail if it is not
data "vcfa_version" "gte_91" {
  condition         = ">= 9.1"
  fail_if_not_match = false
}

output "is_gte_91" {
  value = data.vcfa_version.gte_91.matches_condition # Will show false if we're using a VCFA version < 9.1
}

# This data source will assert that the VCFA version is less than 9.1
data "vcfa_version" "lt_91" {
  condition         = "< 9.1"
  fail_if_not_match = true
}

# This data source will assert that the VCFA version is 9.0.X
data "vcfa_version" "is_90" {
  condition         = "~> 9.0"
  fail_if_not_match = true
}

# This data source will assert that the VCFA version is not 9.1
data "vcfa_version" "not_91" {
  condition         = "!= 9.1"
  fail_if_not_match = true
}

# Output the version and API version of VCFA
data "vcfa_version" "version" {}

output "tenant_manager_version" {
  value = data.vcfa_version.version.tm_version
}

output "tenant_manager_api_version" {
  value = data.vcfa_version.version.tm_api_version
}
```

## Argument Reference

The following arguments are supported:

- `condition` - (Optional) A version constraint to check against the VCFA version
- `fail_if_not_match` - (Optional) Required if `condition` is set. Throws an error if the version constraint set in `condition` is not met.
  Defaults to `false`

## Attribute Reference

- `matches_condition` - It is `true` if the VCFA version matches the constraint set in `condition`
- `tm_version` - The VCFA version
- `tm_api_version` - The maximum supported VCFA API version
