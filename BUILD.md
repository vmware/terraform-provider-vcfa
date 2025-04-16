# Building the Terraform Provider for VMware Cloud Foundation Automation

The instructions outlined below are specific to macOS and Linux only.

If you wish to work on the provider, you'll first need [Go][golang-install] installed on your
machine. Check the [requirements][requirements] before proceeding.

1. Clone the repository to: `$GOPATH/src/github.com/vmware/terraform-provider-vcfa`

   ```shell
   mkdir -p $GOPATH/src/github.com/vmware
   cd $GOPATH/src/github.com/vmware
   git clone git@github.com:vmware/terraform-provider-vcfa.git
   ```

2. Enter the provider directory to build the provider.

   ```shell
   cd $GOPATH/src/github.com/vmware/terraform-provider-vcfa
   make build
   ```

3. Add the following to your `~/.terraformrc`:

   ```hcl
   provider_installation {
     dev_overrides {
       "vmware/vcfa" = "/Users/rainpole/go/bin"
     }

     direct {}
   }
   ```

   Where `/Users/rainpole/go/bin` is your `$GOPATH/bin` path.

[golang-install]: https://golang.org/doc/install
[requirements]: https://github.com/vmware/terraform-provider-vcfa#requirements
