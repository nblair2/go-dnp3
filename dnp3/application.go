package dnp3

import (
	"fmt"
)

// Application - layer abstraction for Request / Response structure.
type Application interface {
	FromBytes(data []byte) error
	ToBytes() ([]byte, error)
	String() string
	GetControl() ApplicationControl
	SetControl(ctl ApplicationControl)
	GetSequence() uint8
	SetSequence(seq uint8) error
	GetFunctionCode() byte
	SetFunctionCode(fc byte)
	GetData() ApplicationData
	SetData(data ApplicationData)
}

// ApplicationControl - header byte for both application types.
type ApplicationControl struct {
	First       bool  `json:"first"`
	Final       bool  `json:"final"`
	Confirm     bool  `json:"confirm"`
	Unsolicited bool  `json:"unsolicited"`
	Sequence    uint8 `json:"sequence"` // 4 bits
}

func (appctl *ApplicationControl) FromByte(value byte) {
	appctl.First = (value & 0b10000000) != 0
	appctl.Final = (value & 0b01000000) != 0
	appctl.Confirm = (value & 0b00100000) != 0
	appctl.Unsolicited = (value & 0b00010000) != 0
	appctl.Sequence = (value & 0b00001111)
}

func (appctl *ApplicationControl) ToByte() (byte, error) {
	var ctlByte byte

	if appctl.First {
		ctlByte |= 0b10000000
	}

	if appctl.Final {
		ctlByte |= 0b01000000
	}

	if appctl.Confirm {
		ctlByte |= 0b00100000
	}

	if appctl.Unsolicited {
		ctlByte |= 0b00010000
	}

	if appctl.Sequence > 15 {
		return 0, fmt.Errorf("sequence number %d exceeds 4 bits", appctl.Sequence)
	}

	ctlByte |= (appctl.Sequence & 0b00001111)

	return ctlByte, nil
}

func (appctl *ApplicationControl) String() string {
	return fmt.Sprintf(`CTL:
	FIR: %t
	FIN: %t
	CON: %t
	UNS: %t
	SEQ: %d`,
		appctl.First, appctl.Final, appctl.Confirm, appctl.Unsolicited, appctl.Sequence)
}
