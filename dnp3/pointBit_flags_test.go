package dnp3_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/nblair2/go-dnp3/v4/dnp3"
)

// bitFlagsObject builds the bytes of a Group 1 Var 2 (binary input status
// with flags) data object: group, variation, qualifier, count, followed by
// the raw per-point bytes.
func bitFlagsObject(qualifier, count byte, points []byte) []byte {
	return append([]byte{0x01, 0x02, qualifier, count}, points...)
}

// wantPointBitFlags asserts a decoded Point is a *dnp3.PointBit with the
// given index (nil for the unprefixed case), value, and Online flag.
func wantPointBitFlags(t *testing.T, point dnp3.Point, wantIndex *int, wantValue, wantOnline bool) {
	t.Helper()

	pointBit, ok := point.(*dnp3.PointBit)
	if !ok {
		t.Fatalf("point is %T, want *dnp3.PointBit", point)
	}

	if pointBit.Value != wantValue {
		t.Fatalf("value = %v, want %v", pointBit.Value, wantValue)
	}

	index, err := point.GetIndex()
	if wantIndex == nil {
		if !errors.Is(err, dnp3.ErrNoIndex) {
			t.Fatalf("GetIndex err = %v, want ErrNoIndex", err)
		}
	} else {
		if err != nil {
			t.Fatalf("GetIndex: %v", err)
		}

		if index != *wantIndex {
			t.Fatalf("index = %d, want %d", index, *wantIndex)
		}
	}

	flags, err := point.GetFlags()
	if err != nil {
		t.Fatalf("GetFlags: %v", err)
	}

	if flags.Online != wantOnline {
		t.Fatalf("online = %v, want %v", flags.Online, wantOnline)
	}
}

// bitFlagsCase is a single decode/round-trip case for TestPointBitFlags_decodeRoundTrip.
type bitFlagsCase struct {
	name       string
	qualifier  byte
	count      byte
	points     []byte
	wantIndex  []*int
	wantValue  []bool
	wantOnline []bool
}

// runBitFlagsCase decodes testCase's wire bytes, checks every point, and
// verifies SerializeTo reproduces the original wire exactly.
func runBitFlagsCase(t *testing.T, testCase bitFlagsCase) {
	t.Helper()

	wire := bitFlagsObject(testCase.qualifier, testCase.count, testCase.points)

	object, err := dnp3.NewDataObjectFromBytes(wire)
	if err != nil {
		t.Fatalf("NewDataObjectFromBytes: %v", err)
	}

	if len(object.Points) != len(testCase.wantValue) {
		t.Fatalf("got %d points, want %d", len(object.Points), len(testCase.wantValue))
	}

	for pointIndex, want := range testCase.wantValue {
		wantPointBitFlags(
			t,
			object.Points[pointIndex],
			testCase.wantIndex[pointIndex],
			want,
			testCase.wantOnline[pointIndex],
		)
	}

	out, err := object.SerializeTo()
	if err != nil {
		t.Fatalf("SerializeTo: %v", err)
	}

	if !slices.Equal(out, wire) {
		t.Fatalf("round-trip mismatch\ngot:  % X\nwant: % X", out, wire)
	}
}

// TestPointBitFlags_decodeRoundTrip covers the issue's repro shape (a single
// index-prefixed point) and confirms the unprefixed count path still works,
// verifying decode matches the wire and re-encodes byte-identical.
func TestPointBitFlags_decodeRoundTrip(t *testing.T) {
	t.Parallel()

	indexFive := 5

	tests := []bitFlagsCase{
		{
			// index out of range [1] with length 1 before the fix: a single
			// byte was passed into DecodeFromBytes regardless of prefSize.
			name:       "issue repro: index-prefixed point",
			qualifier:  0x17, // Index1Octet prefix, Count1
			count:      0x01,
			points:     []byte{0x05, 0x81}, // index 5, value+online
			wantIndex:  []*int{&indexFive},
			wantValue:  []bool{true},
			wantOnline: []bool{true},
		},
		{
			name:       "unprefixed count path",
			qualifier:  0x07, // NoPrefix, Count1
			count:      0x02,
			points:     []byte{0x81, 0x00},
			wantIndex:  []*int{nil, nil},
			wantValue:  []bool{true, false},
			wantOnline: []bool{true, false},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runBitFlagsCase(t, testCase)
		})
	}
}

