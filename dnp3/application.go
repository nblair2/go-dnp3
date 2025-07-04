package dnp3

import "fmt"

// DNP3Application Message abstraction for different Request / Response
// structure
type DNP3Application interface {
	ToBytes() []byte
	String() string
	LayerPayload() []byte
	SetCTL(DNP3ApplicationControl)
	SetSequence(uint8) error
	SetContents([]byte)
	// HACK because Go doesn't let me do proper OO
	IsDNP3Application() bool
}

// DNP3ApplicationRequest is sent from the master to the outstation
type DNP3ApplicationRequest struct {
	CTL DNP3ApplicationControl
	FC  byte
	Raw []byte // Not sure what these are...
}

// DNP3ApplicationResponse is sent from outstation to master
type DNP3ApplicationResponse struct {
	CTL DNP3ApplicationControl
	FC  byte
	IIN DNP3ApplicationIIN
	OBJ []byte // Store all the data here for now
}

// DNP3ApplicationCOntrol is a common header byte for both application types
type DNP3ApplicationControl struct {
	FIR bool
	FIN bool
	CON bool
	UNS bool
	SEQ uint8 //only 4 bits
}

// DNP3ApplicationResponse header contains
type DNP3ApplicationIIN struct {
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
	Reserved1        bool // Should always be set to 0, not enforced
	Reserved2        bool // ^
}

func NewDNP3ApplicationRequest(d []byte) DNP3ApplicationRequest {
	return DNP3ApplicationRequest{
		CTL: NewDNP3ApplicationControl(d[0]),
		FC:  d[1],
		Raw: d[2:],
	}
}

func NewDNP3ApplicationResponse(d []byte) DNP3ApplicationResponse {
	return DNP3ApplicationResponse{
		CTL: NewDNP3ApplicationControl(d[0]),
		FC:  d[1],
		IIN: NewDNP3ApplicationIIN(d[2], d[3]),
		OBJ: d[4:],
	}
}

func NewDNP3ApplicationControl(b byte) DNP3ApplicationControl {
	return DNP3ApplicationControl{
		FIR: (b & 0b10000000) != 0,
		FIN: (b & 0b01000000) != 0,
		CON: (b & 0b00100000) != 0,
		UNS: (b & 0b00010000) != 0,
		SEQ: (b & 0b00001111),
	}
}

func NewDNP3ApplicationIIN(lsb, msb byte) DNP3ApplicationIIN {
	return DNP3ApplicationIIN{
		// IIN 1
		AllStations:   (lsb & 0b00000001) != 0,
		Class1Events:  (lsb & 0b00000010) != 0,
		Class2Events:  (lsb & 0b00000100) != 0,
		Class3Events:  (lsb & 0b00001000) != 0,
		NeedTime:      (lsb & 0b00010000) != 0,
		Local:         (lsb & 0b00100000) != 0,
		DeviceTrouble: (lsb & 0b01000000) != 0,
		Restart:       (lsb & 0b10000000) != 0,
		// IIN 2
		BadFunction:      (msb & 0b00000001) != 0,
		ObjectUnknown:    (msb & 0b00000010) != 0,
		ParameterError:   (msb & 0b00000100) != 0,
		BufferOverflow:   (msb & 0b00001000) != 0,
		AlreadyExiting:   (msb & 0b00010000) != 0,
		BadConfiguration: (msb & 0b00100000) != 0,
		Reserved1:        (msb & 0b01000010) != 0,
		Reserved2:        (msb & 0b10000000) != 0,
	}
}

func (d DNP3ApplicationRequest) ToBytes() []byte {
	return append(append(d.CTL.ToBytes(), d.FC), d.Raw[:]...)
}

func (d DNP3ApplicationResponse) ToBytes() []byte {
	var out []byte

	out = append(out, d.CTL.ToBytes()[:]...)
	out = append(out, d.FC)
	out = append(out, d.IIN.ToBytes()[:]...)
	out = append(out, d.OBJ[:]...)

	return out
}

func (d *DNP3ApplicationControl) ToBytes() []byte {
	var b byte = 0

	if d.FIR {
		b |= 0b10000000
	}
	if d.FIN {
		b |= 0b01000000
	}
	if d.CON {
		b |= 0b00100000
	}
	if d.UNS {
		b |= 0b00010000
	}

	b |= (d.SEQ & 0b00001111)

	return []byte{b}
}

func (d *DNP3ApplicationIIN) ToBytes() []byte {
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
		msb |= 0b01000010
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

func (d DNP3ApplicationRequest) String() string {
	return fmt.Sprintf(`
	Application (Request):
		%s
		FC : %d
		Raw: 0x % X`,
		d.CTL.String(), d.FC, d.Raw)
}

func (d DNP3ApplicationResponse) String() string {
	return fmt.Sprintf(`
	Application (Response):
		%s
		FC : %d
		%s 
		OBJ: 0x % X`,
		d.CTL.String(), d.FC, d.IIN.String(), d.OBJ)
}

func (d DNP3ApplicationControl) String() string {
	return fmt.Sprintf(`CTL:
			FIR: %t
			FIN: %t
			CON: %t
			UNS: %t
			SEQ: %d`,
		d.FIR, d.FIN, d.CON, d.UNS, d.SEQ)
}

func (d DNP3ApplicationIIN) String() string {
	return fmt.Sprintf(`IIN:
			IIN1
			AllStations     : %t
			Class1Events    : %t
			Class2Events    : %t
			Class3Events    : %t
			NeedTime        : %t
			Local           : %t
			DeviceTrouble   : %t
			Restart         : %t
			IIN2
			BadFunction     : %t
			ObjectUnknown   : %t
			ParameterError  : %t
			BufferOverflow  : %t
			AlreadyExiting  : %t
			BadConfiguration: %t
			Reserved1       : %t
			Reserved2       : %t`,
		d.AllStations, d.Class1Events, d.Class2Events, d.Class3Events,
		d.NeedTime, d.Local, d.DeviceTrouble, d.Restart, d.BadFunction,
		d.ObjectUnknown, d.ParameterError, d.BufferOverflow, d.AlreadyExiting,
		d.BadConfiguration, d.Reserved1, d.Reserved2)
}

func (d DNP3ApplicationRequest) LayerPayload() []byte {
	return d.Raw
}

func (d DNP3ApplicationResponse) LayerPayload() []byte {
	return d.OBJ
}

func (d *DNP3ApplicationRequest) SetCTL(c DNP3ApplicationControl) {
	d.CTL = c
}

func (d *DNP3ApplicationResponse) SetCTL(c DNP3ApplicationControl) {
	d.CTL = c
}

func (d *DNP3ApplicationRequest) SetSequence(s uint8) error {
	if s >= 16 {
		return fmt.Errorf("application sequence is only 4 bytes, got %d", s)
	}
	d.CTL.SEQ = s
	return nil
}

func (d *DNP3ApplicationResponse) SetSequence(s uint8) error {
	if s >= 16 {
		return fmt.Errorf("application sequence is only 4 bytes, got %d", s)
	}
	d.CTL.SEQ = s
	return nil
}

func (d *DNP3ApplicationRequest) SetContents(data []byte) {
	d.Raw = data
}

func (d *DNP3ApplicationResponse) SetContents(data []byte) {
	d.OBJ = data
}

// HACK to prevent other types from getting added to DNP3.Application
func (d DNP3ApplicationRequest) IsDNP3Application() bool {
	return true
}

// HACK to prevent other types from getting added to DNP3.Application
func (d DNP3ApplicationResponse) IsDNP3Application() bool {
	return true
}
