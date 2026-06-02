// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package mux

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/vmware/terraform-provider-vcfa/internal/provider"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

// NewMuxServer returns a tfprotov6.ProviderServer that routes
// requests to either the old SDKv2 provider or the new framework-based
// provider depending on which one owns a given resource or
// data source.
func NewMuxServer(ctx context.Context) (tfprotov6.ProviderServer, error) {
	sdkv2Provider := vcfa.Provider()
	upgradedSDKv2Server, err := tf5to6server.UpgradeServer(
		ctx,
		sdkv2Provider.GRPCProvider,
	)
	if err != nil {
		return nil, err
	}

	frameworkProvider := providerserver.NewProtocol6(provider.NewVcfaFrameworkProvider(sdkv2Provider.Meta))

	providers := []func() tfprotov6.ProviderServer{
		func() tfprotov6.ProviderServer { return upgradedSDKv2Server },
		frameworkProvider,
	}

	return tf6muxserver.NewMuxServer(ctx, providers...)
}