// consumedSizeWire builds a 2-point index-prefixed flags object immediately
// followed by an unrelated packed-bit object, so a wrong consumed-size
// return misaligns the second object's header.
func consumedSizeWire() []byte {
	flagsObject := bitFlagsObject(0x17, 0x02, []byte{0x00, 0x81, 0x01, 0x06})
	packedObject := []byte{0x01, 0x01, 0x07, 0x03, 0x05} // group 1 var 1, 3 packed bits: 1,0,1

	return append(slices.Clone(flagsObject), packedObject...)
}

// assertFlagsObject checks the 2-point index-prefixed flags object decoded
// from consumedSizeWire.
func assertFlagsObject(t *testing.T, object dnp3.DataObject) {
	t.Helper()

	indexZero, indexOne := 0, 1

	if object.Header.Group != 1 || object.Header.Variation != 2 {
		t.Fatalf("object 0 = grp %d var %d, want grp 1 var 2",
			object.Header.Group, object.Header.Variation)
	}

	if len(object.Points) != 2 {
		t.Fatalf("object 0: got %d points, want 2", len(object.Points))
	}

	wantPointBitFlags(t, object.Points[0], &indexZero, true, true)
	wantPointBitFlags(t, object.Points[1], &indexOne, false, false)
}

// assertPackedObject checks the trailing packed-bit object decoded from
// consumedSizeWire, confirming it was parsed at the correct offset.
func assertPackedObject(t *testing.T, object dnp3.DataObject) {
	t.Helper()

	if object.Header.Group != 1 || object.Header.Variation != 1 {
		t.Fatalf("object 1 = grp %d var %d, want grp 1 var 1",
			object.Header.Group, object.Header.Variation)
	}

	wantBits := []bool{true, false, true}
	if len(object.Points) != len(wantBits) {
		t.Fatalf("object 1: got %d points, want %d", len(object.Points), len(wantBits))
	}

	for pointIndex, want := range wantBits {
		value, ok := object.Points[pointIndex].GetValue().(bool)
		if !ok {
			t.Fatalf(
				"object 1 point %d: GetValue() = %T, want bool",
				pointIndex,
				object.Points[pointIndex].GetValue(),
			)
		}

		if value != want {
			t.Fatalf("object 1 point %d: value = %v, want %v", pointIndex, value, want)
		}
	}
}

// TestPointBitFlags_consumedSize pins the consumed-size math for
// index-prefixed points: a 2-point object is followed by a second object,
// so a wrong size (num instead of num*(prefSize+1)) misaligns the next
// header and this test fails loudly instead of the objects quietly
// decoding wrong.
func TestPointBitFlags_consumedSize(t *testing.T) {
	t.Parallel()

	wire := consumedSizeWire()

	appData, err := dnp3.NewApplicationDataFromBytes(wire)
	if err != nil {
		t.Fatalf("NewApplicationDataFromBytes: %v", err)
	}

	if len(appData.Objects) != 2 {
		t.Fatalf("got %d objects, want 2", len(appData.Objects))
	}

	assertFlagsObject(t, appData.Objects[0])
	assertPackedObject(t, appData.Objects[1])

	out, err := appData.SerializeTo()
	if err != nil {
		t.Fatalf("SerializeTo: %v", err)
	}

	if !slices.Equal(out, wire) {
		t.Fatalf("round-trip mismatch\ngot:  % X\nwant: % X", out, wire)
	}
}
