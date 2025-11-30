package dnp3

import (
	"errors"
	"fmt"
)

// ApplicationResponse - sent from outstation to master.
type ApplicationResponse struct {
	Control             ApplicationControl             `json:"control"`
	FunctionCode        ResponseFunctionCode           `json:"function_code"`
	InternalIndications ApplicationInternalIndications `json:"internal_indications"`
	Data                ApplicationData                `json:"data"`
}

func (appresp *ApplicationResponse) FromBytes(data []byte) error {
	appresp.Control.FromByte(data[0])

	appresp.FunctionCode = ResponseFunctionCode(data[1])

	err := appresp.InternalIndications.FromBytes(data[2], data[3])
	if err != nil {
		return fmt.Errorf("can't create application response: %w", err)
	}

	err = appresp.Data.FromBytes(data[4:])
	if err != nil {
		return fmt.Errorf("couldn't create AppReq Data FromBytes: %w", err)
	}

	return nil
}

func (appresp *ApplicationResponse) ToBytes() ([]byte, error) {
	var encoded []byte

	ctlByte, err := appresp.Control.ToByte()
	if err != nil {
		return encoded, fmt.Errorf("error encoding application control: %w", err)
	}

	encoded = append(encoded, ctlByte)
	encoded = append(encoded, byte(appresp.FunctionCode))
	encoded = append(encoded, appresp.InternalIndications.ToBytes()...)

	dataBytes, err := appresp.Data.ToBytes()
	if err != nil {
		return encoded, fmt.Errorf("error encoding data: %w", err)
	}

	encoded = append(encoded, dataBytes...)

	return encoded, nil
}

func (appresp *ApplicationResponse) String() string {
	responseString := fmt.Sprintf("Application (Response):\n%s\n\tFC : (%d) %s\n%s",
		indent(appresp.Control.String(), "\t"), appresp.FunctionCode, appresp.FunctionCode.String(),
		indent(appresp.InternalIndications.String(), "\t"))

	dataString := appresp.Data.String()
	if dataString != "" {
		responseString += "\n" + indent(dataString, "\t")
	}

	return responseString
}

func (appresp *ApplicationResponse) GetControl() ApplicationControl {
	return appresp.Control
}

func (appresp *ApplicationResponse) SetControl(ctl ApplicationControl) {
	appresp.Control = ctl
}

func (appresp *ApplicationResponse) GetSequence() uint8 {
	return appresp.Control.Sequence
}

func (appresp *ApplicationResponse) SetSequence(s uint8) error {
	if s >= 0b00001111 {
		return fmt.Errorf("application sequence is only 4 bits, got %d", s)
	}

	appresp.Control.Sequence = s

	return nil
}

func (appresp *ApplicationResponse) GetFunctionCode() byte {
	return byte(appresp.FunctionCode)
}

func (appresp *ApplicationResponse) SetFunctionCode(code byte) {
	appresp.FunctionCode = ResponseFunctionCode(code)
}

func (appresp *ApplicationResponse) GetData() ApplicationData {
	return appresp.Data
}

func (appresp *ApplicationResponse) SetData(payload ApplicationData) {
	appresp.Data = payload
}

// ResponseFunctionCode - specify the action the outstation is taking.
//
//go:generate stringer -type=ResponseFunctionCode
type ResponseFunctionCode byte

const (
	Response ResponseFunctionCode = iota + 0x81
	UnsolicitedResponse
	AuthenticationResponse
)

// ApplicationInternalIndications - information about outstation sate.
type ApplicationInternalIndications struct {
	// IIN 1
	AllStations   bool `json:"all_stations"`
	Class1Events  bool `json:"class_1_events"`
	Class2Events  bool `json:"class_2_events"`
	Class3Events  bool `json:"class_3_events"`
	NeedTime      bool `json:"need_time"`
	Local         bool `json:"local"`
	DeviceTrouble bool `json:"device_trouble"`
	Restart       bool `json:"restart"`
	// IIN 2
	BadFunction      bool `json:"bad_function"`
	ObjectUnknown    bool `json:"object_unknown"`
	ParameterError   bool `json:"parameter_error"`
	BufferOverflow   bool `json:"buffer_overflow"`
	AlreadyExiting   bool `json:"already_exiting"`
	BadConfiguration bool `json:"bad_configuration"`
	Reserved1        bool `json:"reserved_1"` // should be 0
	Reserved2        bool `json:"reserved_2"` // ^
}

func (appiin *ApplicationInternalIndications) FromBytes(lsb, msb byte) error {
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
		return errors.New("IIN 2.6 and 2.7 must be set to 0")
	}

	return nil
}

func boolToBits(bools []bool) byte {
	var out byte

	for i, v := range bools {
		if v {
			out |= 1 << i
		}
	}

	return out
}

func (appiin *ApplicationInternalIndications) ToBytes() []byte {
	lsb := boolToBits([]bool{
		appiin.AllStations,
		appiin.Class1Events,
		appiin.Class2Events,
		appiin.Class3Events,
		appiin.NeedTime,
		appiin.Local,
		appiin.DeviceTrouble,
		appiin.Restart,
	})
	msb := boolToBits([]bool{
		appiin.BadFunction,
		appiin.ObjectUnknown,
		appiin.ParameterError,
		appiin.BufferOverflow,
		appiin.AlreadyExiting,
		appiin.BadConfiguration,
		appiin.Reserved1,
		appiin.Reserved2,
	})

	return []byte{lsb, msb}
}

func (appiin *ApplicationInternalIndications) String() string {
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
