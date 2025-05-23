# Terraform VMware Cloud Foundation Automation Provider

[![Latest Release](https://img.shields.io/github/v/tag/vmware/terraform-provider-vcfa?label=latest%20release&style=for-the-badge)](https://github.com/vmware/terraform-provider-vcfa/releases/latest) [![License](https://img.shields.io/github/license/vmware/terraform-provider-vcfa.svg?style=for-the-badge)](LICENSE)

The Terraform Provider for VMware Cloud Foundation Automation is a plugin for Terraform that allows you to interact with
VMware Cloud Foundation Automation 9+ by Broadcom.

Learn more:

- Read the provider [documentation][provider-documentation]
- Have a look at the [configuration examples][examples]
- This project is using [go-vcloud-director][go-vcd-sdk] Go SDK for making API calls

## Part of Terraform

- Website: <https://www.terraform.io>
- [Hashicorp Discuss](https://discuss.hashicorp.com/c/terraform-core/27)

<!-- markdownlint-disable no-inline-html -->
<img src="https://www.datocms-assets.com/2885/1629941242-logo-terraform-main.svg" alt="Terraform logo" width="600px">

## Requirements

- [Terraform](https://www.terraform.io/downloads.html)
- [Go](https://golang.org/doc/install) 1.24 (to build the provider plugin)

## Building the Provider

**Note:** You *only* need to build the provider plugin if you want to *develop* it. Refer to
[documentation][provider-documentation] for using it. Terraform will
automatically download officially released binaries of this provider plugin on the first run of `terraform init`
command.

```shell
cd ~/mydir
git clone https://github.com/vmware/terraform-provider-vcfa.git
cd terraform-provider-vcfa/
make build
```

## Installing the Provider from source code

**Note:** You *only* need to install the provider from source code if you want to test unreleased features or to develop it. Refer to
[documentation](https://registry.terraform.io/providers/vmware/vcfa/latest/docs) for using it in a standard way. Terraform will
automatically download officially released binaries of this provider plugin on the first run of `terraform init`
command.

```shell
cd ~/mydir
git clone https://github.com/vmware/terraform-provider-vcfa.git
cd terraform-provider-vcfa/
make install
```

This command will build the plugin and transfer it to
`$HOME/.terraform.d/plugins/registry.terraform.io/vmware/vcfa/${VERSION}/${OS}_${ARCH}/terraform-provider-vcfa_v${VERSION}`,
with a name that includes the version (as taken from the `./VERSION` file).

For example, on **macOS**:

```console
$HOME/.terraform.d/
├── checkpoint_cache
├── checkpoint_signature
└── plugins
    └── registry.terraform.io
        └── vmware
            └── vcfa
                └── 1.0.0
                    └── darwin_amd64
                        └── terraform-provider-vcfa_v1.0.0
```

On **Linux**:

```console
├── checkpoint_cache
├── checkpoint_signature
└── plugins
    └── registry.terraform.io
        └── vmware
            └── vcfa
                └── 1.0.0
                    └── linux_amd64
                        └── terraform-provider-vcfa_v1.0.0
```

Once you have installed the plugin as mentioned above, you can simply create a new `config.tf` as defined in [the manual](https://www.terraform.io/docs/providers/vcfa/index.html) and run

```sh
terraform init
terraform plan
terraform apply
```

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
[provider-documentation]: https://registry.terraform.io/providers/vmware/vcfa/latest/docs
[provider-license]: LICENSE
[troubleshooting]: TROUBLESHOOTING.md
