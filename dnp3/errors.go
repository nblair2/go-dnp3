package dnp3

import (
	"errors"
	"fmt"
)

// Library error categories can be matched with errors.Is, including through
// contextual wrappers returned by frame, object, and point operations.
var (
	// ErrInsufficientData indicates that an input buffer is too short.
	ErrInsufficientData = errors.New("insufficient data")
	// ErrInvalidLength indicates an invalid declared length, exact size, or byte width.
	ErrInvalidLength = errors.New("invalid length")
	// ErrInvalidStartBytes indicates incorrect DNP3 synchronization bytes.
	ErrInvalidStartBytes = errors.New("invalid DNP3 start bytes")
	// ErrCRCMismatch indicates a received CRC differs from the calculated CRC.
	ErrCRCMismatch = errors.New("CRC mismatch")
	// ErrInvalidFunctionCode indicates an unrecognized data-link function code.
	ErrInvalidFunctionCode = errors.New("invalid function code")
	// ErrInvalidControl indicates incompatible control fields.
	ErrInvalidControl = errors.New("invalid control")
	// ErrReservedBits indicates a reserved bit is set.
	ErrReservedBits = errors.New("reserved bits must be zero")
	// ErrInvalidQualifier indicates an invalid range code, prefix code, or prefix width.
	ErrInvalidQualifier = errors.New("invalid qualifier")
	// ErrUnsupportedObject indicates a group/variation cannot be decoded or encoded.
	ErrUnsupportedObject = errors.New("unsupported object")
	// ErrValueOutOfRange indicates invalid numeric bounds or a value outside its protocol field's range.
	ErrValueOutOfRange = errors.New("value out of range")
	// ErrInvalidType indicates an unexpected value, point, or range-field type.
	ErrInvalidType = errors.New("invalid type")
	// ErrMissingField indicates a required field is absent, rather than unsupported.
	ErrMissingField = errors.New("missing required field")
)

// UnsupportedObjectError identifies a group/variation unavailable for the
// attempted operation. Additional operation context may wrap this error.
type UnsupportedObjectError struct {
	Group     byte
	Variation byte
}

func (err *UnsupportedObjectError) Error() string {
	return fmt.Sprintf("unsupported group/variation: %d/%d", err.Group, err.Variation)
}

// Unwrap returns ErrUnsupportedObject.
func (*UnsupportedObjectError) Unwrap() error {
	return ErrUnsupportedObject
}
