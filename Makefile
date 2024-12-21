## Versions
COROOT_VERSION ?= latest

### Protobuf Tools
BUF_VERSION := v1.40.1
PROTOC_GEN_GO_VERSION := v1.34.2
PROTOC_GEN_GO_GRPC_VERSION := v1.5.1

## Variables
UI_PATH = front
mk_path  := $(abspath $(lastword $(MAKEFILE_LIST)))
root_dir   := $(dir $(mk_path))
proto_dir := $(root_dir)/api/proto
tool_bin := $(root_dir)/bin

## Top Targets
.PHONY: all
all: lint build test

.PHONY: lint
lint: go-lint ui-lint

.PHONY: build
build: npm-build pb-gen go-build

.PHONY: build-fast
build-fast: go-build

.PHONY: test
test: go-test

## Basic Targets
.PHONY: docker
docker: npm-build
	docker build --build-arg VERSION=$(COROOT_VERSION) -t registry.cn-beijing.aliyuncs.com/obser/coroot:$(COROOT_VERSION) .

.PHONY: docker.debug
docker.debug:
	docker build -f Dockerfile.debug -t registry.cn-beijing.aliyuncs.com/obser/coroot:debug .

.PHONY: go-build
go-build:
	go build -mod=readonly -ldflags "-X main.version=$(COROOT_VERSION)" -o coroot .

.PHONY: go-lint
go-lint: go-mod go-vet go-fmt go-imports

.PHONY: go-mod
go-mod:
	go mod tidy

.PHONY: go-vet
go-vet:
	go vet ./...

.PHONY: go-fmt
go-fmt:
	gofmt -w .

.PHONY: go-imports
go-imports:
	go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .

.PHONY: go-test
go-test:
	go test ./...

.PHONY: ui-lint
ui-lint: npm-install npm-lint npm-fmt

.PHONY: npm-install
npm-install:
	cd $(UI_PATH) && npm ci

.PHONY: npm-lint
npm-lint:
	cd $(UI_PATH) && npm run lint

.PHONY: npm-fmt
npm-fmt:
	cd $(UI_PATH) && npm run fmt

.PHONY: npm-build
npm-build:
	cd $(UI_PATH) && npm run build-prod

BUF := $(tool_bin)/buf
$(BUF):
	@echo "Install proto plugins to $(tool_bin)"
	@mkdir -p $(tool_bin)
	@rm -f $(tool_bin)/protoc-gen-go
	@rm -f $(tool_bin)/protoc-gen-go-grpc
	@rm -f $(tool_bin)/buf
	@GOBIN=$(tool_bin) go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@GOBIN=$(tool_bin) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	@GOBIN=$(tool_bin) go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)

.PHONY: pb-gen
pb-gen: $(BUF)
	@PATH=$(tool_bin):$(proto_dir) $(BUF) generate
