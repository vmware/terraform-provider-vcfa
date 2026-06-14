// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package helpers

import (
	"fmt"

	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

// GetTmClientFromProviderData retrieves the TM Client from the provider data.
// It is designed to be called from a resource's Configure method.
func GetTmClientFromProviderData(providerData any) (*vcfa.VCDClient, error) {
	sdkv2Meta, ok := providerData.(func() any)
	if !ok {
		return nil, fmt.Errorf("unexpected provider data type: expected func() any, got %T", providerData)
	}
	container, ok := sdkv2Meta().(vcfa.ClientContainer)
	if !ok {
		return nil, fmt.Errorf("unexpected provider data, expected `ClientContainer`, got `%T`", providerData)
	}
	return container.GetTMClient(), nil
}
