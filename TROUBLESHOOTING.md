# Troubleshooting Terraform VMware Cloud Foundation Automation Provider

## Table of contents

- [About Terraform](#about-terraform)
  - [State management](#state-management)
  - [Terraform native logging](#terraform-native-logging)
  - [Provider version constraints](#provider-version-constraints)
  - [Terraform architecture and responsibility boundaries](#terraform-architecture-and-responsibility-boundaries)
    - [Terraform Core responsibilities](#terraform-core-responsibilities)
    - [Terraform Provider plugin responsibilities](#terraform-provider-plugin-responsibilities)
    - [Provider version constraints](#provider-version-constraints)
- [Terraform provider VCFA](#about-terraform)
  - [How to enable logging](#how-to-enable-logging)
  - [Common errors](#common-errors)
    - [Login does not work](#login-does-not-work)
    - [Entity Not Found error](#entity-not-found-error)
    - [vcfa_content_library_item creation never finishes (it's stuck)](#vcfa_content_library_item-creation-never-finishes-its-stuck)
    - [External entity connection (vCenter server, NSX Manager) returns certificate error](#external-entity-connection-vcenter-server-nsx-manager-returns-certificate-error)

## About Terraform

This section briefly touches on *Terraform* functionality that is provided within
[`terraform`][terraform] binary by [Hashicorp][hashicorp].

*Note:* For the VCFA provider specific logging see [How to enable logging](#how-to-enable-logging).

### State management

One of core features of Terraform is storing [infrastructure state][terraform-state]. The state can
either be stored [locally, or remotely][state-storage]. The state contains mapping of existing
infrastructure and it serves multiple purposes: drift detection, performance, additional metadata
tracking.

Local state is usually stored in `terraform.tfstate` file, while remote state storage specifics depend on
[state backend][state-backend] being used. `terraform state` subcommand provides CLI capabilities
for state management.

*Note:* Terraform state may contain [sensitive data][state-sensitive-data].

### Terraform native logging

Terraform itself, independently of the provider plugin being used, has some debugging options that
can enable additional logging. They are documented in [Terraform debugging
page][terraform-debugging].

In a nutshell, these are the main variables, but the [debugging page][terraform-debugging] has more details:

```
TF_LOG=trace
TF_LOG_PATH=tf.log
```

### Provider version constraints

Terraform providers are installed automatically by using `terraform init` command. The plugin is
fetched based on [`required_providers` section within `terraform` configuration
block][provider-versioning]. An important part of the configuration block is [version
constraint][provider-version-constraints] that defines which version of the plugin should be
installed.

Terraform provider plugins follow [semantic versioning][semver] pattern and
[MAJOR.MINOR.PATCH][provider-semver].

```
terraform {
  required_providers {
    vcfa = {
      source = "vmware/vcfa"
      version = "~> 1.0.0" # pins major and minor versions, but will accept new patch versions (e.g. 1.0.1)
    }
  }
}

provider "vcfa" {
  # Configuration options
}
```

### Terraform architecture and responsibility boundaries

Terraform consists of three main components:
* Core (`terraform` binary)
* Plugins (called providers) e.g. Terraform Provider VCFA
* Upstream APIs (some Go SDK for the platform)

#### Terraform Core responsibilities

The **Core** that is best identified by `terraform` binary in the consumer's system, is developed by
Hashicorp. It provides:

* HCL (Hashicorp Configuration Language) engine and syntax
* Schema diff that is being shown when performing operations
* Dependency graph for the entities
* All the tooling, including:
  * Terraform enterprise
  * Workspaces
  * CDK
  * Registry and documentation format
* Some fields in any resource:
  * `lifecycle`
  * `provisioner`

#### Terraform Provider plugin responsibilities

Plugin responsibilities, **Terraform provider VCFA** in this case:

* Communication with the platform API (VCFA)
* Implementation of provider plugin using Terraform plugin SDK
  * Architecture (entity schema and granularity of resources/data sources)
  * Implementation of Create, Read, Update, Delete and Import (CRUD+I for each resource
  * Implementation of Read (R) for each data source
* Maintenance of [Go SDK][go-vcloud-director]
* Releasing the provider to [Terraform registry][registry-vcfa]
* Documentation (published to registry)
* Maintenance and support
* Respecting SemVer for the releases

## Terraform provider VCFA

### How to enable logging

The VCFA provider can be configured to write logs into a specific file located in a customised path:

```hcl
provider "vcfa" {
  # ... omitted arguments
  logging      = true              # Enables logging
  logging_file = "my-log-file.log" # Specifies in which file to write logs
}
```

When logs are enabled, all the requests to VCFA and its responses are written into the specified file. Each log entry registers
the following:

- Method call sequence (who called who) to perform the request. This helps to trace the methods that triggered the request to VCFA. 
  - If the method is from `vcfa` package, the call is present in the VCFA Provider source code.
  - If the method is from `govcd` package, the source code is in [go-vcloud-director SDK](https://github.com/vmware/go-vcloud-director).
- HTTP method, request URL and query parameters that were sent to VCFA.
- HTTP request headers. Sensitive information is obfuscated by default. To enable logging of passwords, tokens and other sensitive information,
  users can enable the `GOVCD_LOG_PASSWORDS=1` environment variable right before running Terraform operations.
  - The `X-Vmware-Vcloud-Client-Request-Id` header can be used to improve readability of the requests and responses in big log files, as it tracks
    the request number with a timestamp.
- Method call sequence (who called who) to read the response. This helps to trace the methods that parsed and unmarshalled the response.
- HTTP response headers. The `X-Vmware-Vcloud-Request-Id` header is modified with the VCFA request ID.
- The response itself in raw plain text (JSON/XML).

Example taken from a real log:

```log
2025/03/12 14:14:00 --------------------------------------------------------------------------------
2025/03/12 14:14:00 Request caller: vcfa.getRegionHcl-->govcd.(*VCDClient).GetRegionByName-->govcd.(*VCDClient).GetAllRegions-->govcd.getAllOuterEntities[...]-->govcd.getAllInnerEntities[...]-->govcd.(*Client).OpenApiGetAllItems-->govcd.(*Client).openApiGetAllPages-->govcd.(*Client).newOpenApiRequest
2025/03/12 14:14:00 GET https://my-vcfa-url.com/cloudapi/vcf/regions/?filter=name%3D%3Dtest-region&pageSize=128
2025/03/12 14:14:00 --------------------------------------------------------------------------------
2025/03/12 14:14:00 Req header:
2025/03/12 14:14:00 	User-Agent: [terraform-provider-vcfa/test (darwin/amd64; isProvider:true)]
2025/03/12 14:14:00 	X-Vmware-Vcloud-Client-Request-Id: [41-2025-03-12-14-14-00-054-]
2025/03/12 14:14:00 	X-Vmware-Vcloud-Access-Token: [********]
2025/03/12 14:14:00 	Authorization: [********]
2025/03/12 14:14:00 	X-Vmware-Vcloud-Token-Type: [Bearer]
2025/03/12 14:14:00 	Accept: [application/json;version=40.0]
2025/03/12 14:14:00 	Content-Type: [application/json]
2025/03/12 14:14:00 ################################################################################
2025/03/12 14:14:00 Response caller vcfa.getRegionHcl-->govcd.(*VCDClient).GetRegionByName-->govcd.(*VCDClient).GetAllRegions-->govcd.getAllOuterEntities[...]-->govcd.getAllInnerEntities[...]-->govcd.(*Client).OpenApiGetAllItems-->govcd.(*Client).openApiGetAllPages-->govcd.decodeBody
2025/03/12 14:14:00 Response status 200 OK
2025/03/12 14:14:00 ################################################################################
2025/03/12 14:14:00 Response header:
2025/03/12 14:14:00 	Content-Type: [application/json;version=40.0]
2025/03/12 14:14:00 	X-Vmware-Vcloud-Request-Id: [41-2025-03-12-14-14-00-054--772d833f-bb60-49c9-8a78-5552c87448de]
2025/03/12 14:14:00 	Link: [<https://my-vcfa-url.com/cloudapi/vcf/regions/monitoringToken>;rel="add";type="application/json";model="MonitoringToken" <https://my-vcfa-url.com/cloudapi/vcf/regions>;rel="add";type="application/json";model="Region" <https://my-vcfa-url.com/cloudapi/>;rel="up";type="*/*"]
2025/03/12 14:14:00 	Date: [Wed, 12 Mar 2025 13:14:00 GMT]
2025/03/12 14:14:00 	X-Frame-Options: [SAMEORIGIN]
2025/03/12 14:14:00 	Cache-Control: [no-store, must-revalidate]
2025/03/12 14:14:00 	Vary: [Accept-Encoding]
2025/03/12 14:14:00 	Content-Location: [https://my-vcfa-url.com/cloudapi/vcf/regions/]
2025/03/12 14:14:00 	Strict-Transport-Security: [max-age=63072000; includeSubDomains; preload]
2025/03/12 14:14:00 	X-Vmware-Vcloud-Ceip-Id: [d9dba2ea-6a9e-4424-b3c8-06de619870b8]
2025/03/12 14:14:00 	X-Vmware-Vcloud-Request-Execution-Time: [34]
2025/03/12 14:14:00 Response text: [112]
{
  "resultTotal": 0,
  "pageCount": 0,
  "page": 1,
  "pageSize": 128,
  "associations": null,
  "values": []
}
```

In this log, the reader would see that there was an attempt to fetch a Region by using its name, `test-region`, but
VCFA returned no results, meaning that this Region does not exist.

Logging can also be configured with environment variables `GOVCD_LOG=1` to enable logs and `GOVCD_LOG_FILE=my-log-file.log` to specify
the file path.

To disable HTTP **request** logging, one can set the environment variable `GOVCD_LOG_SKIP_HTTP_REQ=1`.
To disable HTTP **response** logging, one can set the environment variable `GOVCD_LOG_SKIP_HTTP_RESP=1`.

Note that the logs of Terraform itself can also be customised, please take a look [here](https://developer.hashicorp.com/terraform/internals/debugging)
if you need to troubleshoot unexpected behaviors from Terraform.

### Common errors

#### Login does not work

```
Error: something went wrong during authentication: error finding LoginUrl: could not find valid version for login: could not retrieve supported versions: error fetching versions: Get "https://my-vcfa-url.com/api/versions": dial tcp: lookup my-vcfa-url.com: no such host
```

This error is thrown when the VCFA URL is incorrect or can't be reached. Verify that the `url` argument in the VCFA provider
configuration is correct and try again.

```
Error: something went wrong during authentication: error authorizing: received response HTTP 401 (Unauthorized). Please check if your credentials are valid
```

Verify that the `user` and `password` arguments in the VCFA provider configuration are correct and try again.
Also, verify that `org` is correctly set and the provided user has enough permissions to access that Organization.

If you are using an API Token, verify that it is valid and not expired. Set a valid token in `api_token`.

If you are using an API Token from `vcfa_api_token`(https://github.com/vmware/terraform-provider-vcfa/blob/main/providers/vmware/vcfa/latest/docs/resources/api_token), verify that it is valid, not expired and the syntax is correct. Set a valid token file in `api_token_file`.

#### Entity Not Found error

```
Error: error getting Organization by Name 'not-exist': [ENF] entity not found: got zero entities by name 'not-exist'
```

While the error looks trivial (the object does not exist), it may also mean that the user configured in the VCFA Provider
does not have enough Rights to fetch the object. When you are sure that the object exists and the error persists, check
that you are logged in with a user with enough Rights.

#### vcfa_content_library_item creation never finishes (it's stuck)

If `vcfa_content_library_item` creation never ends, it could be that `quarantine_content_library_items` setting is enabled
for the Organization in which the Content Library Item is being uploaded
(see [`vcfa_org_settings`](https://registry.terraform.io/providers/vmware/vcfa/latest/docs/resources/org_settings)).

The property `quarantine_content_library_items` can be set by Providers when configuring the Organization, and it makes uploads
to be blocked until they are approved (or rejected) by an authorized user. If you need to upload Content Library Items with Terraform, you
may need to ask that the Organization has `quarantine_content_library_items` disabled, or that someone approves or denies the
upload when the Terraform script is applied.

#### External entity connection (vCenter server, NSX Manager) returns certificate error

Sample error:

```sh
vcfa_vcenter.demo: Creating...
╷
│ Error: error creating entity vCenter Server. Storing tainted resources ID urn:vcloud:vimserver:39e82b7d-8ca8-4dbf-b13e-16cbc3bf73f6. Task error: task did not complete successfully:  [400:BAD_REQUEST] - [ 17-2025-03-24-16-22-00-804--dbab3b29-dd00-4fbb-9a6c-1953e4b88f1c ] Failed to connect to the vCenter due to org.bouncycastle.tls.TlsFatalAlert: certificate_unknown(46).
│ 
│   with vcfa_vcenter.demo,
│   on main.tf line 21, in resource "vcfa_vcenter" "demo":
│   21: resource "vcfa_vcenter" "demo" {
```

The entities that handle external connections (vCenter server, NSX Manager, LDAP Server
configurations, etc.) must have a valid trusted certificate configured for a destination entity.
Make sure that the certificate of that entity is trusted. Resources have `auto_trust_certificate`
boolean field that can be used to leveraged trusting the certificate automatically.

[terraform]: https://www.terraform.io/
[terraform-state]: https://developer.hashicorp.com/terraform/language/state
[hashicorp]: https://www.hashicorp.com/
[terraform-provider-docs]: https://developer.hashicorp.com/terraform/language/providers
[terraform-debugging]: https://developer.hashicorp.com/terraform/internals/debugging
[state-storage]: https://developer.hashicorp.com/terraform/language/state/remote
[state-sensitive-data]: https://developer.hashicorp.com/terraform/language/state/sensitive-data
[state-backend]: https://developer.hashicorp.com/terraform/language/backend
[provider-versioning]: https://developer.hashicorp.com/terraform/tutorials/configuration-language/provider-versioning
[provider-version-constraints]:https://developer.hashicorp.com/terraform/tutorials/configuration-language/versions#terraform-version-constraints
[provider-semver]: https://developer.hashicorp.com/terraform/plugin/best-practices/versioning
[semver]: https://semver.org/
[go-vcloud-director]: https://github.com/vmware/go-vcloud-director
[registry-vcfa]: http://registry.terraform.io/providers/vmware/vcfa
