# Terraform VMware Cloud Foundation Automation Provider

[![Latest Release](https://img.shields.io/github/v/tag/vmware/terraform-provider-vcfa?label=latest%20release&style=for-the-badge)](https://github.com/vmware/terraform-provider-vcfa/releases/latest) [![License](https://img.shields.io/github/license/vmware/terraform-provider-vcfa.svg?style=for-the-badge)](LICENSE)

The Terraform Provider for VMware Cloud Foundation Automation is a plugin for Terraform that allows you to interact with
VMware Cloud Foundation Automation 9+ by Broadcom.

Learn more:

- Read the provider [documentation][provider-documentation]
- This project is using [go-vcloud-director][go-vcd-sdk] Go SDK for making API calls

## Part of Terraform

- Website: <https://www.terraform.io>
- [Hashicorp Discuss](https://discuss.hashicorp.com/c/terraform-core/27)

<!-- markdownlint-disable no-inline-html -->
<img src="https://www.datocms-assets.com/2885/1629941242-logo-terraform-main.svg" alt="Terraform logo" width="600px">

## Requirements

- [Terraform](https://www.terraform.io/downloads.html)
- [Go](https://golang.org/doc/install) 1.23 (to build the provider plugin)

## Documentation

- Read the official provider documentation [here][provider-documentation]
- Read how to build the provider [here][provider-build]
- Read how to install the provider [here][provider-install]
- Read how to test the provider [here][provider-test]
- You can find configuration examples [here][examples]

## Developing the Provider

When developing `terraform-provider-vcfa` one often needs to modify the underlying `go-vcloud-director` SDK to consume
new methods and types. Go has a convenient [replace](https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive)
directive which can allow you to redirect the import path to your own version of `go-vcloud-director`:

```go
module github.com/vmware/terraform-provider-vcfa
require (
    ...
    github.com/vmware/go-vcloud-director/v3 v3.1.0-alpha.3
)

replace github.com/vmware/go-vcloud-director/v3 v3.1.0-alpha.3 => github.com/my-git-user/go-vcloud-director/v3 v3.1.0-alpha.3    
```

You can also replace pointer to a branch with relative directory:

```go
 module github.com/vmware/terraform-provider-vcfa
 require (
    ...
    github.com/vmware/go-vcloud-director/v3 v3.1.0-alpha.2
)

replace github.com/vmware/go-vcloud-director/v3 v3.1.0-alpha.2 => ../go-vcloud-director
```

See [CODING_GUIDELINES.md][coding-guidelines] for more advice on how to write code for this project.

## Troubleshooting the Provider

Read [TROUBLESHOOTING.md][troubleshooting] to learn how to configure and understand logs, and how to
diagnose common errors.

## License

© Broadcom. All Rights Reserved.
The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.

The Terraform Provider for VMware Cloud Foundation Automation is available under the
[Mozilla Public License, version 2.0][provider-license] license.

[coding-guidelines]: CODING_GUIDELINES.md
[examples]: examples
[go-vcd-sdk]: https://github.com/vmware/go-vcloud-director
[provider-build]: docs/build.md
[provider-documentation]: https://registry.terraform.io/providers/vmware/vcfa/latest/docs
[provider-install]: docs/install.md
[provider-license]: LICENSE
[provider-test]: docs/test.md
[troubleshooting]: TROUBLESHOOTING.md
