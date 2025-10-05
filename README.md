# go-dnp3

[![GoDoc](https://godoc.org/github.com/nblair2/go-dnp3?status.svg)](https://godoc.org/github.com/nblair2/go-dnp3/dnp3)
![Go Version](https://img.shields.io/github/go-mod/go-version/nblair2/go-dnp3?filename=go.mod&style=flat-square)
![License](https://img.shields.io/github/license/nblair2/go-dnp3?style=flat-square)


DNP3 parsing in go.
* **`FromBytes([]byte)`** to parse a byte slice and interpret it as DNP3
* **`ToBytes()`** to go from a struct back to bytes (calculates length, calculates and inserts CRCs on the way)
* **`String()`** to get packet as human-readable indented string (Reserved and CRCs not shown)
* **`json.Marshal(DNP3)`** to get packet as machine-friendly 

## Improvements

* [ ] accept interfaces, return structs
* [ ] consistency on plural / singular of bits / bit and bytes / byte. Functions and attributes should be named based on what they actually are, and be as small as possible
* [ ] errors always available / checked in `FromBytes` and `ToBytes` methods. Should buble these up but not fail if possible.
* [ ] re-write a generic `Point` called `PointBytes` with flags and all possible fields (`Prefix`, `Flags`, `AbsTime`, `RelTime`) rather than the many different types we have now
* [ ] clean up the the `String` method so we aren't always appending newlines, these should be the responsibility of the caller
* [ ] consolidate the 3 different massive switch statements for Group / Variation into one
* [ ] Events / Quality, Prefix parsing
* [ ] `DIR` field / `PRM`

## Test

Run `go test ./dnp3 -v` to check a few different DNP3 messages. You can also use the `-args -pcaps=opendnp3_test1.pcap` argument to pass in a full PCAP. Data taken from [opendnp3 conformance reports](https://dnp3.github.io/conformance/report.html).

## Spec

I don't have access to the DNP3 spec, so working off of Wireshark's Parser and odd PDFs I find around the web ([one](https://www.dnp.org/Portals/0/Public%20Documents/DNP3%20AN2013-004b%20Validation%20of%20Incoming%20DNP3%20Data.pdf))