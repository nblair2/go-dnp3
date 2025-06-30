// DNP3 (Distributed Network Protocol version 3) is a SCADA protocol used in
// industrial automation, especially electric power and water services in
// North America. See dnp.org, IEEE-1815.
// The protocol consists of three layers: A data link layer, a transport layer,
// and an application layer.
package dnp3

import (
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/google/gopacket"
)

// DNP3 holds each of the three layers.
type DNP3 struct {
	DataLink    DNP3DataLink
	Transport   DNP3Transport
	Application DNP3Application
	raw         []byte
}

// DNP3DataLink is the highest layer of DNP3. Each DNP3 frame starts with a
// data link header (8 bytes, 2 byte CRC).
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

// DNP3Transport is the second layer of DNP3, and allows for fragmentation
// and subsequent reassembly of application data. In addition to this header,
// DNP3Transport also intersperses CRC checksums after every 16 bytes.
type DNP3Transport struct {
	FIN bool
	FIR bool
	SEQ uint8 // only 6 bits
}

// DNP3Application is the lowest layer of DNP3, and carries all of the
// data. The application layer is complex.
type DNP3Application struct {
	payload []byte
}

// DNP3Type (required by gopacket)
var DNP3Type = gopacket.RegisterLayerType(20000,
	gopacket.LayerTypeMetadata{
		Name:    "DNP3",
		Decoder: gopacket.DecodeFunc(decodeDNP3),
	},
)

// DNP3Layer type (required by gopacket)
func (d DNP3) LayerType() gopacket.LayerType {
	return DNP3Type
}

// All DNP3 layer bytes (required by gopacket)
func (d DNP3) LayerContents() []byte {
	return d.raw
}

// DNP3 application bytes (required by gopacket)
func (d DNP3) LayerPayload() []byte {
	return d.Application.payload
}

// helper to bridge gopacket and DecodeFromBytes
func decodeDNP3(data []byte, p gopacket.PacketBuilder) error {
	var d DNP3
	if err := d.DecodeFromBytes(data); err != nil {
		return fmt.Errorf("decoding DNP3 from bytes: %w", err)
	}

	p.AddLayer(&d)
	return nil
}

// DecodeFromBytes creates a DNP3 object with appropriate layers based on the
// bytes slice passed to it. Needs at least 10 bytes.
func (d *DNP3) DecodeFromBytes(data []byte) error {
	if len(data) < 10 {
		return fmt.Errorf(
			"not DNP3, only got %d bytes (need at least 10)", len(data))
	}

	var err error
	d.raw = data
	if d.DataLink, err = NewDNP3DataLink(data[:10]); err != nil {
		return fmt.Errorf("can't create DNP3DataLink layer: %w", err)
	}

	// No transport or application
	if len(data) == 10 {
		return nil
	}

	d.Transport = NewDNP3Transport(data[10])

	// No application?
	if len(data) == 11 {
		return nil
	}

	if d.Application, err = NewDNP3Application(data[11:]); err != nil {
		return fmt.Errorf("can't create DNP3Application layer: %w", err)
	}

	return nil
}

// NewDNP3DataLink parses and builds the DNP3 data link header based on the
// first 10 bytes of a byte slice passed to it.
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

// NewDNP3Transport parses and builds the DNP3 Transport header.
func NewDNP3Transport(b byte) DNP3Transport {
	return DNP3Transport{
		FIN: (b & 0b10000000) != 0,
		FIR: (b & 0b01000000) != 0,
		SEQ: (b & 0b00111111),
	}
}

// NewDNP3Application stores the remaining DNP3 data in a byte slice
func NewDNP3Application(data []byte) (DNP3Application, error) {
	return DNP3Application{
			payload: data,
		},
		nil
}

func (d DNP3) String() string {
	return "DNP3:" +
		d.DataLink.String() +
		d.Transport.String() +
		d.Application.String()
}

func (d DNP3DataLink) String() string {
	return fmt.Sprintf(`
	Data Link:
		FRM: 0x % X
		LEN: %d
		CTL: %s
		DST: %d
		SRC: %d
		CRC: 0x % X`,
		d.FRM, d.LEN, d.CTL.String(), d.DST, d.SRC, d.CRC)
}

func (d DNP3DataLinkControl) String() string {
	return fmt.Sprintf(`
			DIR: %t
			PRM: %t
			FCB: %t
			FCV: %t
			FC : %d`,
		d.DIR, d.PRM, d.FCB, d.FCV, d.FC)
}

func (d DNP3Transport) String() string {
	return fmt.Sprintf(`
	Transport:
		FIN: %t
		FIR: %t
		SEQ: %d`,
		d.FIN, d.FIR, d.SEQ)
}

func (d DNP3Application) String() string {
	return fmt.Sprintf(`
	Application:
		payload: 0x % X`,
		d.payload)
}
