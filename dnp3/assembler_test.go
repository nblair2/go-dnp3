package dnp3_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/nblair2/go-dnp3/v3/dnp3"
)

// readClass1230Fragment is the 14-byte application fragment carried by
// readClass1230 (FC 1, Read).
var readClass1230Fragment = []byte{
	0xc5, 0x01, 0x3c, 0x02, 0x06, 0x3c, 0x03, 0x06,
	0x3c, 0x04, 0x06, 0x3c, 0x01, 0x06,
}

// allIINSetFragment is the 4-byte application fragment of an outstation
// response with every IIN bit set.
var allIINSetFragment = []byte{0xc1, 0x81, 0xff, 0x3f}

// segmentFrom builds an unconfirmed-user-data frame carrying one transport
// segment, as a device would put it on the wire.
func segmentFrom(
	source, destination uint16,
	direction, first, final bool,
	sequence uint8,
	payload []byte,
) *dnp3.Frame {
	frame := dnp3.NewFrame()
	frame.DataLink.Control.Primary = true
	frame.DataLink.Control.Direction = direction
	frame.DataLink.Control.FunctionCode = dnp3.UnconfirmedUserData
	frame.DataLink.Source = source
	frame.DataLink.Destination = destination
	frame.Transport.First = first
	frame.Transport.Final = final
	frame.Transport.Sequence = sequence
	frame.Transport.Payload = payload

	return frame
}

// masterSegment builds a segment sent by master 4 to outstation 3.
func masterSegment(first, final bool, sequence uint8, payload []byte) *dnp3.Frame {
	return segmentFrom(4, 3, true, first, final, sequence, payload)
}

// wantPending feeds a segment that must be buffered without completing a
// fragment.
func wantPending(t *testing.T, assembler *dnp3.Assembler, frame *dnp3.Frame) {
	t.Helper()

	fragment, err := assembler.Assemble(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fragment != nil {
		t.Fatalf("expected no fragment, got %d bytes", len(fragment.Data))
	}
}

// wantFragment feeds a segment that must complete a fragment, and returns it.
func wantFragment(t *testing.T, assembler *dnp3.Assembler, frame *dnp3.Frame) *dnp3.Fragment {
	t.Helper()

	fragment, err := assembler.Assemble(frame)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fragment == nil {
		t.Fatal("expected a completed fragment, got nil")
	}

	return fragment
}

// wantErr feeds a segment that must be rejected with the given sentinel error.
func wantErr(t *testing.T, assembler *dnp3.Assembler, frame *dnp3.Frame, sentinel error) {
	t.Helper()

	fragment, err := assembler.Assemble(frame)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected %v, got %v", sentinel, err)
	}

	if fragment != nil {
		t.Fatalf("expected no fragment alongside %v", sentinel)
	}
}

func TestAssembler_singleSegment(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	fragment := wantFragment(t, assembler,
		masterSegment(true, true, 0, readClass1230Fragment))

	if fragment.Segments != 1 {
		t.Fatalf("Segments = %d, want 1", fragment.Segments)
	}

	if !slices.Equal(fragment.Data, readClass1230Fragment) {
		t.Fatalf("Data = %x, want %x", fragment.Data, readClass1230Fragment)
	}

	want := dnp3.SessionKey{Source: 4, Destination: 3, Direction: true}
	if fragment.Session != want {
		t.Fatalf("Session = %s, want %s", fragment.Session, want)
	}

	request, ok := fragment.Application.(*dnp3.ApplicationRequest)
	if !ok {
		t.Fatalf("Application is %T, want *dnp3.ApplicationRequest", fragment.Application)
	}

	if got := request.GetFunctionCode(); got != byte(dnp3.Read) {
		t.Fatalf("function code = %d, want %d", got, byte(dnp3.Read))
	}
}

func TestAssembler_twoSegments(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	wantPending(t, assembler, masterSegment(true, false, 5, readClass1230Fragment[:4]))

	fragment := wantFragment(t, assembler,
		masterSegment(false, true, 6, readClass1230Fragment[4:]))

	if fragment.Segments != 2 {
		t.Fatalf("Segments = %d, want 2", fragment.Segments)
	}

	if fragment.Sequence != 5 {
		t.Fatalf("Sequence = %d, want 5", fragment.Sequence)
	}

	if !slices.Equal(fragment.Data, readClass1230Fragment) {
		t.Fatalf("Data = %x, want %x", fragment.Data, readClass1230Fragment)
	}

	if got := fragment.Application.GetFunctionCode(); got != byte(dnp3.Read) {
		t.Fatalf("function code = %d, want %d", got, byte(dnp3.Read))
	}
}

