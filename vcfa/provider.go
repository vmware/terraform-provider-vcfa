package vcfa

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v3/util"
)

// BuildVersion holds version which is meant to be injected at build time using ldflags
// (e.g. 'go build -ldflags="-X 'github.com/vmware/terraform-provider-vcfa/vcfa.BuildVersion=v1.0.0'"')
var BuildVersion = "unset"

// DataSources is a public function which allows filtering and access all defined data sources
// When 'nameRegexp' is not empty - it will return only those matching the regexp
// When 'includeDeprecated' is false - it will skip out the resources which have a DeprecationMessage set
func DataSources(nameRegexp string, includeDeprecated bool) (map[string]*schema.Resource, error) {
	return vcfaSchemaFilter(globalDataSourceMap, nameRegexp, includeDeprecated)
}

// Resources is a public function which allows filtering and access all defined resources
// When 'nameRegexp' is not empty - it will return only those matching the regexp
// When 'includeDeprecated' is false - it will skip out the resources which have a DeprecationMessage set
func Resources(nameRegexp string, includeDeprecated bool) (map[string]*schema.Resource, error) {
	return vcfaSchemaFilter(globalResourceMap, nameRegexp, includeDeprecated)
}

var globalDataSourceMap = map[string]*schema.Resource{
	"vcfa_tm_version": datasourceVcfaTmVersion(), // 1.0
}

var globalResourceMap = map[string]*schema.Resource{
	"vcfa_deleteme": resourceVcfaDeleteme(), // TODO: VCFA: Delete this (and associated doc) once there's a resource ready
}

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{

			// TODO: VCFA: Revisit and review the existing options

			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_USER", nil),
				Description: "The user name for VCFA API operations.",
			},

			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_PASSWORD", nil),
				Description: "The user password for VCFA API operations.",
			},

			"auth_type": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("VCFA_AUTH_TYPE", "integrated"),
				Description:  "'integrated', 'token', 'api_token', 'api_token_file' and 'service_account_token_file' are supported. 'integrated' is default.",
				ValidateFunc: validation.StringInSlice([]string{"integrated", "token", "api_token", "api_token_file", "service_account_token_file"}, false),
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_TOKEN", nil),
				Description: "The token used instead of username/password for VCFA API operations.",
			},

			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_API_TOKEN", nil),
				Description: "The API token used instead of username/password for VCFA API operations. (Requires VCFA 10.3.1+)",
			},

			"api_token_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_API_TOKEN_FILE", nil),
				Description: "The API token file instead of username/password for VCFA API operations. (Requires VCFA 10.3.1+)",
			},

			"allow_api_token_file": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set this to true if you understand the security risks of using API token files and would like to suppress the warnings",
			},

			"service_account_token_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_SA_TOKEN_FILE", nil),
				Description: "The Service Account API token file instead of username/password for VCFA API operations. (Requires VCFA 9.0+)",
			},

			"allow_service_account_token_file": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set this to true if you understand the security risks of using Service Account token files and would like to suppress the warnings",
			},

			"sysorg": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_SYS_ORG", nil),
				Description: "The VCFA Org for user authentication",
			},

			"org": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_ORG", nil),
				Description: "The VCFA Org for API operations",
			},

			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_VDC", nil),
				Description: "The VDC for API operations",
			},

			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_URL", nil),
				Description: "The VCFA url for VCFA API operations.",
			},

			"allow_unverified_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_ALLOW_UNVERIFIED_SSL", false),
				Description: "If set, VCFAClient will permit unverifiable SSL certificates.",
			},

			"logging": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_API_LOGGING", false),
				Description: "If set, it will enable logging of API requests and responses",
			},

			"logging_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_API_LOGGING_FILE", "go-vcloud-director.log"),
				Description: "Defines the full name of the logging file for API calls (requires 'logging')",
			},
			"import_separator": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCFA_IMPORT_SEPARATOR", "."),
				Description: "Defines the import separation string to be used with 'terraform import'",
			},
		},
		ResourcesMap:         globalResourceMap,
		DataSourcesMap:       globalDataSourceMap,
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	if err := validateProviderSchema(d); err != nil {
		return nil, diag.Errorf("[provider validation] :%s", err)
	}

	// If sysOrg is defined, we use it for authentication.
	// Otherwise, we use the default org defined for regular usage
	connectOrg := d.Get("sysorg").(string)
	if connectOrg == "" {
		connectOrg = d.Get("org").(string)
	}

	config := Config{
		User:                    d.Get("user").(string),
		Password:                d.Get("password").(string),
		Token:                   d.Get("token").(string),
		ApiToken:                d.Get("api_token").(string),
		ApiTokenFile:            d.Get("api_token_file").(string),
		AllowApiTokenFile:       d.Get("allow_api_token_file").(bool),
		ServiceAccountTokenFile: d.Get("service_account_token_file").(string),
		AllowSATokenFile:        d.Get("allow_service_account_token_file").(bool),
		SysOrg:                  connectOrg,            // Connection org
		Org:                     d.Get("org").(string), // Default org for operations
		Vdc:                     d.Get("vdc").(string), // Default vdc
		Href:                    d.Get("url").(string),
		InsecureFlag:            d.Get("allow_unverified_ssl").(bool),
	}

	// auth_type dependent configuration
	authType := d.Get("auth_type").(string)
	switch authType {
	case "token":
		if config.Token == "" {
			return nil, diag.Errorf("empty token detected with 'auth_type' == 'token'")
		}
	case "api_token":
		if config.ApiToken == "" {
			return nil, diag.Errorf("empty API token detected with 'auth_type' == 'api_token'")
		}
	case "service_account_token_file":
		if config.ServiceAccountTokenFile == "" {
			return nil, diag.Errorf("service account token file not provided with 'auth_type' == 'service_account_token_file'")
		}
	case "api_token_file":
		if config.ApiTokenFile == "" {
			return nil, diag.Errorf("api token file not provided with 'auth_type' == 'service_account_token_file'")
		}
	default:
		if config.ApiToken != "" || config.Token != "" {
			return nil, diag.Errorf("to use a token, the appropriate 'auth_type' (either 'token' or 'api_token') must be set")
		}
	}
	if config.ApiToken != "" && config.Token != "" {
		return nil, diag.Errorf("only one of 'token' or 'api_token' should be set")
	}

	var providerDiagnostics diag.Diagnostics
	if config.ServiceAccountTokenFile != "" && !config.AllowSATokenFile {
		providerDiagnostics = append(providerDiagnostics, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The file " + config.ServiceAccountTokenFile + " should be considered sensitive information.",
			Detail: "The file " + config.ServiceAccountTokenFile + " containing the initial service account API " +
				"HAS BEEN UPDATED with a freshly generated token. The initial token was invalidated and the " +
				"token currently in the file will be invalidated at the next usage. In the meantime, it is " +
				"usable by anyone to run operations to the current VCFA. As such, it should be considered SENSITIVE INFORMATION. " +
				"If you would like to remove this warning, add\n\n" + "	allow_service_account_token_file = true\n\nto the provider settings.",
		})
	}

	if config.ApiTokenFile != "" && !config.AllowApiTokenFile {
		providerDiagnostics = append(providerDiagnostics, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The file " + config.ServiceAccountTokenFile + " should be considered sensitive information.",
			Detail: "The file " + config.ServiceAccountTokenFile + " contains the API token which can be used by anyone " +
				"to run operations to the current VCFA. AS such, it should be considered SENSITIVE INFORMATION. " +
				"If you would like to remove this warning, add\n\n" + "	allow_api_token_file = true\n\nto the provider settings.",
		})
	}

	// If the provider includes logging directives,
	// it will activate logging from upstream go-vcloud-director
	logging := d.Get("logging").(bool)
	// Logging is disabled by default.
	// If enabled, we set the log file name and invoke the upstream logging set-up
	if logging {
		loggingFile := d.Get("logging_file").(string)
		if loggingFile != "" {
			util.EnableLogging = true
			util.ApiLogFileName = loggingFile
			util.InitLogging()
		}
	}

	separator := os.Getenv("VCFA_IMPORT_SEPARATOR")
	if separator != "" {
		ImportSeparator = separator
	} else {
		ImportSeparator = d.Get("import_separator").(string)
	}

	vcdClient, err := config.Client()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return vcdClient, providerDiagnostics
}

