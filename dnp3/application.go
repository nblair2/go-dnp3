package dnp3

import (
	"fmt"
)

// DNP3 Application layer abstraction for different Request / Response
// structure.
type Application interface {
	FromBytes(data []byte) error
	ToBytes() ([]byte, error)
	String() string
	GetCTL() ApplicationCTL
	SetCTL(ctl ApplicationCTL)
	GetSequence() uint8
	SetSequence(seq uint8) error
	GetFunctionCode() byte
	SetFunctionCode(fc byte)
	GetData() ApplicationData
	SetData(data ApplicationData)
}

// ApplicationCTL
// a common header byte for both application types.
type ApplicationCTL struct {
	FIR bool
	FIN bool
	CON bool
	UNS bool
	SEQ uint8 // 4 bits
}

func (appctl *ApplicationCTL) FromByte(value byte) {
	appctl.FIR = (value & 0b10000000) != 0
	appctl.FIN = (value & 0b01000000) != 0
	appctl.CON = (value & 0b00100000) != 0
	appctl.UNS = (value & 0b00010000) != 0
	appctl.SEQ = (value & 0b00001111)
}

func (appctl *ApplicationCTL) ToByte() (byte, error) {
	var ctlByte byte

	if appctl.FIR {
		ctlByte |= 0b10000000
	}

	if appctl.FIN {
		ctlByte |= 0b01000000
	}

	if appctl.CON {
		ctlByte |= 0b00100000
	}

	if appctl.UNS {
		ctlByte |= 0b00010000
	}

	if appctl.SEQ > 15 {
		return 0, fmt.Errorf("sequence number %d exceeds 4 bits", appctl.SEQ)
	}

	ctlByte |= (appctl.SEQ & 0b00001111)

	return ctlByte, nil
}

func (appctl *ApplicationCTL) String() string {
	return fmt.Sprintf(`CTL:
	FIR: %t
	FIN: %t
	CON: %t
	UNS: %t
	SEQ: %d`,
		appctl.FIR, appctl.FIN, appctl.CON, appctl.UNS, appctl.SEQ)
}