func TestAssembler_response(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	wantPending(t, assembler, segmentFrom(3, 4, false, true, false, 0, allIINSetFragment[:2]))

	fragment := wantFragment(t, assembler,
		segmentFrom(3, 4, false, false, true, 1, allIINSetFragment[2:]))

	if !slices.Equal(fragment.Data, allIINSetFragment) {
		t.Fatalf("Data = %x, want %x", fragment.Data, allIINSetFragment)
	}

	if _, ok := fragment.Application.(*dnp3.ApplicationResponse); !ok {
		t.Fatalf("Application is %T, want *dnp3.ApplicationResponse", fragment.Application)
	}
}

// TestAssembler_sequenceWrap covers the 6-bit transport sequence rolling over
// from 63 to 0 mid-fragment.
func TestAssembler_sequenceWrap(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	wantPending(t, assembler, masterSegment(true, false, 62, readClass1230Fragment[:4]))
	wantPending(t, assembler, masterSegment(false, false, 63, readClass1230Fragment[4:9]))

	fragment := wantFragment(t, assembler,
		masterSegment(false, true, 0, readClass1230Fragment[9:]))

	if fragment.Segments != 3 {
		t.Fatalf("Segments = %d, want 3", fragment.Segments)
	}

	if fragment.Sequence != 62 {
		t.Fatalf("Sequence = %d, want 62", fragment.Sequence)
	}

	if !slices.Equal(fragment.Data, readClass1230Fragment) {
		t.Fatalf("Data = %x, want %x", fragment.Data, readClass1230Fragment)
	}
}

func TestAssembler_sequenceGap(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	wantPending(t, assembler, masterSegment(true, false, 1, readClass1230Fragment[:4]))
	wantErr(t, assembler,
		masterSegment(false, true, 3, readClass1230Fragment[4:]), dnp3.ErrSequenceMismatch)

	// The session recovers on the next FIR.
	fragment := wantFragment(t, assembler,
		masterSegment(true, true, 4, readClass1230Fragment))

	if fragment.Segments != 1 {
		t.Fatalf("Segments = %d, want 1", fragment.Segments)
	}
}

func TestAssembler_duplicateSegment(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	wantPending(t, assembler, masterSegment(true, false, 1, readClass1230Fragment[:4]))
	wantPending(t, assembler, masterSegment(false, false, 2, readClass1230Fragment[4:9]))
	wantErr(t, assembler,
		masterSegment(false, false, 2, readClass1230Fragment[4:9]), dnp3.ErrSequenceMismatch)
}

func TestAssembler_orphanSegment(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	wantErr(t, assembler,
		masterSegment(false, true, 7, readClass1230Fragment), dnp3.ErrOrphanSegment)
}

// TestAssembler_restartOnFirst verifies FIR abandons a fragment in progress
// rather than erroring.
func TestAssembler_restartOnFirst(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	wantPending(t, assembler, masterSegment(true, false, 1, readClass1230Fragment[:4]))

	fragment := wantFragment(t, assembler,
		masterSegment(true, true, 9, readClass1230Fragment))

	if fragment.Segments != 1 {
		t.Fatalf("Segments = %d, want 1", fragment.Segments)
	}

	if fragment.Sequence != 9 {
		t.Fatalf("Sequence = %d, want 9", fragment.Sequence)
	}

	if !slices.Equal(fragment.Data, readClass1230Fragment) {
		t.Fatalf("Data = %x, want %x", fragment.Data, readClass1230Fragment)
	}
}

// TestAssembler_interleavedSessions verifies segments travelling in both
// directions are reassembled independently.
func TestAssembler_interleavedSessions(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	wantPending(t, assembler, segmentFrom(4, 3, true, true, false, 0, readClass1230Fragment[:4]))
	wantPending(t, assembler, segmentFrom(3, 4, false, true, false, 0, allIINSetFragment[:2]))

	request := wantFragment(t, assembler,
		segmentFrom(4, 3, true, false, true, 1, readClass1230Fragment[4:]))
	response := wantFragment(t, assembler,
		segmentFrom(3, 4, false, false, true, 1, allIINSetFragment[2:]))

	if !slices.Equal(request.Data, readClass1230Fragment) {
		t.Fatalf("request Data = %x, want %x", request.Data, readClass1230Fragment)
	}

	if !request.Session.Direction {
		t.Fatalf("request session = %s, want a request direction", request.Session)
	}

	if !slices.Equal(response.Data, allIINSetFragment) {
		t.Fatalf("response Data = %x, want %x", response.Data, allIINSetFragment)
	}

	if response.Session.Direction {
		t.Fatalf("response session = %s, want a response direction", response.Session)
	}
}

