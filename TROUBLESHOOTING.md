# Troubleshooting Terraform VMware Cloud Foundation Automation Provider

## Table of contents

- [How to enable logging](#how-to-enable-logging)
- [Common errors](#common-errors)

## How to enable logging

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

## Common errors

### Login does not work

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

### Entity Not Found

```
Error: error getting Organization by Name 'not-exist': [ENF] entity not found: got zero entities by name 'not-exist'
```

While the error looks trivial (the object does not exist), it may also mean that the user configured in the VCFA Provider
does not have enough Rights to fetch the object. When you are sure that the object exists and the error persists, check
that you are logged in with a user with enough Rights.
