// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package testutils

import (
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v3/govcd"
)

// minVcfaApiVersion mirrors the minimum API version used by the provider client.
const minVcfaApiVersion = "40.0"

// NewClient builds an authenticated go-vcloud-director client from the shared test
// configuration. It is a provider-independent re-implementation of the connection logic in
// the `vcfa` package, so that helpers in this package do not need to import `vcfa`.
func NewClient(t *testing.T) *govcd.VCDClient {
	t.Helper()
	cfg := GetTestConfig(t)

	authURL, err := url.ParseRequestURI(cfg.Provider.Url)
	if err != nil {
		t.Fatalf("error parsing provider URL %q: %s", cfg.Provider.Url, err)
	}

	userAgent := fmt.Sprintf("terraform-provider-vcfa-test/dev (%s/%s; isProvider:%t)",
		runtime.GOOS, runtime.GOARCH, strings.EqualFold(cfg.Provider.SysOrg, "system"))

	client := govcd.NewVCDClient(*authURL, cfg.Provider.AllowInsecure,
		govcd.WithHttpUserAgent(userAgent),
		govcd.WithAPIVersion(minVcfaApiVersion),
	)

	if err := authenticate(client, cfg); err != nil {
		t.Fatalf("error authenticating VCFA client: %s", err)
	}

	return client
}

// authenticate ports the authentication selection logic used by the provider, choosing the
// appropriate mechanism based on which credentials are present in the configuration.
func authenticate(client *govcd.VCDClient, cfg TestConfig) error {
	p := cfg.Provider
	org := p.SysOrg
	switch {
	case p.ServiceAccountTokenFile != "":
		return client.SetServiceAccountApiToken(org, p.ServiceAccountTokenFile)
	case p.ApiTokenFile != "":
		_, err := client.SetApiTokenFromFile(org, p.ApiTokenFile)
		return err
	case p.ApiToken != "":
		return client.SetToken(org, govcd.ApiTokenHeader, p.ApiToken)
	case p.Token != "":
		if len(p.Token) > 32 {
			return client.SetToken(org, govcd.BearerTokenHeader, p.Token)
		}
		return client.SetToken(org, govcd.AuthorizationHeader, p.Token)
	default:
		return client.Authenticate(p.User, p.Password, org)
	}
}
