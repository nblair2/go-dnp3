package dnp3

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Point is the interface implemented by all DNP3 point types.
type Point interface {
	DecodeFromBytes(data []byte, prefSize int) error
	SerializeTo() ([]byte, error)
	String() string
	Fields() PointFields
	DataType() PointDataType
	GetIndex() (int, error)
	GetFlags() (PointFlags, error)
	GetAbsTime() (AbsoluteTime, error)
	GetRelTime() (RelativeTime, error)
	GetValue() any
	SetIndex(index int) error
	SetFlags(flags PointFlags) error
	SetAbsTime(absTime AbsoluteTime) error
	SetRelTime(relTime RelativeTime) error
	SetValue(value any) error
}

// PointFields describes which optional fields are present on a Point.
type PointFields struct {
	Index        bool `json:"index"`
	Size         bool `json:"size"`
	Flags        bool `json:"flags"`
	Value        bool `json:"value"`
	AbsoluteTime bool `json:"absolute_time"`
	RelativeTime bool `json:"relative_time"`
}

// PointDataType identifies the kind of data a Point holds.
type PointDataType string

const (
	PointDataTypeBit   PointDataType = "1-bit"
	PointDataType2Bits PointDataType = "2-bit"
	PointDataTypeBytes PointDataType = "bytes"
)

type PointsConstructor func([]byte, int, int, PointPrefixCode) ([]Point, int, error)

type PointsPacker func([]Point) ([]byte, error)

// Sentinel errors for unsupported Point field access.
var (
	ErrNoIndex   = errors.New("point does not have an index")
	ErrNoFlags   = errors.New("point type does not have flags")
	ErrNoAbsTime = errors.New("point type does not have an absolute time")
	ErrNoRelTime = errors.New("point type does not have a relative time")
)

// PointFlags describes the quality flags common to most DNP3 point types.
type PointFlags struct {
	Reserved       bool `json:"reserved"` // should be 0
	PointValue     bool `json:"point_value"`
	ReferenceCheck bool `json:"reference_check"`
	OverRange      bool `json:"over_range"`
	LocalForce     bool `json:"local_force"`
	RemoteForce    bool `json:"remote_force"`
	CommFail       bool `json:"comm_fail"`
	Restart        bool `json:"restart"`
	Online         bool `json:"online"`
}

func (f *PointFlags) FromByte(data byte) error {
	f.Reserved = data&0b10000000 != 0
	if f.Reserved {
		return errors.New("reserved bit must be 0")
	}

	f.ReferenceCheck = data&0b01000000 != 0
	f.OverRange = data&0b00100000 != 0
	f.LocalForce = data&0b00010000 != 0
	f.RemoteForce = data&0b00001000 != 0
	f.CommFail = data&0b00000100 != 0
	f.Restart = data&0b00000010 != 0
	f.Online = data&0b00000001 != 0

	return nil
}

func (f *PointFlags) ToByte() byte {
	var flagByte byte
	if f.Reserved {
		flagByte |= 0b10000000
	}

	if f.ReferenceCheck {
		flagByte |= 0b01000000
	}

	if f.OverRange {
		flagByte |= 0b00100000
	}

	if f.LocalForce {
		flagByte |= 0b00010000
	}

	if f.RemoteForce {
		flagByte |= 0b00001000
	}

	if f.CommFail {
		flagByte |= 0b00000100
	}

	if f.Restart {
		flagByte |= 0b00000010
	}

	if f.Online {
		flagByte |= 0b00000001
	}

	return flagByte
}

func (f *PointFlags) String() string {
	return fmt.Sprintf(`Flags:
Reference Check: %t
Over-Range     : %t
Local Force    : %t
Remote Force   : %t
Comm Fail      : %t
Restart        : %t
Online         : %t`,
		f.ReferenceCheck, f.OverRange, f.LocalForce, f.RemoteForce,
		f.CommFail, f.Restart, f.Online)
}

// --- Prefix helpers ---

