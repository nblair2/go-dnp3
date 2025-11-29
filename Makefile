.PHONY: setup generate lint fix test clean

export PATH := $(HOME)/go/bin:$(PATH)

setup:
	git config core.hooksPath .githooks
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/stringer@latest

generate: .generated-stamp

.generated-stamp: $(wildcard *.go)
	GOPATH=$(shell go env GOPATH)
	go generate ./...
	touch $@

fix: generate
	golangci-lint run ./... --fix

lint: generate
	golangci-lint run ./...

test: generate
	go test ./dnp3 -v -args -pcaps=opendnp3_test1.pcap

clean:
	rm -f dnp3/*_string.go
	rm -f .generated-stamp