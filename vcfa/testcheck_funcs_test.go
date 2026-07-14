//go:build api || functional || tm || cci || contentlibrary || org || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

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
