// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package testutils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestSuiteHooks packages the vcfa-specific callbacks that PreTestChecks and
// PostTestChecks invoke at well-defined points. Populate and register once in
// TestMain via SetSuiteHooks before os.Exit(m.Run()).
type TestSuiteHooks struct {
	// RunPriorityTests is called once, before the first non-priority test, to run
	// tests that must complete before shared infrastructure is available.
	// recordResult(name, passed) must be called for each priority test so that
	// PreTestChecks can prevent the same test from running twice.
	RunPriorityTests func(t *testing.T, recordResult func(name string, passed bool))

	// HandlePartitioning implements partition-based test filtering.
	HandlePartitioning func(vcfaVersion, url string, t *testing.T)

	// OnTestFailure is called after a test fails, e.g. to remove leftovers.
	OnTestFailure func(t *testing.T)

	// VcfaVersion and VcfaUrl are passed to HandlePartitioning and used by
	// the test-run-list file naming.
	VcfaVersion string
	VcfaUrl     string

	// SkipAllFile names a file whose presence causes all tests to be skipped.
	// Defaults to "skip_vcfa_tests" when empty.
	SkipAllFile string

	// StartTime is the moment the suite started; defaults to time.Now() at
	// SetSuiteHooks time.
	StartTime time.Time
}

// package-level suite state — owned by testutils, shared across all tests in
// a binary that calls SetSuiteHooks.
var (
	suiteHooks     TestSuiteHooks
	suiteExecuted  sync.Map // map[testName string]bool (true = passed)
	priorityOnce   atomic.Bool
	suiteSkipCount int
	suitePassCount int
	suiteFailCount int
	listFileMu     sync.Mutex
	listFileLocks  = make(map[string]*sync.Mutex)
)

// SetSuiteHooks registers the per-suite callbacks and configuration used by
// PreTestChecks and PostTestChecks. Call once in TestMain after
// RegisterTestFlags and after the test configuration has been loaded (so
// VcfaUrl / VcfaVersion can be populated).
func SetSuiteHooks(hooks TestSuiteHooks) {
	if hooks.SkipAllFile == "" {
		hooks.SkipAllFile = "skip_vcfa_tests"
	}
	if hooks.StartTime.IsZero() {
		hooks.StartTime = time.Now()
	}
	suiteHooks = hooks
}

// SuitePassCount returns the number of tests that have passed this run.
func SuitePassCount() int { return suitePassCount }

// SuiteSkipCount returns the number of tests that have been skipped this run.
func SuiteSkipCount() int { return suiteSkipCount }

// SuiteFailCount returns the number of tests that have failed this run.
func SuiteFailCount() int { return suiteFailCount }

// PreTestChecks is the standard VCFA pre-test gate. It:
//  1. Runs any registered priority tests exactly once (via hooks.RunPriorityTests).
//  2. Prevents a test from running again if it was already executed as a priority sub-test.
//  3. Delegates to hooks.HandlePartitioning for partition-based filtering.
//  4. When VcfaPrePostChecks is enabled: checks timestamps, skip file, skip
//     pattern, env-var skip, pass/fail run lists, and re-run-failed mode.
func PreTestChecks(t *testing.T) {
	t.Helper()

	if !VcfaSkipPriorityTests && suiteHooks.RunPriorityTests != nil {
		if priorityOnce.CompareAndSwap(false, true) {
			suiteHooks.RunPriorityTests(t, func(name string, passed bool) {
				suiteExecuted.Store(name, passed)
			})
		}
	}

	testname := t.Name()
	if strings.Contains(testname, "/") {
		testname = strings.SplitN(testname, "/", 2)[1]
	}

	if result, isSet := suiteExecuted.Load(testname); isSet {
		if !result.(bool) {
			t.Logf("%s already FAILED", testname)
			t.FailNow()
		} else {
			t.Skipf("%s already run with priority", testname)
		}
	}

	if suiteHooks.HandlePartitioning != nil {
		suiteHooks.HandlePartitioning(suiteHooks.VcfaVersion, suiteHooks.VcfaUrl, t)
	}

	if !VcfaPrePostChecks {
		return
	}
	if VcfaShowTimestamp {
		fmt.Printf("Test started at: %s\n", Timestamp())
	}
	if VcfaShowElapsedTime {
		fmt.Printf("Elapsed: %s\n", time.Since(suiteHooks.StartTime).String())
	}
	if fileExists(suiteHooks.SkipAllFile) {
		suiteSkipCount++
		t.Skipf("File '%s' found at %s. Test %s skipped", suiteHooks.SkipAllFile, Timestamp(), t.Name())
	}
	if VcfaSkipPattern != "" {
		re := regexp.MustCompile(VcfaSkipPattern)
		if re.MatchString(t.Name()) {
			suiteSkipCount++
			t.Skipf("Skip pattern '%s' matches test name '%s'", VcfaSkipPattern, t.Name())
		}
	}
	skipEnvVar := fmt.Sprintf("skip-%s", t.Name())
	if TestVerbose {
		fmt.Printf("ENV VAR for %s: %s\n", t.Name(), skipEnvVar)
	}
	if os.Getenv(skipEnvVar) != "" {
		suiteSkipCount++
		t.Skipf("variable '%s' was set.", skipEnvVar)
	}
	if IsTestInFile(t.Name(), "pass") {
		suiteSkipCount++
		t.Skipf("test '%s' found in '%s' ", t.Name(), GetTestListFile("pass"))
	}
	if VcfaReRunFailed {
		if !IsTestInFile(t.Name(), "fail") {
			suiteSkipCount++
			t.Skip("only running tests that have failed at the previous run")
		}
	}
}

