//go:build api || functional || tm || cci || contentlibrary || org || ALL

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vcfa

// This module provides initialization routines for the whole suite

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/ccitypes"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
	"github.com/vmware/go-vcloud-director/v3/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware/terraform-provider-vcfa/internal/testutils"
)

// vsphereProviderVersion specifies the version of vSphere Terraform Provider
var vsphereProviderVersion = "~> 2.11"

func init() {
	// All flag names, env-var mappings, and help text are defined once in
	// testutils.RegisterTestFlags. After flag.Parse(), TestMain syncs the
	// parsed values into this package's local vars.
	testutils.RegisterTestFlags()
}

// TestConfig is a transparent alias for the shared definition in internal/testutils.
// The full struct is defined there so it can be imported by both this package and by
// framework-based test packages (e.g. the VKS cluster test) without creating an import cycle.
type TestConfig = testutils.TestConfig

// names for created resources for all the tests
var (
	// vcfaTestOrgUser enables testing with the Org User
	vcfaTestOrgUser = false

	// vcfaSkipAllFile is the name of the file that will skip all the tests if found during a pre-test check
	vcfaSkipAllFile = "skip_vcfa_tests"

	// skipLeftoversRemoval skips the removal of leftovers at the end of the test suite
	skipLeftoversRemoval = false
	onlyLeftoverRemoval  = false

	// silentLeftoversRemoval omits details while removing leftovers
	silentLeftoversRemoval = false
)

const (
	envVcfaAddProvider           = "VCFA_ADD_PROVIDER"
	envVcfaSkipTemplateWriting   = "VCFA_SKIP_TEMPLATE_WRITING"
	envVcfaRemoveOrgFromTemplate = "REMOVE_ORG_FROM_TEMPLATE"
	envVcfaTestOrgUser           = "VCFA_TEST_ORG_USER"

	// Warning message used for all tests
	acceptanceTestsSkipped = "Acceptance tests skipped unless env 'TF_ACC' set"
)

var (

	// This is a global variable shared across all tests. It contains
	// the information from the configuration file.
	testConfig TestConfig

	// Enables the short test (used by "make test")
	vcfaShortTest = os.Getenv("VCFA_SHORT_TEST") != ""
)

// usingSysAdmin returns true if the current configuration uses a system administrator for
// connections. It delegates to testutils so the predicate logic lives in one place.
func usingSysAdmin() bool {
	return testutils.UsingSysAdmin(testConfig)
}

// skipIfNotSysAdmin skips the calling test if the client is not a system administrator.
func skipIfNotSysAdmin(t *testing.T) { testutils.SkipIfNotSysAdmin(t) }

// skipIfSysAdmin skips the calling test if the client is a system administrator.
func skipIfSysAdmin(t *testing.T) { testutils.SkipIfSysAdmin(t) }

// GetVarsFromTemplate returns all variable names referenced in a Go text/template string.
// It delegates to testutils so the logic lives in one place.
func GetVarsFromTemplate(tmpl string) []string {
	return testutils.GetVarsFromTemplate(tmpl)
}

// templateFill fills a template with data provided as a StringMap and adds `provider`
// configuration. It delegates to testutils.TemplateWriteFill so the logic is defined
// in one place and shared with internal subpackage tests.
func templateFill(tmpl string, inputData StringMap) string {
	return testutils.TemplateWriteFill(testConfig, tmpl, testutils.StringMap(inputData))
}

func getConfigFileName() string {
	// First, we see whether the user has indicated a custom configuration file
	// from a non-standard location
	config := os.Getenv("VCFA_CONFIG")

	// If there was no custom file, we look for the default one
	if config == "" {
		config = getCurrentDir() + "/vcfa_test_config.json"
	}
	// Looks if the configuration file exists before attempting to read it
	if fileExists(config) {
		return config
	}
	return ""
}

