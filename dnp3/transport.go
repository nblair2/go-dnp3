package dnp3

import "fmt"

// DNP3Transport is the second layer of DNP3, and allows for fragmentation
// and subsequent reassembly of application data. In addition to this header,
// DNP3Transport also intersperses CRC checksums after every 16 bytes
type DNP3Transport struct {
	FIN bool
	FIR bool
	SEQ uint8 // only 6 bits
}

// NewDNP3Transport parses and builds the DNP3 Transport header
func NewDNP3Transport(b byte) DNP3Transport {
	return DNP3Transport{
		FIN: (b & 0b10000000) != 0,
		FIR: (b & 0b01000000) != 0,
		SEQ: (b & 0b00111111),
	}
}

func (d DNP3Transport) String() string {
	return fmt.Sprintf(`
	Transport:
		FIN: %t
		FIR: %t
		SEQ: %d`,
		d.FIN, d.FIR, d.SEQ)
}
