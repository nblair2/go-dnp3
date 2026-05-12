// Package dnp3 is a gopacket implementation of the  DNP3 (Distributed
// Network Protocol version 3) protocol. DNP3 is a SCADA protocol used in
// industrial automation, especially electric power and water services in
// North America. See dnp.org, IEEE-1815.
// The protocol consists of three layers: A data link layer, a transport layer,
// and an application layer.
package dnp3

import (
	"fmt"
	"math"

	"github.com/google/gopacket"
)

type Frame struct {
	DataLink    DataLink    `json:"data_link"`
	Transport   Transport   `json:"transport"`
	Application Application `json:"application"`
}

// NewFrame returns a new Frame ready to be populated via FromBytes or by
// setting fields directly. The Application field is nil until populated; it is
// set automatically to the appropriate type (ApplicationRequest or
// ApplicationResponse) when FromBytes is called.
func NewFrame() *Frame {
	return &Frame{}
}

// NewFrameFromBytes returns a new Frame parsed from the given bytes.
func NewFrameFromBytes(data []byte) (*Frame, error) {
	frame := &Frame{}

	err := frame.FromBytes(data)
	if err != nil {
		return nil, err
	}

	return frame, nil
}

// ParseFrames parses all complete DNP3 frames from data.
// It returns the parsed frames, any unconsumed trailing bytes
// (a partial frame), and the first error encountered.
// On error, frames parsed before the error are also returned.
func ParseFrames(data []byte) ([]*Frame, []byte, error) {
	var frames []*Frame

	pos := 0
	for pos < len(data) {
		remaining := data[pos:]
		if len(remaining) < 10 {
			return frames, remaining, nil
		}

		total := frameWireSize(remaining[2])
		if total == 0 {
			return frames, remaining, fmt.Errorf(
				"invalid DNP3 length byte %d at offset %d",
				remaining[2],
				pos,
			)
		}

		if len(remaining) < total {
			return frames, remaining, nil
		}

		frame, err := NewFrameFromBytes(remaining[:total])
		if err != nil {
			return frames, remaining, err
		}

		frames = append(frames, frame)
		pos += total
	}

	return frames, nil, nil
}

// frameWireSize returns the total number of on-wire bytes for a DNP3 frame
// whose DataLink Length field equals lengthByte. Returns 0 for invalid lengths.
func frameWireSize(lengthByte byte) int {
	if lengthByte < 5 {
		return 0
	}

	payloadLen := int(lengthByte) - 5
	numBlocks := (payloadLen + 15) / 16

	return 10 + payloadLen + numBlocks*2
}

// LayerTypeDNP3 is the gopacket layer type for DNP3 (required by gopacket).
var LayerTypeDNP3 = gopacket.RegisterLayerType(20000,
	gopacket.LayerTypeMetadata{
		Name:    "DNP3",
		Decoder: gopacket.DecodeFunc(decodeDNP3),
	},
)

// LayerType returns the gopacket layer type for DNP3 (required by gopacket).
func (*Frame) LayerType() gopacket.LayerType {
	return LayerTypeDNP3
}

// LayerContents returns all DNP3 layer bytes (required by gopacket).
func (dnp *Frame) LayerContents() []byte {
	encodedPacket, err := dnp.ToBytes()
	if err != nil {
		fmt.Printf("error encoding DNP3: %v\n", err)

		return nil
	}

	return encodedPacket
}

// LayerPayload returns the DNP3 application object bytes (required by gopacket).
func (dnp *Frame) LayerPayload() []byte {
	applicationPayload, err := dnp.Application.ToBytes()
	if err != nil {
		fmt.Printf("error encoding DNP3 application: %v\n", err)

		return nil
	}

	return applicationPayload
}

// helper to bridge gopacket and FromBytes.
func decodeDNP3(data []byte, builder gopacket.PacketBuilder) error {
	var decoded Frame

	err := decoded.FromBytes(data)
	if err != nil {
		return fmt.Errorf("decoding DNP3 from bytes: %w", err)
	}

	builder.AddLayer(&decoded)

	return nil
}

// FromBytes creates a DNP3 object with appropriate layers based on the
// bytes slice passed to it. Needs at least 10 bytes.
func (dnp *Frame) FromBytes(data []byte) error {
	var (
		err   error
		clean []byte
	)

	if len(data) < 10 {
		return fmt.Errorf("not DNP3, only got %d bytes (need at least 10)",
			len(data))
	}

	err = dnp.DataLink.FromBytes(data[:10])
	if err != nil {
		return fmt.Errorf("error in DNP3 DataLink layer: %w", err)
	}

	// No transport or application
	if len(data) == 10 {
		return nil
	}

	// Make transport and remove CRCs. Slice to the frame boundary so a second
	// frame in the same buffer cannot corrupt CRC validation.
	payloadLen := int(dnp.DataLink.Length) - 5
	numBlocks := (payloadLen + 15) / 16
	framePayloadBytes := payloadLen + numBlocks*2

	transportData := data[10:]
	if payloadLen > 0 && len(transportData) > framePayloadBytes {
		transportData = transportData[:framePayloadBytes]
	}

	clean, err = dnp.Transport.FromBytes(transportData)
	if err != nil {
		return fmt.Errorf("error in DNP3 Transport layer: %w", err)
	}

	// No application?
	if len(clean) == 0 {
		return nil
	}

	if dnp.DataLink.Control.Direction {
		dnp.Application = &ApplicationRequest{}
	} else {
		dnp.Application = &ApplicationResponse{}
	}

	err = dnp.Application.FromBytes(clean)
	if err != nil {
		return fmt.Errorf("error in DNP3 Application layer: %w", err)
	}

	return nil
}

// ToBytes assembles the DNP3 packet as bytes, in order. It also performs some
// updates to ensure the SYN, LEN, and CRCs are set correctly based on the
// current data.
func (dnp *Frame) ToBytes() ([]byte, error) {
	var transportApplication []byte

	// get these first, for LEN in DL
	transportByte, err := dnp.Transport.ToByte()
	if err != nil {
		return nil, fmt.Errorf("error encoding transport header: %w", err)
	}

	transportApplication = append(transportApplication, transportByte)
	// Application isn't always set
	if dnp.Application != nil {
		applicationBytes, err := dnp.Application.ToBytes()
		if err != nil {
			return transportApplication, fmt.Errorf("error encoding application data: %w", err)
		}

		transportApplication = append(transportApplication, applicationBytes...)
	}
	// len is 5 more bytes in DL, excludes CRCs

	payloadLength := len(transportApplication)
	totalLength := payloadLength + 5

	if totalLength > math.MaxUint16 {
		return nil, fmt.Errorf("transport/application payload too large: %d bytes", payloadLength)
	}

	// #nosec G115 -- guarded by range check above
	dnp.DataLink.Length = uint16(totalLength)

	transportApplication = InsertDNP3CRCs(transportApplication)

	dlBytes, err := dnp.DataLink.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("error encoding data link: %w", err)
	}

	return append(dlBytes, transportApplication...), nil
}

// String outputs the DNP3 packet as an indented string.
func (dnp *Frame) String() string {
	appString := ""
	if dnp.Application != nil {
		appString = indent(dnp.Application.String(), "\t")
	}

	return fmt.Sprintf("DNP3:\n%s\n%s\n%s",
		indent(dnp.DataLink.String(), "\t"),
		indent(dnp.Transport.String(), "\t"),
		appString)
}
