package dnp3_test

import (
	"testing"
	"time"

	"github.com/nblair2/go-dnp3/v4/dnp3"
)

// maxStartStopRequest is a Read request for group 1 var 1 (Binary Input,
// packed format) with a 4-byte start/stop range field where Start == Stop
// == 0xFFFFFFFF, plus one packed-bit data byte for the resulting point.
var maxStartStopRequest = []byte{
	0xc0, 0x01, // control, function code (Read)
	0x01, 0x01, 0x02, // group 1, var 1, qualifier (4-byte start/stop)
	0xff, 0xff, 0xff, 0xff, // start = 0xFFFFFFFF
	0xff, 0xff, 0xff, 0xff, // stop = 0xFFFFFFFF
	0x01, // packed-bit data for the single point
}

// TestUpdateIndexes_startStopMaxDoesNotWrap verifies that a StartStop range
// with Start == Stop == 0xFFFFFFFF decodes promptly with a single index,
// rather than wrapping uint32 and looping forever.
func TestUpdateIndexes_startStopMaxDoesNotWrap(t *testing.T) {
	t.Parallel()

	type result struct {
		request *dnp3.ApplicationRequest
		err     error
	}

	done := make(chan result, 1)

	go func() {
		request, err := dnp3.NewApplicationRequestFromBytes(maxStartStopRequest)
		done <- result{request, err}
	}()

	select {
	case res := <-done:
		if res.err != nil {
			t.Fatalf("DecodeFromBytes: unexpected error: %v", res.err)
		}

		objects := res.request.GetData().Objects
		if len(objects) != 1 {
			t.Fatalf("Objects: got %d, want 1", len(objects))
		}

		indexes := objects[0].Indexes()
		if len(indexes) != 1 {
			t.Fatalf("Indexes: got %d entries, want 1", len(indexes))
		}

		if indexes[0] != 0xffffffff {
			t.Errorf("Indexes[0] = %#x, want 0xffffffff", indexes[0])
		}
	case <-time.After(2 * time.Second):
		t.Fatal(
			"DecodeFromBytes did not return: updateIndexes likely wrapped uint32 and looped forever",
		)
	}
}
