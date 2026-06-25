// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package kubernetes

import (
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// warningCollector implements rest.WarningHandler and accumulates every warning
// header text returned by the Kubernetes API server.  It is safe for concurrent
// use because the dynamic and main client-sets may issue concurrent requests.
type warningCollector struct {
	mu   sync.Mutex
	msgs []string
}

// HandleWarningHeader satisfies rest.WarningHandler.  Only non-empty warning
// texts are recorded; the warn-code and warn-agent are discarded because
// Kubernetes always emits code 299 with agent "-".
func (w *warningCollector) HandleWarningHeader(_ int, _ string, text string) {
	if text == "" {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.msgs = append(w.msgs, text)
}

// drain atomically returns all accumulated warnings and resets the slice so
// that the same collector can be reused across multiple operations on the same
// Client instance.
func (w *warningCollector) drain() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := w.msgs
	w.msgs = nil
	return out
}

// FlushWarnings drains the collector and converts every accumulated warning
// into a Terraform warning diagnostic.  Callers should append the result to
// their resp.Diagnostics immediately after each Kubernetes API call (or at
// the end of the operation using a deferred closure).
func (c *Client) FlushWarnings() diag.Diagnostics {
	var diags diag.Diagnostics
	for _, msg := range c.warnings.drain() {
		diags.AddWarning("Kubernetes API warning", msg)
	}
	return diags
}