// Reads the configuration file and returns its contents as a TestConfig structure
// The default file is called vcfa_test_config.json in the same directory where
// the test files are.
// Users may define a file in a different location using the environment variable
// VCFA_CONFIG
// This function doesn't return an error. It panics immediately because its failure
// will prevent the whole test suite from running
func getConfigStruct(config string) TestConfig {
	var configStruct TestConfig

	// Looks if the configuration file exists before attempting to read it
	if config == "" {
		panic(fmt.Errorf("configuration file %s not found", config))
	}
	jsonFile, err := os.ReadFile(filepath.Clean(config))
	if err != nil {
		panic(fmt.Errorf("could not read config file %s: %v", config, err))
	}
	err = json.Unmarshal(jsonFile, &configStruct)
	if err != nil {
		panic(fmt.Errorf("could not unmarshal json file: %v", err))
	}

	// Sets (or clears) environment variables defined in the configuration file
	if configStruct.EnvVariables != nil {
		for key, value := range configStruct.EnvVariables {
			currentEnvValue := os.Getenv(key)
			debugPrintf("# Setting environment variable '%s' from '%s' to '%s'\n", key, currentEnvValue, value)
			_ = os.Setenv(key, value)
		}
	}
	// Reading the configuration file was successful.
	// Now we fill the environment variables that the library is using for its own initialization.
	if configStruct.Provider.TerraformAcceptanceTests {
		// defined in vendor/github.com/hashicorp/terraform/helper/resource/testing.go
		_ = os.Setenv("TF_ACC", "1")
	}

	if vcfaTestOrgUser {
		orgname := configStruct.Org.Name
		user := configStruct.Org.User
		password := configStruct.Org.Password
		if user == "" || password == "" {
			panic(fmt.Sprintf("%s was enabled, but org user credentials were not found in the configuration file", envVcfaTestOrgUser))
		}
		configStruct.Provider.User = user
		configStruct.Provider.Password = password
		configStruct.Provider.SysOrg = orgname
		fmt.Println("VCFA_TEST_ORG_USER was enabled. Using Org User credentials from configuration file")
	}

	if configStruct.Provider.Token != "" && configStruct.Provider.Password == "" {
		configStruct.Provider.Password = "TOKEN"
	}
	_ = os.Setenv("VCFA_USER", configStruct.Provider.User)
	_ = os.Setenv("VCFA_PASSWORD", configStruct.Provider.Password)
	// VCFA_TOKEN supplied via CLI has bigger priority than configured one
	if os.Getenv("VCFA_TOKEN") == "" {
		_ = os.Setenv("VCFA_TOKEN", configStruct.Provider.Token)
	} else {
		configStruct.Provider.Token = os.Getenv("VCFA_TOKEN")
	}

	_ = os.Setenv("VCFA_URL", configStruct.Provider.Url)
	_ = os.Setenv("VCFA_SYS_ORG", configStruct.Provider.SysOrg)
	_ = os.Setenv("VCFA_ORG", configStruct.Provider.SysOrg)
	if configStruct.Provider.UseConnectionCache {
		enableConnectionCache = true
	}
	if configStruct.Provider.AllowInsecure {
		_ = os.Setenv("VCFA_ALLOW_UNVERIFIED_SSL", "1")
	}

	// Define logging parameters if enabled
	if configStruct.Logging.Enabled {
		util.EnableLogging = true
		if configStruct.Logging.LogFileName != "" {
			util.ApiLogFileName = configStruct.Logging.LogFileName
		}
		if configStruct.Logging.LogHttpResponse {
			util.LogHttpResponse = true
		}
		if configStruct.Logging.LogHttpRequest {
			util.LogHttpRequest = true
		}
		util.InitLogging()
	}

	for i, certificate := range configStruct.Tm.Certificates {
		if certificate.Path != "" {
			certPath, err := filepath.Abs(certificate.Path)
			if err != nil {
				panic(fmt.Sprintf("error retrieving absolute path for certificate %d: %s", i, certificate.PrivateKeyPath))
			}
			configStruct.Tm.Certificates[i].Path = certPath
		}
		if certificate.PrivateKeyPath != "" {
			certPath, err := filepath.Abs(certificate.PrivateKeyPath)
			if err != nil {
				panic(fmt.Sprintf("error retrieving absolute path for private key %d: %s", i, certificate.PrivateKeyPath))
			}
			configStruct.Tm.Certificates[i].PrivateKeyPath = certPath
		}
	}
	if configStruct.Tm.RootCertificatePath != "" {
		certPath, err := filepath.Abs(configStruct.Tm.RootCertificatePath)
		if err != nil {
			panic("error retrieving absolute path for root certificate " + configStruct.Tm.RootCertificatePath)
		}
		configStruct.Tm.RootCertificatePath = certPath
	}
	return configStruct
}

