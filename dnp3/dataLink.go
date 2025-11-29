package dnp3

import (
	"encoding/binary"
	"fmt"
	"slices"
)

// DataLink is the highest layer of DNP3. Each DNP3 frame starts with a
// data link header (8 bytes, 2 byte CRC).
type DataLink struct {
	SYN [2]byte
	LEN uint16
	CTL DataLinkCTL
	DST uint16
	SRC uint16
	CRC [2]byte
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

	dl.SYN = [2]byte{0x05, 0x64}
	dl.LEN = uint16(data[2])
	dl.CTL.FromByte(data[3])
	dl.DST = binary.LittleEndian.Uint16(data[4:6])
	dl.SRC = binary.LittleEndian.Uint16(data[6:8])
	dl.CRC = [2]byte{data[8], data[9]}

	return nil
}

func (dl *DataLink) ToBytes() ([]byte, error) {
	var out []byte

	// Set SYN bytes (in case we initialized an empty packet)
	dl.SYN = [2]byte{0x05, 0x64}

	// LEN needs to be updated externally
	if dl.LEN > 255 {
		return nil, fmt.Errorf("length %d exceeds max byte value", dl.LEN)
	}

	out = append(out, dl.SYN[:]...)
	out = append(out, byte(dl.LEN))
	out = append(out, dl.CTL.ToByte())
	out = binary.LittleEndian.AppendUint16(out, dl.DST)
	out = binary.LittleEndian.AppendUint16(out, dl.SRC)

	dl.CRC = [2]byte(CalculateDNP3CRC(out))
	out = append(out, dl.CRC[:]...)

	return out, nil
}

func (dl *DataLink) String() string {
	return fmt.Sprintf(`Data Link:
	SYN: 0x % X
	LEN: %d
	%s
	DST: %d
	SRC: %d`,
		dl.SYN, dl.LEN, indent(dl.CTL.String(), "\t"), dl.DST, dl.SRC)
}

// DataLinkControl is the 4th byte of the data link header.
type DataLinkCTL struct {
	DIR bool
	PRM bool
	FCB bool              // ignoring checks to enforce dl to 0
	FCV bool              // ignoring use of DFC
	FC  DataLinkPrimaryFC // only 4 bits
}

func (dlctl *DataLinkCTL) FromByte(value byte) {
	dlctl.DIR = (value & 0b10000000) != 0
	dlctl.PRM = (value & 0b01000000) != 0
	dlctl.FCB = (value & 0b00100000) != 0
	dlctl.FCV = (value & 0b00010000) != 0
	dlctl.FC = DataLinkPrimaryFC(value & 0b00001111)
}

func (dlctl *DataLinkCTL) ToByte() byte {
	var controlByte byte

	if dlctl.DIR {
		controlByte |= 0b10000000
	}

	if dlctl.PRM {
		controlByte |= 0b01000000
	}

	if dlctl.FCB {
		controlByte |= 0b00100000
	}

	if dlctl.FCV {
		controlByte |= 0b00010000
	}

	controlByte |= (uint8(dlctl.FC) & 0b00001111)

	return controlByte
}

func (dlctl *DataLinkCTL) String() string {
	return fmt.Sprintf(`CTL:
	DIR: %t
	PRM: %t
	FCB: %t
	FCV: %t
	FC : (%d) %s`,
		dlctl.DIR, dlctl.PRM, dlctl.FCB, dlctl.FCV, dlctl.FC,
		dlctl.FC.String())
}

// Function Codes (PRM set).
//
//go:generate stringer -type=DataLinkPrimaryFC
type DataLinkPrimaryFC uint8

const (
	ResetLinkStates     DataLinkPrimaryFC = 0x0 // FCV 0
	TestLinkStates      DataLinkPrimaryFC = 0x2 //     1
	ConfirmedUserData   DataLinkPrimaryFC = 0x3 //     1
	UnconfirmedUserData DataLinkPrimaryFC = 0x4 //     0
	RequestLinkStatus   DataLinkPrimaryFC = 0x9 //     0
)

// Function Codes (PRM unset).
//
//go:generate stringer -type=DataLinkSecondaryFC
type DataLinkSecondaryFC uint8

const (
	Ack          DataLinkSecondaryFC = 0x0
	Nack         DataLinkSecondaryFC = 0x1
	LinkStatus   DataLinkSecondaryFC = 0xb
	NotSupported DataLinkSecondaryFC = 0xf
)
