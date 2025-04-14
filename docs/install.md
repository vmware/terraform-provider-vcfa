# Installing the Terraform Provider for VMware Cloud Foundation Automation

## Automated Installation (Recommended)

Providers listed on the Terraform Registry can be automatically downloaded when initializing a working directory with `terraform init`.
The Terraform configuration block is used to configure some behaviors of Terraform itself, such as the Terraform version and the required providers and versions.

**Example**: A Terraform configuration block.

```hcl
terraform {
  required_providers {
    vcfa = {
      source = "vmware/vcfa"
    }
  }
  required_version = ">= x.y.z"
}
```

You can use `version` locking and operators to require specific versions of the provider.

**Example**: A Terraform configuration block with the provider versions.

```hcl
terraform {
  required_providers {
    vcfa = {
      source  = "vmware/vcfa"
      version = ">= x.y.z"
    }
  }
  required_version = ">= x.y.z"
}
```

To specify a particular provider version when installing released providers, see the Terraform documentation [on provider versioning][terraform-provider-versioning]

### Verify Terraform Initialization Using the Terraform Registry

To verify the initialization, navigate to the working directory for your Terraform configuration and run `terraform init`.
You should see a message indicating that Terraform has been successfully initialized and has installed the provider from the Terraform Registry.

**Example**: Initialize and Download the Provider.

```console
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Finding vmware/vcfa versions matching ">= x.y.z" ...
- Installing vmware/vcfa x.y.z ...
- Installed vmware/vcfa x.y.z
...

Terraform has been successfully initialized!
```

## Manual Installation

The latest release of the provider can be found in the [releases][releases]. You can download the appropriate version of the provider for your operating system using a command line shell or a browser.

This can be useful in environments that do not allow direct access to the Internet.

### Linux and macOS

1. On a Linux/macOS operating system with Internet access, clone the repository:

    ```shell
    git clone https://github.com/vmware/terraform-provider-vcfa.git
    cd terraform-provider-vcfa/
    ```

2. Run the installation `make` rule:

    ```shell
    make install
    ```

    This command will build the plugin and transfer it to
    `$HOME/.terraform.d/plugins/registry.terraform.io/vmware/vcfa/${VERSION}/${OS}_${ARCH}/terraform-provider-vcfa_v${VERSION}`,
    with a name that includes the version (as taken from the [`VERSION`](../VERSION) file).

    For example, on **macOS**:

    ```console
    $HOME/.terraform.d/
    ├── checkpoint_cache
    ├── checkpoint_signature
    └── plugins
        └── registry.terraform.io
            └── vmware
                └── vcfa
                    └── 0.1.0
                        └── darwin_amd64
                            └── terraform-provider-vcfa_v0.1.0
    ```

    On **Linux**:

    ```console
    ├── checkpoint_cache
    ├── checkpoint_signature
    └── plugins
        └── registry.terraform.io
            └── vmware
                └── vcfa
                    └── 0.1.0
                        └── linux_amd64
                            └── terraform-provider-vcfa_v0.1.0
    ```

### Windows

The following examples use PowerShell on Windows (x64).

1. On a Windows operating system with Internet access, download the plugin using the PowerShell.

   ```powershell
   $RELEASE="x.y.z"
   Invoke-WebRequest -Uri "https://github.com/vmware/terraform-provider-vcfa/releases/download/v${RELEASE}/terraform-provider-vcfa_${RELEASE}_windows_amd64.zip" -OutFile "terraform-provider-vcfa_${RELEASE}_windows_amd64.zip"
   ```

2. Extract the plugin.

   ```powershell
   Expand-Archive terraform-provider-vcfa_${RELEASE}_windows_amd64.zip

   cd terraform-provider-vcfa_${RELEASE}_windows_amd64
   ```

3. Copy the extracted plugin to a target system and move to the Terraform plugins directory.

   > **Note**
   >
   > The directory hierarchy that Terraform uses to precisely determine the source of each provider it finds locally.
   >
   > `$PLUGIN_DIRECTORY/$SOURCEHOSTNAME/$SOURCENAMESPACE/$NAME/$VERSION/$OS_$ARCH/`

   ```powershell
   New-Item $ENV:APPDATA\terraform.d\plugins\local\vmware\vcfa\${RELEASE}\ -Name "windows_amd64" -ItemType "directory"

   Move-Item terraform-provider-vcfa_v${RELEASE}.exe $ENV:APPDATA\terraform.d\plugins\local\vmware\vcfa\${RELEASE}\windows_amd64\terraform-provider-vcfa_v${RELEASE}.exe
   ```

4. Verify the presence of the plugin in the Terraform plugins directory.

   ```powershell
   cd $ENV:APPDATA\terraform.d\plugins\local\vmware\vcfa\${RELEASE}\windows_amd64
   dir
   ```

### Configure the Terraform Configuration Files

A working directory can be initialized with providers that are installed locally on a system by using `terraform init`.
The Terraform configuration block is used to configure some behaviors of Terraform itself, such as the Terraform version and the required providers source and version.

**Example**: A Terraform configuration block.

```hcl
terraform {
  required_providers {
    vcfa = {
      source  = "local/vmware/vcfa"
      version = ">= x.y.z"
    }
  }
  required_version = ">= x.y.z"
}
```

### Verify the Terraform Initialization of a Manually Installed Provider

To verify the initialization, navigate to the working directory for your Terraform configuration and run `terraform init`.
You should see a message indicating that Terraform has been successfully initialized and the installed version of the VMware Cloud Foundation Automation Terraform Provider.

**Example**: Initialize and Use a Manually Installed Provider

```console
$ terraform init

Initializing the backend...

Initializing provider plugins...
- Finding local/vmware/vcfa versions matching ">= x.y.x" ...
- Installing local/vmware/vcfa x.y.x ...
- Installed local/vmware/vcfa x.y.x (unauthenticated)
...

Terraform has been successfully initialized!
```

## Get the Provider Version

To find the provider version, navigate to the working directory of your Terraform configuration and run `terraform version`. You should see a message indicating the provider version.

**Example**: Terraform Provider Version from the Terraform Registry

```console
$ terraform version
Terraform x.y.z
on linux_amd64
+ provider registry.terraform.io/vmware/vcfa x.y.z
```

**Example**: Terraform Provider Version for a Manually Installed Provider

```console
$ terraform version
Terraform x.y.z
on linux_amd64
+ provider local/vmware/vcfa x.y.z
```

[releases]: https://github.com/vmware/terraform-provider-vcfa/releases
[terraform-provider-versioning]: https://developer.hashicorp.com/terraform/language/providers/configuration#version-provider-versions