// setTestEnv enables environment variables that are also used in non-test code
func setTestEnv() {
	if enableDebug {
		_ = os.Setenv("GOVCD_DEBUG", "1")
	}
}

// getVcfaVersion returns the VCFA version (three digits + build number)
// To get the version, we establish a new connection with the credentials
// chosen for the current test.
func getVcfaVersion(config TestConfig) (string, error) {
	tmClient, err := getTestVCFAFromJson(config)
	if tmClient == nil || err != nil {
		return "", err
	}
	err = ProviderAuthenticate(tmClient, config.Provider.User, config.Provider.Password, config.Provider.Token, config.Provider.SysOrg, config.Provider.ApiToken, config.Provider.ApiTokenFile, config.Provider.ServiceAccountTokenFile)
	if err != nil {
		return "", err
	}
	version, _, err := tmClient.Client.GetVcdVersion()
	if err != nil {
		return "", err
	}
	return version, nil
}

// This function is called before any other test
func TestMain(m *testing.M) {

	// Set BuildVersion to have consistent User-Agent for tests:
	// [e.g. terraform-provider-vcfa/test (darwin/amd64; isProvider:true)]
	BuildVersion = "test"

	// Enable custom flags
	flag.Parse()

	// Sync parsed flag values from testutils into vcfa-local vars so every
	// test file in this package can use the short names without importing testutils.
	vcfaShortTest = testutils.VcfaShortTest
	vcfaTestOrgUser = testutils.TestOrgUser
	skipLeftoversRemoval = testutils.SkipLeftoversRemoval
	onlyLeftoverRemoval = testutils.OnlyLeftoverRemoval
	silentLeftoversRemoval = testutils.SilentLeftoversRemoval
	testListFileName = testutils.TestListFileName
	numberOfPartitions = testutils.NumberOfPartitions
	partitionNode = testutils.PartitionNode
	enableDebug = testutils.EnableDebug
	enableTrace = testutils.EnableTrace
	vcfaTestVerbose = testutils.TestVerbose
	vcfaTestTrace = testutils.TestTrace

	setTestEnv()
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if f.Name == "test.v" {
			if f.Value.String() == "false" {
				fmt.Printf("Missing '-v' flag\n")
				os.Exit(1)
			}
		}
	})
	// If -vcfa-help was in the command line
	if testutils.VcfaHelp {
		fmt.Println("vcfa flags:")
		fmt.Println()
		// Prints only the flags defined in this package
		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			if strings.Contains(f.Name, "vcfa-") {
				fmt.Printf("  -%-40s %s (%v)\n", f.Name, f.Usage, f.Value)
			}
		})
		fmt.Println()
		os.Exit(0)
	}
	// If any of the checks is enabled, we enable the pre and post test functions
	if testutils.VcfaSkipPattern != "" || testutils.VcfaShowElapsedTime || testutils.VcfaShowTimestamp ||
		testutils.VcfaRemoveTestList || testutils.VcfaShowCount || testutils.VcfaReRunFailed {
		testutils.VcfaPrePostChecks = true
	}
	if testutils.VcfaPrePostChecks {
		// remove the user-placed skip file
		_ = os.Remove(vcfaSkipAllFile)
	}

	// Fills the configuration variable: it will be available to all tests,
	// or the whole suite will fail if it is not found.
	// If VCFA_SHORT_TEST is defined, it means that "make test" is called,
	// and we won't really run any tests involving vcfa connections.
	configFile := getConfigFileName()
	if configFile != "" {
		testConfig = getConfigStruct(configFile)
	}

	// Register suite hooks after testConfig is loaded so VcfaUrl/VcfaVersion are available.
	testutils.SetSuiteHooks(testutils.TestSuiteHooks{
		RunPriorityTests: func(t *testing.T, recordResult func(string, bool)) {
			runPriorityTestsOnce(t, recordResult)
		},
		HandlePartitioning: handlePartitioning,
		OnTestFailure: func(t *testing.T) {
			tmClient, err := getTestVCFAFromJson(testConfig)
			if err != nil {
				t.Logf("error getting a govcd client: %s\n", err)
				return
			}
			if err = ProviderAuthenticate(tmClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken, testConfig.Provider.ApiTokenFile, testConfig.Provider.ServiceAccountTokenFile); err != nil {
				t.Logf("error authenticating provider: %s\n", err)
				return
			}
			if err = removeLeftovers(tmClient, !silentLeftoversRemoval, false); err != nil {
				t.Logf("error during leftover removal: %s\n", err)
			}
		},
		VcfaVersion: testConfig.Provider.VcfaVersion,
		VcfaUrl:     testConfig.Provider.Url,
		SkipAllFile: vcfaSkipAllFile,
	})

	if testutils.VcfaRemoveTestList {
		for _, ft := range []string{"pass", "fail"} {
			err := testutils.RemoveTestRunList(ft)
			if err != nil {
				fmt.Printf("Error removing testRunList: %s", err)
				fmt.Printf("You should remove the file %s manually before trying again", testutils.GetTestListFile(ft))
				os.Exit(0)
			}
		}
	}
	if !vcfaShortTest {

		if configFile == "" {
			fmt.Println("No configuration file found")
			os.Exit(1)
		}
		versionInfo, err := getVcfaVersion(testConfig)
		if err == nil {
			versionInfo = fmt.Sprintf("(version %s)", versionInfo)
		}
		fmt.Printf("Connecting to %s %s\n", testConfig.Provider.Url, versionInfo)

		authentication := "password"
		// Token based auth has priority over other types
		if testConfig.Provider.Token != "" {
			authentication = "token"
		}
		if testConfig.Provider.ApiToken != "" {
			authentication = "API-token"
		}

		fmt.Printf("as user %s@%s (using %s)\n", testConfig.Provider.User, testConfig.Provider.SysOrg, authentication)
		// Provider initialization moved here from provider_test.init
		testAccProvider = Provider()
		testAccProviders = map[string]func() (*schema.Provider, error){
			"vcfa": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		}
	}

	if onlyLeftoverRemoval {
		if vcfaTestVerbose {
			fmt.Println("# Running only leftover cleanup")
		}
		var exitCode int
		tmClient, err := getTestVCFAFromJson(testConfig)
		if err != nil {
			fmt.Printf("error getting a govcd client: %s\n", err)
			exitCode = 1
		} else {
			err = ProviderAuthenticate(tmClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken, testConfig.Provider.ApiTokenFile, testConfig.Provider.ServiceAccountTokenFile)
			if err != nil {
				fmt.Printf("error authenticating provider: %s\n", err)
				exitCode = 1
			}
			err := removeLeftovers(tmClient, !silentLeftoversRemoval, true)
			if err != nil {
				fmt.Printf("error during leftover removal: %s\n", err)
				exitCode = 1
			}
		}
		// Exiting early
		os.Exit(exitCode)
	}

	// Runs all test functions
	exitCode := m.Run()

	if numberOfPartitions != 0 {
		entTestFileName := getTestFileName("end", testConfig.Provider.VcfaVersion)
		err := os.WriteFile(entTestFileName, []byte(fmt.Sprintf("%d", exitCode)), 0600)
		if err != nil {
			fmt.Printf("error writing to file '%s': %s\n", entTestFileName, err)
		}
	}
	if testutils.VcfaShowCount {
		fmt.Printf("Pass: %5d - Skip: %5d - Fail: %5d\n", testutils.SuitePassCount(), testutils.SuiteSkipCount(), testutils.SuiteFailCount())
	}

	if skipLeftoversRemoval || vcfaShortTest {
		os.Exit(exitCode)
	}
	tmClient, err := getTestVCFAFromJson(testConfig)
	if err != nil {
		fmt.Printf("error getting a govcd client: %s\n", err)
		exitCode = 1
	} else {
		err = ProviderAuthenticate(tmClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken, testConfig.Provider.ApiTokenFile, testConfig.Provider.ServiceAccountTokenFile)
		if err != nil {
			fmt.Printf("error authenticating provider: %s\n", err)
			exitCode = 1
		}
		err := removeLeftovers(tmClient, !silentLeftoversRemoval, true)
		if err != nil {
			fmt.Printf("error during leftover removal: %s\n", err)
			exitCode = 1
		}
	}

	// If there were some priority tests - cleanup their things
	if priorityTestCleanupFunc != nil {
		err := priorityTestCleanupFunc()
		if err != nil {
			fmt.Printf("# got error while cleaning up vCenter / NSX Manager: %s", err)
		}
	}

	os.Exit(exitCode)
}

