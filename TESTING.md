# Testing terraform-provider-vcfa

## Table of contents

- [Meeting prerequisites: Building the test environment](#meeting-prerequisites-building-the-test-environment)
- [VCFA test environment configuration](#vcfa-test-environment-configuration)
- [Running tests](#running-tests)
- [Test prioritization and sharing core components (vCenter, NSX Manager)](#test-prioritization-and-sharing-core-components)
- [Adding new tests](#adding-new-tests)
- [Parallelism considerations](#parallelism-considerations)
- [Binary testing](#binary-testing)
- [Handling failures in binary tests](#handling-failures-in-binary-tests)
- [Conditional running of tests](#conditional-running-of-tests)
- [Leftovers removal](#leftovers-removal)
- [Environment variables and corresponding flags](#environment-variables-and-corresponding-flags)
- [Troubleshooting code issues](#troubleshooting-code-issues)

## Meeting prerequisites: Building the test environment

To run the tests, your VCFA needs to have the following:

* Classic Tenancy feature flag enabled

## VCFA test environment configuration

**VCFA** tests are executed based on `vcfa_test_config.json` (see `sample_vcfa_test_config.json` for
example) configuration that can be either be put into working directory or its path can be set using
`VCFA_CONFIG` environment variable.

## Running tests

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests for **VCFA**, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

### Unit tests

You can run the unit tests directly with

```sh
make testunit
```

## Test prioritization and sharing core components

The test suite will prioritize testing core infrastructure component resources such *vCenter server*
and *NSX Manager*. After the prioritized tests are run, it will create these components so that they
can be shared with the next tests that rely on it. This saves a lot of time because almost every
tests relies on core components and their creation/removal for each test takes time. The test
snippet helpers in `vcfa_common_test.go` are flexible and can either return `data` or `resource`
snippet for these components. As long as these helpers are used, the test will take advantage of
sharing components.

The prioritization and core component sharing functionality **can be disabled** by using
`-vcfa-skip-priority-tests` flag. The penalty is prolonged test execution time due to each test will
create its own components.

**Note:** Go testing framework does not directly provide functionality to prioritize tests,
therefore the priority tests are always executed as subtests of whichever test was triggered. Their
state is stored and later, when the same test is picked for run, it will be skipped with its
previously reported state. This functionality can be skipped with `-vcfa-skip-priority-tests` flag.

In the below test output, the three prioritized tests `TestAccVcfaNsxManager`, `TestAccVcfaVcenter`
and `TestAccVcfaVcenterInvalid` are being run as subtests of `TestAccDataSourceNotFound` because it
was the first test that was execute in the test suite. When the actual test runs are triggered, they
are skipped with a note that they were already executed.

```sh
TestAccDataSourceNotFound
# Running priority tests before shared vCenter and NSX Manager is created, so they do not collide later (can be skipped with '-vcfa-skip-priority-tests' flag)
# Running priority test 'TestAccVcfaNsxManager' as a subtest of 'TestAccDataSourceNotFound':


TestAccDataSourceNotFound/TestAccVcfaNsxManager
# Running priority test 'TestAccVcfaVcenter' as a subtest of 'TestAccDataSourceNotFound':


TestAccDataSourceNotFound/TestAccVcfaVcenter
# Running priority test 'TestAccVcfaVcenterInvalid' as a subtest of 'TestAccDataSourceNotFound':


TestAccDataSourceNotFound/TestAccVcfaVcenterInvalid
# Setting up shared vCenter and NSX Manager
# Done setting up shared vCenter and NSX Manager
# Continuing run of TestAccDataSourceNotFound test after priority tests are now done


TestAccDataSourceNotFound/vcfa_certificate

......
......

TestAccVcfaNsxManager
config_test.go:1135: TestAccVcfaNsxManager already run with priority
SKIP: TestAccVcfaNsxManager (0.00s)

TestAccVcfaVcenter
config_test.go:1135: TestAccVcfaVcenter already run with priority
SKIP: TestAccVcfaVcenter (0.00s)


TestAccVcfaVcenterInvalid
config_test.go:1135: TestAccVcfaVcenterInvalid already run with priority
SKIP: TestAccVcfaVcenterInvalid (0.00s)
```

## Adding new tests

All tests need to have a build tag. The tag should be the first line of the file, followed by a blank line

```go
// +build functional featurename ALL

package vcfa
```

Tests that integrate in the functional suite use the tag `functional`. Using that tag, we can run
all functional tests at once.
We define as `functional` the tests that need a live *VCFA* to run.

1. The test should always define the `ALL` tag:

* ALL :       Runs all the tests

2. The test should also always define either the `unit` or `functional` tag:

* functional: Runs all the tests that use a live VCFA (acceptance tests)
* unit:       Runs unit tests that do not need a live VCFA

3. Finally, the test could define the feature tag. For example:

* contentlibrary: Runs content library related tests
* region:         Runs region related tests

The `ALL` tag includes tests that use a different framework. At the moment, this is useful to run a global compilation test.
Depending on which additional tests we will implement, we may change the dependency on the `ALL` tag if we detect
clashes between frameworks.

If the test file defines a new feature tag (i.e. one that has not been used before) the file should also implement an
`init` function that sets the tag in the global tag list.
This information is used by the main tag test in `api_test.go` to determine which tags were activated.

```go
func init() {
	testingTags["newtag"] = "filename_test.go"
}
```

### Parallelism considerations

When writing Terraform acceptance tests there are two ways to define tests. Either using
`resource.Test` or `resource.ParallelTest`. The former runs tests sequentially one by one while the
later runs all the tests (which are defined to be run in parallel) instantiated this way in
parallel. This is useful because it can speed up total test execution time. However one must be sure
that the tests defined for parallel run are not clashing with each other.

By default `make testacc` runs acceptance tests with parallelism enabled (for the tests which are
defined with `resource.ParallelTest`). If there is a need to troubleshoot or simply force the tests
to run sequentially - `make seqtestacc` can be used to achieve it. Only a minority of tests are
running in parallel mode due to there can't be multiple resources created without having large lab.

## Binary testing

By *binary testing* we mean the tests that run using Terraform binary executable, as opposed to running the test through the Go framework.
This test runs the same tasks that run in the acceptance test, but instead of running them directly, they are fed to the
terraform tool through a shell script, and for every test we run

* `terraform init`
* `terraform plan`
* `terraform apply -auto-approve`
* `terraform plan -detailed-exitcode` (for ensuring that `plan` is empty right after `apply`)
* `terraform destroy -auto-approve`

Running **VCFA** binary tests, using:

```bash
make test-binary
```

All the tests run unattended, stopping only if there is an error.

It is possible to customise running of the binary tests by preparing them and then running the test
script from the `tests-artifacts` directory. 

The command to prepare binary test snippets is:

```
make test-binary-prepare
```


The following commands can be used to run tests with the generated binary test snippets:

```
cd ./vcfa/test-artifacts
./test-binary.sh help

# OR
./test-binary.sh pause verbose

# OR
./test-binary.sh pause verbose tags "region contentlibrary"
```

The "pause" option will stop the test after every call to the terraform tool, waiting for user input.

When the test runs unattended, it is possible to stop it gracefully by creating a file named `pause` inside the
`test-artifacts` directory. When such file exists, the test execution stops at the next `terraform` command, waiting
for user input.

## Handling failures in binary tests

When one test fails, the binary test script will attempt to recover it, by running `terraform destroy`. If the recovery
fails, the whole test halts. If recovery succeeds, the names of the failed test are recorded inside 
`./vcfa/test-artifacts/failed_tests.txt` and the summary at the end of the test will show them.

If the test runs with `make test-binary`, the output is captured inside `./vcfa/test-artifacts/test-binary-TIME.txt` (where
`TIME` has the format `YYYY-MM-DD-HH-MM`). To see the actual failure, open the output file and search for the name of
the test that failed.

For example, the test ends with this annotation :

```
# ---------------------------------------------------------
# Operations dir: /path/to/terraform-provider-vcfa/vcfa/test-artifacts/tmp
# Started:        Thu Mar 12 14:10:43 CET 2025
# Ended:          Thu Mar 12 14:12:16 CET 2025
# Elapsed:        1m:33s (93 sec)
# exit code:      0
# ---------------------------------------------------------
# ---------------------------------------------------------
# FAILED TESTS    4
# ---------------------------------------------------------
Thu Mar 12 14:11:02 CET 2020 - vcfa.TestAccVcfaContentLibraryItemProvider.tf (apply)
Thu Mar 12 14:11:08 CET 2020 - vcfa.TestAccVcfaContentLibraryItemProvider.tf (plancheck)
Thu Mar 12 14:11:30 CET 2020 - vcfa.TestAccVcfaGlobalRole.tf (apply)
Thu Mar 12 14:11:36 CET 2020 - vcfa.TestAccVcfaGlobalRole.tf (plancheck)
# ---------------------------------------------------------
```

In the output file (in the directory `./vcfa/test-artifacts`), look for `vcfa.TestAccVcfaContentLibraryItemProvider.tf`
and you will see the operations occurring with the actual errors.


## Conditional running of tests

The whole test suite takes several hours to run. If some errors happen during the run, we need to clean up and try again
from the beginning, which is not always convenient.
There are a few tags that help us gain some control on the flow:

* `-vcfa-pre-post-checks`    Global switch enabling checks before and after tests (false). Also activated by using any of the flags below.
* `-vcfa-re-run-failed`      Run only tests that failed in a previous run (false)
* `-vcfa-remove-test-list`   Remove list of test runs (false)
* `-vcfa-show-count`         Show number of pass/fail tests (false)
* `-vcfa-show-elapsed-time`  Show elapsed time since the start of the suite in pre and post checks (false)
* `-vcfa-show-timestamp`     Show timestamp in pre and post checks (false)
* `-vcfa-skip-pattern`       Skip tests that match the pattern (implies vcfa-pre-post-checks ()

When `-vcfa-pre-post-checks` is used, we have several advantages:

1. After each successful test, the test name gets recorded in a file `VCFA_test_pass_list_{vcfa_IP}.txt`, and each failed
   test goes to `VCFA_test_fail_list_{vcfa_IP}.txt`. When running the suite on the same vcfa a second time, all tests in
   the `pass` list are skipped. If the test run was interrupted (see #2 below), we can only run the tests that did not
   run in the previous attempt.
2. We can **gracefully** interrupt the tests by creating a file `skip_vcfa_tests` in the `./vcfa` directory. 
   When this file is found by the pre-run routine, all the tests are skipped. The file `skip_vcfa_tests` will be removed
   automatically at the next run.
3. We can skip one or more tests conditionally, using `-vcfa-skip-pattern="{REGEXP}"`. All the test with a name that
   matches the pattern are skipped.
4. We can re-run only the tests that failed in the previous run, using `-vcfa-re-run-failed`.
5. We can add monitoring information with `-vcfa-show-count`, `-vcfa-show-elapsed-time`, `-vcfa-show-timestamp`.

If we use `-vcfa-pre-post-checks` and the run was successful, the next run will skip all tests, because the test names
would be all found in `VCFA_test_pass_list_{vcfa_IP}.txt`. To run again the test from scratch, we could either remove
the file manually, or use the tag `-vcfa-remove-test-list`.
<!--  -->
**VERY IMPORTANT**: for the conditional running to work, each test must have a call to `preTestChecks(t)` at the beginning
and immediately defer `postTestChecks(t)` right before the end.

## Leftovers removal

After the test suite runs, an automated process will scan the vcfa and remove any resources that may have been
left behind because of test failure or environment issues.
The procedure can be skipped by using the flag `-vcfa-skip-leftovers-removal`. If you want the operation to omit
details of the scanning, you can use `-vcfa-silent-leftovers-removal`.

To run the removal only, without running the full suite, use the command

```
$ make cleanup
```

## Environment variables and corresponding flags

There are several environment variables that can affect the tests. Many of them have a corresponding flag
that can be used in combination with the `go test` command. You can see them using the `-vcfa-help` flag.

* `TF_ACC=1` enables the acceptance tests. It is also set when you run `make testacc`.
* `GOVCD_DEBUG=1` (`-govcd-debug`) enables debug output of the test suite
* `GOVCD_TRACE=1` (`-govcd-trace`) enables function calls tracing
* `VCFA_SKIP_TEMPLATE_WRITING=1`  (`-vcfa-skip-template-write`) skips the production of test templates into `./vcfa/test-artifacts`
* `VCFA_ADD_PROVIDER=1` (`-vcfa-add-provider`) Adds the full provider definition to the snippets inside `./vcfa/test-artifacts`.
   **WARNING**: the provider definition includes your VCFA credentials.
* `VCFA_SHORT_TEST=1` (`-vcfa-short`) Will not execute the tests themselves, but only generate snippets in `./vcfa/test-artifacts`.
* `VCFA_CONFIG=FileName` sets the file name for the test configuration file.
* `VCFA_TEST_SUITE_CLEANUP=1` will clean up testing resources that were created in previous test runs.
* `VCFA_TEST_VERBOSE=1` (`-vcfa-test-verbose`) enables verbose output in some tests, such as the list of used tags, or the version
used in the documentation index.
* `VCFA_TEST_TRACE=1` (`vcfa-test-trace`) enable trace output in some tests that is not shown in verbose mode
* `VCFA_SKIP_PRIORITY_TESTS=1` (`vcfa-skip-priority-tests`) will disable using test prioritization
  and core infrastructure component sharing
* `VCFA_TEST_ORG_USER=1` (`-vcfa-test-org-user`) will enable tests with Org User, using the credentials from the configuration file
  (`testEnvBuild.OrgUser` and `testEnvBuild.OrgUserPassword`)
* `VCFA_TOKEN=string` : specifies the authentication token to use instead of username/password
   (Use `./scripts/get_token.sh` to retrieve one)
* `VCFA_TEST_DATA_GENERATION=1` generates some sample catalog items for data source filter engine test
* `GOVCD_KEEP_TEST_OBJECTS=1` does not delete test objects created with `VCFA_TEST_DATA_GENERATION`
* `VCFA_MAX_ITEMS=number` during filter engine tests, limits the collection of data sources of a given type to the number
  indicated. The default is 5. The maximum is 100.
* `VCFA_PRE_POST_CHECKS` (`-vcfa-pre-post-checks`) Perform checks before and after tests (false)
* `VCFA_RE_RUN_FAILED` (`-vcfa-re-run-failed`) Run only tests that failed in a previous run (false)
* `VCFA_REMOVE_TEST_LIST` (`-vcfa-remove-test-list`) Remove list of test runs (false)
* `VCFA_SHOW_COUNT` (`-vcfa-show-count`) Show number of pass/fail tests (false)
* `VCFA_SHOW_ELAPSED_TIME` (`-vcfa-show-elapsed-time`) Show elapsed time since the start of the suite in pre and post checks (false)
* `VCFA_SHOW_TIMESTAMP` (`-vcfa-show-timestamp`) Show timestamp in pre and post checks (false)
* `VCFA_SKIP_PATTERN` (`-vcfa-skip-pattern`) Skip tests that match the pattern (implies vcfa-pre-post-checks ()
* `VCFA_SKIP_LEFTOVERS_REMOVAL` (`-vcfa-skip-leftover-removal`) Do not run the leftovers removal at the end of the suite
* `VCFA_SILENT_LEFTOVERS_REMOVAL` (`-vcfa-silent-leftover-removal`) Omit details during leftovers removal.

When both the environment variable and the command line option are possible, the environment variable gets evaluated first.

## Troubleshooting code issues

### Functions for dumping state and pause during acceptance testing

These functions match signature of Terraform's own `resource.TestCheckResourceAttr` and can be
dropped in for troubleshooting problems. 

This function will dump the state at the test run (while executing all field evaluations). It can
help troubleshooting why some fields fail and find typos, wrong state, etc.

```go
func stateDumper() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		spew.Dump(s)
		return nil
	}
}
```

This function can pause test run in the middle which gives the chance to investigate environment
(UI, API calls, etc)

```go
func sleepTester(d time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Printf("sleeping %s\n", d.String())
		time.Sleep(d)
		fmt.Println("finished sleeping")
		return nil
	}
}
```
