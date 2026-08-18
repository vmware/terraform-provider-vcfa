// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/vmware/terraform-provider-vcfa/internal/provider/vkscluster"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/vksclusterclass"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/vksclusterkubeconfig"
	"github.com/vmware/terraform-provider-vcfa/internal/provider/vkskubernetesrelease"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &VcfaFrameworkProvider{}
)

type VcfaFrameworkProvider struct {
	// SDKv2Meta is a function which returns the meta struct from the SDKv2 provider
	SDKv2Meta func() any
}

func NewVcfaFrameworkProvider(sdkv2Meta func() any) provider.Provider {
	return &VcfaFrameworkProvider{
		SDKv2Meta: sdkv2Meta,
	}
}

func (p *VcfaFrameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vcfa"
}

func (p *VcfaFrameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	// The SDKv2 provider (vcfa.Provider()) declares "org" and "url" as Required with an
	// EnvDefaultFunc fallback. terraform-plugin-sdk's schema-to-protocol conversion flips such
	// attributes to Optional whenever the environment variable is set (see
	// helper/schema.Schema.coreConfigSchemaAttribute). Both providers are muxed together and
	// their schemas must be byte-identical, so this mirrors that same env-dependent behavior here.
	orgAttribute := schema.StringAttribute{
		Description: "The VCFA Org for API operations",
	}
	if os.Getenv("VCFA_ORG") != "" {
		orgAttribute.Optional = true
	} else {
		orgAttribute.Required = true
	}

	urlAttribute := schema.StringAttribute{
		Description: "The VCFA url for VCFA API operations.",
	}
	if os.Getenv("VCFA_URL") != "" {
		urlAttribute.Optional = true
	} else {
		urlAttribute.Required = true
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Optional:    true,
				Description: "The user name for VCFA API operations.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Description: "The user password for VCFA API operations.",
			},
			"auth_type": schema.StringAttribute{
				Optional:    true,
				Description: "'integrated', 'token', 'api_token', 'api_token_file' and 'service_account_token_file' are supported. 'integrated' is default.",
				Validators: []validator.String{
					stringvalidator.OneOf("integrated", "token", "api_token", "api_token_file", "service_account_token_file"),
				},
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Description: "The token used instead of username/password for VCFA API operations.",
			},
			"api_token": schema.StringAttribute{
				Optional:    true,
				Description: "The API token used instead of username/password for VCFA API operations",
			},
			"api_token_file": schema.StringAttribute{
				Optional:    true,
				Description: "The API token file instead of username/password for VCFA API operations",
			},
			"allow_api_token_file": schema.BoolAttribute{
				Optional:    true,
				Description: "Set this to true if you understand the security risks of using API token files and would like to suppress the warnings",
			},
			"service_account_token_file": schema.StringAttribute{
				Optional:    true,
				Description: "The Service Account API token file instead of username/password for VCFA API operations. (Requires VCFA 9.0+)",
			},
			"allow_service_account_token_file": schema.BoolAttribute{
				Optional:    true,
				Description: "Set this to true if you understand the security risks of using Service Account token files and would like to suppress the warnings",
			},
			"sysorg": schema.StringAttribute{
				Optional:    true,
				Description: "The VCFA Org for user authentication",
			},
			"org": orgAttribute,
			"url": urlAttribute,
			"allow_unverified_ssl": schema.BoolAttribute{
				Optional:    true,
				Description: "If set, VCFAClient will permit unverifiable SSL certificates.",
			},
			"logging": schema.BoolAttribute{
				Optional:    true,
				Description: "If set, it will enable logging of API requests and responses",
			},
			"logging_file": schema.StringAttribute{
				Optional:    true,
				Description: "Defines the full name of the logging file for API calls (requires 'logging')",
			},
			"import_separator": schema.StringAttribute{
				Optional:    true,
				Description: "Defines the import separation string to be used with 'terraform import'",
			},
		},
	}
}

func (p *VcfaFrameworkProvider) Configure(_ context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Re-use the SDKv2 configuration until all datasources and resources have been migrated to the framework provider
	resp.ResourceData = p.SDKv2Meta
	resp.DataSourceData = p.SDKv2Meta
}

// Resources returns the list of framework-based resources.
func (p *VcfaFrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		vkscluster.NewVcfaVksClusterResource,
	}
}

// DataSources returns the list of framework-based data sources.
func (p *VcfaFrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		vksclusterclass.NewVcfaVksClusterClassDataSource,
		vkscluster.NewVcfaVksClusterDataSource,
		vkskubernetesrelease.NewVcfaVksKubernetesReleaseDataSource,
		vksclusterkubeconfig.NewVcfaVksClusterKubeconfigDataSource,
	}
}
