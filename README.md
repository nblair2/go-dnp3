# go-dnp3

[![GoDoc](https://godoc.org/github.com/nblair2/go-dnp3?status.svg)](https://godoc.org/github.com/nblair2/go-dnp3/dnp3)
![Go Version](https://img.shields.io/github/go-mod/go-version/nblair2/go-dnp3?filename=go.mod&style=flat-square)
![License](https://img.shields.io/github/license/nblair2/go-dnp3?style=flat-square)

`go-dnp3` is a Go library for parsing and encoding DNP3 (Distributed Network Protocol) frames. 

![DNP3 Gopher](.media/dnp3-gopher.png)

## Key Methods

*   **Parsing**: Use `FromBytes([]byte)` to parse raw byte slices into structured DNP3 objects
*   **Encoding**: Use `ToBytes()` to serialize structs back into bytes. This automatically handles length calculations and inserts CRCs on the fly
*   **Inspection**: Use `String()` to generate human-readable, indented string representations of packets (excluds reserved fields and CRCs)
*   **Serialization**: Full support for `json.Marshal()` to convert packets into machine-friendly JSON formats.

## Installation

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/nblair2/go-dnp3/dnp3"
)

func main() {
	// Example DNP3 bytes
	data := []byte{
		0x05, 0x64, 0x05, 0xC0, 0x01, 0x00, 0x00, 0x04, 0xE9, 0x21,
	}

	frame := dnp3.Frame{}
	if err := frame.FromBytes(data); err != nil {
		log.Fatalf("Failed to parse frame: %v", err)
	}

	fmt.Println(frame.String())
}
```

## Development

### Setup
Run `make setup` to install development tools used by this repository. 

### Testing
> Data for tests is sourced from [opendnp3 conformance reports](https://dnp3.github.io/conformance/report.html)

Run `make test` to run basic tests.

#### PCAP Testing
You can pass a full PCAP file using the `-args` option `-pcaps=comma.pcap,delimited.pcap,list.pcap`

```bash
go test ./dnp3 -v -args -pcaps=my-custom.pcap
```

#### Printing Strings
You can view the string outputs of packets using the `-args` flag `-print`.

```bash
go test ./dnp3 -args -print
```

### Linting
[`golangci-lint`](https://golangci-lint.run/) is used for lint and format checking. Run `make lint` to check for errors.


## Disclaimer

This implementation is based on Wireshark's parser and publicly available documents (such as [this validation guide](https://www.dnp.org/Portals/0/Public%20Documents/DNP3%20AN2013-004b%20Validation%20of%20Incoming%20DNP3%20Data.pdf)), as access to the official DNP3 specification is restricted.