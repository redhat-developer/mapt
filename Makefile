VERSION ?= 0.9.0-dev
CONTAINER_MANAGER ?= podman

# Image URL to use all building/pushing image targets
IMG ?= quay.io/redhat-developer/mapt:v${VERSION}
TKN_IMG ?= quay.io/redhat-developer/mapt:v${VERSION}-tkn

# Integrations
# renovate: datasource=github-releases depName=cirruslabs/cirrus-cli
CIRRUS_CLI ?= v0.135.0
# renovate: datasource=github-releases depName=actions/runner
GITHUB_RUNNER ?= 2.317.0

# Go and compilation related variables
GOPATH ?= $(shell go env GOPATH)
BUILD_DIR ?= out
SOURCE_DIRS = cmd pkg
SOURCES := $(shell find . -name "*.go" -not -path "./vendor/*")
# repo
ORG := github.com/redhat-developer
MODULEPATH = $(ORG)/mapt
# Linker flags
VERSION_VARIABLES := -X $(MODULEPATH)/pkg/manager/context.OCI=$(IMG) \
	-X $(MODULEPATH)/pkg/integrations/cirrus.version=$(CIRRUS_CLI) \
	-X $(MODULEPATH)/pkg/integrations/github.runnerVersion=$(GITHUB_RUNNER)
LDFLAGS := $(VERSION_VARIABLES) ${GO_EXTRA_LDFLAGS}
GCFLAGS := all=-N -l
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)



# Tools
TOOLS_DIR := tools
include tools/tools.mk

# Functions
define tkn_update
	rm tkn/*.yaml 
	sed -e 's%<IMAGE>%$(1)%g' -e 's%<VERSION>%$(2)%g' tkn/template/infra-aws-fedora.yaml > tkn/infra-aws-fedora.yaml
	sed -e 's%<IMAGE>%$(1)%g' -e 's%<VERSION>%$(2)%g' tkn/template/infra-aws-mac.yaml > tkn/infra-aws-mac.yaml
	sed -e 's%<IMAGE>%$(1)%g' -e 's%<VERSION>%$(2)%g' tkn/template/infra-aws-rhel.yaml > tkn/infra-aws-rhel.yaml
	sed -e 's%<IMAGE>%$(1)%g' -e 's%<VERSION>%$(2)%g' tkn/template/infra-aws-windows-server.yaml > tkn/infra-aws-windows-server.yaml
	sed -e 's%<IMAGE>%$(1)%g' -e 's%<VERSION>%$(2)%g' tkn/template/infra-azure-aks.yaml > tkn/infra-azure-aks.yaml
	sed -e 's%<IMAGE>%$(1)%g' -e 's%<VERSION>%$(2)%g' tkn/template/infra-azure-rhel.yaml > tkn/infra-azure-rhel.yaml
	sed -e 's%<IMAGE>%$(1)%g' -e 's%<VERSION>%$(2)%g' tkn/template/infra-azure-fedora.yaml > tkn/infra-azure-fedora.yaml
	sed -e 's%<IMAGE>%$(1)%g' -e 's%<VERSION>%$(2)%g' tkn/template/infra-azure-windows-desktop.yaml > tkn/infra-azure-windows-desktop.yaml
endef

# Add default target
.PHONY: default
default: install

# Create and update the vendor directory
.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

.PHONY: check
check: build test lint

# Start of the actual build targets

.PHONY: install
install: $(SOURCES)
	go install -ldflags="$(LDFLAGS)" $(GO_EXTRA_BUILDFLAGS) ./cmd/mapt

$(BUILD_DIR)/mapt: $(SOURCES)
	GOOS="$(GOOS)" GOARCH=$(GOARCH) go build -gcflags="$(GCFLAGS)" -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/mapt $(GO_EXTRA_BUILDFLAGS) ./cmd/mapt
 
.PHONY: build
build: $(BUILD_DIR)/mapt

.PHONY: test
test:
	CGO_ENABLED=1 go test -race --tags build -v -ldflags="$(VERSION_VARIABLES)" ./pkg/... ./cmd/...

.PHONY: clean ## Remove all build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(GOPATH)/bin/mapt

.PHONY: fmt
fmt:
	@gofmt -l -w $(SOURCE_DIRS)

# Run golangci-lint against code
.PHONY: lint 
lint: $(TOOLS_BINDIR)/golangci-lint
	"$(TOOLS_BINDIR)"/golangci-lint run -v --timeout 10m

# Build the container image
.PHONY: oci-build
oci-build: clean oci-build-amd64 oci-build-arm64

# Build for amd64 architecture only
.PHONY: oci-build-amd64
oci-build-amd64: clean
	# Build the container image for amd64
	${CONTAINER_MANAGER} build --platform linux/amd64 --manifest $(IMG)-amd64 -f oci/Containerfile .

# Build for arm64 architecture only
.PHONY: oci-build-arm64
oci-build-arm64: clean
	# Build the container image for arm64
	${CONTAINER_MANAGER} build --platform linux/arm64 --manifest $(IMG)-arm64 -f oci/Containerfile .

# Save images for amd64 architecture only
.PHONY: oci-save-amd64
oci-save-amd64:
	${CONTAINER_MANAGER} save -m -o $(MAPT_SAVE)-amd64.tar $(IMG)-amd64

# Save images for arm64 architecture only
.PHONY: oci-save-arm64
oci-save-arm64:
	${CONTAINER_MANAGER} save -m -o $(MAPT_SAVE)-arm64.tar $(IMG)-arm64


MAPT_SAVE ?= mapt
.PHONY: oci-save 
oci-save: oci-save-amd64 oci-save-arm64

oci-load:
	${CONTAINER_MANAGER} load -i $(MAPT_SAVE)-arm64/$(MAPT_SAVE)-arm64.tar 
	${CONTAINER_MANAGER} load -i $(MAPT_SAVE)-amd64/$(MAPT_SAVE)-amd64.tar 

# Push the docker image
.PHONY: oci-push
oci-push:
	${CONTAINER_MANAGER} push $(IMG)-arm64
	${CONTAINER_MANAGER} push $(IMG)-amd64
	${CONTAINER_MANAGER} manifest create $(IMG)
	${CONTAINER_MANAGER} manifest add $(IMG) docker://$(IMG)-arm64
	${CONTAINER_MANAGER} manifest add $(IMG) docker://$(IMG)-amd64
	${CONTAINER_MANAGER} manifest push --all $(IMG)

# Update tekton with new version
.PHONY: tkn-update
tkn-update:
	$(call tkn_update,$(IMG),$(VERSION))

# Create tekton task bundle
.PHONY: tkn-push
tkn-push: install-out-of-tree-tools
	$(TOOLS_BINDIR)/tkn bundle push $(TKN_IMG) \
		-f tkn/infra-aws-fedora.yaml \
		-f tkn/infra-aws-mac.yaml \
		-f tkn/infra-aws-rhel.yaml \
		-f tkn/infra-aws-windows-server.yaml \
		-f tkn/infra-azure-rhel.yaml \
		-f tkn/infra-azure-windows-desktop.yaml
