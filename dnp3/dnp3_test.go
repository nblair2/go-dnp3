package dnp3_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/nblair2/go-dnp3/dnp3"
)

var (
	customPcapsFlag string
	customPcaps     []string
	printStringFlag bool
	printJSONFlag   bool

	tests = []struct {
		name  string
		input []byte
	}{
		{
			"Request/ReadClass1230",
			[]byte{
				0x05, 0x64, 0x14, 0xc4, 0x04, 0x00, 0x03, 0x00,
				0xc7, 0x17, 0xc4, 0xc5, 0x01, 0x3c, 0x02, 0x06,
				0x3c, 0x03, 0x06, 0x3c, 0x04, 0x06, 0x3c, 0x01,
				0x06, 0xa3, 0x61,
			},
		},
		{
			"Request/ReadBinaryInputChange",
			[]byte{
				0x05, 0x64, 0x0b, 0xc4, 0x00, 0x04, 0x01, 0x00,
				0xca, 0x8a, 0xc0, 0xc1, 0x01, 0x02, 0x00, 0x06,
				0x95, 0x76,
			},
		},
		{
			"Request/WriteTime",
			[]byte{
				0x05, 0x64, 0x12, 0xc4, 0x04, 0x00, 0x03, 0x00,
				0x1e, 0x7c, 0xc1, 0xc1, 0x02, 0x32, 0x01, 0x07,
				0x01, 0xeb, 0xe4, 0x5a, 0x87, 0xff, 0x00, 0x28,
				0x01,
			},
		},
		{
			"Request/Select",
			[]byte{
				0x05, 0x64, 0x1a, 0xc4, 0x04, 0x00, 0x03, 0x00,
				0xc2, 0xe6, 0xd2, 0xc3, 0x03, 0x0c, 0x01, 0x28,
				0x01, 0x00, 0x9f, 0x86, 0x03, 0x01, 0x64, 0x00,
				0x00, 0x00, 0xec, 0x41, 0x64, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x5b,
			},
		},
		{
			"Response/AllIINSet",
			[]byte{
				0x05, 0x64, 0x0a, 0x44, 0x03, 0x00, 0x04, 0x00,
				0x7c, 0xae, 0xe7, 0xc1, 0x81, 0xff, 0x3f, 0x1c,
				0x48,
			},
		},
		{
			"Response/GV_02-02",
			[]byte{
				0x05, 0x64, 0x2a, 0x44, 0x01, 0x00, 0x00, 0x04,
				0xe5, 0x79, 0xc1, 0xe2, 0x81, 0x90, 0x00, 0x02,
				0x02, 0x28, 0x03, 0x00, 0x00, 0x00, 0x81, 0xda,
				0x33, 0xd2, 0xdf, 0xe5, 0x64, 0x71, 0x01, 0x00,
				0x00, 0x01, 0xda, 0x33, 0xd2, 0x64, 0x71, 0x01,
				0xff, 0xff, 0x81, 0xdb, 0xdd, 0x14, 0x33, 0xd2,
				0x64, 0x71, 0x01, 0x38, 0x5d,
			},
		},
		{
			"Response/GV_01-01_10-02_20-05_21-09_30-03",
			[]byte{
				0x05, 0x64, 0x4e, 0x44, 0x03, 0x00, 0x04, 0x00,
				0x6f, 0x4d, 0xc7, 0xc7, 0x81, 0x00, 0x00, 0x01,
				0x01, 0x00, 0x00, 0x05, 0x19, 0x0a, 0x02, 0x00,
				0x00, 0x05, 0xc3, 0x47, 0x81, 0x01, 0x81, 0x81,
				0x01, 0x01, 0x14, 0x05, 0x00, 0x00, 0x00, 0x20,
				0x00, 0x00, 0x00, 0x15, 0xf1, 0x7b, 0x09, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1e, 0x03,
				0x00, 0x00, 0x06, 0xca, 0x00, 0x00, 0x18, 0x7e,
				0x00, 0xcb, 0x00, 0x00, 0x00, 0xc9, 0x00, 0x00,
				0x00, 0xff, 0xff, 0xff, 0xff, 0x66, 0x21, 0x00,
				0xd6, 0xf3, 0x00, 0x59, 0x21, 0x00, 0x00, 0x4b,
				0x21, 0x00, 0x00, 0xe0, 0x51,
			},
		},
	}
)

