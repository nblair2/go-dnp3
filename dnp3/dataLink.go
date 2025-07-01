package dnp3

import (
	"encoding/binary"
	"fmt"
	"slices"
)

// DNP3DataLink is the highest layer of DNP3. Each DNP3 frame starts with a
// data link header (8 bytes, 2 byte CRC)
type DNP3DataLink struct {
	FRM [2]byte
	LEN byte
	CTL DNP3DataLinkControl
	DST uint16
	SRC uint16
	CRC [2]byte
}

// DNP3DataLinkControl is the 4th byte of the data link header.
type DNP3DataLinkControl struct {
	DIR bool
	PRM bool
	FCB bool
	FCV bool
	FC  uint8 //only 4 bits
}

// NewDNP3DataLink parses and builds the DNP3 data link header based on the
// first 10 bytes of a byte slice passed to it
func NewDNP3DataLink(data []byte) (DNP3DataLink, error) {
	if data[0] != 0x05 || data[1] != 0x64 {
		return DNP3DataLink{}, fmt.Errorf(
			"first 2 bytes %#X don't match the magic bytes (0x0564)",
			data[:2])
	}
	crc := CalculateDNP3CRC(data[:8])
	if slices.Equal(crc, data[8:10]) {
		return DNP3DataLink{}, fmt.Errorf(
			"data link checksum %#X doesn't match CRC (%#X)", crc, data[8:10])
	}

	d := DNP3DataLink{
		LEN: data[2],
		CTL: NewDNP3DataLinkControl(data[3]),
		DST: binary.LittleEndian.Uint16(data[4:6]),
		SRC: binary.LittleEndian.Uint16(data[6:8]),
		CRC: [2]byte{data[8], data[9]},
	}

	return d, nil
}

// NewDNP3DataLinkControl parses and builds the CTL byte of the DNP3DataLink
// header.
func NewDNP3DataLinkControl(b byte) DNP3DataLinkControl {
	return DNP3DataLinkControl{
		DIR: (b & 0b10000000) != 0,
		PRM: (b & 0b01000000) != 0,
		FCB: (b & 0b00100000) != 0,
		FCV: (b & 0b00010000) != 0,
		FC:  (b & 0b00001111),
	}
}

func (d DNP3DataLink) String() string {
	return fmt.Sprintf(`
	Data Link:
		FRM: 0x % X
		LEN: %d
		%s
		DST: %d
		SRC: %d
		CRC: 0x % X`,
		d.FRM, d.LEN, d.CTL.String(), d.DST, d.SRC, d.CRC)
}

func (d DNP3DataLinkControl) String() string {
	return fmt.Sprintf(`CTL:
			DIR: %t
			PRM: %t
			FCB: %t
			FCV: %t
			FC : %d`,
		d.DIR, d.PRM, d.FCB, d.FCV, d.FC)
}