// vcfaSchemaFilter is a function which allows to filters and export type 'map[string]*schema.Resource' which may hold
// Terraform's native resource or data source list
// When 'nameRegexp' is not empty - it will return only those matching the regexp
// When 'includeDeprecated' is false - it will skip out the resources which have a DeprecationMessage set
func vcfaSchemaFilter(schemaMap map[string]*schema.Resource, nameRegexp string, includeDeprecated bool) (map[string]*schema.Resource, error) {
	var (
		err error
		re  *regexp.Regexp
	)
	filteredResources := make(map[string]*schema.Resource)

	// validate regex if it was provided
	if nameRegexp != "" {
		re, err = regexp.Compile(nameRegexp)
		if err != nil {
			return nil, fmt.Errorf("unable to compile regexp: %s", err)
		}
	}

	// copy the map with filtering out unwanted object
	for resourceName, schemaResource := range schemaMap {

		// Skip deprecated resources if it was requested so
		if !includeDeprecated && schemaResource.DeprecationMessage != "" {
			continue
		}
		// If regex was defined - try to filter based on it
		if re != nil {
			// if it does not match regex - skip it
			doesNotmatchRegex := !re.MatchString(resourceName)
			if doesNotmatchRegex {
				continue
			}

		}

		filteredResources[resourceName] = schemaResource
	}

	return filteredResources, nil
}

func validateProviderSchema(d *schema.ResourceData) error {

	// Validate org and sys org
	sysOrg := d.Get("sysorg").(string)
	org := d.Get("org").(string)
	if sysOrg == "" && org == "" {
		return fmt.Errorf(`both "org" and "sysorg" properties are empty`)
	}

	return nil
}
