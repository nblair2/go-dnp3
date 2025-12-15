package dnp3

import (
	"encoding/binary"
	"fmt"
	"slices"
)

// DataLink is the highest layer of DNP3. Each DNP3 frame starts with a
// data link header (8 bytes, 2 byte CRC).
type DataLink struct {
	Synchronize [2]byte         `json:"synchronize"`
	Length      uint16          `json:"length"`
	Control     DataLinkControl `json:"control"`
	Destination uint16          `json:"destination"`
	Source      uint16          `json:"source"`
	Checksum    [2]byte         `json:"checksum"`
}

func (dl *DataLink) FromBytes(data []byte) error {
	if data[0] != 0x05 || data[1] != 0x64 {
		return fmt.Errorf(
			"first 2 bytes %#X don't match the magic bytes (0x0564)", data[:2])
	}

	crc := CalculateDNP3CRC(data[:8])
	if !slices.Equal(crc, data[8:10]) {
		return fmt.Errorf(
			"data link checksum %#X doesn't match CRC (%#X)", crc, data[8:10])
	}

	dl.Synchronize = [2]byte{0x05, 0x64}

	dl.Length = uint16(data[2])

	err := dl.Control.FromByte(data[3])
	if err != nil {
		return err
	}

	dl.Destination = binary.LittleEndian.Uint16(data[4:6])
	dl.Source = binary.LittleEndian.Uint16(data[6:8])
	dl.Checksum = [2]byte{data[8], data[9]}

	return nil
}

func (dl *DataLink) ToBytes() ([]byte, error) {
	var out []byte

	// Set SYN bytes (in case we initialized an empty packet)
	dl.Synchronize = [2]byte{0x05, 0x64}

	// LEN needs to be updated externally
	if dl.Length > 255 {
		return nil, fmt.Errorf("length %d exceeds max byte value", dl.Length)
	}

	out = append(out, dl.Synchronize[:]...)
	out = append(out, byte(dl.Length))

	ctlByte, err := dl.Control.ToByte()
	if err != nil {
		return nil, err
	}

	out = append(out, ctlByte)
	out = binary.LittleEndian.AppendUint16(out, dl.Destination)
	out = binary.LittleEndian.AppendUint16(out, dl.Source)

	dl.Checksum = [2]byte(CalculateDNP3CRC(out))
	out = append(out, dl.Checksum[:]...)

	return out, nil
}

func (dl *DataLink) String() string {
	return fmt.Sprintf(`Data Link:
	SYN: 0x % X
	LEN: %d
	%s
	DST: %d
	SRC: %d`,
		dl.Synchronize, dl.Length, indent(dl.Control.String(), "\t"), dl.Destination, dl.Source)
}

// DataLinkControl is the 4th byte of the data link header.
type DataLinkControl struct {
	Direction       bool             `json:"direction"`
	Primary         bool             `json:"primary"`
	FrameCountBit   bool             `json:"frame_count_bit"`   // ignoring checks to enforce dl to 0
	FrameCountValid bool             `json:"frame_count_valid"` // ignoring use of DFC
	FunctionCode    DataLinkFunction `json:"function_code"`     // only 4 bits
}

func (dlctl *DataLinkControl) FromByte(value byte) error {
	dlctl.Direction = (value & 0b10000000) != 0
	dlctl.Primary = (value & 0b01000000) != 0
	dlctl.FrameCountBit = (value & 0b00100000) != 0
	dlctl.FrameCountValid = (value & 0b00010000) != 0

	functionCode := value & 0b00001111
	if dlctl.Primary {
		code := DataLinkPrimaryFunctionCode(functionCode)
		if !isValidPrimaryFunctionCode(code) {
			return fmt.Errorf("unknown primary function code 0x%X", functionCode)
		}

		if !checkPrimaryFunctionCodeFCVValidity(code, dlctl.FrameCountValid) {
			return fmt.Errorf(
				"invalid FCV value %t for primary function code 0x%X",
				dlctl.FrameCountValid,
				functionCode,
			)
		}

		dlctl.FunctionCode = code
	} else {
		code := DataLinkSecondaryFunctionCode(functionCode)
		if !isValidSecondaryFunctionCode(code) {
			return fmt.Errorf("unknown secondary function code 0x%X", functionCode)
		}

		dlctl.FunctionCode = code
	}

	return nil
}

