---
layout: "vcfa"
page_title: "VMware Cloud Foundation Automation: vcfa_tm_version"
sidebar_current: "docs-vcfa-data-source-version"
description: |-
  Provides a VCD version data source.
---

# vcfa\_version

Provides a VMware Cloud Foundation Automation version data source to fetch the VCD version, the maximum supported API version and
perform some optional checks with version constraints.

Supported in provider *v3.12+*. Requires System Administrator privileges.

## Example Usage

```hcl
# This data source will assert that the VCD version is exactly 10.5.1, otherwise it will fail
data "vcfa_tm_version" "eq_1051" {
  condition         = "= 10.5.1"
  fail_if_not_match = true
}

# This data source will assert that the VCD version is greater than or equal to 10.4.2, but it won't fail if it is not
data "vcfa_tm_version" "gte_1042" {
  condition         = ">= 10.4.2"
  fail_if_not_match = false
}

output "is_gte_1042" {
  value = data.vcfa_tm_version.gte_1042.matches_condition # Will show false if we're using a VCD version < 10.4.2
}

# This data source will assert that the VCD version is less than 10.5.0
data "vcfa_tm_version" "lt_1050" {
  condition         = "< 10.5.0"
  fail_if_not_match = true
}

# This data source will assert that the VCD version is 10.5.X
data "vcfa_tm_version" "is_105" {
  condition         = "~> 10.5"
  fail_if_not_match = true
}

# This data source will assert that the VCD version is not 10.5.1
data "vcfa_tm_version" "not_1051" {
  condition         = "!= 10.5.1"
  fail_if_not_match = true
}
```

## Argument Reference

The following arguments are supported:

* `condition` - (Optional) A version constraint to check against the VCD version
* `fail_if_not_match` - (Optional) Required if `condition` is set. Throws an error if the version constraint set in `condition` is not met

## Attribute Reference

* `matches_condition` - It is true if the VCD version matches the constraint set in `condition`
* `vcfa_tm_version` - The VCD version
* `api_version` - The maximum supported API version
