package dnp3

import (
	"fmt"
)

// Transport is the second layer of DNP3, and allows for fragmentation
// and subsequent reassembly of application data. In addition to trans header,
// DNP3Transport also intersperses CRC checksums after every 16 bytes.
type Transport struct {
	FIN bool
	FIR bool
	SEQ uint8 // only 6 bits
	CRC [][]byte
}

func (trans *Transport) FromBytes(data []byte) ([]byte, error) {
	crcs, clean, err := RemoveDNP3CRCs(data)
	if err != nil {
		return nil, fmt.Errorf("can't remove crcs: %w", err)
	}

	trans.FIN = (data[0] & 0b10000000) != 0
	trans.FIR = (data[0] & 0b01000000) != 0
	trans.SEQ = (data[0] & 0b00111111)
	trans.CRC = crcs

	return clean[1:], nil
}

func (trans *Transport) ToByte() (byte, error) {
	var transportByte byte

	if trans.FIN {
		transportByte |= 0b10000000
	}

	if trans.FIR {
		transportByte |= 0b01000000
	}

	if trans.SEQ > 63 {
		return 0, fmt.Errorf("transport sequence number %d exceeds 6 bits", trans.SEQ)
	}

	transportByte |= (trans.SEQ & 0b00111111)

	return transportByte, nil
}

func (trans *Transport) String() string {
	return fmt.Sprintf(`Transport:
	FIN: %t
	FIR: %t
	SEQ: %d`,
		trans.FIN, trans.FIR, trans.SEQ)
}