func TestMain(m *testing.M) {
	flag.StringVar(&customPcapsFlag, "pcaps", "", "Comma-separated list of pcap files to read")
	flag.BoolVar(&printStringFlag, "print-string", false, "Print packet string output")
	flag.BoolVar(&printJSONFlag, "print-json", false, "Print packet json output")
	flag.Parse()

	if customPcapsFlag != "" {
		customPcaps = splitComma(customPcapsFlag)
	}

	os.Exit(m.Run())
}

func TestDNP3(t *testing.T) {
	t.Parallel()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testRoundTrip(t, tc.input)
		})
	}
}

func TestCustomPcaps(t *testing.T) {
	t.Parallel()
	flag.Parse()

	if len(customPcaps) == 0 {
		t.Skip("No custom pcap file provided")
	}

	for _, pcapFile := range customPcaps {
		t.Run(pcapFile, func(t *testing.T) {
			t.Parallel()

			handle, err := pcap.OpenOffline(pcapFile)
			if err != nil {
				t.Skipf("Error opening PCAP: %v", err)
			}
			defer handle.Close()

			pcap := gopacket.NewPacketSource(handle, handle.LinkType())

			packetIndex := 0
			for pkt := range pcap.Packets() {
				packetIndex++

				var input []byte
				// Prefer the auto-decoded DNP3 layer (via TCP/UDP port 20000
				// registration). Fall back to raw TCP payload for non-standard
				// ports or multi-frame segments.
				if dnp3Layer := pkt.Layer(dnp3.LayerTypeDNP3); dnp3Layer != nil {
					input = dnp3Layer.LayerContents()
				} else if tcpLayer := pkt.Layer(layers.LayerTypeTCP); tcpLayer != nil {
					tcp, _ := tcpLayer.(*layers.TCP)
					input = tcp.Payload
				}

				if len(input) < 10 {
					continue
				}

				t.Run(fmt.Sprintf("Packet%d", packetIndex), func(t *testing.T) {
					testRoundTrip(t, input)
				})
			}
		})
	}
}

// serializeFrame is a test helper that runs a frame through
// gopacket.SerializeLayers and returns the resulting bytes.
func serializeFrame(t *testing.T, frame *dnp3.Frame) []byte {
	t.Helper()

	buf := gopacket.NewSerializeBuffer()

	err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{}, frame)
	if err != nil {
		t.Fatal("SerializeLayers:", err)
	}

	return buf.Bytes()
}

func testRoundTrip(t *testing.T, input []byte) {
	t.Helper()

	packet, err := dnp3.NewFrameFromBytes(input)
	if err != nil {
		t.Fatal("NewFrameFromBytes:", err)
	}

	if packet.Application != nil {
		data := packet.Application.GetData()
		if data.HasExtra() {
			t.Fatalf(
				"application parser used the `Extra` field meaining a Group/Variation is not supported or some error",
			)
		}
	}

	output := serializeFrame(t, packet)

	if !slices.Equal(input, output) {
		t.Fatal("Input and Output not equal")
	}

	str := packet.String()
	if printStringFlag {
		fmt.Println(str)
	}

	jsonBytes, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		t.Fatal("MarshalIndent:", err)
	}

	if printJSONFlag {
		fmt.Println(string(jsonBytes))
	}
}

func splitComma(s string) []string {
	var out []string

	for v := range strings.SplitSeq(s, ",") {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}

	return out
}

// readBinaryInputChange is a 18-byte frame used across ParseFrames tests.
var readBinaryInputChange = []byte{
	0x05, 0x64, 0x0b, 0xc4, 0x00, 0x04, 0x01, 0x00,
	0xca, 0x8a, 0xc0, 0xc1, 0x01, 0x02, 0x00, 0x06,
	0x95, 0x76,
}

