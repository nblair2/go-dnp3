package dnp3_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/nblair2/go-dnp3/v4/dnp3"
)

// layoutCase is a single group/variation round-trip: raw wire bytes for one
// point go in, decoded field values are asserted, then re-encoding must
// reproduce the input exactly.
type layoutCase struct {
	name string

	group     byte
	variation byte
	point     []byte

	flagsByte  *byte
	statusByte *byte
	absTime    []byte // raw 6-byte field, nil if absent
	wantValue  []byte
}

// buildObjectBytes assembles a single-point object header (group,
// variation, a NoPrefix/1-byte-count qualifier of 0x07, and a count of 1)
// followed by the point's raw field bytes.
func buildObjectBytes(group, variation byte, point []byte) []byte {
	return append([]byte{group, variation, 0x07, 0x01}, point...)
}

var layoutCases = []layoutCase{
	{
		name:      "g32v1_FlagsValue",
		group:     32,
		variation: 1,
		point:     []byte{0x01, 0x11, 0x22, 0x33, 0x44},
		flagsByte: new(byte(0x01)),
		wantValue: []byte{0x11, 0x22, 0x33, 0x44},
	},
	{
		name:      "g32v3_FlagsValueAbsTime",
		group:     32,
		variation: 3,
		point:     []byte{0x09, 0xAA, 0xBB, 0xCC, 0xDD, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		flagsByte: new(byte(0x09)),
		wantValue: []byte{0xAA, 0xBB, 0xCC, 0xDD},
		absTime:   []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
	},
	{
		name:      "g33v1_FlagsValue",
		group:     33,
		variation: 1,
		point:     []byte{0x04, 0xDE, 0xAD, 0xBE, 0xEF},
		flagsByte: new(byte(0x04)),
		wantValue: []byte{0xDE, 0xAD, 0xBE, 0xEF},
	},
	{
		name:      "g42v1_FlagsValue",
		group:     42,
		variation: 1,
		point:     []byte{0x10, 0x00, 0x01, 0x02, 0x03},
		flagsByte: new(byte(0x10)),
		wantValue: []byte{0x00, 0x01, 0x02, 0x03},
	},
	{
		// Status is trailing and >= 0x80: if it were routed through
		// PointFlags, the reserved-bit check would reject it.
		name:       "g41v1_ValueStatus",
		group:      41,
		variation:  1,
		point:      []byte{0x01, 0x02, 0x03, 0x04, 0x93},
		wantValue:  []byte{0x01, 0x02, 0x03, 0x04},
		statusByte: new(byte(0x93)),
	},
	{
		// Status is leading and >= 0x80, same reasoning as g41v1.
		name:       "g43v1_StatusValue",
		group:      43,
		variation:  1,
		point:      []byte{0x87, 0x05, 0x06, 0x07, 0x08},
		statusByte: new(byte(0x87)),
		wantValue:  []byte{0x05, 0x06, 0x07, 0x08},
	},
	{
		name:       "g43v3_StatusValueAbsTime",
		group:      43,
		variation:  3,
		point:      []byte{0x91, 0x09, 0x0A, 0x0B, 0x0C, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60},
		statusByte: new(byte(0x91)),
		wantValue:  []byte{0x09, 0x0A, 0x0B, 0x0C},
		absTime:    []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60},
	},
	{
		// Distinguishable value/time byte patterns so a regression of the
		// flags->value->time field order would fail this case loudly.
		name:      "g22v5_FlagsValueAbsTime",
		group:     22,
		variation: 5,
		point:     []byte{0x02, 0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
		flagsByte: new(byte(0x02)),
		wantValue: []byte{0xDE, 0xAD, 0xBE, 0xEF},
		absTime:   []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
	},
}

func TestPointBytesLayouts(t *testing.T) {
	t.Parallel()

	for _, tc := range layoutCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runLayoutCase(t, tc)
		})
	}
}

func runLayoutCase(t *testing.T, testCase layoutCase) {
	t.Helper()

	raw := buildObjectBytes(testCase.group, testCase.variation, testCase.point)

	obj, err := dnp3.NewDataObjectFromBytes(raw)
	if err != nil {
		t.Fatalf("NewDataObjectFromBytes: %v", err)
	}

	if len(obj.Points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(obj.Points))
	}

	point := obj.Points[0]

	assertFlags(t, point, testCase.flagsByte)
	assertStatus(t, point, testCase.statusByte)
	assertAbsTime(t, point, testCase.absTime)

	value, ok := point.GetValue().([]byte)
	if !ok {
		t.Fatalf("GetValue: expected []byte, got %T", point.GetValue())
	}

	if !slices.Equal(value, testCase.wantValue) {
		t.Fatalf("GetValue: got 0x % X, want 0x % X", value, testCase.wantValue)
	}

	out, err := obj.SerializeTo()
	if err != nil {
		t.Fatalf("SerializeTo: %v", err)
	}

	if !slices.Equal(out, raw) {
		t.Fatalf("round trip mismatch: got 0x % X, want 0x % X", out, raw)
	}
}

func assertFlags(t *testing.T, point dnp3.Point, want *byte) {
	t.Helper()

	if want == nil {
		_, err := point.GetFlags()
		if !errors.Is(err, dnp3.ErrNoFlags) {
			t.Fatalf("GetFlags: expected ErrNoFlags, got %v", err)
		}

		return
	}

	var wantFlags dnp3.PointFlags

	err := wantFlags.FromByte(*want)
	if err != nil {
		t.Fatalf("test setup: FromByte(0x%02X): %v", *want, err)
	}

	gotFlags, err := point.GetFlags()
	if err != nil {
		t.Fatalf("GetFlags: %v", err)
	}

	if gotFlags != wantFlags {
		t.Fatalf("GetFlags: got %+v, want %+v", gotFlags, wantFlags)
	}
}

func assertStatus(t *testing.T, point dnp3.Point, want *byte) {
	t.Helper()

	if want == nil {
		_, err := point.GetStatus()
		if !errors.Is(err, dnp3.ErrNoStatus) {
			t.Fatalf("GetStatus: expected ErrNoStatus, got %v", err)
		}

		return
	}

	gotStatus, err := point.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}

	if gotStatus != dnp3.CommandStatus(*want) {
		t.Fatalf("GetStatus: got %s, want %s", gotStatus, dnp3.CommandStatus(*want))
	}
}

func assertAbsTime(t *testing.T, point dnp3.Point, want []byte) {
	t.Helper()

	if want == nil {
		_, err := point.GetAbsTime()
		if !errors.Is(err, dnp3.ErrNoAbsTime) {
			t.Fatalf("GetAbsTime: expected ErrNoAbsTime, got %v", err)
		}

		return
	}

	wantTime, err := dnp3.BytesToDNP3TimeAbsolute(want)
	if err != nil {
		t.Fatalf("test setup: BytesToDNP3TimeAbsolute: %v", err)
	}

	gotTime, err := point.GetAbsTime()
	if err != nil {
		t.Fatalf("GetAbsTime: %v", err)
	}

	if !gotTime.Time().Equal(wantTime.Time()) {
		t.Fatalf("GetAbsTime: got %s, want %s", gotTime.Time(), wantTime.Time())
	}
}
