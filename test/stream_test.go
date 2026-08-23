package corpus_test

import (
	"net"
	"testing"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/nblair2/go-dnp3/v2/dnp3"
)

// twoSegmentWire is two link frames carrying one Read request split across
// two transport segments, as used by ExampleAssembler.
var twoSegmentWire = []byte{
	0x05, 0x64, 0x0a, 0xc4, 0x03, 0x00, 0x04, 0x00,
	0x08, 0xcf, 0x40, 0xc5, 0x01, 0x3c, 0x02, 0x6f,
	0xda, 0x05, 0x64, 0x10, 0xc4, 0x03, 0x00, 0x04,
	0x00, 0xa2, 0x0b, 0x81, 0x06, 0x3c, 0x03, 0x06,
	0x3c, 0x04, 0x06, 0x3c, 0x01, 0x06, 0x74, 0xa9,
}

// dnp3Stream turns one reassembled TCP stream into DNP3 application
// fragments.
type dnp3Stream struct {
	assembler dnp3.Assembler
	buffer    []byte
	fragments *[]*dnp3.Fragment
}

// Reassembled implements tcpassembly.Stream.
func (stream *dnp3Stream) Reassembled(chunks []tcpassembly.Reassembly) {
	for _, chunk := range chunks {
		// A nonzero Skip means bytes were lost to a TCP gap; drop any
		// partial frame rather than splice across it.
		if chunk.Skip != 0 {
			stream.buffer = stream.buffer[:0]
		}

		// tcpassembly reuses chunk.Bytes once this call returns.
		stream.buffer = append(stream.buffer, chunk.Bytes...)
	}

	frames, remainder, err := dnp3.ParseFrames(stream.buffer)

	// Frames parsed before an error are still good.
	for _, frame := range frames {
		fragment, assembleErr := stream.assembler.Assemble(frame)
		if assembleErr == nil && fragment != nil {
			*stream.fragments = append(*stream.fragments, fragment)
		}
	}

	if err != nil {
		// A real consumer would resync on the 0x0564 start bytes.
		stream.buffer = stream.buffer[:0]

		return
	}

	// remainder aliases buffer, so copy it down before the next append.
	stream.buffer = append(stream.buffer[:0], remainder...)
}

// ReassemblyComplete implements tcpassembly.Stream.
func (*dnp3Stream) ReassemblyComplete() {}

// dnp3StreamFactory hands tcpassembly a fresh DNP3 stream per TCP connection.
type dnp3StreamFactory struct {
	fragments *[]*dnp3.Fragment
}

// New implements tcpassembly.StreamFactory.
//
//nolint:ireturn // tcpassembly.StreamFactory requires returning the interface.
func (factory dnp3StreamFactory) New(_, _ gopacket.Flow) tcpassembly.Stream {
	return &dnp3Stream{fragments: factory.fragments}
}

// tcpSegment builds a synthetic TCP segment carrying payload.
func tcpSegment(seq uint32, syn bool, payload []byte) *layers.TCP {
	tcp := &layers.TCP{
		SrcPort: 40000,
		DstPort: 20000,
		Seq:     seq,
		SYN:     syn,
	}
	tcp.Payload = payload

	// TransportFlow reads the raw port bytes, which only decoding fills in.
	tcp.SetInternalPortsForTesting()

	return tcp
}

// TestAssemblerTCPStream drives a dnp3.Assembler from gopacket/tcpassembly,
// the usual shape for reassembling DNP3 out of a live capture or pcap.
func TestAssemblerTCPStream(t *testing.T) {
	t.Parallel()

	flow, err := gopacket.FlowFromEndpoints(
		layers.NewIPEndpoint(net.IPv4(10, 0, 0, 4)),
		layers.NewIPEndpoint(net.IPv4(10, 0, 0, 3)))
	if err != nil {
		t.Fatal(err)
	}

	var fragments []*dnp3.Fragment

	tcpAssembler := tcpassembly.NewAssembler(
		tcpassembly.NewStreamPool(dnp3StreamFactory{fragments: &fragments}))
	seen := time.Unix(0, 0)
	seq := uint32(1001)

	// tcpassembly buffers a stream until it sees the start of it.
	tcpAssembler.AssembleWithTimestamp(flow, tcpSegment(seq-1, true, nil), seen)

	// Split mid-frame so the stream parser has to carry a remainder.
	split := len(twoSegmentWire) / 2
	for _, chunk := range [][]byte{twoSegmentWire[:split], twoSegmentWire[split:]} {
		tcpAssembler.AssembleWithTimestamp(flow, tcpSegment(seq, false, chunk), seen)

		// #nosec G115 -- synthetic payloads are a few dozen bytes
		seq += uint32(len(chunk))
	}

	tcpAssembler.FlushAll()

	if len(fragments) != 1 {
		t.Fatalf("reassembled %d fragments, want 1", len(fragments))
	}

	fragment := fragments[0]
	if fragment.Segments != 2 || len(fragment.Data) != 14 {
		t.Fatalf("got %d segments of %d bytes, want 2 of 14",
			fragment.Segments, len(fragment.Data))
	}

	if fragment.Application == nil ||
		fragment.Application.GetFunctionCode() != byte(dnp3.Read) {
		t.Fatalf("Application = %v, want a Read request", fragment.Application)
	}
}
