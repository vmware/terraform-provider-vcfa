// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/vmware/go-vcloud-director/v3/util"
)

var (
	// Text lines used for logging of http requests and responses
	lineLength int    = 80
	dashLine   string = strings.Repeat("-", lineLength)
	hashLine   string = strings.Repeat("#", lineLength)
)

// kubernetesloggingRoundTripper wraps any http.RoundTripper and logs every
// Kubernetes API request and response using the shared util.Logger.
type kubernetesloggingRoundTripper struct {
	wrapped http.RoundTripper
}

func (kl *kubernetesloggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// ── Request ──────────────────────────────────────────────────────────────
	util.Logger.Printf("[K8S] %s", dashLine)
	util.Logger.Printf("[K8S] %s %s", req.Method, req.URL.String())
	util.Logger.Printf("[K8S] %s", dashLine)
	util.Logger.Printf("[K8S] Request header:")
	for k, v := range util.SanitizedHeader(req.Header) {
		util.Logger.Printf("[K8S]   %s: %v", k, v)
	}
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err == nil && len(bodyBytes) > 0 {
			util.Logger.Printf("[K8S] Request body: [%d]\n%s", len(bodyBytes), util.HideSensitive(string(bodyBytes), false))
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// ── Execute ──────────────────────────────────────────────────────────────
	resp, err := kl.wrapped.RoundTrip(req)

	// ── Response ─────────────────────────────────────────────────────────────
	util.Logger.Printf("[K8S] %s", hashLine)
	if err != nil {
		util.Logger.Printf("[K8S] Response error: %v", err)
		util.Logger.Printf("[K8S] %s", hashLine)
		return nil, err
	}

	util.Logger.Printf("[K8S] Response status: %s", resp.Status)
	util.Logger.Printf("[K8S] Response headers:")
	for k, v := range util.SanitizedHeader(resp.Header) {
		util.Logger.Printf("[K8S]   %s: %v", k, v)
	}
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err == nil {
			util.Logger.Printf("[K8S] Response body: [%d]\n%s", len(bodyBytes), util.HideSensitive(string(bodyBytes), false))
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	return resp, nil
}
