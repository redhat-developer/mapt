VERSION ?= 0.6.8
CONTAINER_MANAGER ?= podman
# Image URL to use all building/pushing image targets
IMG ?= quay.io/redhat-developer/mapt:v${VERSION}
TKN_IMG ?= quay.io/redhat-developer/mapt:v${VERSION}-tkn

# Go and compilation related variables
GOPATH ?= $(shell go env GOPATH)
BUILD_DIR ?= out
SOURCE_DIRS = cmd pkg
SOURCES := $(shell find . -name "*.go" -not -path "./vendor/*")
# https://golang.org/cmd/link/
# LDFLAGS := $(VERSION_VARIABLES) -extldflags='-static' ${GO_EXTRA_LDFLAGS}
LDFLAGS := $(VERSION_VARIABLES) ${GO_EXTRA_LDFLAGS}
GCFLAGS := all=-N -l
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

TOOLS_DIR := tools
include tools/tools.mk

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
oci-build: clean
	${CONTAINER_MANAGER} build -t ${IMG} -f oci/Containerfile .

# Push the docker image
.PHONY: oci-push
oci-push:
	${CONTAINER_MANAGER} push ${IMG}
	
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
