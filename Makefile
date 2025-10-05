.PHONY: hooks lint fix test

hooks:
	git config core.hooksPath .githooks

lint:
	golangci-lint run ./...

fix:
	golangci-lint run ./... --fix

test:
	go test ./dnp3 -v -args -pcaps=opendnp3_test1.pcap