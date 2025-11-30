package dnp3

import (
	"fmt"
)

// Transport is the second layer of DNP3, and allows for fragmentation
// and subsequent reassembly of application data. In addition to trans header,
// DNP3Transport also intersperses CRC checksums after every 16 bytes.
type Transport struct {
	Final     bool     `json:"final"`
	First     bool     `json:"first"`
	Sequence  uint8    `json:"sequence"` // only 6 bits
	Checksums [][]byte `json:"checksums"`
}

func (trans *Transport) FromBytes(data []byte) ([]byte, error) {
	crcs, clean, err := RemoveDNP3CRCs(data)
	if err != nil {
		return nil, fmt.Errorf("can't remove crcs: %w", err)
	}

	trans.Final = (data[0] & 0b10000000) != 0
	trans.First = (data[0] & 0b01000000) != 0
	trans.Sequence = (data[0] & 0b00111111)
	trans.Checksums = crcs

	return clean[1:], nil
}

func (trans *Transport) ToByte() (byte, error) {
	var transportByte byte

	if trans.Final {
		transportByte |= 0b10000000
	}

	if trans.First {
		transportByte |= 0b01000000
	}

	if trans.Sequence > 63 {
		return 0, fmt.Errorf("transport sequence number %d exceeds 6 bits", trans.Sequence)
	}

	transportByte |= (trans.Sequence & 0b00111111)

	return transportByte, nil
}

func (trans *Transport) String() string {
	return fmt.Sprintf(`Transport:
	FIN: %t
	FIR: %t
	SEQ: %d`,
		trans.Final, trans.First, trans.Sequence)
}
