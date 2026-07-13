// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package testutils

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// AcceptanceTestsSkipped is the warning message used when acceptance tests are skipped.
const AcceptanceTestsSkipped = "Acceptance tests skipped unless env 'TF_ACC' set"

// VcfaShortTest reports whether the short test mode is enabled (used by "make test").
var VcfaShortTest = os.Getenv("VCFA_SHORT_TEST") != ""

// UsingSysAdmin returns true when the given configuration authenticates as a System
// administrator (i.e. SysOrg == "system"). It is a pure predicate extracted so that both
// vcfa package tests (which pass their already-initialised testConfig) and framework tests
// (which pass the result of GetTestConfig) share a single implementation.
func UsingSysAdmin(cfg TestConfig) bool {
	return strings.EqualFold(cfg.Provider.SysOrg, "system")
}

// SkipIfShort skips the calling test when running in short mode.
func SkipIfShort(t *testing.T) {
	t.Helper()
	if VcfaShortTest {
		t.Skip(AcceptanceTestsSkipped)
	}
}

// SkipIfSysAdmin skips the calling test if the configuration uses a System administrator,
// matching the behaviour required by tenant-scoped resources.
func SkipIfSysAdmin(t *testing.T) {
	t.Helper()
	if UsingSysAdmin(GetTestConfig(t)) {
		t.Skip(t.Name() + " requires org (tenant) privileges")
	}
}

// SkipIfNotSysAdmin skips the calling test if the configuration does not use a System
// administrator, matching the behaviour required by system-admin-only resources.
func SkipIfNotSysAdmin(t *testing.T) {
	t.Helper()
	if !UsingSysAdmin(GetTestConfig(t)) {
		t.Skip(t.Name() + " requires system admin privileges")
	}
}

// CheckAttrNonEmptyList returns a TestCheckFunc that fails when the named count
// attribute (e.g. "some_list.#") is absent, zero, or empty.
//
// This is intentionally a raw state walk rather than a wrapper around
// TestCheckResourceAttrWith: that helper internally runs testCheckResourceAttrSet
// before invoking any custom function, so it would still emit "expected to be set"
// when the attribute key is missing from the flat state (which happens for null
// list attributes in framework-based providers).
func CheckAttrNonEmptyList(resourceName, attr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("CheckAttrNonEmptyList: resource %q not found in state", resourceName)
		}
		is := rs.Primary
		if is == nil {
			return fmt.Errorf("CheckAttrNonEmptyList: %s has no primary instance", resourceName)
		}
		value := is.Attributes[attr] // empty string when key is absent
		if value == "" || value == "0" {
			return fmt.Errorf("%s: attribute %q expected to be a non-empty list, got %q",
				resourceName, attr, value)
		}
		return nil
	}
}

// CheckAttrNonEmptySet returns a TestCheckFunc that fails when the named count
// attribute (e.g. "some_set.#") is absent, zero, or empty.
//
// Like CheckAttrNonEmptyList, this is a raw state walk rather than a wrapper
// around TestCheckResourceAttrWith; that helper calls testCheckResourceAttrSet
// internally before invoking any custom function, so it emits "expected to be
// set" whenever the attribute key is missing from the flat state (which happens
// for null set attributes in framework-based providers).
func CheckAttrNonEmptySet(resourceName, attr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("CheckAttrNonEmptySet: resource %q not found in state", resourceName)
		}
		is := rs.Primary
		if is == nil {
			return fmt.Errorf("CheckAttrNonEmptySet: %s has no primary instance", resourceName)
		}
		value := is.Attributes[attr] // empty string when key is absent
		if value == "" || value == "0" {
			return fmt.Errorf("%s: attribute %q expected to be a non-empty set, got %q",
				resourceName, attr, value)
		}
		return nil
	}
}
