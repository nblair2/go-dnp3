package dnp3

import (
	"errors"
	"fmt"
)

// maxFragmentSize bounds reassembly memory. IEEE 1815 sets no hard cap on
// application fragment size; most devices advertise 2048 bytes.
const maxFragmentSize = 65536

// Sentinel errors from Assemble. A segment that trips one is discarded along
// with any fragment in progress for its session.
var (
	ErrOrphanSegment    = errors.New("transport segment without FIR and no fragment in progress")
	ErrSequenceMismatch = errors.New("transport segment out of sequence")
	ErrFragmentTooLarge = errors.New("application fragment exceeds maximum size")
)

// SessionKey identifies a transport session: IEEE 1815 runs one transport
// sequence per direction between a pair of link addresses.
type SessionKey struct {
	Source      uint16 `json:"source"`
	Destination uint16 `json:"destination"`
	Direction   bool   `json:"direction"` // data link DIR bit: true is master to outstation
}

// String outputs the session as "source -> destination (direction)".
func (key SessionKey) String() string {
	direction := "response"
	if key.Direction {
		direction = "request"
	}

	return fmt.Sprintf("%d -> %d (%s)", key.Source, key.Destination, direction)
}

// Fragment is a complete application fragment reassembled from one or more
// transport segments.
type Fragment struct {
	Session     SessionKey  `json:"session"`
	Sequence    uint8       `json:"sequence"`    // transport SEQ of the FIR segment
	Segments    int         `json:"segments"`    // link frames consumed
	Data        []byte      `json:"data"`        // reassembled application bytes, CRCs removed
	Application Application `json:"application"` // nil when Data could not be decoded
}

// String outputs the fragment as an indented string.
func (frag *Fragment) String() string {
	appString := ""
	if frag.Application != nil {
		appString = indent(frag.Application.String(), "\t")
	}

	return fmt.Sprintf("Fragment: %s, SEQ %d, %d segments, %d bytes\n%s",
		frag.Session, frag.Sequence, frag.Segments, len(frag.Data), appString)
}

// Assembler reassembles application fragments from the transport segments of
// one or more sessions, keyed by source, destination, and direction. The zero
// value is ready to use. It is not safe for concurrent use; create one per
// stream or serialize calls.
type Assembler struct {
	sessions map[SessionKey]*fragmentState
}

// fragmentState is the reassembly progress of one session.
type fragmentState struct {
	data     []byte
	segments int
	sequence uint8 // SEQ of the FIR segment
	next     uint8 // SEQ the next segment must carry
}

// Assemble feeds one link frame into the reassembly state machine.
//
// It returns (nil, nil) when the frame carried no transport segment, when the
// segment was buffered but the fragment is still incomplete, or when a
// completed fragment held no application bytes. It returns (fragment, nil)
// once a fragment is complete and its application layer decoded, and
// (fragment, err) when the fragment is complete but the application decode
// failed, in which case Data is still valid. It returns (nil, err) when the
// segment violated the transport rules (ErrOrphanSegment, ErrSequenceMismatch,
// ErrFragmentTooLarge); the segment and any fragment in progress are dropped.
//
//nolint:nilnil // (nil, nil) means "no fragment yet"; documented above.
func (asm *Assembler) Assemble(frame *Frame) (*Fragment, error) {
	if !frame.DataLink.Control.carriesUserData() {
		return nil, nil
	}

	// Too short to hold a transport byte. Length 0 is a manually built frame,
	// which is sized only on serialization.
	if frame.DataLink.Length > 0 && frame.DataLink.Length < 6 {
		return nil, nil
	}

	key := SessionKey{
		Source:      frame.DataLink.Source,
		Destination: frame.DataLink.Destination,
		Direction:   frame.DataLink.Control.Direction,
	}

	state, err := asm.session(key, frame)
	if err != nil {
		return nil, err
	}

	state.data = append(state.data, frame.Transport.Payload...)
	if len(state.data) > maxFragmentSize {
		delete(asm.sessions, key)

		return nil, fmt.Errorf("%w: %s has %d bytes", ErrFragmentTooLarge, key, len(state.data))
	}

	// Transport SEQ is 6 bits and wraps from 63 to 0.
	state.next = (frame.Transport.Sequence + 1) & 0b00111111
	state.segments++

	if !frame.Transport.Final {
		return nil, nil
	}

	delete(asm.sessions, key)

	return asm.complete(key, state, frame)
}

