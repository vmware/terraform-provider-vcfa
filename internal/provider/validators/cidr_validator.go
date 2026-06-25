// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type cidrValidator struct{}

// IsValidCIDR returns a validator.String that accepts any valid IPv4 or IPv6 CIDR
// prefix (e.g. "192.168.0.0/16" or "2001:db8::/32").
func IsValidCIDR() validator.String {
	return cidrValidator{}
}

func (v cidrValidator) Description(_ context.Context) string {
	return "must be a valid IPv4 or IPv6 CIDR block (e.g. \"192.168.0.0/16\")"
}

func (v cidrValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v cidrValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	if _, _, err := net.ParseCIDR(val); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid CIDR block",
			fmt.Sprintf("%q is not a valid IPv4 or IPv6 CIDR block", val),
		)
	}
}
