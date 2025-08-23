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
var LayerTypeDNP3 = gopacket.RegisterLayerType(20000,
	gopacket.LayerTypeMetadata{
		Name:    "DNP3",
		Decoder: gopacket.DecodeFunc(decodeDNP3),
	},
)

// DNP3Layer type (required by gopacket)
func (DNP3) LayerType() gopacket.LayerType {
	return LayerTypeDNP3
}

// All DNP3 layer bytes (required by gopacket)
func (dnp3 DNP3) LayerContents() []byte {
	b, err := dnp3.ToBytes()
	if err != nil {
		fmt.Printf("error encoding DNP3: %v\n", err)
		return nil
	}
	return b
}

// DNP3 application object bytes (required by gopacket)
func (dnp3 DNP3) LayerPayload() []byte {
	b, err := dnp3.Application.ToBytes()
	if err != nil {
		fmt.Printf("error encoding DNP3 application: %v\n", err)
		return nil
	}
	return b
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
func (d *DNP3) FromBytes(data []byte) error {
	var (
		err   error
		clean []byte
	)
	if len(data) < 10 {
		return fmt.Errorf("not DNP3, only got %d bytes (need at least 10)",
			len(data))
	}

	if err = d.DataLink.FromBytes(data[:10]); err != nil {
		return fmt.Errorf("error in DNP3 DataLink layer: %w", err)
	}

	// No transport or application
	if len(data) == 10 {
		return nil
	}

	// Make transport and remove CRCs
	if clean, err = d.Transport.FromBytes(data[10:]); err != nil {
		return fmt.Errorf("error in DNP3 Transport layer: %w", err)
	}

	// No application?
	if len(clean) <= 0 {
		return nil
	}

	if d.DataLink.CTL.DIR {
		d.Application = &ApplicationRequest{}
	} else {
		d.Application = &ApplicationResponse{}
	}
	if err := d.Application.FromBytes(clean); err != nil {
		return fmt.Errorf("error in DNP3 Application layer: %w", err)
	}

	return nil
}

// ToBytes assembles the DNP3 packet as bytes, in order. It also performs some
// updates to ensure the SYN, LEN, and CRCs are set correctly based on the
// current data
func (d *DNP3) ToBytes() ([]byte, error) {
	var ta []byte

	// get these first, for LEN in DL
	ta = append(ta, d.Transport.ToByte())
	// Application isn't always set
	if d.Application != nil {
		b, err := d.Application.ToBytes()
		if err != nil {
			return ta, fmt.Errorf("error encoding application data: %v", err)
		}
		ta = append(ta, b...)
	}
	// len is 5 more bytes in DL, excludes CRCs
	d.DataLink.LEN = uint16(len(ta) + 5)

	ta = InsertDNP3CRCs(ta)

	return append(d.DataLink.ToBytes(), ta[:]...), nil
}

// String outputs the DNP3 packet as an indented string
func (d *DNP3) String() string {
	return fmt.Sprintf("DNP3:%s%s%s",
		d.DataLink.String(),
		d.Transport.String(),
		d.Application.String())
}