func (dlctl *DataLinkControl) ToByte() (byte, error) {
	var controlByte byte

	if dlctl.Direction {
		controlByte |= 0b10000000
	}

	if dlctl.Primary {
		controlByte |= 0b01000000
	}

	if dlctl.FrameCountBit {
		controlByte |= 0b00100000
	}

	if dlctl.FrameCountValid {
		controlByte |= 0b00010000
	}

	if dlctl.FunctionCode != nil {
		controlByte |= (dlctl.FunctionCode.Byte() & 0b00001111)
	}

	return controlByte, nil
}

func (dlctl *DataLinkControl) String() string {
	var codeVal byte

	codeName := "n/a"

	if dlctl.FunctionCode != nil {
		codeVal = dlctl.FunctionCode.Byte()
		codeName = dlctl.FunctionCode.String()
	}

	return fmt.Sprintf(
		`CTL:
	DIR: %t
	PRM: %t
	FCB: %t
	FCV: %t
	FC : (%d) %s`,
		dlctl.Direction,
		dlctl.Primary,
		dlctl.FrameCountBit,
		dlctl.FrameCountValid,
		codeVal,
		codeName,
	)
}

type DataLinkFunction interface {
	fmt.Stringer
	Byte() byte
}

// DataLinkPrimaryFunctionCode - Function Codes for master to outsation
// (PRM set).
//
//go:generate stringer -type=DataLinkPrimaryFunctionCode
type DataLinkPrimaryFunctionCode byte

func (fc DataLinkPrimaryFunctionCode) Byte() byte {
	return byte(fc)
}

const (
	ResetLinkStates     DataLinkPrimaryFunctionCode = 0x0 // FCV 0
	TestLinkStates      DataLinkPrimaryFunctionCode = 0x2 //     1
	ConfirmedUserData   DataLinkPrimaryFunctionCode = 0x3 //     1
	UnconfirmedUserData DataLinkPrimaryFunctionCode = 0x4 //     0
	RequestLinkStatus   DataLinkPrimaryFunctionCode = 0x9 //     0
)

func isValidPrimaryFunctionCode(code DataLinkPrimaryFunctionCode) bool {
	switch code {
	case ResetLinkStates, TestLinkStates, ConfirmedUserData, UnconfirmedUserData, RequestLinkStatus:
		return true
	default:
		return false
	}
}

func checkPrimaryFunctionCodeFCVValidity(code DataLinkPrimaryFunctionCode, fcv bool) bool {
	switch code {
	case ResetLinkStates, UnconfirmedUserData, RequestLinkStatus:
		return !fcv
	case TestLinkStates, ConfirmedUserData:
		return fcv
	}

	return false
}

// DataLinkSecondaryFunctionCode - Function Codes for outstation to master
// (PRM unset).
//
//go:generate stringer -type=DataLinkSecondaryFunctionCode
type DataLinkSecondaryFunctionCode byte

func (fc DataLinkSecondaryFunctionCode) Byte() byte {
	return byte(fc)
}

const (
	Ack          DataLinkSecondaryFunctionCode = 0x0
	Nack         DataLinkSecondaryFunctionCode = 0x1
	LinkStatus   DataLinkSecondaryFunctionCode = 0xb
	NotSupported DataLinkSecondaryFunctionCode = 0xf
)

func isValidSecondaryFunctionCode(code DataLinkSecondaryFunctionCode) bool {
	switch code {
	case Ack, Nack, LinkStatus, NotSupported:
		return true
	default:
		return false
	}
}
