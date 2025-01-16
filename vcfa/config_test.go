//go:build api || functional || tm || ALL

package vcfa

// This module provides initialization routines for the whole suite

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/util"
)

// #nosec G101 -- These credentials are fake for testing purposes
const (
	envSecondVcfaUrl      = "VCFA_URL2"
	envSecondVcfaUser     = "VCFA_USER2"
	envSecondVcfaPassword = "VCFA_PASSWORD2"
	envSecondVcfaSysOrg   = "VCFA_SYSORG2"
)

func init() {

	// To list the flags when we run "go test -tags functional -vcfa-help", the flag name must start with "vcfa"
	// They will all appear alongside the native flags when we use an invalid one
	setBoolFlag(&vcfaHelp, "vcfa-help", "VCFA_HELP", "Show vcfa flags")
	setBoolFlag(&vcfaRemoveTestList, "vcfa-remove-test-list", "VCFA_REMOVE_TEST_LIST", "Remove list of test runs")
	setBoolFlag(&vcfaPrePostChecks, "vcfa-pre-post-checks", "VCFA_PRE_POST_CHECKS", "Perform checks before and after tests")
	setBoolFlag(&vcfaShowTimestamp, "vcfa-show-timestamp", "VCFA_SHOW_TIMESTAMP", "Show timestamp in pre and post checks")
	setBoolFlag(&vcfaShowElapsedTime, "vcfa-show-elapsed-time", "VCFA_SHOW_ELAPSED_TIME", "Show elapsed time since the start of the suite in pre and post checks")
	setBoolFlag(&vcfaShowCount, "vcfa-show-count", "VCFA_SHOW_COUNT", "Show number of pass/fail tests")
	setBoolFlag(&vcfaReRunFailed, "vcfa-re-run-failed", "VCFA_RE_RUN_FAILED", "Run only tests that failed in a previous run")
	setBoolFlag(&testDistributedNetworks, "vcfa-test-distributed", "", "enables testing of distributed network")
	setBoolFlag(&enableDebug, "vcfa-debug", "GOVCD_DEBUG", "enables debug output")
	setBoolFlag(&vcfaTestVerbose, "vcfa-verbose", "TEST_VERBOSE", "enables verbose output")
	setBoolFlag(&enableTrace, "vcfa-trace", "GOVCD_TRACE", "enables function calls tracing")
	setBoolFlag(&vcfaShortTest, "vcfa-short", "VCFA_SHORT_TEST", "runs short test")
	setBoolFlag(&vcfaAddProvider, "vcfa-add-provider", envVcfaAddProvider, "add provider to test scripts")
	setBoolFlag(&vcfaSkipTemplateWriting, "vcfa-skip-template-write", envVcfaSkipTemplateWriting, "Skip writing templates to file")
	setBoolFlag(&vcfaRemoveOrgVdcFromTemplate, "vcfa-remove-org-vdc-from-template", envVcfaRemoveOrgVdcFromTemplate, "Remove org and VDC from template")
	setBoolFlag(&vcfaTestOrgUser, "vcfa-test-org-user", envVcfaTestOrgUser, "Run tests with org user")
	setStringFlag(&vcfaSkipPattern, "vcfa-skip-pattern", "VCFA_SKIP_PATTERN", "Skip tests that match the pattern (implies vcfa-pre-post-checks")
	setBoolFlag(&skipLeftoversRemoval, "vcfa-skip-leftovers-removal", "VCFA_SKIP_LEFTOVERS_REMOVAL", "Do not attempt removal of leftovers at the end of the test suite")
	setBoolFlag(&silentLeftoversRemoval, "vcfa-silent-leftovers-removal", "VCFA_SILENT_LEFTOVERS_REMOVAL", "Omit details during removal of leftovers")
	setStringFlag(&testListFileName, "vcfa-partition-tests-file", "VCFA_PARTITION_TESTS_FILE", "Name of the file containing the tests to run in the current partition node")
	setIntFlag(&numberOfPartitions, "vcfa-partitions", "VCFA_PARTITIONS", "")
	setIntFlag(&partitionNode, "vcfa-partition-node", "VCFA_PARTITION_NODE", "")
}

