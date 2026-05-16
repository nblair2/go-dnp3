# go-dnp3

[![GoDoc](https://godoc.org/github.com/nblair2/go-dnp3/v2?status.svg)](https://godoc.org/github.com/nblair2/go-dnp3/v2/dnp3)
![Go Version](https://img.shields.io/github/go-mod/go-version/nblair2/go-dnp3?filename=go.mod&style=flat-square)
![License](https://img.shields.io/github/license/nblair2/go-dnp3?style=flat-square)

`go-dnp3` is a Go library for parsing and encoding DNP3 (Distributed Network Protocol) frames. 

![DNP3 Gopher](.media/dnp3-gopher.png)

## Usage

`*dnp3.Frame` implements the standard gopacket interfaces (`Layer`, `DecodingLayer`, `SerializableLayer`, `ApplicationLayer`). TCP and UDP port 20000 (DNP3-over-IP) are auto-registered, so `gopacket.NewPacket` decodes DNP3 automatically.

*   **Parsing**: Use `gopacket.NewPacket(data, dnp3.LayerTypeDNP3, gopacket.Default)`, or `dnp3.NewFrameFromBytes(data)` for raw frame bytes, or `frame.DecodeFromBytes(data, df)` to drive `gopacket.DecodingLayerParser`.
*   **Encoding**: Use `gopacket.SerializeLayers(buf, opts, frame)`. `Frame.SerializeTo` recomputes `DataLink.Length` and inserts DNP3 CRCs on the fly.
*   **Stream parsing**: Use `dnp3.ParseFrames(data)` to consume multiple DNP3 frames out of a single TCP read (handles partial trailing frames).
*   **Inspection**: Use `String()` for a human-readable, indented packet dump (excludes reserved fields and CRCs).
*   **Serialization**: Full support for `json.Marshal()` to convert packets into machine-friendly JSON.

### Example

```go
// Auto-decode (NewPacket dispatches on LayerTypeDNP3 directly, or on TCP/UDP port 20000 in a pcap):
pkt := gopacket.NewPacket(input, dnp3.LayerTypeDNP3, gopacket.Default)
frame := pkt.Layer(dnp3.LayerTypeDNP3).(*dnp3.Frame)

// Build outbound:
buf := gopacket.NewSerializeBuffer()
gopacket.SerializeLayers(buf, gopacket.SerializeOptions{}, frame)
wire := buf.Bytes()
```

See [`example.go`](example.go) for a full end-to-end demo, including in-place point mutation and round-tripping.

## Development

### Setup
Run `make setup` to install development tools used by this repository.

### Testing
> Data for tests is sourced from [opendnp3 conformance reports](https://dnp3.github.io/conformance/report.html)

Run `make test` to run basic tests.

#### PCAP Testing
Pass a full PCAP file using the `-args` option `-pcaps=comma.pcap,delimited.pcap,list.pcap`.

```bash
go test ./dnp3 -v -args -pcaps=my-custom.pcap
```

#### Printing Strings
View the string and json outputs of test cases using the `-args` flag `-print-string` and `-print-json`.

```bash
go test ./dnp3 -args -print
```

### Linting
[`golangci-lint`](https://golangci-lint.run/) is used for lint and format checking. Run `make lint` to check for errors, and `make fix` to try to automatically fix linting or formatting errors.

## Implementation

Based on Wireshark's parser and publicly available documents (such as [this validation guide](https://www.dnp.org/Portals/0/Public%20Documents/DNP3%20AN2013-004b%20Validation%20of%20Incoming%20DNP3%20Data.pdf)), as access to the official DNP3 specification is restricted.
