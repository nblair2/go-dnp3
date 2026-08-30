.PHONY: help setup install-deps generate lint spell corpus test example bump-major check-major clean

include .github/versions.env

export PATH := $(HOME)/.local/bin:$(HOME)/go/bin:$(PATH)

# The module major version is repeated in go.mod, every import path, and the
# README doc link, so bump and check them as a set.
MODULE      = $(shell sed -n 's/^module //p' go.mod)
MAJOR       = $(patsubst v%,%,$(notdir $(MODULE)))
UNVERSIONED = $(patsubst %/,%,$(dir $(MODULE)))
VERSIONED   = go.mod README.md $(shell git ls-files '*.go')
NEW_MAJOR  ?= $(shell expr $(MAJOR) + 1)

help:
	@echo "Available targets:"
	@echo "  setup        - Install dependencies, tools, and set up git hooks"
	@echo "  install-deps - Install build/test deps (libpcap-dev + stringer)"
	@echo "  generate     - Generate code using go generate"
	@echo "  lint         - Run all prek hooks (lint, spellcheck, format) on all files"
	@echo "  spell        - Check for spelling errors in the codebase"
	@echo "  corpus       - Fetch the pcap test corpus (see test/corpus.txt)"
	@echo "  test         - Run tests with generated code"
	@echo "  example      - Run the example program (example.go)"
	@echo "  bump-major   - Bump the module major version (go.mod, imports, README)"
	@echo "  check-major  - Verify go.mod, imports, and README agree on the major"
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
	@mkdir -p test/corpus
	@while read -r name sha url; do \
		case "$$name" in ''|\#*) continue;; esac; \
		f="test/corpus/$$name"; \
		echo "$$sha  $$f" | sha256sum --check --quiet --status - 2>/dev/null && continue; \
		echo "fetching $$name"; \
		curl -sSfL "$$url" -o "$$f" || exit 1; \
		echo "$$sha  $$f" | sha256sum --check --quiet - || { rm -f "$$f"; exit 1; }; \
	done < test/corpus.txt

test: generate
	go test -v ./...

example: generate
	go run .

bump-major:
	sed -i 's|$(UNVERSIONED)/v$(MAJOR)|$(UNVERSIONED)/v$(NEW_MAJOR)|g' $(VERSIONED)
	@echo "bumped module major: v$(MAJOR) -> v$(NEW_MAJOR)"

check-major:
	@stale=$$(grep -n '$(UNVERSIONED)/v[0-9][0-9]*' $(VERSIONED) \
		| grep -v '$(UNVERSIONED)/v$(MAJOR)\b') || true; \
	if [ -n "$$stale" ]; then \
		echo "module major is v$(MAJOR), but these disagree:"; \
		echo "$$stale"; \
		exit 1; \
	fi

clean:
	rm -f **/*_string.go .generated-canary