// AssemblePayload parses every complete DNP3 link frame from a byte slice
// (typically one TCP read) and feeds them, in wire order, into the Assembler.
// It returns the frames parsed, every application fragment that completed while
// consuming them, any trailing partial-frame bytes (to prepend to the next
// read), and the first error encountered.
//
// Prefer this over Assemble when a read may hold more than one link frame: a
// fragment whose transport segments are concatenated into a single payload is
// then reassembled in one call, rather than silently stalling if only the first
// frame is fed. A parse error (malformed length or CRC) or an Assemble error
// (ErrOrphanSegment, ErrSequenceMismatch, ErrFragmentTooLarge) stops
// consumption and is returned with whatever frames and fragments came first.
func (asm *Assembler) AssemblePayload(payload []byte) ([]*Frame, []*Fragment, []byte, error) {
	frames, rest, err := ParseFrames(payload)

	var fragments []*Fragment

	for _, frame := range frames {
		fragment, assembleErr := asm.Assemble(frame)
		if assembleErr != nil {
			return frames, fragments, rest, assembleErr
		}

		if fragment != nil {
			fragments = append(fragments, fragment)
		}
	}

	return frames, fragments, rest, err
}

// session returns the state a segment belongs to, starting a new fragment on
// FIR and otherwise requiring an in-progress fragment with a matching SEQ.
func (asm *Assembler) session(key SessionKey, frame *Frame) (*fragmentState, error) {
	if frame.Transport.First {
		// FIR restarts unconditionally, abandoning any fragment in progress.
		state := &fragmentState{sequence: frame.Transport.Sequence}

		if asm.sessions == nil {
			asm.sessions = make(map[SessionKey]*fragmentState)
		}

		asm.sessions[key] = state

		return state, nil
	}

	state, ok := asm.sessions[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s SEQ %d",
			ErrOrphanSegment, key, frame.Transport.Sequence)
	}

	// Catches gaps, duplicates, and reordering alike.
	if frame.Transport.Sequence != state.next {
		delete(asm.sessions, key)

		return nil, fmt.Errorf("%w: %s wanted SEQ %d, got %d",
			ErrSequenceMismatch, key, state.next, frame.Transport.Sequence)
	}

	return state, nil
}

// complete builds the Fragment for a session whose FIN segment just arrived.
//
//nolint:nilnil // an empty fragment yields no result; see Assemble.
func (*Assembler) complete(key SessionKey, state *fragmentState, frame *Frame) (*Fragment, error) {
	fragment := &Fragment{
		Session:  key,
		Sequence: state.sequence,
		Segments: state.segments,
		Data:     state.data,
	}

	// A single-segment fragment was already decoded by Frame.DecodeFromBytes.
	if state.segments == 1 && frame.Application != nil {
		fragment.Application = frame.Application

		return fragment, nil
	}

	if len(fragment.Data) == 0 {
		return nil, nil
	}

	if len(fragment.Data) < 2 {
		return fragment, fmt.Errorf("application fragment too short: %d bytes",
			len(fragment.Data))
	}

	var app Application
	if key.Direction {
		app = &ApplicationRequest{}
	} else {
		app = &ApplicationResponse{}
	}

	err := app.DecodeFromBytes(fragment.Data)
	if err != nil {
		return fragment, fmt.Errorf("error in DNP3 Application layer: %w", err)
	}

	fragment.Application = app

	return fragment, nil
}
