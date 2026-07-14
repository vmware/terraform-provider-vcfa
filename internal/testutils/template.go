// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package testutils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"

	"github.com/vmware/go-vcloud-director/v3/util"
)

// StringMap is used to simplify reading resource definitions, matching the type used by the
// `vcfa` package test helpers.
type StringMap map[string]interface{}

const (
	// TestArtifactsDirectory is the directory where filled templates are written.
	TestArtifactsDirectory = "test-artifacts"

	// Provider alias names and their Terraform alias strings for multi-provider tests.
	ProviderVcfaOrg1      = "vcfaorg1"
	ProviderVcfaOrg1Alias = "vcfa.org1"
	ProviderVcfaOrg2      = "vcfaorg2"
	ProviderVcfaOrg2Alias = "vcfa.org2"
	ProviderVcfaSystem2   = "vcfasys2"
	ProviderVcfaSys2Alias = "vcfa.sys2"
)

// providerTemplate is prepended to test snippets by TemplateFill.
const providerTemplate = `
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
}
`

// fullProviderTemplate is the metadata-rich template used by TemplateWriteFill.
const fullProviderTemplate = `
# tags {{.Tags}}
# dirname {{.DirName}}
# comment {{.Comment}}
# date {{.Timestamp}}
# file {{.CallerFileName}}
# VCFA version {{.VcfaVersion}}
# API version {{.ApiVersion}}

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

var (
	testArtifactNames   = make(map[string]string)
	testArtifactNamesMu sync.Mutex
)

// TestParamsNotEmpty skips the test if any of the provided params is an empty string.
func TestParamsNotEmpty(t *testing.T, params StringMap) {
	t.Helper()
	for key, value := range params {
		if value == "" {
			t.Skipf("[%s] %s must be set for acceptance tests", t.Name(), key)
		}
	}
}

// TemplateFill renders the given Terraform template with the provided data, prepending a
// `provider "vcfa"` block populated from the shared test configuration. It returns
// ready-to-use Terraform configuration text.
func TemplateFill(t *testing.T, tmpl string, inputData StringMap) string {
	t.Helper()
	cfg := GetTestConfig(t)

	// Copy the input data to avoid mutating the caller's map when adding provider fields.
	data := make(StringMap, len(inputData)+12)
	for k, v := range inputData {
		data[k] = v
	}

	tmpl = providerTemplate + tmpl

	data["PrUser"] = cfg.Provider.User
	data["PrPassword"] = cfg.Provider.Password
	data["Token"] = cfg.Provider.Token
	data["ApiToken"] = cfg.Provider.ApiToken
	data["PrUrl"] = cfg.Provider.Url
	data["PrSysOrg"] = cfg.Provider.SysOrg
	data["PrOrg"] = cfg.Provider.SysOrg
	data["AllowInsecure"] = cfg.Provider.AllowInsecure

	switch {
	case cfg.Provider.Token != "":
		data["AuthType"] = "token"
	case cfg.Provider.ApiToken != "":
		data["AuthType"] = "api_token"
	default:
		data["AuthType"] = "integrated"
	}

	parsed, err := template.New(t.Name()).Parse(tmpl)
	if err != nil {
		t.Fatalf("error parsing template: %s", err)
	}

	buf := &bytes.Buffer{}
	if err := parsed.Execute(buf, data); err != nil {
		t.Fatalf("error executing template: %s", err)
	}

	return buf.String()
}

// TemplateWriteFill is the full-featured template renderer used by the vcfa package tests.
// It conditionally prepends a provider block, handles multi-org provider aliases, writes
// the rendered template to TestArtifactsDirectory for troubleshooting, and supports
// org-removal mode. It mirrors the vcfa package's templateFill function.
func TemplateWriteFill(cfg TestConfig, tmpl string, inputData StringMap) string {
	// Copy the input data to prevent side effects on the caller's map.
	data := make(StringMap)
	for k, v := range inputData {
		data[k] = v
	}

	caller := callerTestFuncName()
	realCaller := caller
	caller = filepath.Base(caller)
	caller = strings.ReplaceAll(caller, "/", "-")

	_, callerFileName, _, _ := runtime.Caller(1)

	varList := GetVarsFromTemplate(tmpl)
	for _, capture := range varList {
		if _, ok := data[capture]; !ok {
			data[capture] = fmt.Sprintf("*** MISSING FIELD [%s] from func %s", capture, caller)
		}
	}

	prefix := "vcfa"
	if p, ok := data["Prefix"]; ok {
		prefix = p.(string)
	}
	if fn, ok := data["FuncName"]; ok {
		caller = prefix + "." + fn.(string)
	}

	if VcfaAddProvider {
		tmpl = fullProviderTemplate + tmpl

		data["PrUser"] = cfg.Provider.User
		data["PrPassword"] = cfg.Provider.Password
		data["Token"] = cfg.Provider.Token
		data["ApiToken"] = cfg.Provider.ApiToken
		data["PrUrl"] = cfg.Provider.Url
		data["PrSysOrg"] = cfg.Provider.SysOrg
		data["PrOrg"] = cfg.Provider.SysOrg
		data["AllowInsecure"] = cfg.Provider.AllowInsecure
		data["VersionRequired"] = providerMajorVersion()
		data["Logging"] = cfg.Logging.Enabled
		if cfg.Logging.LogFileName != "" {
			data["LoggingFile"] = cfg.Logging.LogFileName
		} else {
			data["LoggingFile"] = util.ApiLogFileName
		}

		switch {
		case cfg.Provider.Token != "":
			data["AuthType"] = "token"
		case cfg.Provider.ApiToken != "":
			data["AuthType"] = "api_token"
		default:
			data["AuthType"] = "integrated"
		}
	}

	if _, ok := data["Tags"]; !ok {
		data["Tags"] = "ALL"
	}
	for _, item := range []string{"Comment", "DirName"} {
		if _, ok := data[item]; !ok {
			data[item] = "n/a"
		}
	}
	if _, ok := data["CallerFileName"]; !ok {
		data["CallerFileName"] = callerFileName
	}
	data["Timestamp"] = time.Now().Format("2006-01-02 15:04")
	data["VcfaVersion"] = cfg.Provider.VcfaVersion
	data["ApiVersion"] = cfg.Provider.ApiVersion

	unfilledTemplate := template.Must(template.New(caller).Parse(tmpl))
	buf := &bytes.Buffer{}
	if err := unfilledTemplate.Execute(buf, data); err != nil {
		return ""
	}

	templateWriting := !VcfaSkipTemplateWriting
	populatedStr := buf.Bytes()

	if VcfaRemoveOrgFromTemplate {
		reOrg := regexp.MustCompile(`\sorg\s*=`)
		populatedStr = reOrg.ReplaceAll(buf.Bytes(), []byte("# org = "))
	}

	if templateWriting {
		if !dirExists(TestArtifactsDirectory) {
			if err := os.Mkdir(TestArtifactsDirectory, 0750); err != nil {
				panic(fmt.Errorf("error creating directory %s: %s", TestArtifactsDirectory, err))
			}
		}

		reProvider1 := regexp.MustCompile(`\bprovider\s*=\s*` + ProviderVcfaOrg1)
		reProvider2 := regexp.MustCompile(`\bprovider\s*=\s*` + ProviderVcfaOrg2)
		reSystemProvider2 := regexp.MustCompile(`\bprovider\s*=\s*` + ProviderVcfaSystem2)

		templateText := string(populatedStr)
		usingProvider1 := reProvider1.MatchString(templateText)
		usingProvider2 := reProvider2.MatchString(templateText)
		usingSysProvider2 := reSystemProvider2.MatchString(templateText)

		if VcfaAddProvider && (usingProvider1 || usingProvider2 || usingSysProvider2) {
			if usingProvider1 {
				templateText = fmt.Sprintf("%s\n%s", templateText, orgProviderText(cfg, "org1", cfg.Provider.SysOrg))
				templateText = strings.ReplaceAll(templateText, ProviderVcfaOrg1, ProviderVcfaOrg1Alias)
			}
			if usingProvider2 {
				templateText = fmt.Sprintf("%s\n%s", templateText, orgProviderText(cfg, "org2", cfg.Provider.SysOrg))
				templateText = strings.ReplaceAll(templateText, ProviderVcfaOrg2, ProviderVcfaOrg2Alias)
			}
			if usingSysProvider2 {
				templateText = fmt.Sprintf("%s\n%s", templateText, sysProviderText("sys2"))
				templateText = strings.ReplaceAll(templateText, ProviderVcfaSystem2, ProviderVcfaSys2Alias)
			}
		}

		resourceFile := path.Join(TestArtifactsDirectory, caller) + ".tf"
		testArtifactNamesMu.Lock()
		storedFunc, alreadyWritten := testArtifactNames[resourceFile]
		if alreadyWritten {
			testArtifactNamesMu.Unlock()
			panic(fmt.Sprintf("File %s was already used from function %s", resourceFile, storedFunc))
		}
		testArtifactNames[resourceFile] = realCaller
		testArtifactNamesMu.Unlock()

		file, err := os.Create(filepath.Clean(resourceFile))
		if err != nil {
			panic(fmt.Errorf("error creating file %s: %s", resourceFile, err))
		}
		writer := bufio.NewWriter(file)
		count, err := writer.Write([]byte(templateText))
		if err != nil || count == 0 {
			panic(fmt.Errorf("error writing to file %s. Reported %d bytes written. %s", resourceFile, count, err))
		}
		if err = writer.Flush(); err != nil {
			panic(fmt.Errorf("error flushing file %s. %s", resourceFile, err))
		}
		_ = file.Close()
	}

	return string(populatedStr)
}

// DebugPrintf displays conditional debug messages when GOVCD_DEBUG is enabled.
func DebugPrintf(format string, args ...interface{}) {
	if os.Getenv("GOVCD_DEBUG") != "" {
		fmt.Printf(format, args...)
	}
}

// GetVarsFromTemplate returns the list of variable names referenced in a Go text/template
// string (i.e. all occurrences of {{.VarName}}). This is useful for detecting missing
// template fields before rendering.
func GetVarsFromTemplate(tmpl string) []string {
	var varList []string
	reTemplateVar := regexp.MustCompile(`{{\.([^{]+)}}`)
	for _, capture := range reTemplateVar.FindAllStringSubmatch(tmpl, -1) {
		varList = append(varList, capture[1])
	}
	return varList
}

// callerTestFuncName walks the call stack to find the nearest function whose name
// contains ".Test" and is not inside the testutils package itself.
func callerTestFuncName() string {
	pcs := make([]uintptr, 25)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		fn := frame.Function
		base := filepath.Base(fn)
		if !strings.Contains(fn, "testutils") &&
			!strings.HasPrefix(base, "testing.") &&
			strings.Contains(fn, ".Test") {
			return fn
		}
		if !more {
			break
		}
	}
	return "unknown"
}

// providerMajorVersion reads the VERSION file from the module root and returns the
// major.minor version string (e.g. "2.0").
func providerMajorVersion() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "unknown"
	}
	// internal/testutils/ is two directories below the module root
	moduleRoot := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	versionFile := filepath.Join(moduleRoot, "VERSION")
	raw, err := os.ReadFile(filepath.Clean(versionFile))
	if err != nil {
		return "unknown"
	}
	reVersion := regexp.MustCompile(`v(\d+\.\d+)\.\d+`)
	matches := reVersion.FindAllStringSubmatch(strings.TrimSpace(string(raw)), -1)
	if len(matches) == 0 || len(matches[0]) < 2 {
		return "unknown"
	}
	return matches[0][1]
}

// dirExists returns true if the given path exists and is a directory.
func dirExists(path string) bool {
	f, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return f.Mode().IsDir()
}

// orgProviderText builds a Terraform provider alias block for an org (tenant) provider.
func orgProviderText(cfg TestConfig, providerName, orgName string) string {
	const orgProviderTmpl = `
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
		"VcfaUrl":         cfg.Provider.Url,
		"SysOrg":          orgName,
		"Org":             orgName,
		"ProviderName":    providerName,
		"OrgUser":         "",
		"OrgUserPassword": "",
		"Alias":           "",
	}
	parsed := template.Must(template.New("orgProvider").Parse(orgProviderTmpl))
	buf := &bytes.Buffer{}
	if err := parsed.Execute(buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// sysProviderText builds a Terraform provider alias block for a second system provider,
// reading credentials from the VCFA_URL2 / VCFA_USER2 / VCFA_PASSWORD2 / VCFA_SYSORG2 env vars.
func sysProviderText(providerName string) string {
	const sysProviderTmpl = `
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
		"User":         os.Getenv("VCFA_USER2"),
		"UserPassword": os.Getenv("VCFA_PASSWORD2"),
		"VcfaUrl":      os.Getenv("VCFA_URL2"),
		"SysOrg":       os.Getenv("VCFA_SYSORG2"),
		"Org":          os.Getenv("VCFA_SYSORG2"),
		"ProviderName": providerName,
		"Alias":        "",
	}
	parsed := template.Must(template.New("sysProvider").Parse(sysProviderTmpl))
	buf := &bytes.Buffer{}
	if err := parsed.Execute(buf, data); err != nil {
		return ""
	}
	return buf.String()
}
