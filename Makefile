.PHONY: setup generate lint fix test clean

export PATH := $(HOME)/go/bin:$(PATH)

setup:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.10.1
	go install golang.org/x/tools/cmd/stringer@latest
	sudo apt-get install -y libpcap-dev
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
	go test ./dnp3 -v -args -pcaps=opendnp3_test1.pcap

clean:
	rm -f **/*_string.go .generated-canary