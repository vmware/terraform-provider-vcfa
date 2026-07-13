// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package testutils

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"
)

// Exported flag variables — single source of truth for every VCFA test flag.
//
// RegisterTestFlags wires each variable to its CLI flag and env-var default.
// The vcfa package registers flags by calling RegisterTestFlags and then copies
// these vars into its own package-level names in TestMain, so the flag
// definitions (names, env vars, help text) live only here.
var (
	// Flags with direct effect in both the vcfa and internal test binaries.
	TestOrgUser bool
	EnableDebug bool
	EnableTrace bool
	TestVerbose bool
	TestTrace   bool
	// VcfaShortTest is declared in checks.go; RegisterTestFlags wires it here.

	// Flags consumed primarily by the vcfa package.  They are exported so that
	// vcfa/config_test.go can sync them into its own package-level vars without
	// re-declaring the flag metadata.
	VcfaHelp                  bool
	VcfaRemoveTestList        bool
	VcfaPrePostChecks         bool
	VcfaShowTimestamp         bool
	VcfaShowElapsedTime       bool
	VcfaShowCount             bool
	VcfaReRunFailed           bool
	VcfaSkipPriorityTests     bool
	VcfaAddProvider           bool
	VcfaSkipTemplateWriting   bool
	VcfaRemoveOrgFromTemplate bool
	VcfaSkipPattern           string
	SkipLeftoversRemoval      bool
	OnlyLeftoverRemoval       bool
	SilentLeftoversRemoval    bool
	TestListFileName          string
	NumberOfPartitions        int
	PartitionNode             int
)

var registerFlagsOnce sync.Once

// RegisterTestFlags registers the full set of VCFA CLI flags in the calling
// test binary. It is safe to call multiple times; registration happens only once.
func RegisterTestFlags() {
	registerFlagsOnce.Do(registerFlags)
}

func registerFlags() {
	registerBoolFlag(&TestOrgUser, "vcfa-test-org-user", "VCFA_TEST_ORG_USER", "Run tests with org user credentials from the config file")
	registerBoolFlag(&EnableDebug, "vcfa-debug", "GOVCD_DEBUG", "Enable debug output from the API client")
	registerBoolFlag(&EnableTrace, "govcd-trace", "GOVCD_TRACE", "Enable function calls tracing")
	registerBoolFlag(&TestVerbose, "vcfa-test-verbose", "VCFA_TEST_VERBOSE", "Enable verbose test output")
	registerBoolFlag(&TestTrace, "vcfa-test-trace", "VCFA_TEST_TRACE", "Enable trace test output")
	registerBoolFlag(&VcfaShortTest, "vcfa-short", "VCFA_SHORT_TEST", "Run short tests only")

	registerBoolFlag(&VcfaHelp, "vcfa-help", "VCFA_HELP", "Show vcfa flags")
	registerBoolFlag(&VcfaRemoveTestList, "vcfa-remove-test-list", "VCFA_REMOVE_TEST_LIST", "Remove list of test runs")
	registerBoolFlag(&VcfaPrePostChecks, "vcfa-pre-post-checks", "VCFA_PRE_POST_CHECKS", "Perform checks before and after tests")
	registerBoolFlag(&VcfaShowTimestamp, "vcfa-show-timestamp", "VCFA_SHOW_TIMESTAMP", "Show timestamp in pre and post checks")
	registerBoolFlag(&VcfaShowElapsedTime, "vcfa-show-elapsed-time", "VCFA_SHOW_ELAPSED_TIME", "Show elapsed time since the start of the suite in pre and post checks")
	registerBoolFlag(&VcfaShowCount, "vcfa-show-count", "VCFA_SHOW_COUNT", "Show number of pass/fail tests")
	registerBoolFlag(&VcfaReRunFailed, "vcfa-re-run-failed", "VCFA_RE_RUN_FAILED", "Run only tests that failed in a previous run")
	registerBoolFlag(&VcfaSkipPriorityTests, "vcfa-skip-priority-tests", "VCFA_SKIP_PRIORITY_TESTS", "Skip priority tests")
	registerBoolFlag(&VcfaAddProvider, "vcfa-add-provider", "VCFA_ADD_PROVIDER", "Add provider to test scripts")
	registerBoolFlag(&VcfaSkipTemplateWriting, "vcfa-skip-template-write", "VCFA_SKIP_TEMPLATE_WRITING", "Skip writing templates to file")
	registerBoolFlag(&VcfaRemoveOrgFromTemplate, "vcfa-remove-org-from-template", "REMOVE_ORG_FROM_TEMPLATE", "Remove org from template")
	registerStringFlag(&VcfaSkipPattern, "vcfa-skip-pattern", "VCFA_SKIP_PATTERN", "Skip tests that match the pattern (implies vcfa-pre-post-checks)")
	registerBoolFlag(&SkipLeftoversRemoval, "vcfa-skip-leftovers-removal", "VCFA_SKIP_LEFTOVERS_REMOVAL", "Do not attempt removal of leftovers at the end of the test suite")
	registerBoolFlag(&OnlyLeftoverRemoval, "vcfa-only-leftover-removal", "VCFA_ONLY_LEFTOVER_REMOVAL", "Only do leftover cleanup")
	registerBoolFlag(&SilentLeftoversRemoval, "vcfa-silent-leftovers-removal", "VCFA_SILENT_LEFTOVERS_REMOVAL", "Omit details during removal of leftovers")
	registerStringFlag(&TestListFileName, "vcfa-partition-tests-file", "VCFA_PARTITION_TESTS_FILE", "Name of the file containing the tests to run in the current partition node")
	registerIntFlag(&NumberOfPartitions, "vcfa-partitions", "VCFA_PARTITIONS", "Number of partitions")
	registerIntFlag(&PartitionNode, "vcfa-partition-node", "VCFA_PARTITION_NODE", "Partition node number")
}

// registerBoolFlag binds a CLI flag to a boolean variable. If the
// corresponding environment variable is already set the variable is pre-set to
// true before the flag is bound, so env-var driven behaviour works even when
// flag.Parse() has not been called yet.
func registerBoolFlag(varPointer *bool, name, envVar, help string) {
	if envVar != "" && os.Getenv(envVar) != "" {
		*varPointer = true
	}
	flag.BoolVar(varPointer, name, *varPointer, help)
}

func registerStringFlag(varPointer *string, name, envVar, help string) {
	if envVar != "" && os.Getenv(envVar) != "" {
		*varPointer = os.Getenv(envVar)
	}
	flag.StringVar(varPointer, name, *varPointer, help)
}

func registerIntFlag(varPointer *int, name, envVar, help string) {
	if envVar != "" && os.Getenv(envVar) != "" {
		// Ignore parse errors; the default of 0 is safe.
		_, _ = fmt.Sscanf(os.Getenv(envVar), "%d", varPointer)
	}
	flag.IntVar(varPointer, name, *varPointer, help)
}

// RunTestMain is a convenience wrapper for packages whose TestMain only needs
// flag registration and parsing. Call it as the sole body of TestMain:
//
//	func TestMain(m *testing.M) { testutils.RunTestMain(m) }
func RunTestMain(m *testing.M) {
	RegisterTestFlags()
	flag.Parse()
	os.Exit(m.Run())
}