func setupVcAndNsx() (func() error, error) {
	tmClient := createSystemTemporaryVCFAConnection()
	nsxManager, nsxCleanup, err := getOrCreateNsxtManager(tmClient.VCDClient)
	if err != nil {
		return nil, fmt.Errorf("got error after NSX Manager creation: %s", err)
	}
	if nsxManager == nil {
		return nil, fmt.Errorf("nil NSX Manager after creation")
	}

	vc, vcCleanup, err := getOrCreateVCenter(tmClient.VCDClient)
	if err != nil {
		return nil, fmt.Errorf("got error after vCenter creation: %s", err)
	}
	if vc == nil {
		return nil, fmt.Errorf("nil vCenter after creation")
	}

	cleanupFunc := func() error {
		if vcfaTestVerbose {
			fmt.Println("# Cleaning up shared vCenter and NSX Manager")
		}
		err := nsxCleanup()
		if err != nil {
			return fmt.Errorf("error cleaning up deferred NSX Manager: %s", err)
		}
		err = vcCleanup()
		if err != nil {
			return fmt.Errorf("error cleaning up deferred vCenter: %s", err)
		}

		return nil

	}

	return cleanupFunc, nil
}

func getOrCreateNsxtManager(tmClient *govcd.VCDClient) (*govcd.NsxtManagerOpenApi, func() error, error) {
	nsxtManager, err := tmClient.GetNsxtManagerOpenApiByUrl(testConfig.Tm.NsxManagerUrl)
	if err == nil {
		return nsxtManager, nil, nil
	}
	if !govcd.ContainsNotFound(err) {
		return nil, nil, err
	}
	if !testConfig.Tm.CreateNsxManager {
		return nil, nil, fmt.Errorf("NSX manager creation disabled")
	}

	if vcfaTestVerbose {
		fmt.Printf("# Creating NSX Manager %s\n", testConfig.Tm.NsxManagerUrl)
	}
	nsxtCfg := &types.NsxtManagerOpenApi{
		Name:     "test-tf-shared-nsx",
		Username: testConfig.Tm.NsxManagerUsername,
		Password: testConfig.Tm.NsxManagerPassword,
		Url:      testConfig.Tm.NsxManagerUrl,
	}
	// Certificate must be trusted before adding NSX Manager
	url, err := url.Parse(nsxtCfg.Url)
	if err != nil {
		return nil, nil, err
	}
	_, err = tmClient.AutoTrustHttpsCertificate(url, nil)
	if err != nil {
		return nil, nil, err
	}
	nsxtManager, err = tmClient.CreateNsxtManagerOpenApi(nsxtCfg)
	if err != nil {
		return nil, nil, err
	}
	nsxtManagerCreated := true

	return nsxtManager, func() error {
		if !nsxtManagerCreated {
			return nil
		}
		if vcfaTestVerbose {
			fmt.Printf("# Deleting NSX Manager %s\n", nsxtManager.NsxtManagerOpenApi.Name)
		}

		nsxManager, err := tmClient.GetNsxtManagerOpenApiByName(nsxtCfg.Name)
		if err != nil {
			if govcd.ContainsNotFound(err) {
				return nil // does not exist, nothing to remove
			}
			return err
		}

		err = nsxManager.Delete()
		if err != nil {
			return err
		}
		return nil
	}, nil
}

