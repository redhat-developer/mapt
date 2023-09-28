VERSION ?= 0.0.6-dev
CONTAINER_MANAGER ?= podman
# Image URL to use all building/pushing image targets
IMG ?= quay.io/rhqp/qenvs:v${VERSION}
TKN_IMG ?= quay.io/rhqp/qenvs-tkn:v${VERSION}

# Go and compilation related variables
GO_VERSION ?= 1.19
GOPATH ?= $(shell go env GOPATH)
BUILD_DIR ?= out
SOURCE_DIRS = cmd pkg test
# https://golang.org/cmd/link/
# LDFLAGS := $(VERSION_VARIABLES) -extldflags='-static' ${GO_EXTRA_LDFLAGS}
LDFLAGS := $(VERSION_VARIABLES) ${GO_EXTRA_LDFLAGS}
GCFLAGS := all=-N -l 

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
	go install -ldflags="$(LDFLAGS)" $(GO_EXTRA_BUILDFLAGS) ./cmd

$(BUILD_DIR)/qenvs: $(SOURCES)
	go build -x -o $(BUILD_DIR)/qenvs $(GO_EXTRA_BUILDFLAGS) ./cmd
 
.PHONY: build 
build: $(BUILD_DIR)/qenvs

.PHONY: test
test:
	go test -race --tags build -v ./pkg/... ./cmd/...

.PHONY: clean ## Remove all build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(GOPATH)/bin/qenvs

.PHONY: fmt
fmt:
	@gofmt -l -w $(SOURCE_DIRS)

.PHONY: lint
lint: 
	${CONTAINER_MANAGER} run -it --rm -v ${PWD}:/workspace:z --workdir=/workspace quay.io/app-sre/golangci-lint:v1.53.2-alpine \
		run -v --timeout 10m

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
		-f tkn/infra-management-aws.yaml \
		-f tkn/infra-management-azure.yaml