package dnp3

import (
	"fmt"
)

// DNP3ApplicationResponse is sent from outstation to master
type ApplicationResponse struct {
	CTL  ApplicationCTL
	FC   ResponseFC
	IIN  ApplicationIIN
	Data ApplicationData
}

func (appresp *ApplicationResponse) FromBytes(d []byte) error {
	appresp.CTL.FromByte(d[0])
	appresp.FC = ResponseFC(d[1])
	if err := appresp.IIN.FromBytes(d[2], d[3]); err != nil {
		return fmt.Errorf("can't create application response: %w", err)
	}
	if err := appresp.Data.FromBytes(d[4:]); err != nil {
		return fmt.Errorf("couldn't create AppReq Data FromBytes: %v", err)
	}
	return nil
}

func (appresp *ApplicationResponse) ToBytes() ([]byte, error) {
	var o []byte
	o = append(o, appresp.CTL.ToByte())
	o = append(o, byte(appresp.FC))
	o = append(o, appresp.IIN.ToBytes()...)
	b, err := appresp.Data.ToBytes()
	if err != nil {
		return o, fmt.Errorf("error encoding data: %v", err)
	}
	o = append(o, b...)
	return o, nil
}

func (appresp *ApplicationResponse) String() string {
	o := fmt.Sprintf(`
	Application (Response):
		%s
		FC : (%d) %s
		%s`,
		appresp.CTL.String(), appresp.FC, appresp.FC.String(),
		appresp.IIN.String())
	d := appresp.Data.String()
	if d != "" {
		o += "\n\t\t" + d
	}
	return o
}

func (appresp *ApplicationResponse) GetCTL() ApplicationCTL {
	return appresp.CTL
}

func (appresp *ApplicationResponse) SetCTL(ctl ApplicationCTL) {
	appresp.CTL = ctl
}

func (appresp *ApplicationResponse) GetSequence() uint8 {
	return appresp.CTL.SEQ
}

func (appresp *ApplicationResponse) SetSequence(s uint8) error {
	if s >= 0b00001111 {
		return fmt.Errorf("application sequence is only 4 bits, got %d", s)
	}
	appresp.CTL.SEQ = s
	return nil
}

func (appresp *ApplicationResponse) GetFunctionCode() byte {
	return byte(appresp.FC)
}

func (appresp *ApplicationResponse) SetFunctionCode(d byte) {
	appresp.FC = ResponseFC(d)
}

func (appresp *ApplicationResponse) GetData() ApplicationData {
	return appresp.Data
}

func (appresp *ApplicationResponse) SetData(d ApplicationData) {
	appresp.Data = d
}

// DNP3 Application ResponseFC specify the action the outstation is taking
type ResponseFC byte

const (
	Response ResponseFC = iota + 0x81
	UnsolicitedResponse
	AuthenticationResponse
)

var ResponseFCNames = map[ResponseFC]string{
	Response:               "RESPONSE",
	UnsolicitedResponse:    "UNSOLICITED_RESPONSE",
	AuthenticationResponse: "AUTHENTICATION_RESPONSE",
}

func (fc ResponseFC) String() string {
	if name, ok := ResponseFCNames[fc]; ok {
		return name
	}
	return fmt.Sprintf("unknown Function Code %d", fc)
}

// DNP3ApplicationResponse header (information about outstation)
type ApplicationIIN struct {
	// IIN 1
	AllStations   bool
	Class1Events  bool
	Class2Events  bool
	Class3Events  bool
	NeedTime      bool
	Local         bool
	DeviceTrouble bool
	Restart       bool
	// IIN 2
	BadFunction      bool
	ObjectUnknown    bool
	ParameterError   bool
	BufferOverflow   bool
	AlreadyExiting   bool
	BadConfiguration bool
	Reserved1        bool // should be 0
	Reserved2        bool // ^
}

func (appiin *ApplicationIIN) FromBytes(lsb, msb byte) error {
	appiin.AllStations = (lsb & 0b00000001) != 0
	appiin.Class1Events = (lsb & 0b00000010) != 0
	appiin.Class2Events = (lsb & 0b00000100) != 0
	appiin.Class3Events = (lsb & 0b00001000) != 0
	appiin.NeedTime = (lsb & 0b00010000) != 0
	appiin.Local = (lsb & 0b00100000) != 0
	appiin.DeviceTrouble = (lsb & 0b01000000) != 0
	appiin.Restart = (lsb & 0b10000000) != 0
	appiin.BadFunction = (msb & 0b00000001) != 0
	appiin.ObjectUnknown = (msb & 0b00000010) != 0
	appiin.ParameterError = (msb & 0b00000100) != 0
	appiin.BufferOverflow = (msb & 0b00001000) != 0
	appiin.AlreadyExiting = (msb & 0b00010000) != 0
	appiin.BadConfiguration = (msb & 0b00100000) != 0
	appiin.Reserved1 = (msb & 0b01000000) != 0
	appiin.Reserved2 = (msb & 0b10000000) != 0
	if (msb & 0b11000000) != 0 {
		return fmt.Errorf("IIN 2.6 and 2.7 must be set to 0")
	}
	return nil
}

func (d *ApplicationIIN) ToBytes() []byte {
	var lsb, msb byte

	// IIN 1 (lsb)
	if d.AllStations {
		lsb |= 0b00000001
	}
	if d.Class1Events {
		lsb |= 0b00000010
	}
	if d.Class2Events {
		lsb |= 0b00000100
	}
	if d.Class3Events {
		lsb |= 0b00001000
	}
	if d.NeedTime {
		lsb |= 0b00010000
	}
	if d.Local {
		lsb |= 0b00100000
	}
	if d.DeviceTrouble {
		lsb |= 0b01000000
	}
	if d.Restart {
		lsb |= 0b10000000
	}

	// IIN 2 (msb)
	if d.BadFunction {
		msb |= 0b00000001
	}
	if d.ObjectUnknown {
		msb |= 0b00000010
	}
	if d.ParameterError {
		msb |= 0b00000100
	}
	if d.BufferOverflow {
		msb |= 0b00001000
	}
	if d.AlreadyExiting {
		msb |= 0b00010000
	}
	if d.BadConfiguration {
		msb |= 0b00100000
	}
	if d.Reserved1 {
		msb |= 0b01000000
	}
	if d.Reserved2 {
		msb |= 0b10000000
	}

	return []byte{lsb, msb}
}

func (appiin *ApplicationIIN) String() string {
	return fmt.Sprintf(`IIN:
			IIN1:
			AllStations     : %t
			Class1Events    : %t
			Class2Events    : %t
			Class3Events    : %t
			NeedTime        : %t
			Local           : %t
			DeviceTrouble   : %t
			Restart         : %t
			IIN2:
			BadFunction     : %t
			ObjectUnknown   : %t
			ParameterError  : %t
			BufferOverflow  : %t
			AlreadyExiting  : %t
			BadConfiguration: %t`,
		appiin.AllStations, appiin.Class1Events, appiin.Class2Events,
		appiin.Class3Events, appiin.NeedTime, appiin.Local, appiin.DeviceTrouble,
		appiin.Restart, appiin.BadFunction, appiin.ObjectUnknown,
		appiin.ParameterError, appiin.BufferOverflow, appiin.AlreadyExiting,
		appiin.BadConfiguration,
	)
}