func getOrCreateVCenter(tmClient *govcd.VCDClient) (*govcd.VCenter, func() error, error) {
	vc, err := tmClient.GetVCenterByUrl(testConfig.Tm.VcenterUrl)
	if err == nil {
		if !vc.VSphereVCenter.IsEnabled {
			if vcfaTestVerbose {
				fmt.Printf("# vCenter with %s found. Enabling it.\n", testConfig.Tm.VcenterUrl)
			}
			vc.VSphereVCenter.IsEnabled = true
			vc, err = vc.Update(vc.VSphereVCenter)
			if err != nil {
				return nil, nil, err
			}
			err = waitForListenerStatusConnected(vc)
			if err != nil {
				return nil, nil, err
			}
			err = vc.Refresh()
			if err != nil {
				return nil, nil, err
			}
			err = vc.RefreshStorageProfiles()
			if err != nil {
				return nil, nil, err
			}
		}

		return vc, nil, nil
	}
	if !govcd.ContainsNotFound(err) {
		return nil, nil, err
	}
	if !testConfig.Tm.CreateVcenter {
		return nil, nil, fmt.Errorf("vCenter creation disabled")
	}
	printfVerbose("# Creating vCenter %s\n", testConfig.Tm.VcenterUrl)

	vcCfg := &types.VSphereVirtualCenter{
		Name:      "test-tf-shared-vc",
		Username:  testConfig.Tm.VcenterUsername,
		Password:  testConfig.Tm.VcenterPassword,
		Url:       testConfig.Tm.VcenterUrl,
		IsEnabled: true,
	}
	// Certificate must be trusted before adding vCenter
	url, err := url.Parse(vcCfg.Url)
	if err != nil {
		return nil, nil, err
	}
	_, err = tmClient.AutoTrustHttpsCertificate(url, nil)
	if err != nil {
		return nil, nil, err
	}

	vc, err = tmClient.CreateVcenter(vcCfg)
	if err != nil {
		return nil, nil, err
	}

	printfTrace("# Waiting for listener status to become 'CONNECTED'\n")
	err = waitForListenerStatusConnected(vc)
	if err != nil {
		return nil, nil, err
	}

	afterConnectedSleep := 4 * time.Second
	printfTrace("# Sleeping %s after vCenter is 'CONNECTED' \n", afterConnectedSleep.String())

	time.Sleep(afterConnectedSleep) // TODO: TM: Re-evaluate need for sleep
	// Refresh connected vCenter to be sure that all artifacts are loaded
	printfTrace("# Refreshing vCenter %s\n", vc.VSphereVCenter.Url)

	err = vc.RefreshVcenter()
	if err != nil {
		return nil, nil, err
	}

	printfTrace("# Refreshing Storage Profiles in vCenter %s\n", vc.VSphereVCenter.Url)
	err = vc.RefreshStorageProfiles()
	if err != nil {
		return nil, nil, err
	}

	afterRefreshSleep := 30 * time.Second
	printfTrace("# Sleeping %s after vCenter refreshes \n", afterRefreshSleep.String())

	time.Sleep(afterRefreshSleep) // TODO: TM: Re-evaluate need for sleep
	vCenterCreated := true

	return vc, func() error {
		if !vCenterCreated {
			return nil
		}
		vc, err := tmClient.GetVCenterByName(vcCfg.Name)
		if govcd.ContainsNotFound(err) {
			return nil // does not exist, nothing to remove
		}
		printfVerbose("# Disabling and deleting vCenter %s\n", testConfig.Tm.VcenterUrl)

		err = vc.Disable()
		if err != nil {
			return err
		}
		err = vc.Delete()
		if err != nil {
			return err
		}
		return nil
	}, nil
}

