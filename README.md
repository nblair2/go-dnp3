# go-dnp3

Partial DNP3 layer for gopacket.

### Status

This is better than the [Scapy DNP3](https://github.com/nrodofile/ScapyDNP3_lib) implementation which stops at the application data, but not as good as the [C++ (archived)](https://github.com/dnp3/opendnp3) / [Rust (paid)](https://github.com/stepfunc/dnp3) OpenDNP3 implementations. 

### Improvements

* build out custom Group/Variation structs so we can interpret the data better
* the couple of times we are are getting a "raw" / extra are from bad packets (a reserved field is set)