func TestAssembler_maxFragmentSize(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}
	payload := make([]byte, 40000)

	wantPending(t, assembler, masterSegment(true, false, 0, payload))
	wantErr(t, assembler, masterSegment(false, true, 1, payload), dnp3.ErrFragmentTooLarge)

	// The oversized fragment was dropped, so the session is gone.
	wantErr(t, assembler, masterSegment(false, true, 2, payload), dnp3.ErrOrphanSegment)
}

// TestAssembler_linkControlFrame verifies frames without user data are ignored
// and leave no session behind.
func TestAssembler_linkControlFrame(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	linkStatus := dnp3.NewFrame()
	linkStatus.DataLink.Control.Primary = true
	linkStatus.DataLink.Control.Direction = true
	linkStatus.DataLink.Control.FunctionCode = dnp3.RequestLinkStatus
	linkStatus.DataLink.Source = 4
	linkStatus.DataLink.Destination = 3

	wantPending(t, assembler, linkStatus)
	wantErr(t, assembler,
		masterSegment(false, true, 0, readClass1230Fragment), dnp3.ErrOrphanSegment)
}

// TestAssembler_emptyFragment verifies a complete but empty fragment yields no
// result.
func TestAssembler_emptyFragment(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	wantPending(t, assembler, masterSegment(true, true, 0, nil))
}

// checkSegmentFrame verifies a parsed segment frame decoded no application
// layer of its own and re-encodes to the bytes it came from.
func checkSegmentFrame(t *testing.T, index int, frame *dnp3.Frame, want []byte) {
	t.Helper()

	// No single segment holds a whole fragment, so none decodes alone.
	if frame.Application != nil {
		t.Fatalf("frame %d: Application should be nil for a partial fragment", index)
	}

	if got := serializeFrame(t, frame); !slices.Equal(got, want) {
		t.Fatalf("frame %d: re-encode mismatch\ngot:  %x\nwant: %x", index, got, want)
	}
}

// checkReadFragment verifies a reassembled fragment carries the whole
// readClass1230Fragment.
func checkReadFragment(t *testing.T, fragment *dnp3.Fragment, segments int) {
	t.Helper()

	if fragment == nil {
		t.Fatal("expected a completed fragment after the final segment")
	}

	if fragment.Segments != segments {
		t.Fatalf("Segments = %d, want %d", fragment.Segments, segments)
	}

	if !slices.Equal(fragment.Data, readClass1230Fragment) {
		t.Fatalf("Data = %x, want %x", fragment.Data, readClass1230Fragment)
	}

	if got := fragment.Application.GetFunctionCode(); got != byte(dnp3.Read) {
		t.Fatalf("function code = %d, want %d", got, byte(dnp3.Read))
	}
}

// TestAssembler_wireRoundTrip is the end-to-end regression test: segments of a
// multi-frame fragment must survive serialization, ParseFrames, and
// re-serialization byte-for-byte, then reassemble.
func TestAssembler_wireRoundTrip(t *testing.T) {
	t.Parallel()

	segments := []*dnp3.Frame{
		masterSegment(true, false, 10, readClass1230Fragment[:4]),
		masterSegment(false, false, 11, readClass1230Fragment[4:9]),
		masterSegment(false, true, 12, readClass1230Fragment[9:]),
	}

	var wire []byte

	encoded := make([][]byte, 0, len(segments))

	for _, segment := range segments {
		wireBytes := serializeFrame(t, segment)
		encoded = append(encoded, wireBytes)
		wire = append(wire, wireBytes...)
	}

	frames, remainder, err := dnp3.ParseFrames(wire)
	if err != nil {
		t.Fatal("ParseFrames:", err)
	}

	if len(frames) != len(segments) {
		t.Fatalf("parsed %d frames, want %d", len(frames), len(segments))
	}

	if len(remainder) != 0 {
		t.Fatalf("expected empty remainder, got %d bytes", len(remainder))
	}

	assembler := &dnp3.Assembler{}

	var fragment *dnp3.Fragment

	for index, frame := range frames {
		checkSegmentFrame(t, index, frame, encoded[index])

		fragment, err = assembler.Assemble(frame)
		if err != nil {
			t.Fatalf("frame %d: Assemble: %v", index, err)
		}
	}

	checkReadFragment(t, fragment, len(segments))
}

// wirePayload serializes segments and concatenates them into a single on-wire
// payload, as a device would pack multiple link frames into one TCP read.
func wirePayload(t *testing.T, segments ...*dnp3.Frame) []byte {
	t.Helper()

	// 292 bytes is the maximum on-wire size of a DNP3 link frame.
	wire := make([]byte, 0, len(segments)*292)
	for _, segment := range segments {
		wire = append(wire, serializeFrame(t, segment)...)
	}

	return wire
}

