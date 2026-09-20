package dnp3_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/nblair2/go-dnp3/v4/dnp3"
)

func TestFrameErrorCategories(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		offset int
		value  byte
		want   error
	}{
		{"start bytes", 0, 0, dnp3.ErrInvalidStartBytes},
		{"length", 2, 4, dnp3.ErrInvalidLength},
		{"function code", 3, 0xc1, dnp3.ErrInvalidFunctionCode},
		{"FCV", 3, 0xd4, dnp3.ErrInvalidControl},
		{"header CRC", 8, readClass1230[8] ^ 1, dnp3.ErrCRCMismatch},
		{
			"body CRC",
			len(readClass1230) - 1,
			readClass1230[len(readClass1230)-1] ^ 1,
			dnp3.ErrCRCMismatch,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			data := slices.Clone(readClass1230)

			data[test.offset] = test.value
			if test.offset < 8 {
				copy(data[8:10], dnp3.CalculateDNP3CRC(data[:8]))
			}

			_, err := dnp3.NewFrameFromBytes(data)
			if !errors.Is(err, test.want) {
				t.Fatalf("NewFrameFromBytes: got %v, want %v", err, test.want)
			}
		})
	}
}

func TestValidationErrorCategories(t *testing.T) {
	t.Parallel()

	iin := new(dnp3.ApplicationInternalIndications)
	point := new(dnp3.PointBit)
	_, missing := new(dnp3.ObjectHeader).SerializeTo()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{"short IIN", iin.DecodeFromBytes([]byte{0}), dnp3.ErrInsufficientData},
		{"long IIN", iin.DecodeFromBytes([]byte{0, 0, 0}), dnp3.ErrInvalidLength},
		{"reserved bits", iin.DecodeFromBytes([]byte{0, 0x40}), dnp3.ErrReservedBits},
		{"packed prefix", point.DecodeFromBytes([]byte{0}, 1), dnp3.ErrInvalidQualifier},
		{"value type", point.SetValue(1), dnp3.ErrInvalidType},
		{"missing range", missing, dnp3.ErrMissingField},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if !errors.Is(test.err, test.want) {
				t.Fatalf("got %v, want %v", test.err, test.want)
			}
		})
	}
}
