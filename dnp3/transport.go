package dnp3

import "fmt"

// DNP3Transport is the second layer of DNP3, and allows for fragmentation
// and subsequent reassembly of application data. In addition to trans header,
// DNP3Transport also intersperses CRC checksums after every 16 bytes
type Transport struct {
	FIN bool
	FIR bool
	SEQ uint8 // only 6 bits
	CRC [][]byte
}

func (trans *Transport) FromBytes(d []byte) ([]byte, error) {
	crcs, clean, err := RemoveDNP3CRCs(d)
	if err != nil {
		return nil, fmt.Errorf("can't remove crcs: %w", err)
	}

	trans.FIN = (d[0] & 0b10000000) != 0
	trans.FIR = (d[0] & 0b01000000) != 0
	trans.SEQ = (d[0] & 0b00111111)
	trans.CRC = crcs

	return clean[1:], nil
}

func (trans *Transport) ToByte() byte {
	var o byte = 0

	if trans.FIN {
		o |= 0b10000000
	}
	if trans.FIR {
		o |= 0b01000000
	}

	o |= (trans.SEQ & 0b00111111)

	return o
}

func (trans Transport) String() string {
	return fmt.Sprintf(`
	Transport:
		FIN: %t
		FIR: %t
		SEQ: %d
		CRC: 0x % X`,
		trans.FIN, trans.FIR, trans.SEQ, trans.CRC)
}