func waitForListenerStatusConnected(v *govcd.VCenter) error {
	startTime := time.Now()
	tryCount := 20
	for c := 0; c < tryCount; c++ {
		err := v.Refresh()
		if err != nil {
			return fmt.Errorf("error refreshing vCenter: %s", err)
		}

		if v.VSphereVCenter.ListenerState == "CONNECTED" {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("waiting for listener state to become 'CONNECTED' expired after %d tries (%d seconds), got '%s'",
		tryCount, int(time.Since(startTime)/time.Second), v.VSphereVCenter.ListenerState)
}

// Creates a VCDClient based on the endpoint given in the TestConfig argument.
// TestConfig struct can be obtained by calling GetConfigStruct. Throws an error
// if endpoint given is not a valid url.
func getTestVCFAFromJson(testConfig TestConfig) (*govcd.VCDClient, error) {
	configUrl, err := url.ParseRequestURI(testConfig.Provider.Url)
	if err != nil {
		return &govcd.VCDClient{}, fmt.Errorf("could not parse Url: %s", err)
	}
	tmClient := govcd.NewVCDClient(*configUrl, true,
		govcd.WithHttpUserAgent(buildUserAgent("test", testConfig.Provider.SysOrg)),
		govcd.WithAPIVersion(minVcfaApiVersion),
	)
	return tmClient, nil
}

// preTestChecks is to be called at the beginning of each test function.
func preTestChecks(t *testing.T) { testutils.PreTestChecks(t) }

// postTestChecks is to be called at the end of each test function.
func postTestChecks(t *testing.T) { testutils.PostTestChecks(t) }

func printfVerbose(format string, args ...interface{}) {
	if vcfaTestVerbose {
		fmt.Printf(format, args...)
	}
}

func printfTrace(format string, args ...interface{}) {
	if vcfaTestTrace {
		fmt.Printf(format, args...)
	}
}

var priorityTestCleanupFunc func() error

type priorityTest struct {
	Name string
	Test func(*testing.T)
}

var registeredPriorityTests = []priorityTest{}

// runPriorityTestsOnce runs each registered priority test and records the result
// via recordResult so that testutils.PreTestChecks can skip a test that already ran.
// It also sets up the shared vCenter / NSX Manager infrastructure.
func runPriorityTestsOnce(t *testing.T, recordResult func(name string, passed bool)) {
	fmt.Printf("# Running priority tests before shared vCenter and NSX Manager is created, so they do not collide later (can be skipped with '-vcfa-skip-priority-tests' flag)\n")
	for _, test := range registeredPriorityTests {
		fmt.Printf("# Running priority test '%s' as a subtest of '%s':\n", test.Name, t.Name())
		t.Run(test.Name, test.Test)
		printfTrace("## Storing test to executed test list '%s'\n", test.Name)
		recordResult(test.Name, !t.Failed())
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

func createProject(t *testing.T, orgName string, username string, password string, projectName string) {
	tmClient := createTemporaryOrgConnection(orgName, username, password)
	projectCfg := &ccitypes.Project{
		TypeMeta: v1.TypeMeta{
			Kind:       ccitypes.ProjectKind,
			APIVersion: ccitypes.ProjectAPI + "/" + ccitypes.ProjectVersion,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: projectName,
		},
		Spec: ccitypes.ProjectSpec{
			Description: fmt.Sprintf("Terraform test project [%s]", projectName),
		},
	}

	newProjectAddr, err := tmClient.Client.GetEntityUrl(ccitypes.ProjectsURL)
	if err != nil {
		t.Fatalf("error creating URL for new project: %s", err)
	}

	newProject := &ccitypes.Project{}
	err = tmClient.Client.PostEntity(newProjectAddr, nil, projectCfg, newProject, nil)
	if err != nil {
		t.Fatalf("error creating project %s: %s", projectCfg.Name, err)
	}

	err = os.Setenv("TF_VAR_project_id", fmt.Sprintf("urn:vcloud:projectAssignment:%s", string(newProject.UID)))
	if err != nil {
		t.Fatalf("error setting project_id environment variable: %s", err)
	}
}

func removeProject(t *testing.T, orgName string, username string, password string, projectName string) {
	tmClient := createTemporaryOrgConnection(orgName, username, password)

	projectAddr, err := tmClient.Client.GetEntityUrl(ccitypes.ProjectsURL, "/", projectName)
	if err != nil {
		t.Fatalf("error creating URL for get project: %s", err)
	}

	err = tmClient.Client.DeleteEntity(projectAddr, nil, nil)
	if err != nil {
		t.Fatalf("failed removing Project: %s", err)
	}
}
