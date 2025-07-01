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
	DataLink    DNP3DataLink
	Transport   DNP3Transport
	Application DNP3Application
	raw         []byte
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

// DNP3 application object bytes (required by gopacket)
func (d DNP3) LayerPayload() []byte {
	return d.Application.OBJ
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

	d.Application = NewDNP3Application(data[11:])

	return nil
}

func (d DNP3) String() string {
	return "DNP3:" +
		d.DataLink.String() +
		d.Transport.String() +
		d.Application.String()
}
