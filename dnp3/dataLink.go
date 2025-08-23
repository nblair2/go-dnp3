package dnp3

import (
	"encoding/binary"
	"fmt"
	"slices"
)

// DataLink is the highest layer of DNP3. Each DNP3 frame starts with a
// data link header (8 bytes, 2 byte CRC)
type DataLink struct {
	SYN [2]byte
	LEN uint16
	CTL DataLinkCTL
	DST uint16
	SRC uint16
	CRC [2]byte
}

func (dl *DataLink) FromBytes(d []byte) error {
	if d[0] != 0x05 || d[1] != 0x64 {
		return fmt.Errorf(
			"first 2 bytes %#X don't match the magic bytes (0x0564)", d[:2])
	}

	crc := CalculateDNP3CRC(d[:8])
	if !slices.Equal(crc, d[8:10]) {
		return fmt.Errorf(
			"data link checksum %#X doesn't match CRC (%#X)", crc, d[8:10])
	}

	dl.SYN = [2]byte{0x05, 0x64}
	dl.LEN = uint16(d[2])
	dl.CTL.FromByte(d[3])
	dl.DST = binary.LittleEndian.Uint16(d[4:6])
	dl.SRC = binary.LittleEndian.Uint16(d[6:8])
	dl.CRC = [2]byte{d[8], d[9]}

	return nil
}

func (d *DataLink) ToBytes() []byte {
	var out []byte

	// Set SYN bytes (in case we initialized an empty packet)
	d.SYN = [2]byte{0x05, 0x64}

	// LEN needs to be updated externally

	out = append(out, d.SYN[:]...)
	out = append(out, byte(d.LEN))
	out = append(out, d.CTL.ToByte())
	out = binary.LittleEndian.AppendUint16(out, d.DST)
	out = binary.LittleEndian.AppendUint16(out, d.SRC)

	d.CRC = [2]byte(CalculateDNP3CRC(out))
	out = append(out, d.CRC[:]...)

	return out
}

func (d DataLink) String() string {
	return fmt.Sprintf(`
	Data Link:
		SYN: 0x % X
		LEN: %d
		%s
		DST: %d
		SRC: %d`,
		d.SYN, d.LEN, d.CTL.String(), d.DST, d.SRC)
}

// DataLinkControl is the 4th byte of the data link header.
type DataLinkCTL struct {
	DIR bool
	PRM bool
	FCB bool              // ignoring checks to enforce dl to 0
	FCV bool              // ignoring use of DFC
	FC  DataLinkPrimaryFC //only 4 bits
}

func (dlctl *DataLinkCTL) FromByte(d byte) {
	dlctl.DIR = (d & 0b10000000) != 0
	dlctl.PRM = (d & 0b01000000) != 0
	dlctl.FCB = (d & 0b00100000) != 0
	dlctl.FCV = (d & 0b00010000) != 0
	dlctl.FC = DataLinkPrimaryFC(d & 0b00001111)
}

func (dlctl *DataLinkCTL) ToByte() byte {
	var o byte = 0

	if dlctl.DIR {
		o |= 0b10000000
	}
	if dlctl.PRM {
		o |= 0b01000000
	}
	if dlctl.FCB {
		o |= 0b00100000
	}
	if dlctl.FCV {
		o |= 0b00010000
	}

	o |= (uint8(dlctl.FC) & 0b00001111)

	return o
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

// Function Codes (PRM set)
type DataLinkPrimaryFC uint8

const (
	resetLinkStates     = 0x0 // FCV 0
	testLinkStates      = 0x2 //     1
	confirmedUserData   = 0x3 //     1
	unconfirmedUserData = 0x4 //     0
	requestLinkStatus   = 0x9 //     0
)

var DataLinkPrimaryFCNames = map[DataLinkPrimaryFC]string{
	resetLinkStates:     "RESET_LINK_STATES",
	testLinkStates:      "TEST_LINK_STATES",
	confirmedUserData:   "CONFIRMED_USER_DATA",
	unconfirmedUserData: "UNCONFIRMED_USER_DATA",
	requestLinkStatus:   "REQUEST_LINK_STATUS",
}

func (fc DataLinkPrimaryFC) String() string {
	if name, ok := DataLinkPrimaryFCNames[fc]; ok {
		return name
	}
	return fmt.Sprintf("unknown Function Code %d", fc)
}

// Function Codes (PRM unset)
type DataLinkSecondaryFC uint8

const (
	ack          = 0x0
	nack         = 0x1
	linkStatus   = 0xb
	notSupported = 0xf
)

var DataLinkSecondaryFCNames = map[DataLinkSecondaryFC]string{
	ack:          "ACK",
	nack:         "NACK",
	linkStatus:   "LINK_STATUS",
	notSupported: "NOT_SUPPORTED",
}

func (fc DataLinkSecondaryFC) String() string {
	if name, ok := DataLinkSecondaryFCNames[fc]; ok {
		return name
	}
	return fmt.Sprintf("unknown Function Code %d", fc)
}
