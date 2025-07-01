package dnp3

import "fmt"

// DNP3Application is the lowest layer of DNP3, and carries all of the
// data. The application layer is complex.
type DNP3Application struct {
	CTL DNP3ApplicationControl
	FC  byte
	IIN DNP3ApplicationIIN
	OBJ []byte // Store all the data here for now
}

// DNP3ApplicationControl is the first byte of the application layer, and
// is used to keep application messages synchronized across frames
type DNP3ApplicationControl struct {
	FIR bool
	FIN bool
	CON bool
	UNS bool
	SEQ uint8 //only 4 bits
}

// DNP3ApplicationIIN communicate information about the state of the device
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
	Reserved1        bool
	Reserved2        bool
}

// NewDNP3Application parses and builds the DNP3 Application layer
func NewDNP3Application(d []byte) DNP3Application {
	return DNP3Application{
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
		AllStations:   (lsb & 0b10000000) != 0,
		Class1Events:  (lsb & 0b01000000) != 0,
		Class2Events:  (lsb & 0b00100000) != 0,
		Class3Events:  (lsb & 0b00010000) != 0,
		NeedTime:      (lsb & 0b00001000) != 0,
		Local:         (lsb & 0b00000100) != 0,
		DeviceTrouble: (lsb & 0b00000010) != 0,
		Restart:       (lsb & 0b00000001) != 0,
		// IIN 2
		BadFunction:      (msb & 0b10000000) != 0,
		ObjectUnknown:    (msb & 0b01000000) != 0,
		ParameterError:   (msb & 0b00100000) != 0,
		BufferOverflow:   (msb & 0b00010000) != 0,
		AlreadyExiting:   (msb & 0b00001000) != 0,
		BadConfiguration: (msb & 0b00000100) != 0,
		Reserved1:        (msb & 0b00000010) != 0,
		Reserved2:        (msb & 0b00000001) != 0,
	}
}

func (d DNP3Application) String() string {
	return fmt.Sprintf(`
	Application:
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
