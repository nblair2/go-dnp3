package dnp3

import (
	"fmt"
)

// DNP3 ApplicationRequest are sent from the master to the outstation
type ApplicationRequest struct {
	CTL  ApplicationCTL
	FC   RequestFC
	Data ApplicationData
}

func (appreq *ApplicationRequest) FromBytes(d []byte) error {

	appreq.CTL.FromByte(d[0])
	appreq.FC = RequestFC(d[1])
	if err := appreq.Data.FromBytes(d[2:]); err != nil {
		fmt.Printf("error decoding: %v\n", err)
	}
	return nil
}

func (appreq *ApplicationRequest) ToBytes() []byte {
	var o []byte
	o = append(o, appreq.CTL.ToByte())
	o = append(o, byte(appreq.FC))
	o = append(o, appreq.Data.ToBytes()...)
	return o
}

func (appreq *ApplicationRequest) String() string {
	o := fmt.Sprintf(`
	Application (Request):
		%s
		FC : (%d) %s`,
		appreq.CTL.String(), appreq.FC, appreq.FC.String())
	d := appreq.Data.String()
	if d != "" {
		o += "\n\t\t" + d
	}
	return o
}

func (appreq *ApplicationRequest) GetSequence() uint8 {
	return appreq.CTL.SEQ
}

func (appreq *ApplicationRequest) SetSequence(s uint8) error {
	if s >= 0b00001111 {
		return fmt.Errorf("application sequence is only 4 bits, got %d", s)
	}
	appreq.CTL.SEQ = s
	return nil
}

func (appreq *ApplicationRequest) GetFunctionCode() byte {
	return byte(appreq.FC)
}

func (appreq *ApplicationRequest) SetFunctionCode(d byte) {
	appreq.FC = RequestFC(d)
}
func (appreq *ApplicationRequest) GetData() []byte {
	return appreq.Data.ToBytes()
}

func (appreq *ApplicationRequest) SetData(data []byte) error {
	err := appreq.Data.FromBytes(data)
	return err
}

// DNP3 Application RequestFC specify the action the master is directing the
// outstation to take
type RequestFC byte

const (
	Confirm RequestFC = iota // 0x0
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
	AuthenticationRequestNoAck //0x21
)

var RequestFCNames = map[RequestFC]string{
	Confirm:               "CONFIRM",
	Read:                  "READ",
	Write:                 "WRITE",
	Select:                "SELECT",
	Operate:               "OPERATE",
	DirOperate:            "DIRECT_OPERATE",
	DirOperateNoAck:       "DIRECT_OPERATE_NO_ACK",
	Freeze:                "FREEZE",
	FreezeNoAck:           "FREEZE_NO_ACK",
	FreezeClear:           "FREEZE_CLEAR",
	FreezeClearNoAck:      "FREEZE_CLEAR_NO_ACK",
	FreezeAtTime:          "FREEZE_AT_TIME",
	FreezeAtTimeNoAck:     "FREEZE_AT_TIME_NO_ACK",
	ColdRestart:           "COLD_RESTART",
	WarmRestart:           "WARM_RESTART",
	InitializedData:       "INITIALIZE_DATA",
	InitializeApplication: "INITIALIZE_APPLICATION",
	StartApplication:      "START_APPLICATION",
	StopApplication:       "STOP_APPLICATION",
	SaveConfiguration:     "SAVE_CONFIGURATION",
	EnableUnsolicited:     "ENABLE_UNSOLICITED",
	DisableUnsolicited:    "DISABLE_UNSOLICITED",
	AssignClass:           "ASSIGN_CLASS",
	DelayMeasurement:      "DELAY_MEASUREMENT",
	RecordCurrentTime:     "RECORD_CURRENT_TIME",
	OpenFile:              "OPEN_FILE",
	CloseFile:             "CLOSE_FILE",
	DeleteFile:            "DELETE_FILE",
	GetFileInformation:    "GET_FILE_INFORMATION",
	AuthenticateFile:      "AUTHENTICATE_FILE",
	AbortFile:             "ABORT_FILE",
	ActivateConfig:        "ACTIVATE_CONFIG",
	AuthenticationRequest: "AUTHENTICATION_REQUEST",
}

func (fc RequestFC) String() string {
	if name, ok := RequestFCNames[fc]; ok {
		return name
	}
	return fmt.Sprintf("Unknown Function Code %d", fc)
}
