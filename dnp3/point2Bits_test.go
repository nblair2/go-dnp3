package dnp3_test

import (
	"slices"
	"testing"

	"github.com/nblair2/go-dnp3/v4/dnp3"
)

// point2BitsObject builds the bytes of a Group 3 Var 1 (packed double-bit
// binary input) data object: a StartStop1 header over indexes [0, stop]
// followed by the packed point bytes.
func point2BitsObject(stop uint8, packed []byte) []byte {
	return append([]byte{0x03, 0x01, 0x00, 0x00, stop}, packed...)
}

// wantPoint2Bits asserts a decoded Point is a *dnp3.Point2Bits with the given value.
func wantPoint2Bits(t *testing.T, point dnp3.Point, want [2]bool) {
	t.Helper()

	point2Bits, ok := point.(*dnp3.Point2Bits)
	if !ok {
		t.Fatalf("point is %T, want *dnp3.Point2Bits", point)
	}

	if point2Bits.Value != want {
		t.Fatalf("value = %v, want %v", point2Bits.Value, want)
	}
}

// TestPoint2Bits_decodeRoundTrip covers the issue's repro shape (2 points
// packed into a single byte) and a 5-point run that crosses a byte boundary,
// verifying decode matches the packer's own bit layout and re-encodes
// byte-identical.
func TestPoint2Bits_decodeRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		packed []byte
		want   [][2]bool
	}{
		{
			// data[pointIndex%8] on a 1-byte buffer panics for pointIndex 1
			// before the fix; byte 0b00001001 packs point0={true,false} in
			// bits 0-1 and point1={false,true} in bits 2-3.
			name:   "issue repro: 2 points in 1 byte",
			packed: []byte{0b00001001},
			want:   [][2]bool{{true, false}, {false, true}},
		},
		{
			name:   "5 points crossing a byte boundary",
			packed: []byte{0x39, 0x01},
			want: [][2]bool{
				{true, false},
				{false, true},
				{true, true},
				{false, false},
				{true, false},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // test data, len(want) is small
			wire := point2BitsObject(uint8(len(testCase.want)-1), testCase.packed)

			object, err := dnp3.NewDataObjectFromBytes(wire)
			if err != nil {
				t.Fatalf("NewDataObjectFromBytes: %v", err)
			}

			if len(object.Points) != len(testCase.want) {
				t.Fatalf("got %d points, want %d", len(object.Points), len(testCase.want))
			}

			for i, want := range testCase.want {
				wantPoint2Bits(t, object.Points[i], want)
			}

			out, err := object.SerializeTo()
			if err != nil {
				t.Fatalf("SerializeTo: %v", err)
			}

			if !slices.Equal(out, wire) {
				t.Fatalf("round-trip mismatch\ngot:  % X\nwant: % X", out, wire)
			}
		})
	}
}
