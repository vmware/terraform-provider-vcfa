//go:build tm || ALL || functional

package vcfa

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

var priorityTestsExecuted atomic.Bool
var executedTests sync.Map
var priorityTestCleanupFunc func() error

type priorityTest struct {
	Name string
	Test func(*testing.T)
}

func runPriorityTestsOnce(t *testing.T) {
	notExecuted := priorityTestsExecuted.CompareAndSwap(false, true)
	if notExecuted {
		tests := []priorityTest{
			{Name: "TestAccVcfaNsxManager", Test: TestAccVcfaNsxManager},
			{Name: "TestAccVcfaVcenter", Test: TestAccVcfaVcenter},
			{Name: "TestAccVcfaVcenterInvalid", Test: TestAccVcfaVcenterInvalid},
		}

		fmt.Printf("# Running priority tests before shared vCenter and NSX Manager is created, so they do not collide later (can be skipped with '-vcfa-skip-priority-tests' flag)\n")
		for _, test := range tests {
			fmt.Printf("# Running priority test '%s' as a subtest of '%s':\n", test.Name, t.Name())
			t.Run(test.Name, test.Test)
			printfTrace("## Storing test to executed test list '%s'\n", test.Name)
			executedTests.Store(test.Name, !t.Failed())
		}

		// setup shared components for other tests
		fmt.Printf("# Setting up shared vCenter and NSX Manager\n")
		cleanup, err := setupVcAndNsx()
		if err != nil {
			fmt.Printf("error setting up shared VC and NSX: %s", err)
		}
		fmt.Printf("# Done setting up shared vCenter and NSX Manager\n")

		priorityTestCleanupFunc = cleanup
		fmt.Printf("# Continuing run of %s test after priority tests are now done\n", t.Name())
	}
}
