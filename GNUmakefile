TEST?=$$(go list ./... )
GOFMT_FILES?=$$(find . -name '*.go' )
WEBSITE_REPO=github.com/hashicorp/terraform-website
GIT_DESCRIBE=$(shell git describe --tags)
PKG_NAME=vcfa

default: build

# runs a Go format check
fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

# builds the plugin injecting output of `git describe` to BuildVersion variable
build: fmtcheck
	go install -ldflags="-X 'github.com/vmware/terraform-provider-vcfa/v1/vcfa.BuildVersion=$(GIT_DESCRIBE)'"

# builds and deploys the plugin
install: build
	@sh -c "'$(CURDIR)/scripts/install-plugin.sh'"

fmt:
	gofmt -s -w $(GOFMT_FILES)