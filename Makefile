## Versions
COROOT_VERSION ?= latest

## Variables
UI_PATH = front

## Top Targets
.PHONY: all
all: lint build test

.PHONY: lint
lint: go-lint ui-lint

.PHONY: build
build: npm-build go-build

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
