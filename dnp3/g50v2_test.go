package dnp3_test

import (
	"slices"
	"testing"

	"github.com/nblair2/go-dnp3/v3/dnp3"
)

// TestG50V2AbsoluteTimeBeforeValue verifies that a Group 50 Variation 2
// point decodes the 6-byte absolute time before the 4-byte interval,
// matching the wire order, and that re-encoding reproduces the input.
func TestG50V2AbsoluteTimeBeforeValue(t *testing.T) {
	t.Parallel()

	absTimeBytes := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
	intervalBytes := []byte{0xAA, 0xBB, 0xCC, 0xDD}

	point := slices.Concat(absTimeBytes, intervalBytes)
	// Header: group 50, variation 2, NoPrefix/1-octet-count qualifier, count 1.
	raw := slices.Concat([]byte{50, 2, 0x07, 0x01}, point)

	obj, err := dnp3.NewDataObjectFromBytes(raw)
	if err != nil {
		t.Fatalf("NewDataObjectFromBytes: %v", err)
	}

	if len(obj.Points) != 1 {
		t.Fatalf("expected 1 point, got %d", len(obj.Points))
	}

	got := obj.Points[0]

	wantTime, err := dnp3.BytesToDNP3TimeAbsolute(absTimeBytes)
	if err != nil {
		t.Fatalf("test setup: BytesToDNP3TimeAbsolute: %v", err)
	}

	gotTime, err := got.GetAbsTime()
	if err != nil {
		t.Fatalf("GetAbsTime: %v", err)
	}

	if !gotTime.Time().Equal(wantTime.Time()) {
		t.Fatalf("GetAbsTime: got %s, want %s", gotTime.Time(), wantTime.Time())
	}

	gotValue, ok := got.GetValue().([]byte)
	if !ok {
		t.Fatalf("GetValue: expected []byte, got %T", got.GetValue())
	}

	if !slices.Equal(gotValue, intervalBytes) {
		t.Fatalf("GetValue: got 0x % X, want 0x % X", gotValue, intervalBytes)
	}

	out, err := obj.SerializeTo()
	if err != nil {
		t.Fatalf("SerializeTo: %v", err)
	}

	if !slices.Equal(out, raw) {
		t.Fatalf("round trip mismatch: got 0x % X, want 0x % X", out, raw)
	}
}
