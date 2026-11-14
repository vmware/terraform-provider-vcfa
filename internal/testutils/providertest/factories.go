//go:build functional || org || tm || cci || vks || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

// Package providertest provides shared acceptance-test helpers that depend on internal/mux.
// It lives in a sub-package of testutils (rather than inside testutils itself) so that the
// vcfa package tests can import internal/testutils without pulling in internal/mux, which
// imports vcfa — an import cycle Go would not allow.
package providertest

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/vmware/terraform-provider-vcfa/internal/mux"
)

// ProtoV6ProviderFactories wires up the muxed provider server (SDKv2 + framework) for
// acceptance tests that exercise framework-based resources such as vcfa_vks_cluster.
// Framework test packages should use this instead of declaring the map locally.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"vcfa": newMuxServer,
	// A second, separate provider server for tests that also log in as a tenant user. The test
	// harness runs a single server per provider name, so two provider configurations with the
	// same name would share, and overwrite, their connection.
	"vcfatenant": newMuxServer,
}

func newMuxServer() (tfprotov6.ProviderServer, error) {
	return mux.NewMuxServer(context.Background())
}
