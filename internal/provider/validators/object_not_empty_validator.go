// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type objectNotEmptyValidator struct{}

// ObjectNotEmpty returns a validator.Object that rejects a configured block
// where every attribute is null. Users must either set at least one attribute
// or omit the block entirely.
//
// This prevents "Provider produced inconsistent result after apply" errors
// that arise when a provider returns null for a nested object once all of its
// fields are absent: if the user writes an empty block (e.g. `deletion {}`),
// the framework plans it as a non-null object with all-null fields, but the
// read path returns null, causing a plan-vs-actual mismatch.
// Rejecting the empty block at plan time eliminates the ambiguity.
//
// Note: objectvalidator.AtLeastOneOf cannot be used here because it always
// includes the attribute it is applied to in the check, so the condition is
// trivially satisfied whenever the block is non-null.
func ObjectNotEmpty() validator.Object {
	return objectNotEmptyValidator{}
}

func (objectNotEmptyValidator) Description(_ context.Context) string {
	return "block must set at least one attribute; omit the block entirely if no attributes are needed"
}

func (v objectNotEmptyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (objectNotEmptyValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	for _, attr := range req.ConfigValue.Attributes() {
		if !attr.IsNull() {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Empty block",
		"This block must set at least one attribute. "+
			"To disable it, remove the block entirely rather than leaving it empty.",
	)
}
