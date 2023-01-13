ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
export PATH := "$(ROOT_DIR)/bin:$(PATH)"
SHELL := env PATH=$(PATH) /bin/sh

.PHONY: install
install: visibility/v1/visibility.pb.go
	go install .

.PHONY: test
test: visibility/v1/visibility.pb.go
	go test -v -race -cover ./...

.PHONY: test-coverage
test-coverage: visibility/v1/visibility.pb.go
	go test -v -race -covermode atomic -coverprofile cover.out ./...

visibility/v1/visibility.pb.go: bin/buf bin/protoc-gen-go
	./bin/buf generate

bin/protoc-gen-visibility: visibility/v1/visibility.pb.go
	go build -o "$@" .

bin/buf:
	go build -o "$@" github.com/bufbuild/buf/cmd/buf

bin/protoc-gen-go:
	go build -o "$@" google.golang.org/protobuf/cmd/protoc-gen-go

.PHONY: clean
clean: cleanbin cleangen

.PHONY: cleanbin
cleanbin:
	rm -rf ./bin

.PHONY: cleangen
cleangen:
	rm -rf ./visibility/v1/visibility.pb.go