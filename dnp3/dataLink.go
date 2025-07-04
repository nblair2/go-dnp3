package dnp3

import (
	"encoding/binary"
	"fmt"
	"slices"
)

// DNP3DataLink is the highest layer of DNP3. Each DNP3 frame starts with a
// data link header (8 bytes, 2 byte CRC)
type DNP3DataLink struct {
	SYN [2]byte
	LEN uint16
	CTL DNP3DataLinkControl
	DST uint16
	SRC uint16
	CRC [2]byte
}

// DNP3DataLinkControl is the 4th byte of the data link header.
type DNP3DataLinkControl struct {
	DIR bool
	PRM bool
	FCB bool  // ignoring checks to enforce this to 0
	FCV bool  // ignoring use of DFC
	FC  uint8 //only 4 bits
}

func NewDNP3DataLink(data []byte) (DNP3DataLink, error) {
	if data[0] != 0x05 || data[1] != 0x64 {
		return DNP3DataLink{}, fmt.Errorf(
			"first 2 bytes %#X don't match the magic bytes (0x0564)",
			data[:2])
	}
	crc := CalculateDNP3CRC(data[:8])
	if !slices.Equal(crc, data[8:10]) {
		return DNP3DataLink{}, fmt.Errorf(
			"data link checksum %#X doesn't match CRC (%#X)", crc, data[8:10])
	}

	d := DNP3DataLink{
		SYN: [2]byte{0x05, 0x64},
		LEN: uint16(data[2]),
		CTL: NewDNP3DataLinkControl(data[3]),
		DST: binary.LittleEndian.Uint16(data[4:6]),
		SRC: binary.LittleEndian.Uint16(data[6:8]),
		CRC: [2]byte{data[8], data[9]},
	}

	return d, nil
}

func NewDNP3DataLinkControl(b byte) DNP3DataLinkControl {
	return DNP3DataLinkControl{
		DIR: (b & 0b10000000) != 0,
		PRM: (b & 0b01000000) != 0,
		FCB: (b & 0b00100000) != 0,
		FCV: (b & 0b00010000) != 0,
		FC:  (b & 0b00001111),
	}
}

func (d *DNP3DataLink) ToBytes() []byte {
	var out []byte

	// Set SYN bytes (in case we initialized an empty packet)
	d.SYN = [2]byte{0x05, 0x64}

	// LEN needs to be updated externally (application len)

	out = append(out, d.SYN[:]...)
	out = append(out, byte(d.LEN))
	out = append(out, d.CTL.ToBytes()[:]...)
	out = binary.LittleEndian.AppendUint16(out, d.DST)
	out = binary.LittleEndian.AppendUint16(out, d.SRC)

	d.CRC = [2]byte(CalculateDNP3CRC(out))
	out = append(out, d.CRC[:]...)

	return out
}

func (d *DNP3DataLinkControl) ToBytes() []byte {
	var b byte = 0

	if d.DIR {
		b |= 0b10000000
	}
	if d.PRM {
		b |= 0b01000000
	}
	if d.FCB {
		b |= 0b00100000
	}
	if d.FCV {
		b |= 0b00010000
	}

	b |= (d.FC & 0b00001111)

	return []byte{b}
}

func (d DNP3DataLink) String() string {
	return fmt.Sprintf(`
	Data Link:
		SYN: 0x % X
		LEN: %d
		%s
		DST: %d
		SRC: %d
		CRC: 0x % X`,
		d.SYN, d.LEN, d.CTL.String(), d.DST, d.SRC, d.CRC)
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
