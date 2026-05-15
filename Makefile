.PHONY: help setup install-deps generate lint fix spell check test example clean

include .github/versions.env

export PATH := $(HOME)/go/bin:$(PATH)

help:
	@echo "Available targets:"
	@echo "  setup        - Install dependencies, tools, and set up git hooks"
	@echo "  install-deps - Install build/test deps (libpcap-dev + stringer)"
	@echo "  generate     - Generate code using go generate"
	@echo "  fix          - Run linters and fix issues automatically"
	@echo "  lint         - Run linters to check for issues"
	@echo "  spell        - Check for spelling errors in the codebase"
	@echo "  check        - Run both lint and spell checks"
	@echo "  test         - Run tests with generated code"
	@echo "  example      - Run the example program (example.go)"
	@echo "  clean        - Remove generated files and canary"

install-deps:
	sudo apt-get update && sudo apt-get install -y libpcap-dev
	go install golang.org/x/tools/cmd/stringer@$(STRINGER_VERSION)

setup: install-deps
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	git config core.hooksPath .githooks

generate: .generated-canary

.generated-canary: $(wildcard *.go)
	@GOPATH=$(shell go env GOPATH)
	go generate ./...
	@touch $@

fix: generate
	codespell . -w
	golangci-lint run ./... --fix

lint: generate
	golangci-lint run ./...

spell:
	codespell .

check: lint spell

test: generate
	go test ./dnp3 -v -args -pcaps=opendnp3_test1.pcap -print-string -print-json

example: generate
	go run .

clean:
	rm -f **/*_string.go .generated-canary
