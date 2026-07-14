// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package testutils

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestCachedFieldValue stores a field value captured during one TestStep so it can be
// compared in a later TestStep. This is useful for fields with computed values (for example
// a name derived from a name_prefix).
type TestCachedFieldValue struct {
	fieldValue string
}

// FieldValue returns the cached value.
func (c *TestCachedFieldValue) FieldValue() string {
	return c.fieldValue
}

// String satisfies the fmt.Stringer interface.
func (c *TestCachedFieldValue) String() string {
	return c.fieldValue
}

// CacheTestResourceFieldValue has the same signature as builtin Terraform test functions, but
// stores the captured field value in the receiver for later comparison.
func (c *TestCachedFieldValue) CacheTestResourceFieldValue(res, field string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("resource not found: %s", res)
		}
		value, exists := rs.Primary.Attributes[field]
		if !exists {
			return fmt.Errorf("field %s in resource %s does not exist", field, res)
		}
		c.fieldValue = value
		return nil
	}
}

// TestCheckCachedResourceFieldValue verifies that the current field value equals the
// previously cached value.
func (c *TestCachedFieldValue) TestCheckCachedResourceFieldValue(res, field string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("resource not found: %s", res)
		}
		value, exists := rs.Primary.Attributes[field]
		if !exists {
			return fmt.Errorf("field %s in resource %s does not exist", field, res)
		}
		if value != c.fieldValue {
			return fmt.Errorf("got '%s - %s' field value %s, expected: %s", res, field, value, c.fieldValue)
		}
		return nil
	}
}

// TestCheckCachedResourceFieldValuePattern verifies that the current field value equals the
// pattern formatted with the previously cached value.
func (c *TestCachedFieldValue) TestCheckCachedResourceFieldValuePattern(res, field, pattern string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("resource not found: %s", res)
		}
		value, exists := rs.Primary.Attributes[field]
		if !exists {
			return fmt.Errorf("field %s in resource %s does not exist", field, res)
		}
		expectedValue := fmt.Sprintf(pattern, c.fieldValue)
		if value != expectedValue {
			return fmt.Errorf("got '%s - %s' field value %s, expected: %s", res, field, value, expectedValue)
		}
		return nil
	}
}

// ResourceFieldsEqual checks that secondObject has all the fields and values set on
// firstObject, except those listed in excludeFields. It is useful to verify a data source
// exposes the same values as its resource.
func ResourceFieldsEqual(firstObject, secondObject string, excludeFields []string) resource.TestCheckFunc {
	return ResourceFieldsEqualCustom(firstObject, secondObject, excludeFields, slices.Contains)
}

// ResourceFieldsEqualCustom is like ResourceFieldsEqual but accepts a custom exclusion
// checker, allowing callers to implement arbitrary field-name matching logic.
func ResourceFieldsEqualCustom(firstObject, secondObject string, excludeFields []string, exclusionChecker func(list []string, str string) bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource1, ok := s.RootModule().Resources[firstObject]
		if !ok {
			return fmt.Errorf("unable to find %s", firstObject)
		}
		resource2, ok := s.RootModule().Resources[secondObject]
		if !ok {
			return fmt.Errorf("unable to find %s", secondObject)
		}

		for fieldName := range resource1.Primary.Attributes {
			if excludeFields != nil && exclusionChecker(excludeFields, fieldName) {
				continue
			}
			if !reflect.DeepEqual(resource1.Primary.Attributes[fieldName], resource2.Primary.Attributes[fieldName]) {
				return fmt.Errorf("field %s differs in resources %s (value %s) and %s (value %s)",
					fieldName, firstObject, resource1.Primary.Attributes[fieldName], secondObject, resource2.Primary.Attributes[fieldName])
			}
		}
		return nil
	}
}
