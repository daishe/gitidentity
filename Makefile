
BIN_DIR = $(shell pwd)/bin

.PHONY: all
all: test

dependencies: go.mod go.sum
	go mod download
	touch dependencies

.PHONY: test
test: dependencies
	go test -timeout 5m ./...

.PHONY: tools-all
tools-all: bin/buf bin/protoc-gen-go

.PHONY: tools-clean
tools-clean:
	rm -rf bin tools/dependencies

tools/dependencies: tools/go.mod tools/go.sum tools/tools.go
	cd tools && go mod download
	touch tools/dependencies

bin/buf: tools/dependencies
	mkdir -p bin
	cd tools && go build -o ../bin/buf github.com/bufbuild/buf/cmd/buf

bin/protoc-gen-go: tools/dependencies
	mkdir -p bin
	cd tools && go build -o ../bin/protoc-gen-go google.golang.org/protobuf/cmd/protoc-gen-go

.PHONY: proto
proto: bin/buf bin/protoc-gen-go
	PATH="$(BIN_DIR):$$PATH" go generate -tags proto ./...
