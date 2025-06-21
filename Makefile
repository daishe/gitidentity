
BIN_DIR = $(shell pwd)/bin
GIT_TAG ?= "0.0.0-dev"
GIT_COMMIT ?= "?"

GO_SOURCE := $(shell find . -type f -name '*.go')

define goBuild
GOOS=$(1) GOARCH=$(2) go build -ldflags="-X 'main.Version=$(GIT_TAG)' -X 'main.Commit=$(GIT_COMMIT)'" -o $(3) -v .
endef

.PHONY: all
all: lint test build

.PHONY: clean
clean: dist-clean tools-clean
	rm -rf dependencies

dependencies: go.mod go.sum
	go mod download
	touch dependencies

.PHONY: lint
lint: bin/golangci-lint
	bin/golangci-lint config verify
	bin/golangci-lint run

.PHONY: test
test: dependencies
	go test -timeout 5m ./...

.PHONY: dist-all
dist-all: build dist/gitidentity-linux-amd64 dist/gitidentity-linux-arm64 dist/gitidentity-windows-amd64.exe dist/gitidentity-windows-arm64.exe dist/gitidentity-darwin-amd64 dist/gitidentity-darwin-arm64

.PHONY: dist-clean
dist-clean:
	rm -rf dist

.PHONY: build
build: $(GO_SOURCE) dependencies
	$(call goBuild,,,dist/gitidentity)

dist/gitidentity-linux-amd64: $(GO_SOURCE) dependencies
	$(call goBuild,linux,amd64,dist/gitidentity-linux-amd64)

dist/gitidentity-linux-arm64: $(GO_SOURCE) dependencies
	$(call goBuild,linux,arm64,dist/gitidentity-linux-arm64)

dist/gitidentity-windows-amd64.exe: $(GO_SOURCE) dependencies
	$(call goBuild,windows,amd64,dist/gitidentity-windows-amd64.exe)

dist/gitidentity-windows-arm64.exe: $(GO_SOURCE) dependencies
	$(call goBuild,windows,arm64,dist/gitidentity-windows-arm64.exe)

dist/gitidentity-darwin-amd64: $(GO_SOURCE) dependencies
	$(call goBuild,darwin,amd64,dist/gitidentity-darwin-amd64)

dist/gitidentity-darwin-arm64: $(GO_SOURCE) dependencies
	$(call goBuild,darwin,arm64,dist/gitidentity-darwin-arm64)

.PHONY: dist/gitidentity-windows-amd64
dist/gitidentity-windows-amd64: dist/gitidentity-windows-amd64.exe

.PHONY: dist/gitidentity-windows-arm64
dist/gitidentity-windows-arm64: dist/gitidentity-windows-arm64.exe

.PHONY: tools-all
tools-all: bin/buf bin/golangci-lint bin/protoc-gen-go

.PHONY: tools-clean
tools-clean:
	rm -rf bin tools/dependencies

tools/dependencies: tools/go.mod tools/go.sum tools/tools.go
	cd tools && go mod download
	touch tools/dependencies

bin/buf: tools/dependencies
	cd tools && go build -o ../bin/buf github.com/bufbuild/buf/cmd/buf

bin/golangci-lint: tools/dependencies
	cd tools && go build -o ../bin/golangci-lint github.com/golangci/golangci-lint/v2/cmd/golangci-lint

bin/protoc-gen-go: tools/dependencies
	cd tools && go build -o ../bin/protoc-gen-go google.golang.org/protobuf/cmd/protoc-gen-go

.PHONY: proto
proto: bin/buf bin/protoc-gen-go
	PATH="$(BIN_DIR):$$PATH" go generate -tags proto ./...
