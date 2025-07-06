# go-dnp3

Partial DNP3 layer for gopacket.

### Status

This is better than the [Scapy DNP3](https://github.com/nrodofile/ScapyDNP3_lib) implementation which stops at the application data, but not as good as the [C++ (archived)](https://github.com/dnp3/opendnp3) / [Rust (paid)](https://github.com/stepfunc/dnp3) OpenDNP3 implementations. 

### Improvements

* [ ] Currently hangs parsing packet 91 of example (binary points)
* [ ] Points are read out as []byte, including their indexes, flags, etc.
* [ ] Support for Binary points
* [ ] Support for more common groups and variations 