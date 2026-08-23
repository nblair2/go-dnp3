.PHONY: help setup install-deps generate lint spell corpus test example clean

include .github/versions.env

export PATH := $(HOME)/.local/bin:$(HOME)/go/bin:$(PATH)

help:
	@echo "Available targets:"
	@echo "  setup        - Install dependencies, tools, and set up git hooks"
	@echo "  install-deps - Install build/test deps (libpcap-dev + stringer)"
	@echo "  generate     - Generate code using go generate"
	@echo "  lint         - Run all prek hooks (lint, spellcheck, format) on all files"
	@echo "  spell        - Check for spelling errors in the codebase"
	@echo "  corpus       - Fetch the pcap test corpus (see test/testdata/corpus.txt)"
	@echo "  test         - Run tests with generated code"
	@echo "  example      - Run the example program (example.go)"
	@echo "  clean        - Remove generated files and canary"

install-deps:
	sudo apt-get update && sudo apt-get install -y libpcap-dev
	go install golang.org/x/tools/cmd/stringer@$(STRINGER_VERSION)

setup: install-deps
	curl --proto '=https' --tlsv1.2 -LsSf https://github.com/j178/prek/releases/download/v$(PREK_VERSION)/prek-installer.sh | sh
	prek install

generate: .generated-canary

.generated-canary: $(wildcard *.go dnp3/*.go test/*.go)
	go generate ./...
	@touch $@

lint: generate
	prek run --all-files

spell:
	prek run codespell --all-files

corpus:
	@mkdir -p test/testdata/corpus
	@while read -r name sha url; do \
		case "$$name" in ''|\#*) continue;; esac; \
		f="test/testdata/corpus/$$name"; \
		echo "$$sha  $$f" | sha256sum --check --quiet --status - 2>/dev/null && continue; \
		echo "fetching $$name"; \
		curl -sSfL "$$url" -o "$$f" || exit 1; \
		echo "$$sha  $$f" | sha256sum --check --quiet - || { rm -f "$$f"; exit 1; }; \
	done < test/testdata/corpus.txt

test: generate
	go test -v ./...

example: generate
	go run .

clean:
	rm -f **/*_string.go .generated-canary