// TestAssemblePayload_multiFrameSinglePayload is the regression for the pepper-puddle
// stall: a whole fragment whose transport segments are concatenated into one
// payload must reassemble in a single call, not stall after the first frame.
func TestAssemblePayload_multiFrameSinglePayload(t *testing.T) {
	t.Parallel()

	payload := wirePayload(t,
		masterSegment(true, false, 10, readClass1230Fragment[:4]),
		masterSegment(false, false, 11, readClass1230Fragment[4:9]),
		masterSegment(false, true, 12, readClass1230Fragment[9:]),
	)

	assembler := &dnp3.Assembler{}

	frames, fragments, rest, err := assembler.AssemblePayload(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frames) != 3 {
		t.Fatalf("parsed %d frames, want 3", len(frames))
	}

	if len(rest) != 0 {
		t.Fatalf("rest = %x, want empty", rest)
	}

	if len(fragments) != 1 {
		t.Fatalf("completed %d fragments, want 1", len(fragments))
	}

	checkReadFragment(t, fragments[0], 3)
}

// TestAssemblePayload_singleFrame covers a payload holding one self-contained
// FIR+FIN frame.
func TestAssemblePayload_singleFrame(t *testing.T) {
	t.Parallel()

	payload := wirePayload(t, masterSegment(true, true, 0, readClass1230Fragment))

	assembler := &dnp3.Assembler{}

	frames, fragments, rest, err := assembler.AssemblePayload(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frames) != 1 || len(fragments) != 1 || len(rest) != 0 {
		t.Fatalf("frames=%d fragments=%d rest=%d, want 1/1/0",
			len(frames), len(fragments), len(rest))
	}

	checkReadFragment(t, fragments[0], 1)
}

// TestAssemblePayload_splitAcrossPayloads verifies the assembler carries state
// between calls when a fragment spans two reads.
func TestAssemblePayload_splitAcrossPayloads(t *testing.T) {
	t.Parallel()

	assembler := &dnp3.Assembler{}

	first := wirePayload(t, masterSegment(true, false, 5, readClass1230Fragment[:4]))

	_, fragments, _, err := assembler.AssemblePayload(first)
	if err != nil {
		t.Fatalf("unexpected error on first payload: %v", err)
	}

	if len(fragments) != 0 {
		t.Fatalf("completed %d fragments on first payload, want 0", len(fragments))
	}

	second := wirePayload(t, masterSegment(false, true, 6, readClass1230Fragment[4:]))

	_, fragments, _, err = assembler.AssemblePayload(second)
	if err != nil {
		t.Fatalf("unexpected error on second payload: %v", err)
	}

	if len(fragments) != 1 {
		t.Fatalf("completed %d fragments on second payload, want 1", len(fragments))
	}

	checkReadFragment(t, fragments[0], 2)
}

// TestAssemblePayload_trailingPartial verifies a whole frame plus an incomplete
// trailing frame surfaces the parsed frame and the leftover bytes without error.
func TestAssemblePayload_trailingPartial(t *testing.T) {
	t.Parallel()

	whole := wirePayload(t, masterSegment(true, false, 0, readClass1230Fragment[:4]))
	partial := []byte{0x05, 0x64, 0x10, 0xc4} // a fresh start byte pair, truncated
	payload := append(append([]byte{}, whole...), partial...)

	assembler := &dnp3.Assembler{}

	frames, fragments, rest, err := assembler.AssemblePayload(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frames) != 1 {
		t.Fatalf("parsed %d frames, want 1", len(frames))
	}

	if len(fragments) != 0 {
		t.Fatalf("completed %d fragments, want 0 (FIR only)", len(fragments))
	}

	if !slices.Equal(rest, partial) {
		t.Fatalf("rest = %x, want %x", rest, partial)
	}
}

// TestAssemblePayload_orphanError verifies an Assemble error stops consumption
// and is returned.
func TestAssemblePayload_orphanError(t *testing.T) {
	t.Parallel()

	payload := wirePayload(t, masterSegment(false, true, 7, readClass1230Fragment))

	assembler := &dnp3.Assembler{}

	_, fragments, _, err := assembler.AssemblePayload(payload)
	if !errors.Is(err, dnp3.ErrOrphanSegment) {
		t.Fatalf("expected ErrOrphanSegment, got %v", err)
	}

	if len(fragments) != 0 {
		t.Fatalf("completed %d fragments, want 0", len(fragments))
	}
}
