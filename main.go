// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"log"
	"os"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"

	"github.com/vmware/terraform-provider-vcfa/internal/mux"
)

const providerAddress = "registry.terraform.io/vmware/vcfa"

func main() {
	ctx := context.Background()

	muxServer, err := mux.NewMuxServer(ctx)
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	opts := []tf6server.ServeOpt{}
	tf6server.Serve(providerAddress, func() tfprotov6.ProviderServer { return muxServer }, opts...)
}
