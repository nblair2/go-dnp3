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
	"github.com/google/gopacket/layers"
)

type Frame struct {
	DataLink    DataLink    `json:"data_link"`
	Transport   Transport   `json:"transport"`
	Application Application `json:"application"`

	// contents caches the on-wire bytes captured during DecodeFromBytes so
	// LayerContents can return them without re-encoding.
	contents []byte `json:"-"`
}

// Compile-time interface assertions for gopacket compliance.
var (
	_ gopacket.Layer             = (*Frame)(nil)
	_ gopacket.DecodingLayer     = (*Frame)(nil)
	_ gopacket.SerializableLayer = (*Frame)(nil)
	_ gopacket.ApplicationLayer  = (*Frame)(nil)
)

// NewFrame returns a new Frame ready to be populated via DecodeFromBytes or by
// setting fields directly. The Application field is nil until populated; it is
// set automatically to the appropriate type (ApplicationRequest or
// ApplicationResponse) when DecodeFromBytes is called.
func NewFrame() *Frame {
	return &Frame{}
}

// NewFrameFromBytes returns a new Frame parsed from the given bytes.
func NewFrameFromBytes(data []byte) (*Frame, error) {
	frame := &Frame{}

	err := frame.DecodeFromBytes(data, gopacket.NilDecodeFeedback)
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

// init registers DNP3 as the auto-decoder for TCP and UDP port 20000, the
// well-known DNP3-over-IP port per IEEE-1815-2012. With this in place,
// gopacket.NewPacket will surface a *dnp3.Frame layer automatically for
// matching packets.
//
//nolint:gochecknoinits // gopacket port registration must run at package init so stream parsers auto-decode DNP3.
func init() {
	layers.RegisterTCPPortLayerType(20000, LayerTypeDNP3)
	layers.RegisterUDPPortLayerType(20000, LayerTypeDNP3)
}

// LayerType returns the gopacket layer type for DNP3 (required by gopacket).
func (*Frame) LayerType() gopacket.LayerType {
	return LayerTypeDNP3
}

// LayerContents returns the on-wire bytes captured during DecodeFromBytes.
// Returns nil for frames built manually (use SerializeTo to encode those).
func (dnp *Frame) LayerContents() []byte {
	return dnp.contents
}

// LayerPayload returns nil. DNP3 is a terminal application-layer protocol;
// nothing is nested inside it from gopacket's perspective. Use the
// Application accessor to reach into protocol-internal application data.
func (*Frame) LayerPayload() []byte {
	return nil
}

// Payload implements gopacket.ApplicationLayer. DNP3 is the terminal
// application protocol, so it returns nil.
func (*Frame) Payload() []byte {
	return nil
}

// CanDecode implements gopacket.DecodingLayer.
//
//nolint:ireturn // gopacket.DecodingLayer.CanDecode requires returning the LayerClass interface.
func (*Frame) CanDecode() gopacket.LayerClass {
	return LayerTypeDNP3
}

// NextLayerType implements gopacket.DecodingLayer. DNP3 is terminal.
func (*Frame) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypeZero
}

// helper to bridge gopacket and DecodeFromBytes.
func decodeDNP3(data []byte, builder gopacket.PacketBuilder) error {
	decoded := &Frame{}

	err := decoded.DecodeFromBytes(data, builder)
	if err != nil {
		return fmt.Errorf("decoding DNP3 from bytes: %w", err)
	}

	builder.AddLayer(decoded)
	builder.SetApplicationLayer(decoded)

	return nil
}

// DecodeFromBytes parses a DNP3 frame from data, populating dnp. It implements
// gopacket.DecodingLayer. If data is shorter than the frame's declared wire
// size, df.SetTruncated() is called before the error is returned.
func (dnp *Frame) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	total, err := dnp.checkFrameBounds(data, df)
	if err != nil {
		return err
	}

	// Cache only the frame's wire-size slice so concatenated frames don't
	// leak into LayerContents.
	dnp.contents = append([]byte(nil), data[:total]...)

	err = dnp.DataLink.DecodeFromBytes(data[:10])
	if err != nil {
		return fmt.Errorf("error in DNP3 DataLink layer: %w", err)
	}

	// No transport or application
	if total == 10 {
		return nil
	}

	return dnp.decodeTransportAndApplication(data[10:total])
}

