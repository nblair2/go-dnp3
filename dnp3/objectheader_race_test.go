package dnp3_test

import (
	"sync"
	"testing"

	"github.com/nblair2/go-dnp3/v3/dnp3"
)

// numRaceWorkers is the goroutine count used to race concurrent readers
// against each other below.
const numRaceWorkers = 64

// newRaceHeader builds an ObjectHeader through exported fields only, so the
// unexported objectType cache starts out nil, exactly as it does after
// decoding a group/variation ObjectHeader.String has no cached entry for yet.
func newRaceHeader(t *testing.T) dnp3.ObjectHeader {
	t.Helper()

	return dnp3.ObjectHeader{
		Group:      30,
		Variation:  1,
		RangeField: &dnp3.AllRangeField{},
	}
}

// TestObjectHeaderStringRace guards against ObjectHeader.String reintroducing
// a lazy write to the unexported objectType field. It passes today because
// String only reads the objectTypes map, but fails under -race if a write is
// ever reintroduced (issue #43).
func TestObjectHeaderStringRace(t *testing.T) {
	t.Parallel()

	header := newRaceHeader(t)

	var readers sync.WaitGroup

	for range numRaceWorkers {
		readers.Go(func() {
			_ = header.String()
		})
	}

	readers.Wait()
}

// TestDataObjectSerializeToRace guards against DataObject.SerializeTo
// reintroducing a lazy write to the header's unexported objectType field. It
// passes today because SerializeTo only reads the objectTypes map, but fails
// under -race if a write is ever reintroduced (issue #43).
func TestDataObjectSerializeToRace(t *testing.T) {
	t.Parallel()

	obj := &dnp3.DataObject{
		Header: newRaceHeader(t),
		Points: []dnp3.Point{&dnp3.PointBytes{}},
	}

	var readers sync.WaitGroup

	for range numRaceWorkers {
		readers.Go(func() {
			_ = obj.String()

			_, err := obj.SerializeTo()
			if err != nil {
				t.Errorf("SerializeTo: %v", err)
			}
		})
	}

	readers.Wait()
}
