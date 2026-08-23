# go-dnp3

[![Release](https://img.shields.io/github/v/release/nblair2/go-dnp3?style=flat-square)](https://github.com/nblair2/go-dnp3/releases/latest)
[![GoDoc](https://godoc.org/github.com/nblair2/go-dnp3?status.svg)](https://godoc.org/github.com/nblair2/go-dnp3/v2/dnp3)
[![Go Version](https://img.shields.io/github/go-mod/go-version/nblair2/go-dnp3?filename=go.mod&style=flat-square)](go.mod)
[![License](https://img.shields.io/github/license/nblair2/go-dnp3?style=flat-square)](LICENSE.txt)

**`go-dnp3` is a Go library for parsing and encoding DNP3 (Distributed Network Protocol) frames**

![DNP3 Gopher](.media/dnp3-gopher.png)

## Usage

`*dnp3.Frame` implements the standard gopacket interfaces (`Layer`, `DecodingLayer`, `SerializableLayer`, `ApplicationLayer`). TCP and UDP port 20000 (DNP3-over-IP) are auto-registered, so `gopacket.NewPacket` decodes DNP3 automatically.

*   **Parsing**: Use `gopacket.NewPacket(data, dnp3.LayerTypeDNP3, gopacket.Default)`, or `dnp3.NewFrameFromBytes(data)` for raw frame bytes, or `frame.DecodeFromBytes(data, df)` to drive `gopacket.DecodingLayerParser`.
*   **Encoding**: Use `gopacket.SerializeLayers(buf, opts, frame)`. `Frame.SerializeTo` recomputes `DataLink.Length` and inserts DNP3 CRCs on the fly.
*   **Stream parsing**: Use `dnp3.ParseFrames(data)` to consume multiple DNP3 frames out of a single TCP read (handles partial trailing frames).
*   **Reassembly**: Use `dnp3.Assembler` to rebuild application fragments that span multiple transport segments, tracked per session (source, destination, direction). See `ExampleAssembler`, and `test/stream_test.go` for wiring it up to `gopacket/tcpassembly`.
*   **Inspection**: Use `String()` for a human-readable, indented packet dump (excludes reserved fields and CRCs).
*   **Serialization**: Full support for `json.Marshal()` to convert packets into machine-friendly JSON.

## Development

| Target | Description |
|--------|-------------|
| `make setup` | Install development tools: `prek`, `stringer`, `libpcap-dev` |
| `make generate` | Generate code using go generate |
| `make lint` | Run all prek hooks (lint, spellcheck, format) on all files |
| `make corpus` | Fetch the pcap test corpus (see test/testdata/corpus.txt) |
| `make test` | Run tests with generated code |
| `make clean` | Remove generated files and canary |

>  pcap corpus (pinned + checksummed in [`test/testdata/corpus.txt`](test/testdata/corpus.txt),
> sourced from the CC-BY-4.0 [ITI/ICS-Security-Tools](https://github.com/ITI/ICS-Security-Tools) captures

#### PCAP Testing
Round-trip ad-hoc PCAP files using the `-pcaps` flag:

```bash
go test ./test -v -args -pcaps=my-custom.pcap,another.pcap
```

#### Printing Strings
View the string and json outputs of test cases using the `-args` flags `-print-string` and `-print-json`.

```bash
go test ./dnp3 -args -print-string -print-json
```
