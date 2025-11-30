package dnp3

import (
	"fmt"
)

// ApplicationRequest - sent from the master to the outstation.
type ApplicationRequest struct {
	Control      ApplicationControl  `json:"control"`
	FunctionCode RequestFunctionCode `json:"function_code"`
	Data         ApplicationData     `json:"data"`
}

func (appreq *ApplicationRequest) FromBytes(data []byte) error {
	appreq.Control.FromByte(data[0])

	appreq.FunctionCode = RequestFunctionCode(data[1])

	err := appreq.Data.FromBytes(data[2:])
	if err != nil {
		return fmt.Errorf("couldn't create AppReq Data FromBytes: %w", err)
	}

	return nil
}

func (appreq *ApplicationRequest) ToBytes() ([]byte, error) {
	var encoded []byte

	ctlByte, err := appreq.Control.ToByte()
	if err != nil {
		return encoded, fmt.Errorf("error encoding application control: %w", err)
	}

	encoded = append(encoded, ctlByte)
	encoded = append(encoded, byte(appreq.FunctionCode))

	dataBytes, err := appreq.Data.ToBytes()
	if err != nil {
		return encoded, fmt.Errorf("couldn't convert AppReq Data ToBytes: %w", err)
	}

	encoded = append(encoded, dataBytes...)

	return encoded, nil
}

func (appreq *ApplicationRequest) String() string {
	requestString := fmt.Sprintf("Application (Request):\n%s\n\tFC : (%d) %s",
		indent(appreq.Control.String(), "\t"), appreq.FunctionCode, appreq.FunctionCode.String())

	dataString := appreq.Data.String()
	if dataString != "" {
		requestString += "\n" + indent(dataString, "\t")
	}

	return requestString
}

func (appreq *ApplicationRequest) GetControl() ApplicationControl {
	return appreq.Control
}

func (appreq *ApplicationRequest) SetControl(ctl ApplicationControl) {
	appreq.Control = ctl
}

func (appreq *ApplicationRequest) GetSequence() uint8 {
	return appreq.Control.Sequence
}

func (appreq *ApplicationRequest) SetSequence(s uint8) error {
	if s >= 0b00001111 {
		return fmt.Errorf("application sequence is only 4 bits, got %d", s)
	}

	appreq.Control.Sequence = s

	return nil
}

func (appreq *ApplicationRequest) GetFunctionCode() byte {
	return byte(appreq.FunctionCode)
}

func (appreq *ApplicationRequest) SetFunctionCode(code byte) {
	appreq.FunctionCode = RequestFunctionCode(code)
}

func (appreq *ApplicationRequest) GetData() ApplicationData {
	return appreq.Data
}

func (appreq *ApplicationRequest) SetData(payload ApplicationData) {
	appreq.Data = payload
}

// RequestFunctionCode - specify the action the master is directing the
// outstation to take.
//
//go:generate stringer -type=RequestFunctionCode
type RequestFunctionCode byte

const (
	Confirm RequestFunctionCode = iota // 0x0
	Read
	Write
	Select
	Operate
	DirOperate
	DirOperateNoAck
	Freeze
	FreezeNoAck
	FreezeClear
	FreezeClearNoAck
	FreezeAtTime
	FreezeAtTimeNoAck
	ColdRestart
	WarmRestart
	InitializedData
	InitializeApplication
	StartApplication
	StopApplication
	SaveConfiguration
	EnableUnsolicited
	DisableUnsolicited
	AssignClass
	DelayMeasurement
	RecordCurrentTime
	OpenFile
	CloseFile
	DeleteFile
	GetFileInformation
	AuthenticateFile
	AbortFile
	ActivateConfig
	AuthenticationRequest
	AuthenticationRequestNoAck // 0x21
)