// Structure to get info from a config json file that the user specifies
type TestConfig struct {
	Provider struct {
		User                    string `json:"user"`
		Password                string `json:"password"`
		Token                   string `json:"token,omitempty"`
		ApiToken                string `json:"api_token,omitempty"`
		ApiTokenFile            string `json:"api_token_file,omitempty"`
		ServiceAccountTokenFile string `json:"service_account_token_file,omitempty"`

		// Tenant Manager version and API version, they allow tests to
		// check for compatibility without using an extra connection
		TmVersion    string `json:"tmVersion,omitempty"`
		TmApiVersion string `json:"tmApiVersion,omitempty"`

		Url                      string `json:"url"`
		SysOrg                   string `json:"sysOrg"`
		AllowInsecure            bool   `json:"allowInsecure"`
		TerraformAcceptanceTests bool   `json:"tfAcceptanceTests"`
		UseConnectionCache       bool   `json:"useConnectionCache"`
	} `json:"provider"`
	Tm struct {
		Org            string `json:"org"` // temporary field to make skipIfNotTm work
		CreateRegion   bool   `json:"createRegion"`
		Region         string `json:"region"`
		StorageClass   string `json:"storageClass"`
		Vdc            string `json:"vdc"`
		ContentLibrary string `json:"contentLibrary"`

		CreateNsxtManager   bool   `json:"createNsxtManager"`
		NsxtManagerUsername string `json:"nsxtManagerUsername"`
		NsxtManagerPassword string `json:"nsxtManagerPassword"`
		NsxtManagerUrl      string `json:"nsxtManagerUrl"`
		NsxtEdgeCluster     string `json:"nsxtEdgeCluster"`
		NsxtTier0Gateway    string `json:"nsxtTier0Gateway"`

		CreateVcenter         bool   `json:"createVcenter"`
		VcenterUsername       string `json:"vcenterUsername"`
		VcenterPassword       string `json:"vcenterPassword"`
		VcenterUrl            string `json:"vcenterUrl"`
		VcenterStorageProfile string `json:"vcenterStorageProfile"`
		VcenterSupervisor     string `json:"vcenterSupervisor"`
		VcenterSupervisorZone string `json:"vcenterSupervisorZone"`
	} `json:"tm,omitempty"`
	Logging struct {
		Enabled         bool   `json:"enabled,omitempty"`
		LogFileName     string `json:"logFileName,omitempty"`
		LogHttpRequest  bool   `json:"logHttpRequest,omitempty"`
		LogHttpResponse bool   `json:"logHttpResponse,omitempty"`
	} `json:"logging"`
	Certificates struct {
		Certificate1Path           string `json:"certificate1Path,omitempty"`           // absolute path to pem file
		Certificate1PrivateKeyPath string `json:"certificate1PrivateKeyPath,omitempty"` // absolute path to private key pem file
		Certificate1Pass           string `json:"certificate1Pass,omitempty"`           // pass phrase for private key
		Certificate2Path           string `json:"certificate2Path,omitempty"`           // absolute path to pem file
		Certificate2PrivateKeyPath string `json:"certificate2PrivateKeyPath,omitempty"` // absolute path to private key pem file
		Certificate2Pass           string `json:"certificate2Pass,omitempty"`           // absolute path to pem file
		RootCertificatePath        string `json:"rootCertificatePath,omitempty"`        // absolute path to pem file
	} `json:"certificates"`
	EnvVariables map[string]string `json:"envVariables,omitempty"`
}

// names for created resources for all the tests
var (
	// vcfaAddProvider will add the provide section to the template
	vcfaAddProvider = os.Getenv(envVcfaAddProvider) != ""

	// vcfaSkipTemplateWriting disable the writing of the template to a .tf file
	vcfaSkipTemplateWriting = false

	// vcfaRemoveOrgVdcFromTemplate removes org and vdc from template, thus tetsing with the
	// variables in provider section
	vcfaRemoveOrgVdcFromTemplate = false

	// vcfaTestOrgUser enables testing with the Org User
	vcfaTestOrgUser = false

	// vcfaRemoveTestList triggers the removal of the test run list if present
	vcfaRemoveTestList = false

	// vcfaPrePostChecks enables pre and post checks for all tests
	vcfaPrePostChecks = false

	// vcfaReRunFailed will run only tests that failed in a previous run
	vcfaReRunFailed = false

	// vcfaShowTimestamp shows a time stamp at the start of each test
	vcfaShowTimestamp = false

	// vcfaShowElapsedTime shows the elapsed time since the start od the suite
	vcfaShowElapsedTime = false

	// vcfaShowCount shows the count of pass/skip/fail at the end of the suite
	vcfaShowCount = false

	// vcfaSkipPattern will skip all tests with a name that matches a given pattern
	vcfaSkipPattern string

	// vcfaSkipAllFile is the name of the file that will skip all the tests if found during a pre-test check
	vcfaSkipAllFile = "skip_vcfa_tests"

	// vcfaStartTime is he time when the tests started
	vcfaStartTime = time.Now()

	// vcfaPassCount, vcfaFailCount, vcfaSkipCount are the global counters for the tests result
	vcfaPassCount = 0
	vcfaFailCount = 0
	vcfaSkipCount = 0

	// vcfaHelp displays the vcfa-* flags
	vcfaHelp = false

	// Distributed networks require an edge gateway with distributed routing enabled,
	// which in turn requires a NSX controller. To run the distributed test, users
	// need to set the environment variable VCFA_TEST_DISTRIBUTED_NETWORK
	testDistributedNetworks = false

	// runTestRunListFileLock regulates access to the list of run tests
	runTestRunListFileLock = newMutexKVSilent()

	// skipLeftoversRemoval skips the removal of leftovers at the end of the test suite
	skipLeftoversRemoval = false

	// silentLeftoversRemoval omits details while removing leftovers
	silentLeftoversRemoval = false
)

