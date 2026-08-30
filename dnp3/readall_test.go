package dnp3_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/nblair2/go-dnp3/v4/dnp3"
)

// noConstructorGroups are the var-0 "Any Variations" groups that have no
// Constructor in the object type table. A read-all request for one of these
// must still decode as a zero-point object.
var noConstructorGroups = []uint8{
	1, 3, 4, 10, 11, 12, 13, 20, 21, 22, 23, 30, 31, 32, 33, 34, 40, 41, 42, 43,
}

// readAllRequest builds a Read request for group/variation 0, qualifier 0x06
// (no range field, implies all points).
func readAllRequest(group uint8) []byte {
	return []byte{0xC1, byte(dnp3.Read), group, 0x00, 0x06}
}

func TestReadAllVariationZero(t *testing.T) {
	t.Parallel()

	for _, group := range noConstructorGroups {
		t.Run(fmt.Sprintf("Group%d", group), func(t *testing.T) {
			t.Parallel()

			input := readAllRequest(group)
			request := &dnp3.ApplicationRequest{}

			err := request.DecodeFromBytes(input)
			if err != nil {
				t.Fatalf("DecodeFromBytes: %v", err)
			}

			data := request.GetData()
			if data.HasExtra() {
				t.Fatalf("unexpected extra: % X", data.GetExtra())
			}

			if len(data.Objects) != 1 {
				t.Fatalf("Objects = %d, want 1", len(data.Objects))
			}

			object := data.Objects[0]
			if object.Header.Group != group || object.Header.Variation != 0 {
				t.Fatalf("got %d/%d, want %d/0",
					object.Header.Group, object.Header.Variation, group)
			}

			if len(object.Points) != 0 {
				t.Fatalf("Points = %d, want 0", len(object.Points))
			}

			output, err := request.SerializeTo()
			if err != nil {
				t.Fatalf("SerializeTo: %v", err)
			}

			if !slices.Equal(output, input) {
				t.Fatalf("round-trip mismatch\n got: % X\nwant: % X", output, input)
			}
		})
	}
}

// TestReadAllVariationZero_hasConstructor guards against a regression on
// {2,0}, the one var-0 entry that already had a Constructor.
func TestReadAllVariationZero_hasConstructor(t *testing.T) {
	t.Parallel()

	input := readAllRequest(2)
	request := &dnp3.ApplicationRequest{}

	err := request.DecodeFromBytes(input)
	if err != nil {
		t.Fatalf("DecodeFromBytes: %v", err)
	}

	data := request.GetData()
	if len(data.Objects) != 1 || data.Objects[0].Header.Variation != 0 {
		t.Fatalf("unexpected objects: %+v", data.Objects)
	}

	output, err := request.SerializeTo()
	if err != nil {
		t.Fatalf("SerializeTo: %v", err)
	}

	if !slices.Equal(output, input) {
		t.Fatalf("round-trip mismatch\n got: % X\nwant: % X", output, input)
	}
}

// TestReadAllVariationZero_unknownGroup checks that a group/variation absent
// from the object type table is still rejected, even with a zero-point
// range field.
func TestReadAllVariationZero_unknownGroup(t *testing.T) {
	t.Parallel()

	request := &dnp3.ApplicationRequest{}

	err := request.DecodeFromBytes(readAllRequest(255))
	if err == nil {
		t.Fatal("expected an error for an unknown group/variation")
	}

	if !strings.Contains(err.Error(), "unsupported group/variation") {
		t.Fatalf("error = %v, want it to mention unsupported group/variation", err)
	}
}