// prefixToInt decodes a little-endian prefix byte slice as an int.
func prefixToInt(prefix []byte) (int, error) {
	switch len(prefix) {
	case 0:
		return 0, ErrNoIndex
	case 1:
		return int(prefix[0]), nil
	case 2:
		return int(binary.LittleEndian.Uint16(prefix)), nil
	case 3:
		var padded [4]byte
		copy(padded[:3], prefix)

		return int(binary.LittleEndian.Uint32(padded[:])), nil
	case 4:
		return int(binary.LittleEndian.Uint32(prefix)), nil
	default:
		return 0, fmt.Errorf("unsupported prefix size: %d bytes", len(prefix))
	}
}

// intToPrefixSized encodes an int as a little-endian byte slice of exactly
// the given width (1, 2, 3, or 4 bytes).
func intToPrefixSized(value int, size int) ([]byte, error) {
	switch size {
	case 1:
		if value > 0xFF {
			return nil, fmt.Errorf("value %d exceeds 1-byte index max (255)", value)
		}

		// #nosec G115 -- value range is clamped above
		return []byte{byte(value)}, nil
	case 2:
		if value > 0xFFFF {
			return nil, fmt.Errorf("value %d exceeds 2-byte index max (65535)", value)
		}

		b := make([]byte, 2)
		// #nosec G115 -- value range is clamped above
		binary.LittleEndian.PutUint16(b, uint16(value))

		return b, nil
	case 3:
		if value > 0xFFFFFF {
			return nil, fmt.Errorf("value %d exceeds 3-byte index max (16777215)", value)
		}

		b := make([]byte, 4)
		//nolint:gosec // G115 - value range is clamped above
		binary.LittleEndian.PutUint32(b, uint32(value))

		return b[:3], nil
	case 4:
		b := make([]byte, 4)
		//nolint:gosec // G115 - value range is clamped by int
		binary.LittleEndian.PutUint32(b, uint32(value))

		return b, nil
	default:
		return nil, fmt.Errorf("unsupported index size: %d bytes", size)
	}
}

// intToPrefix encodes an int as a little-endian byte slice. If size is
// positive the encoding uses exactly that many bytes; otherwise the
// minimal DNP3 prefix width (1, 2, or 4 bytes) is chosen automatically.
func intToPrefix(value int, size int) ([]byte, error) {
	if value < 0 {
		return nil, fmt.Errorf("index value must be non-negative, got %d", value)
	}

	if size == 0 {
		switch {
		case value <= 0xFF:
			size = 1
		case value <= 0xFFFF:
			size = 2
		default:
			size = 4
		}
	}

	return intToPrefixSized(value, size)
}

// setIndex validates and stores an index value, auto-determining
// the encoding width if one has not been set.
func setIndex(index **int, indexSize *int, value int) error {
	if value < 0 {
		return fmt.Errorf("index must be non-negative, got %d", value)
	}

	if *indexSize > 0 {
		_, err := intToPrefix(value, *indexSize)
		if err != nil {
			return err
		}
	} else {
		switch {
		case value <= 0xFF:
			*indexSize = 1
		case value <= 0xFFFF:
			*indexSize = 2
		default:
			*indexSize = 4
		}
	}

	*index = &value

	return nil
}

// --- General-purpose constructors/packers ---

func packPointsBytes(points []Point) ([]byte, error) {
	var encoded []byte

	for _, point := range points {
		b, err := point.SerializeTo()
		if err != nil {
			return nil, fmt.Errorf("error packing points: %w", err)
		}

		encoded = append(encoded, b...)
	}

	return encoded, nil
}

func constructorNoPoints(_ []byte, num, _ int, _ PointPrefixCode) ([]Point, int, error) {
	if num != 0 {
		return nil, 0, fmt.Errorf("no points expected, got %d", num)
	}

	return nil, 0, nil
}

func packNoPoints(points []Point) ([]byte, error) {
	if len(points) != 0 {
		return nil, fmt.Errorf("no points expected, got %d", len(points))
	}

	return nil, nil
}