const (
	providerVcfaOrg1      = "vcfaorg1"
	providerVcfaOrg1Alias = "vcfa.org1"
	providerVcfaOrg2      = "vcfaorg2"
	providerVcfaOrg2Alias = "vcfa.org2"
	providerVcfaSystem2   = "vcfasys2"
	providerVcfaSys2Alias = "vcfa.sys2"

	testArtifactsDirectory          = "test-artifacts"
	envVcfaAddProvider              = "VCFA_ADD_PROVIDER"
	envVcfaSkipTemplateWriting      = "VCFA_SKIP_TEMPLATE_WRITING"
	envVcfaRemoveOrgVdcFromTemplate = "REMOVE_ORG_VDC_FROM_TEMPLATE"
	envVcfaTestOrgUser              = "VCFA_TEST_ORG_USER"

	// Warning message used for all tests
	acceptanceTestsSkipped = "Acceptance tests skipped unless env 'TF_ACC' set"
	// This template will be added to test resource snippets on demand
	providerTemplate = `
# tags {{.Tags}}
# dirname {{.DirName}}
# comment {{.Comment}}
# date {{.Timestamp}}
# file {{.CallerFileName}}
# VCFA version {{.TmVersion}}
# API version {{.TmApiVersion}}

provider "vcfa" {
  user                 = "{{.PrUser}}"
  password             = "{{.PrPassword}}"
  token                = "{{.Token}}"
  api_token            = "{{.ApiToken}}"
  auth_type            = "{{.AuthType}}"
  url                  = "{{.PrUrl}}"
  sysorg               = "{{.PrSysOrg}}"
  org                  = "{{.PrOrg}}"
  allow_unverified_ssl = "{{.AllowInsecure}}"
  logging              = {{.Logging}}
  logging_file         = "{{.LoggingFile}}"
}
`
)

var (

	// This is a global variable shared across all tests. It contains
	// the information from the configuration file.
	testConfig TestConfig

	// Enables the short test (used by "make test")
	vcfaShortTest = os.Getenv("VCFA_SHORT_TEST") != ""

	// Keeps track of test artifact names, to avoid duplicates
	testArtifactNames = make(map[string]string)
)

// usingSysAdmin returns true if the current configuration uses a system administrator for connections
func usingSysAdmin() bool {
	return strings.ToLower(testConfig.Provider.SysOrg) == "system"
}

// skipIfNotSysAdmin skips the calling test if the client is not a system administrator
func skipIfNotSysAdmin(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}
}

// Gets a list of all variables mentioned in a template
func GetVarsFromTemplate(tmpl string) []string {
	var varList []string

	// Regular expression to match a template variable
	// Two opening braces       {{
	// one dot                  \.
	// non-closing-brace chars  [^}]+
	// Two closing braces       }}
	reTemplateVar := regexp.MustCompile(`{{\.([^{]+)}}`)
	captureList := reTemplateVar.FindAllStringSubmatch(tmpl, -1)
	if len(captureList) > 0 {
		for _, capture := range captureList {
			varList = append(varList, capture[1])
		}
	}
	return varList
}

