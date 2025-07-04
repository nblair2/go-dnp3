package dnp3

import "fmt"

// DNP3Transport is the second layer of DNP3, and allows for fragmentation
// and subsequent reassembly of application data. In addition to this header,
// DNP3Transport also intersperses CRC checksums after every 16 bytes
type DNP3Transport struct {
	FIN bool
	FIR bool
	SEQ uint8 // only 6 bits
	CRC [][]byte
}

// NewDNP3Transport parses and builds the DNP3 Transport header and remove
// the CRCs interspersed in the application
func NewDNP3Transport(data []byte) (DNP3Transport, []byte, error) {
	var (
		crcs  [][]byte
		clean []byte
		err   error
	)

	crcs, clean, err = RemoveDNP3CRCs(data)
	if err != nil {
		return DNP3Transport{}, nil, fmt.Errorf("can't remove crcs: %w", err)
	}

	t := DNP3Transport{
		FIN: (data[0] & 0b10000000) != 0,
		FIR: (data[0] & 0b01000000) != 0,
		SEQ: (data[0] & 0b00111111),
		CRC: crcs,
	}

	return t, clean[1:], nil
}

func (d *DNP3Transport) ToBytes() []byte {
	var b byte = 0

	if d.FIN {
		b |= 0b10000000
	}
	if d.FIR {
		b |= 0b01000000
	}

	b |= (d.SEQ & 0b00111111)

	return []byte{b}
}

func (d DNP3Transport) String() string {
	return fmt.Sprintf(`
	Transport:
		FIN: %t
		FIR: %t
		SEQ: %d
		CRC: 0x % X`,
		d.FIN, d.FIR, d.SEQ, d.CRC)
}
