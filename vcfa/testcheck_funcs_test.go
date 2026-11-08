//go:build api || functional || tm || cci || contentlibrary || org || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"regexp"
	"strings"
	"sync"

	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/vmware/terraform-provider-vcfa/internal/testutils"
)

// testCachedFieldValue wraps testutils.TestCachedFieldValue, embedding the shared
// implementation while exposing the unexported method names expected by existing vcfa tests.
type testCachedFieldValue struct {
	testutils.TestCachedFieldValue
}

// cacheTestResourceFieldValue delegates to the shared implementation.
func (c *testCachedFieldValue) cacheTestResourceFieldValue(res, field string) resource.TestCheckFunc {
	return c.CacheTestResourceFieldValue(res, field)
}

// testCheckCachedResourceFieldValue delegates to the shared implementation.
func (c *testCachedFieldValue) testCheckCachedResourceFieldValue(res, field string) resource.TestCheckFunc {
	return c.TestCheckCachedResourceFieldValue(res, field)
}

// testCheckCachedResourceFieldValuePattern delegates to the shared implementation.
func (c *testCachedFieldValue) testCheckCachedResourceFieldValuePattern(res, field, pattern string) resource.TestCheckFunc {
	return c.TestCheckCachedResourceFieldValuePattern(res, field, pattern)
}

// resourceFieldsEqual delegates to the shared implementation.
var resourceFieldsEqual = testutils.ResourceFieldsEqual

var (
	vcfaRuntimeVersion     *semver.Version
	vcfaRuntimeVersionOnce sync.Once
)

// vcfaVersionAtLeast returns true if the VCFA instance under test is at least the given version (e.g. "9.1.2").
// The build number and any suffix are ignored. If the version can't be determined it returns false, so that
// version-dependent checks are skipped instead of failing.
func vcfaVersionAtLeast(minimum string) bool {
	vcfaRuntimeVersionOnce.Do(func() {
		raw, err := getVcfaVersion(testConfig)
		if err != nil || raw == "" {
			raw = testConfig.Provider.VcfaVersion
		}
		parts := regexp.MustCompile(`\d+`).FindAllString(raw, 3)
		if len(parts) < 3 {
			return
		}
		v, err := semver.NewVersion(strings.Join(parts, "."))
		if err == nil {
			vcfaRuntimeVersion = v
		}
	})
	if vcfaRuntimeVersion == nil {
		return false
	}
	min, err := semver.NewVersion(minimum)
	if err != nil {
		return false
	}
	return vcfaRuntimeVersion.GreaterThanOrEqual(min)
}

// ifVcfaVersionAtLeast runs the given checks only if the VCFA instance under test is at least the given version.
// Otherwise, the checks are skipped. The version is evaluated when the checks are executed.
func ifVcfaVersionAtLeast(minimum string, checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !vcfaVersionAtLeast(minimum) {
			return nil
		}
		return resource.ComposeAggregateTestCheckFunc(checks...)(s)
	}
}