// readClass1230 is a 27-byte frame used across ParseFrames tests.
var readClass1230 = []byte{
	0x05, 0x64, 0x14, 0xc4, 0x04, 0x00, 0x03, 0x00,
	0xc7, 0x17, 0xc4, 0xc5, 0x01, 0x3c, 0x02, 0x06,
	0x3c, 0x03, 0x06, 0x3c, 0x04, 0x06, 0x3c, 0x01,
	0x06, 0xa3, 0x61,
}

func TestParseFrames_empty(t *testing.T) {
	t.Parallel()

	frames, remainder, err := dnp3.ParseFrames(nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(frames) != 0 {
		t.Fatalf("expected 0 frames, got %d", len(frames))
	}

	if len(remainder) != 0 {
		t.Fatalf("expected empty remainder, got %d bytes", len(remainder))
	}
}

func TestParseFrames_single(t *testing.T) {
	t.Parallel()

	frames, remainder, err := dnp3.ParseFrames(readBinaryInputChange)
	if err != nil {
		t.Fatal(err)
	}

	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	if len(remainder) != 0 {
		t.Fatalf("expected empty remainder, got %d bytes", len(remainder))
	}
}

func TestParseFrames_two(t *testing.T) {
	t.Parallel()

	input := append(slices.Clone(readBinaryInputChange), readClass1230...)

	frames, remainder, err := dnp3.ParseFrames(input)
	if err != nil {
		t.Fatal(err)
	}

	if len(frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames))
	}

	if len(remainder) != 0 {
		t.Fatalf("expected empty remainder, got %d bytes", len(remainder))
	}
}

func TestParseFrames_partialHeader(t *testing.T) {
	t.Parallel()

	// First frame complete; trailing 5 bytes are a partial header.
	partial := readClass1230[:5]
	input := append(slices.Clone(readBinaryInputChange), partial...)

	frames, remainder, err := dnp3.ParseFrames(input)
	if err != nil {
		t.Fatal(err)
	}

	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	if !slices.Equal(remainder, partial) {
		t.Fatalf("expected remainder %x, got %x", partial, remainder)
	}
}

func TestParseFrames_partialBody(t *testing.T) {
	t.Parallel()

	// First frame complete; trailing bytes have a valid header but incomplete body.
	partial := readClass1230[:12] // 10-byte header + 2 body bytes, needs 27 total
	input := append(slices.Clone(readBinaryInputChange), partial...)

	frames, remainder, err := dnp3.ParseFrames(input)
	if err != nil {
		t.Fatal(err)
	}

	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	if !slices.Equal(remainder, partial) {
		t.Fatalf("expected remainder %x, got %x", partial, remainder)
	}
}

// TestDecode_concatenatedFrames is a regression test for the CRC-corruption
// bug: passing two frames in one buffer must not corrupt the first frame's CRC
// validation.
func TestDecode_concatenatedFrames(t *testing.T) {
	t.Parallel()

	// Two valid frames concatenated in a single buffer.
	buf := append(slices.Clone(readBinaryInputChange), readClass1230...)

	frame, err := dnp3.NewFrameFromBytes(buf)
	if err != nil {
		t.Fatalf("DecodeFromBytes failed on concatenated buffer: %v", err)
	}

	// Re-encode and compare only the first frame's bytes.
	got := serializeFrame(t, frame)

	if !slices.Equal(got, readBinaryInputChange) {
		t.Fatalf("re-encoded frame mismatch\nwant: %x\n got: %x", readBinaryInputChange, got)
	}
}

func TestApplicationData_GetSetExtra(t *testing.T) {
	t.Parallel()

	appData := dnp3.NewApplicationData()
	if appData.HasExtra() {
		t.Fatal("new ApplicationData should have no extra")
	}

	if appData.GetExtra() != nil {
		t.Fatal("GetExtra should return nil when not set")
	}

	extra := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	appData.SetExtra(extra)

	if !appData.HasExtra() {
		t.Fatal("HasExtra should be true after SetExtra")
	}

	if !slices.Equal(appData.GetExtra(), extra) {
		t.Fatalf("GetExtra returned %x, want %x", appData.GetExtra(), extra)
	}

	// SerializeTo should include the extra bytes in output.
	out, err := appData.SerializeTo()
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(out, extra) {
		t.Fatalf("SerializeTo returned %x, want %x", out, extra)
	}
}

