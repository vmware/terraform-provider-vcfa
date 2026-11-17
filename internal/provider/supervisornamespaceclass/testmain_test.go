//go:build tm || ALL || functional

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisornamespaceclass_test

import (
	"testing"

	"github.com/vmware/terraform-provider-vcfa/internal/testutils"
)

func TestMain(m *testing.M) { testutils.RunTestMain(m) }