// templateFill fills a template with data provided as a StringMap and adds `provider`
// configuration.
// Returns the text of a ready-to-use Terraform directive. It also saves the filled
// template to a file, for further troubleshooting.
func templateFill(tmpl string, inputData StringMap) string {

	// Copying the input data, to prevent side effects in the original string map:
	// When we use the option -vcfa-add-provider, the data will also contain the fields
	// needed to populate the provider. Some of those fields are empty (e.g. "Token")
	// If the data is evaluated (testParamsNotEmpty) after filling the template, the
	// test gets skipped for what happen to be mysterious reasons.
	data := make(StringMap)
	for k, v := range inputData {
		data[k] = v
	}

	// Gets the name of the function containing the template
	caller := callFuncName()
	realCaller := caller
	// Removes the full path to the function, leaving only package + function name
	caller = filepath.Base(caller)

	_, callerFileName, _, _ := runtime.Caller(1)
	// First, we get all variables in the pattern {{.VarName}}
	varList := GetVarsFromTemplate(tmpl)
	if len(varList) > 0 {
		for _, capture := range varList {
			// For each variable in the template text, we look whether it is
			// in the map
			_, ok := data[capture]
			if !ok {
				data[capture] = fmt.Sprintf("*** MISSING FIELD [%s] from func %s", capture, caller)
			}
		}
	}
	prefix := "vcfa"
	_, ok := data["Prefix"]

	if ok {
		prefix = data["Prefix"].(string)
	}

	// If the call comes from a function that does not have a good descriptive name,
	// (for example when it's an auxiliary function that builds the template but does not
	// run the test) users can add the function name in the data, and it will be used instead of
	// the caller name.
	funcName, ok := data["FuncName"]
	if ok {
		caller = prefix + "." + funcName.(string)
	}

	// If requested, the provider defined in testConfig will be added to test snippets.
	if vcfaAddProvider {
		// the original template is prefixed with the provider template
		tmpl = providerTemplate + tmpl

		// The data structure used to fill the template is integrated with
		// provider data
		data["PrUser"] = testConfig.Provider.User
		data["PrPassword"] = testConfig.Provider.Password
		data["Token"] = testConfig.Provider.Token
		data["ApiToken"] = testConfig.Provider.ApiToken
		data["PrUrl"] = testConfig.Provider.Url
		data["PrSysOrg"] = testConfig.Provider.SysOrg
		data["PrOrg"] = testConfig.Provider.SysOrg
		data["AllowInsecure"] = testConfig.Provider.AllowInsecure
		data["VersionRequired"] = currentProviderVersion
		data["Logging"] = testConfig.Logging.Enabled
		if testConfig.Logging.LogFileName != "" {
			data["LoggingFile"] = testConfig.Logging.LogFileName
		} else {
			data["LoggingFile"] = util.ApiLogFileName
		}

		// Pick correct auth_type
		switch {
		case testConfig.Provider.Token != "":
			data["AuthType"] = "token"
		case testConfig.Provider.ApiToken != "":
			data["AuthType"] = "api_token"
		default:
			data["AuthType"] = "integrated" // default AuthType for local and LDAP users
		}
	}
	if _, ok := data["Tags"]; !ok {
		data["Tags"] = "ALL"
	}
	nullableItems := []string{"Comment", "DirName"}
	for _, item := range nullableItems {
		if _, ok := data[item]; !ok {
			data[item] = "n/a"
		}
	}
	if _, ok := data["CallerFileName"]; !ok {
		data["CallerFileName"] = callerFileName
	}
	data["Timestamp"] = time.Now().Format("2006-01-02 15:04")
	data["TmVersion"] = testConfig.Provider.TmVersion
	data["TmApiVersion"] = testConfig.Provider.TmApiVersion

	// Creates a template. The template gets the same name of the calling function, to generate a better
	// error message in case of failure
	unfilledTemplate := template.Must(template.New(caller).Parse(tmpl))
	buf := &bytes.Buffer{}

	// If an error occurs, returns an empty string
	if err := unfilledTemplate.Execute(buf, data); err != nil {
		return ""
	}
	// Writes the populated template into a directory named "test-artifacts"
	// These templates will help investigate failed tests using Terraform
	// Writing is enabled by default. It can be skipped using an environment variable.
	TemplateWriting := true
	if vcfaSkipTemplateWriting {
		TemplateWriting = false
	}
	var populatedStr = buf.Bytes()

	// This is a quick way of enabling an alternate testing mode:
	// When REMOVE_ORG_VDC_FROM_TEMPLATE is set, the terraform
	// templates will be changed on-the-fly, to comment out the
	// definitions of org and vdc. This will force the test to
	// borrow org and vcfa from the provider.
	if vcfaRemoveOrgVdcFromTemplate {
		reOrg := regexp.MustCompile(`\sorg\s*=`)
		buf2 := reOrg.ReplaceAll(buf.Bytes(), []byte("# org = "))
		reVdc := regexp.MustCompile(`\svdc\s*=`)
		buf2 = reVdc.ReplaceAll(buf2, []byte("# vdc = "))
		populatedStr = buf2
	}
	if TemplateWriting {
		if !dirExists(testArtifactsDirectory) {
			err := os.Mkdir(testArtifactsDirectory, 0750)
			if err != nil {
				panic(fmt.Errorf("error creating directory %s: %s", testArtifactsDirectory, err))
			}
		}
		reProvider1 := regexp.MustCompile(`\bprovider\s*=\s*` + providerVcfaOrg1)
		reProvider2 := regexp.MustCompile(`\bprovider\s*=\s*` + providerVcfaOrg2)
		reSystemProvider2 := regexp.MustCompile(`\bprovider\s*=\s*` + providerVcfaSystem2)

		templateText := string(populatedStr)

		usingProvider1 := reProvider1.MatchString(templateText)
		usingProvider2 := reProvider2.MatchString(templateText)
		usingSysProvider2 := reSystemProvider2.MatchString(templateText)
		// Since the integrated test framework does not support aliases, but the Terraform tool
		// requires them, we change the explicit provider names used in the framework
		// with properly aliased ones (for use in the binary tests)
		if vcfaAddProvider && (usingProvider1 || usingProvider2 || usingSysProvider2) {
			if usingProvider1 {
				templateText = fmt.Sprintf("%s\n%s", templateText, getOrgProviderText("org1", testConfig.Provider.SysOrg))
				templateText = strings.Replace(templateText, providerVcfaOrg1, providerVcfaOrg1Alias, -1)
			}
			if usingProvider2 {
				templateText = fmt.Sprintf("%s\n%s", templateText, getOrgProviderText("org2", testConfig.Provider.SysOrg))
				templateText = strings.Replace(templateText, providerVcfaOrg2, providerVcfaOrg2Alias, -1)
			}
			if usingSysProvider2 {
				templateText = fmt.Sprintf("%s\n%s", templateText, getSysProviderText("sys2"))
				templateText = strings.Replace(templateText, providerVcfaSystem2, providerVcfaSys2Alias, -1)
			}
		}
		resourceFile := path.Join(testArtifactsDirectory, caller) + ".tf"
		storedFunc, alreadyWritten := testArtifactNames[resourceFile]
		if alreadyWritten {
			panic(fmt.Sprintf("File %s was already used from function %s", resourceFile, storedFunc))
		}
		testArtifactNames[resourceFile] = realCaller

		file, err := os.Create(filepath.Clean(resourceFile))
		if err != nil {
			panic(fmt.Errorf("error creating file %s: %s", resourceFile, err))
		}
		writer := bufio.NewWriter(file)
		count, err := writer.Write([]byte(templateText))
		if err != nil || count == 0 {
			panic(fmt.Errorf("error writing to file %s. Reported %d bytes written. %s", resourceFile, count, err))
		}
		err = writer.Flush()
		if err != nil {
			panic(fmt.Errorf("error flushing file %s. %s", resourceFile, err))
		}
		_ = file.Close()
	}
	// Returns the populated template
	return string(populatedStr)
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

	if configStruct.Certificates.Certificate1Path != "" {
		certificatePath1Path, err := filepath.Abs(configStruct.Certificates.Certificate1Path)
		if err != nil {
			panic("error retrieving absolute path for certificate 1 path " + configStruct.Certificates.Certificate1Path)
		}
		configStruct.Certificates.Certificate1Path = certificatePath1Path
	}
	if configStruct.Certificates.Certificate2Path != "" {
		certificatePath2Path, err := filepath.Abs(configStruct.Certificates.Certificate2Path)
		if err != nil {
			panic("error retrieving absolute path for certificate 2 path " + configStruct.Certificates.Certificate2Path)
		}
		configStruct.Certificates.Certificate2Path = certificatePath2Path
	}
	if configStruct.Certificates.Certificate1PrivateKeyPath != "" {
		certificatePrivatePath1Path, err := filepath.Abs(configStruct.Certificates.Certificate1PrivateKeyPath)
		if err != nil {
			panic("error retrieving absolute path for private certificate 1 path " + configStruct.Certificates.Certificate1PrivateKeyPath)
		}
		configStruct.Certificates.Certificate1PrivateKeyPath = certificatePrivatePath1Path
	}
	if configStruct.Certificates.Certificate2PrivateKeyPath != "" {
		certificatePrivatePath2Path, err := filepath.Abs(configStruct.Certificates.Certificate2PrivateKeyPath)
		if err != nil {
			panic("error retrieving absolute path for private certificate 2 path " + configStruct.Certificates.Certificate2PrivateKeyPath)
		}
		configStruct.Certificates.Certificate2PrivateKeyPath = certificatePrivatePath2Path
	}
	if configStruct.Certificates.RootCertificatePath != "" {
		rootCertificatePath2Path, err := filepath.Abs(configStruct.Certificates.RootCertificatePath)
		if err != nil {
			panic("error retrieving absolute path for certificate 2 path " + configStruct.Certificates.Certificate2Path)
		}
		configStruct.Certificates.RootCertificatePath = rootCertificatePath2Path
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
	vcdClient, err := getTestVCFAFromJson(config)
	if vcdClient == nil || err != nil {
		return "", err
	}
	err = ProviderAuthenticate(vcdClient, config.Provider.User, config.Provider.Password, config.Provider.Token, config.Provider.SysOrg, config.Provider.ApiToken, config.Provider.ApiTokenFile, config.Provider.ServiceAccountTokenFile)
	if err != nil {
		return "", err
	}
	version, _, err := vcdClient.Client.GetVcdVersion()
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
	if vcfaHelp {
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
	if vcfaSkipPattern != "" || vcfaShowElapsedTime || vcfaShowTimestamp || vcfaRemoveTestList ||
		vcfaShowCount || vcfaReRunFailed {
		vcfaPrePostChecks = true
	}
	if vcfaPrePostChecks {
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
	if vcfaRemoveTestList {
		for _, ft := range []string{"pass", "fail"} {
			err := removeTestRunList(ft)
			if err != nil {
				fmt.Printf("Error removing testRunList: %s", err)
				fmt.Printf("You should remove the file %s manually before trying again", getTestListFile(ft))
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

	// Runs all test functions
	exitCode := m.Run()

	if numberOfPartitions != 0 {
		entTestFileName := getTestFileName("end", testConfig.Provider.TmVersion)
		err := os.WriteFile(entTestFileName, []byte(fmt.Sprintf("%d", exitCode)), 0600)
		if err != nil {
			fmt.Printf("error writing to file '%s': %s\n", entTestFileName, err)
		}
	}
	if vcfaShowCount {
		fmt.Printf("Pass: %5d - Skip: %5d - Fail: %5d\n", vcfaPassCount, vcfaSkipCount, vcfaFailCount)
	}

	if skipLeftoversRemoval || vcfaShortTest {
		os.Exit(exitCode)
	}
	govcdClient, err := getTestVCFAFromJson(testConfig)
	if err != nil {
		fmt.Printf("error getting a govcd client: %s\n", err)
		exitCode = 1
	} else {
		err = ProviderAuthenticate(govcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken, testConfig.Provider.ApiTokenFile, testConfig.Provider.ServiceAccountTokenFile)
		if err != nil {
			fmt.Printf("error authenticating provider: %s\n", err)
			exitCode = 1
		}
		err := removeLeftovers(govcdClient, !silentLeftoversRemoval)
		if err != nil {
			fmt.Printf("error during leftover removal: %s\n", err)
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}

// Creates a VCDClient based on the endpoint given in the TestConfig argument.
// TestConfig struct can be obtained by calling GetConfigStruct. Throws an error
// if endpoint given is not a valid url.
func getTestVCFAFromJson(testConfig TestConfig) (*govcd.VCDClient, error) {
	configUrl, err := url.ParseRequestURI(testConfig.Provider.Url)
	if err != nil {
		return &govcd.VCDClient{}, fmt.Errorf("could not parse Url: %s", err)
	}
	vcdClient := govcd.NewVCDClient(*configUrl, true,
		govcd.WithHttpUserAgent(buildUserAgent("test", testConfig.Provider.SysOrg)))
	return vcdClient, nil
}

// setBoolFlag binds a flag to a boolean variable (passed as pointer)
// it also uses an optional environment variable that, if set, will
// update the variable before binding it to the flag.
func setBoolFlag(varPointer *bool, name, envVar, help string) {
	if envVar != "" && os.Getenv(envVar) != "" {
		*varPointer = true
	}
	flag.BoolVar(varPointer, name, *varPointer, help)
}

// setStringFlag binds a flag to a string variable (passed as pointer)
// it also uses an optional environment variable that, if set, will
// update the variable before binding it to the flag.
func setStringFlag(varPointer *string, name, envVar, help string) {
	if envVar != "" && os.Getenv(envVar) != "" {
		*varPointer = os.Getenv(envVar)
	}
	flag.StringVar(varPointer, name, *varPointer, help)
}

func setIntFlag(varPointer *int, name, envVar, help string) {
	if envVar != "" && os.Getenv(envVar) != "" {
		var err error
		value := os.Getenv(envVar)
		*varPointer, err = strconv.Atoi(value)
		if err != nil {
			panic(fmt.Sprintf("error converting value '%s' to integer: %s", value, err))
		}
	}
	flag.IntVar(varPointer, name, *varPointer, help)
}

func timeStamp() string {
	now := time.Now()
	return now.Format(time.RFC3339)
}

// preTestChecks is to be called at the beginning of a test function.
// It allows for several skipping mechanisms:
//
//  1. It will skip if the file 'skip_vcfa_tests' is found.
//     This allows to interrupt the test suite in  a clean way, by creating the skipping trigger file
//     during the test run
//     When the user creates such file, the tests still running will continue until their natural end
//     and the other tests will skip
//
// 2) if the file 'skip_vcfa_tests' contains a pattern, only the tests with a name that match such pattern will skip
//
//  3. It will skip if a test has already run successfully. This is useful when the suite was interrupted,
//     so that we can repeat the run without repeating the tests that have succeeded
//
// 4) It will skip the test if a given environment variable was set
//
//  5. It will skip the test if the option -vcfa-skip-pattern or the environment variable 'VCFA_SKIP_PATTERN'
//     contains a pattern that matches the test name.
//  6. If the flag -vcfa-re-run-failed is true, it will only run the tests that failed in the previous run
func preTestChecks(t *testing.T) {
	handlePartitioning(testConfig.Provider.TmVersion, testConfig.Provider.Url, t)
	// if the test runs without -vcfa-pre-post-checks, all post-checks will be skipped
	if !vcfaPrePostChecks {
		return
	}
	if vcfaShowTimestamp {
		fmt.Printf("Test started at: %s\n", timeStamp())
	}
	if vcfaShowElapsedTime {
		elapsed := time.Since(vcfaStartTime)
		fmt.Printf("Elapsed: %s\n", elapsed.String())
	}
	if fileExists(vcfaSkipAllFile) {
		vcfaSkipCount += 1
		t.Skipf("File '%s' found at %s. Test %s skipped", vcfaSkipAllFile, timeStamp(), t.Name())
	}
	if vcfaSkipPattern != "" {
		re := regexp.MustCompile(vcfaSkipPattern)
		if re.MatchString(t.Name()) {
			vcfaSkipCount += 1
			t.Skipf("Skip pattern '%s' matches test name '%s'", vcfaSkipPattern, t.Name())
		}
	}
	skipEnvVar := fmt.Sprintf("skip-%s", t.Name())

	if vcfaTestVerbose {
		fmt.Printf("ENV VAR for %s: %s\n", t.Name(), skipEnvVar)
	}
	if os.Getenv(skipEnvVar) != "" {
		vcfaSkipCount += 1
		t.Skipf("variable '%s' was set.", skipEnvVar)
	}
	// If this test has run already, we skip it
	if isTestInFile(t.Name(), "pass") {
		vcfaSkipCount += 1
		t.Skipf("test '%s' found in '%s' ", t.Name(), getTestListFile("pass"))
	}
	if vcfaReRunFailed {
		if !isTestInFile(t.Name(), "fail") {
			vcfaSkipCount += 1
			t.Skip("only running tests that have failed at the previous run")
		}
	}
}

// postTestChecks runs checks after the test
// It performs the following:
// 1) shows a time stamp (if enabled by -vcfa-show-timestamp
// 2) stores file name in the "pass" or "fail" list, depending on their outcome. The lists are distinct by VCFA IP
// 3) increments the pass/fail counters
func postTestChecks(t *testing.T) {
	// if the test runs without -vcfa-pre-post-checks, all post-checks will be skipped
	if !vcfaPrePostChecks {
		return
	}
	if vcfaShowTimestamp {
		fmt.Printf("Test ended at at: %s\n", timeStamp())
	}
	var err error
	var fileType = "pass"
	if t.Failed() {
		fileType = "fail"
		vcfaFailCount += 1
	} else {
		vcfaPassCount += 1
	}
	err = addToTestRunList(t.Name(), fileType)
	if err != nil {
		fmt.Printf("WARNING: error adding test name '%s' to file '%s'\n", t.Name(), getTestListFile(fileType))
	}
}

// getTestListFile returns the name of the file containing the wanted (pass/fail) list
// for the VCFA being tested
func getTestListFile(fileType string) string {
	if testConfig.Provider.Url == "" {
		return ""
	}
	testingVcfaIp := strings.Replace(testConfig.Provider.Url, "https://", "", -1)
	testingVcfaIp = strings.Replace(testingVcfaIp, "/api", "", -1)
	testingVcfaIp = strings.Replace(testingVcfaIp, "/", "", -1)
	testingVcfaIp = strings.Replace(testingVcfaIp, ".", "-", -1)
	return fmt.Sprintf("vcfa_test_%s_list-%s.txt", fileType, testingVcfaIp)
}

// isTestInFile returns true if a given test name is found in the wanted (pass/fail) list
func isTestInFile(testName, fileType string) bool {
	fileName := getTestListFile(fileType)
	if fileName == "" {
		return false
	}
	runTestRunListFileLock.kvLock(fileName)
	defer runTestRunListFileLock.kvUnlock(fileName)
	if !fileExists(fileName) {
		return false
	}
	f, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		return false
	}
	defer safeClose(f)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == testName {
			return true
		}
	}
	return false
}

// removeTestRunList removes the wanted (pass/fail) list for the VCFA being tested
// This operation is triggered by -vcfa-remove-test-list, and it is needed to run
// a test again after running with -vcfa-pre-post-checks
func removeTestRunList(fileType string) error {
	fileName := getTestListFile(fileType)
	runTestRunListFileLock.kvLock(fileName)
	defer runTestRunListFileLock.kvUnlock(fileName)
	if fileExists(vcfaSkipAllFile) {
		err := os.Remove(vcfaSkipAllFile)
		if err != nil {
			return err
		}
	}
	if !fileExists(fileName) {
		fmt.Printf("[removeTestRunList] '%s' not found\n", fileName)
		return nil
	}
	return os.Remove(fileName)
}

// addToTestRunList adds a given test name to a wanted (pass/fail) list
func addToTestRunList(testName, fileType string) error {
	fileName := getTestListFile(fileType)
	if fileName == "" {
		return nil
	}
	runTestRunListFileLock.kvLock(fileName)
	defer runTestRunListFileLock.kvUnlock(fileName)

	var file *os.File
	var err error
	if fileExists(fileName) {
		file, err = os.OpenFile(filepath.Clean(fileName), os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	} else {
		file, err = os.Create(filepath.Clean(fileName))
	}
	if err != nil {
		return err
	}
	defer safeClose(file)

	w := bufio.NewWriter(file)
	_, err = fmt.Fprintf(w, "%s\n", testName)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %s", fileName, err)
	}
	return w.Flush()
}

func getOrgProviderText(providerName, orgName string) string {
	orgProviderTemplate := `
provider "vcfa" {
  alias                = "{{.ProviderName}}"
  user                 = "{{.OrgUser}}"
  password             = "{{.OrgUserPassword}}"
  auth_type            = "integrated"
  url                  = "{{.VcfaUrl}}"
  sysorg               = "{{.Org}}"
  org                  = "{{.Org}}"
  allow_unverified_ssl = "true"
  logging              = true
  logging_file         = "go-vcloud-director-{{.Org}}.log"
}`

	data := StringMap{
		"VcfaUrl":      testConfig.Provider.Url,
		"SysOrg":       orgName,
		"Org":          orgName,
		"ProviderName": providerName,
		"Alias":        "",
	}
	unfilledTemplate := template.Must(template.New("getOrgProvider").Parse(orgProviderTemplate))
	buf := &bytes.Buffer{}

	// If an error occurs, returns an empty string
	if err := unfilledTemplate.Execute(buf, data); err != nil {
		return ""
	}
	return buf.String()
}

func getSysProviderText(providerName string) string {
	providerSysTemplate := `
provider "vcfa" {
  alias                = "{{.ProviderName}}"
  user                 = "{{.User}}"
  password             = "{{.UserPassword}}"
  auth_type            = "integrated"
  url                  = "{{.VcfaUrl}}"
  sysorg               = "{{.Org}}"
  org                  = "{{.Org}}"
  allow_unverified_ssl = "true"
  logging              = true
  logging_file         = "go-vcloud-director-{{.Org}}.log"
}`

	data := StringMap{
		"User":         os.Getenv(envSecondVcfaUser),
		"UserPassword": os.Getenv(envSecondVcfaPassword),
		"VcfaUrl":      os.Getenv(envSecondVcfaUrl),
		"SysOrg":       os.Getenv(envSecondVcfaSysOrg),
		"Org":          os.Getenv(envSecondVcfaSysOrg),
		"ProviderName": providerName,
		"Alias":        "",
	}
	unfilledTemplate := template.Must(template.New("getSysProvider").Parse(providerSysTemplate))
	buf := &bytes.Buffer{}

	// If an error occurs, returns an empty string
	if err := unfilledTemplate.Execute(buf, data); err != nil {
		return ""
	}
	return buf.String()
}
