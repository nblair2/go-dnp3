# go-dnp3

DNP3 parsing in go.
* **`FromBytes([]byte)`** to parse a byte slice and interpret it as DNP3
* **`ToBytes()`** to go from a struct back to bytes (calculates length, calculates and inserts CRCs on the way)
* **`String()`** to get packet as human-readable indented string (Reserved and CRCs not shown)
* **`json.MarshalIndent(dnp, "", "  ")`** to get packet as machine-friendly 

## Improvements

* [ ] consistency on plural / singular of bits / bit and bytes / byte. Functions and attributes should be named based on what they actually are, and be as small as possible
* [ ] errors always available / checked in `FromBytes` and `ToBytes` methods. Should buble these up but not fail if possible.
* [ ] re-write a generic `Point` called `PointBytes` with flags and all possible fields (`Prefix`, `Flags`, `AbsTime`, `RelTime`) rather than the many different types we have now
* [ ] clean up the the `String` method so we aren't always appending newlines, these should be the responsibility of the caller
* [ ] consolidate the 3 different massive switch statements for Group / Variation into one
* [ ] Always more point improvements (Events / Quality, Prefix parsing)
* [ ] `DIR` field / `PRM`

## Test

Run `make test` to check `examples/*.pcap` for errors. Data taken from [opendnp3 conformance reports](https://dnp3.github.io/conformance/report.html) and [ITI ICS Security Tools Repository](https://github.com/ITI/ICS-Security-Tools/tree/master/pcaps/dnp3)

## Spec

I don't have access to the DNP3 spec, so working off of Wireshark's Parser and odd PDFs I find around the web ([one](https://www.dnp.org/Portals/0/Public%20Documents/DNP3%20AN2013-004b%20Validation%20of%20Incoming%20DNP3%20Data.pdf))