// truncFeedback captures gopacket.DecodeFeedback.SetTruncated calls.
type truncFeedback struct {
	truncated bool
}

func (tf *truncFeedback) SetTruncated() { tf.truncated = true }

// TestDecodingLayer exercises Frame's gopacket.DecodingLayer surface: it
// verifies CanDecode, NextLayerType, LayerContents (cached on-wire bytes),
// LayerPayload (nil), and Payload (nil).
func TestDecodingLayer(t *testing.T) {
	t.Parallel()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var frame dnp3.Frame

			err := frame.DecodeFromBytes(testCase.input, gopacket.NilDecodeFeedback)
			if err != nil {
				t.Fatal("DecodeFromBytes:", err)
			}

			if got := frame.CanDecode(); got != dnp3.LayerTypeDNP3 {
				t.Fatalf("CanDecode = %v, want %v", got, dnp3.LayerTypeDNP3)
			}

			if got := frame.NextLayerType(); got != gopacket.LayerTypeZero {
				t.Fatalf("NextLayerType = %v, want %v", got, gopacket.LayerTypeZero)
			}

			if got := frame.LayerContents(); !slices.Equal(got, testCase.input) {
				t.Fatalf("LayerContents mismatch\ngot:  %x\nwant: %x", got, testCase.input)
			}

			if got := frame.LayerPayload(); got != nil {
				t.Fatalf("LayerPayload should be nil, got %d bytes", len(got))
			}

			if got := frame.Payload(); got != nil {
				t.Fatalf("Payload should be nil, got %d bytes", len(got))
			}
		})
	}
}

// TestDecodingLayer_truncated verifies DecodeFromBytes calls
// DecodeFeedback.SetTruncated when input is shorter than the frame's wire size.
func TestDecodingLayer_truncated(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data []byte
	}{
		{"shorter than header", []byte{0x05, 0x64, 0x14}},
		{"header but truncated body", readClass1230[:15]},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			feedback := &truncFeedback{}

			var frame dnp3.Frame

			err := frame.DecodeFromBytes(testCase.data, feedback)
			if err == nil {
				t.Fatal("expected error on truncated input")
			}

			if !feedback.truncated {
				t.Fatal("SetTruncated was not called")
			}
		})
	}
}

// TestNewPacket_appLayer verifies gopacket.NewPacket surfaces the DNP3 frame
// as both pkt.Layer(LayerTypeDNP3) and pkt.ApplicationLayer().
func TestNewPacket_appLayer(t *testing.T) {
	t.Parallel()

	pkt := gopacket.NewPacket(readBinaryInputChange, dnp3.LayerTypeDNP3, gopacket.Default)
	if errLayer := pkt.ErrorLayer(); errLayer != nil {
		t.Fatalf("decode error: %v", errLayer.Error())
	}

	layer := pkt.Layer(dnp3.LayerTypeDNP3)
	if layer == nil {
		t.Fatal("DNP3 layer not found in packet")
	}

	frame, ok := layer.(*dnp3.Frame)
	if !ok {
		t.Fatalf("layer is not *dnp3.Frame, got %T", layer)
	}

	if !slices.Equal(frame.LayerContents(), readBinaryInputChange) {
		t.Fatal("LayerContents mismatch")
	}

	appLayer := pkt.ApplicationLayer()
	if appLayer == nil {
		t.Fatal("ApplicationLayer is nil")
	}

	if appLayer.LayerType() != dnp3.LayerTypeDNP3 {
		t.Fatalf("ApplicationLayer type = %v, want %v", appLayer.LayerType(), dnp3.LayerTypeDNP3)
	}
}

// TestSerializeLayers verifies that running each test-case frame through
// gopacket.SerializeLayers reproduces the original bytes exactly.
func TestSerializeLayers(t *testing.T) {
	t.Parallel()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			frame, err := dnp3.NewFrameFromBytes(testCase.input)
			if err != nil {
				t.Fatal("NewFrameFromBytes:", err)
			}

			got := serializeFrame(t, frame)

			if !slices.Equal(got, testCase.input) {
				t.Fatalf("round-trip mismatch\ngot:  %x\nwant: %x", got, testCase.input)
			}
		})
	}
}
