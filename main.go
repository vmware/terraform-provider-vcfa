// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/vmware/terraform-provider-vcfa/vcfa"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vcfa.Provider})
}