// PostTestChecks is the standard VCFA post-test recorder. It:
//  1. Records the test result in the executed-test map so PreTestChecks can
//     detect duplicate runs.
//  2. On failure, calls hooks.OnTestFailure (e.g. leftover removal).
//  3. When VcfaPrePostChecks is enabled: shows a timestamp and updates the
//     pass/fail run-list files.
func PostTestChecks(t *testing.T) {
	testname := t.Name()
	if strings.Contains(testname, "/") {
		testname = strings.SplitN(testname, "/", 2)[1]
	}

	PrintfTrace("# postTestChecks storing testname %s state\n", testname)
	suiteExecuted.Store(testname, !t.Failed())

	if t.Failed() && !SkipLeftoversRemoval && suiteHooks.OnTestFailure != nil {
		suiteHooks.OnTestFailure(t)
	}

	if !VcfaPrePostChecks {
		return
	}
	if VcfaShowTimestamp {
		fmt.Printf("Test ended at: %s\n", Timestamp())
	}
	fileType := "pass"
	if t.Failed() {
		fileType = "fail"
		suiteFailCount++
	} else {
		suitePassCount++
	}
	if err := AddToTestRunList(t.Name(), fileType); err != nil {
		fmt.Printf("WARNING: error adding test name '%s' to file '%s'\n", t.Name(), GetTestListFile(fileType))
	}
}

// Timestamp returns the current time formatted as RFC 3339.
func Timestamp() string {
	return time.Now().Format(time.RFC3339)
}

// PrintfVerbose prints when TestVerbose is set.
func PrintfVerbose(format string, args ...interface{}) {
	if TestVerbose {
		fmt.Printf(format, args...)
	}
}

// PrintfTrace prints when TestTrace is set.
func PrintfTrace(format string, args ...interface{}) {
	if TestTrace {
		fmt.Printf(format, args...)
	}
}

// GetTestListFile returns the filename used to record pass/fail run-list entries
// for the registered VcfaUrl. Returns "" when no URL is registered.
func GetTestListFile(fileType string) string {
	if suiteHooks.VcfaUrl == "" {
		return ""
	}
	ip := strings.ReplaceAll(suiteHooks.VcfaUrl, "https://", "")
	ip = strings.ReplaceAll(ip, "/api", "")
	ip = strings.ReplaceAll(ip, "/", "")
	ip = strings.ReplaceAll(ip, ".", "-")
	return fmt.Sprintf("vcfa_test_%s_list-%s.txt", fileType, ip)
}

// IsTestInFile reports whether testName appears in the given fileType run-list.
func IsTestInFile(testName, fileType string) bool {
	fileName := GetTestListFile(fileType)
	if fileName == "" {
		return false
	}
	mu := listFileMutex(fileName)
	mu.Lock()
	defer mu.Unlock()
	if !fileExists(fileName) {
		return false
	}
	f, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == testName {
			return true
		}
	}
	return false
}

// AddToTestRunList appends testName to the given fileType run-list.
func AddToTestRunList(testName, fileType string) error {
	fileName := GetTestListFile(fileType)
	if fileName == "" {
		return nil
	}
	mu := listFileMutex(fileName)
	mu.Lock()
	defer mu.Unlock()

	var f *os.File
	var err error
	if fileExists(fileName) {
		f, err = os.OpenFile(filepath.Clean(fileName), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	} else {
		f, err = os.Create(filepath.Clean(fileName))
	}
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	w := bufio.NewWriter(f)
	if _, err = fmt.Fprintf(w, "%s\n", testName); err != nil {
		return fmt.Errorf("error writing to file %s: %s", fileName, err)
	}
	return w.Flush()
}

// RemoveTestRunList deletes the pass/fail run-list file and also removes the
// skip-all file if present.
func RemoveTestRunList(fileType string) error {
	fileName := GetTestListFile(fileType)
	mu := listFileMutex(fileName)
	mu.Lock()
	defer mu.Unlock()

	if fileExists(suiteHooks.SkipAllFile) {
		if err := os.Remove(suiteHooks.SkipAllFile); err != nil {
			return err
		}
	}
	if !fileExists(fileName) {
		fmt.Printf("[RemoveTestRunList] '%s' not found\n", fileName)
		return nil
	}
	return os.Remove(fileName)
}

// listFileMutex returns a per-filename mutex (creating it on first use).
func listFileMutex(fileName string) *sync.Mutex {
	listFileMu.Lock()
	defer listFileMu.Unlock()
	mu, ok := listFileLocks[fileName]
	if !ok {
		mu = &sync.Mutex{}
		listFileLocks[fileName] = mu
	}
	return mu
}
