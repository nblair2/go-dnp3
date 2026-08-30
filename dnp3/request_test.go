package dnp3_test

import (
	"testing"

	"github.com/nblair2/go-dnp3/v4/dnp3"
)

// TestApplicationRequest_DecodeFromBytes_shortInput verifies short input is
// rejected with an error rather than panicking.
func TestApplicationRequest_DecodeFromBytes_shortInput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input []byte
	}{
		{"nil", nil},
		{"empty", []byte{}},
		{"oneByte", []byte{0x00}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			appreq := &dnp3.ApplicationRequest{}

			err := appreq.DecodeFromBytes(testCase.input)
			if err == nil {
				t.Fatalf("DecodeFromBytes(%x): expected error, got nil", testCase.input)
			}
		})
	}
}

// TestApplicationRequest_DecodeFromBytes_minimalHeader verifies a 2-byte
// header (control + function code, no data) still decodes.
func TestApplicationRequest_DecodeFromBytes_minimalHeader(t *testing.T) {
	t.Parallel()

	appreq := &dnp3.ApplicationRequest{}

	err := appreq.DecodeFromBytes([]byte{0xc4, byte(dnp3.Read)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := appreq.GetFunctionCode(); got != byte(dnp3.Read) {
		t.Fatalf("function code = %d, want %d", got, byte(dnp3.Read))
	}
}
