package dnp3

import "fmt"

// ApplicationCTL
// a common header byte for both application types
type ApplicationCTL struct {
	FIR bool
	FIN bool
	CON bool
	UNS bool
	SEQ uint8 // 4 bits
}

func (appctl *ApplicationCTL) FromByte(d byte) {
	appctl.FIR = (d & 0b10000000) != 0
	appctl.FIN = (d & 0b01000000) != 0
	appctl.CON = (d & 0b00100000) != 0
	appctl.UNS = (d & 0b00010000) != 0
	appctl.SEQ = (d & 0b00001111)
}

func (appctl *ApplicationCTL) ToByte() byte {
	var o byte = 0

	if appctl.FIR {
		o |= 0b10000000
	}
	if appctl.FIN {
		o |= 0b01000000
	}
	if appctl.CON {
		o |= 0b00100000
	}
	if appctl.UNS {
		o |= 0b00010000
	}

	o |= (appctl.SEQ & 0b00001111)

	return o
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