// SerializeTo implements gopacket.SerializableLayer. It assembles the DNP3
// packet (DataLink + Transport + Application), recomputes DataLink.Length
// from the payload, inserts the per-block DNP3 CRCs, and prepends the result
// onto b. DNP3 CRCs are mandatory by protocol; opts.ComputeChecksums=false is
// ignored. opts.FixLengths is honored implicitly since the length is always
// recomputed from the current payload.
func (dnp *Frame) SerializeTo(buf gopacket.SerializeBuffer, _ gopacket.SerializeOptions) error {
	var transportApplication []byte

	// get these first, for LEN in DL
	transportByte, err := dnp.Transport.ToByte()
	if err != nil {
		return fmt.Errorf("error encoding transport header: %w", err)
	}

	transportApplication = append(transportApplication, transportByte)
	// Application isn't always set
	if dnp.Application != nil {
		applicationBytes, err := dnp.Application.SerializeTo()
		if err != nil {
			return fmt.Errorf("error encoding application data: %w", err)
		}

		transportApplication = append(transportApplication, applicationBytes...)
	}
	// len is 5 more bytes in DL, excludes CRCs

	payloadLength := len(transportApplication)
	totalLength := payloadLength + 5

	if totalLength > math.MaxUint16 {
		return fmt.Errorf("transport/application payload too large: %d bytes", payloadLength)
	}

	// #nosec G115 -- guarded by range check above
	dnp.DataLink.Length = uint16(totalLength)

	transportApplication = InsertDNP3CRCs(transportApplication)

	dlBytes, err := dnp.DataLink.SerializeTo()
	if err != nil {
		return fmt.Errorf("error encoding data link: %w", err)
	}

	dst, err := buf.PrependBytes(len(dlBytes) + len(transportApplication))
	if err != nil {
		return fmt.Errorf("prepending DNP3 bytes: %w", err)
	}

	copy(dst, dlBytes)
	copy(dst[len(dlBytes):], transportApplication)

	return nil
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

// checkFrameBounds validates that data contains a complete DNP3 frame and
// returns its total wire size. Calls df.SetTruncated() when data is short.
func (*Frame) checkFrameBounds(data []byte, feedback gopacket.DecodeFeedback) (int, error) {
	if len(data) < 10 {
		feedback.SetTruncated()

		return 0, fmt.Errorf("not DNP3, only got %d bytes (need at least 10)",
			len(data))
	}

	total := frameWireSize(data[2])
	if total == 0 {
		return 0, fmt.Errorf("invalid DNP3 length byte: %d", data[2])
	}

	if len(data) < total {
		feedback.SetTruncated()

		return 0, fmt.Errorf("truncated DNP3 frame: have %d bytes, need %d",
			len(data), total)
	}

	return total, nil
}

// decodeTransportAndApplication parses the post-header portion of a frame.
// data must already be sliced to a single frame's wire bytes (excluding the
// 10-byte DataLink header).
func (dnp *Frame) decodeTransportAndApplication(data []byte) error {
	// Slice transport bytes to the payload boundary so a second frame in the
	// same buffer cannot corrupt CRC validation.
	payloadLen := int(dnp.DataLink.Length) - 5
	numBlocks := (payloadLen + 15) / 16
	framePayloadBytes := payloadLen + numBlocks*2

	transportData := data
	if payloadLen > 0 && len(transportData) > framePayloadBytes {
		transportData = transportData[:framePayloadBytes]
	}

	clean, err := dnp.Transport.DecodeFromBytes(transportData)
	if err != nil {
		return fmt.Errorf("error in DNP3 Transport layer: %w", err)
	}

	if len(clean) == 0 {
		return nil
	}

	if dnp.DataLink.Control.Direction {
		dnp.Application = &ApplicationRequest{}
	} else {
		dnp.Application = &ApplicationResponse{}
	}

	err = dnp.Application.DecodeFromBytes(clean)
	if err != nil {
		return fmt.Errorf("error in DNP3 Application layer: %w", err)
	}

	return nil
}
