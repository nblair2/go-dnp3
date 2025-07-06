// partial implementation of DNP3 for gopacket
package dnp3

import (
	"fmt"

	"github.com/google/gopacket"
)

// DNP3 (Distributed Network Protocol version 3) is a SCADA protocol used in
// industrial automation, especially electric power and water services in
// North America. See dnp.org, IEEE-1815.
// The protocol consists of three layers: A data link layer, a transport layer,
// and an application layer.
type DNP3 struct {
	DataLink    DataLink
	Transport   Transport
	Application Application
}

// DNP3Type (required by gopacket)
var DNP3Type = gopacket.RegisterLayerType(20000,
	gopacket.LayerTypeMetadata{
		Name:    "DNP3",
		Decoder: gopacket.DecodeFunc(decodeDNP3),
	},
)

// DNP3Layer type (required by gopacket)
func (DNP3) LayerType() gopacket.LayerType {
	return DNP3Type
}

// All DNP3 layer bytes (required by gopacket)
func (dnp3 DNP3) LayerContents() []byte {
	return dnp3.ToBytes()
}

// DNP3 application object bytes (required by gopacket)
func (dnp3 DNP3) LayerPayload() []byte {
	return dnp3.Application.ToBytes()
}

// helper to bridge gopacket and FromBytes
func decodeDNP3(data []byte, p gopacket.PacketBuilder) error {
	var d DNP3
	if err := d.FromBytes(data); err != nil {
		return fmt.Errorf("decoding DNP3 from bytes: %w", err)
	}

	p.AddLayer(&d)
	return nil
}

// FromBytes creates a DNP3 object with appropriate layers based on the
// bytes slice passed to it. Needs at least 10 bytes.
func (dnp3 *DNP3) FromBytes(data []byte) error {
	var (
		err   error
		clean []byte
	)
	if len(data) < 10 {
		return fmt.Errorf("not DNP3, only got %d bytes (need at least 10)",
			len(data))
	}

	if err = dnp3.DataLink.FromBytes(data[:10]); err != nil {
		return fmt.Errorf("can't create DNP3 DataLink layer: %w", err)
	}

	// No transport or application
	if len(data) == 10 {
		return nil
	}

	if clean, err = dnp3.Transport.FromBytes(data[10:]); err != nil {
		return fmt.Errorf("can't create DNP3 Transport layer: %w", err)
	}

	// No application?
	if len(clean) <= 0 {
		return nil
	}

	if dnp3.DataLink.CTL.DIR {
		dnp3.Application = &ApplicationRequest{}
	} else {
		dnp3.Application = &ApplicationResponse{}
	}
	dnp3.Application.FromBytes(clean)

	return nil
}

// ToBytes assembles the DNP3 packet as bytes, in order. It also performs some
// updates to ensure the SYN, LEN, and CRCs are set correctly based on the
// current data
func (dnp3 *DNP3) ToBytes() []byte {
	var ta []byte

	// get these first, for LEN in DL
	ta = append(ta, dnp3.Transport.ToByte())
	ta = append(ta, dnp3.Application.ToBytes()...)
	// len is 5 more bytes in DL, excludes CRCs
	dnp3.DataLink.LEN = uint16(len(ta) + 5)

	ta = InsertDNP3CRCs(ta)

	return append(dnp3.DataLink.ToBytes(), ta[:]...)
}

// String outputs the DNP3 packet as an indented string
func (dnp3 *DNP3) String() string {
	return fmt.Sprintf("DNP3:%s%s%s",
		dnp3.DataLink.String(),
		dnp3.Transport.String(),
		dnp3.Application.String())
}

// DNP3 Application layer abstraction for different Request / Response
// structure
type Application interface {
	FromBytes([]byte) error
	ToBytes() []byte
	String() string
	SetSequence(uint8) error
	SetContents([]byte)
}
