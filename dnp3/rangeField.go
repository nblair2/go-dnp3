package dnp3

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// RangeSpecCode - describes rangefield is format and size.
//
//go:generate stringer -type=RangeSpecCode
type RangeSpecCode uint8 // only 4 bits

const (
	StartStop1 RangeSpecCode = iota // 0
	StartStop2
	StartStop4
	VirtualStartStop1
	VirtualStartStop2
	VirtualStartStop4
	NoRangeField
	Count1
	Count2
	Count4
	ReservedA
	Count1Variable // B
	ReservedC
	ReservedD
	ReservedE
	ReservedF
)

type RangeField interface {
	SerializeTo() ([]byte, error)
	DecodeFromBytes(data []byte) error
	String() string
	NumObjects() int
	Size() int
}

type rangeFieldConstructor func() RangeField

var rangeFieldConstructors = map[RangeSpecCode]rangeFieldConstructor{
	StartStop1: func() RangeField { return &StartStopRangeField{byteWidth: 1, code: StartStop1} },
	StartStop2: func() RangeField { return &StartStopRangeField{byteWidth: 2, code: StartStop2} },
	StartStop4: func() RangeField { return &StartStopRangeField{byteWidth: 4, code: StartStop4} },
	VirtualStartStop1: func() RangeField {
		return &StartStopRangeField{byteWidth: 1, virtual: true, code: VirtualStartStop1}
	},
	VirtualStartStop2: func() RangeField {
		return &StartStopRangeField{byteWidth: 2, virtual: true, code: VirtualStartStop2}
	},
	VirtualStartStop4: func() RangeField {
		return &StartStopRangeField{byteWidth: 4, virtual: true, code: VirtualStartStop4}
	},
	NoRangeField:   func() RangeField { return &AllRangeField{} },
	Count1:         func() RangeField { return &CountRangeField{byteWidth: 1, code: Count1} },
	Count2:         func() RangeField { return &CountRangeField{byteWidth: 2, code: Count2} },
	Count4:         func() RangeField { return &CountRangeField{byteWidth: 4, code: Count4} },
	Count1Variable: func() RangeField { return &CountRangeField{byteWidth: 1, variable: true, code: Count1Variable} },
}

var reservedRangeSpecifiers = map[RangeSpecCode]struct{}{
	ReservedA: {},
	ReservedC: {},
	ReservedD: {},
	ReservedE: {},
	ReservedF: {},
}

func rangeFieldConstructorFor(code RangeSpecCode) (rangeFieldConstructor, error) {
	if _, invalid := reservedRangeSpecifiers[code]; invalid {
		return nil, fmt.Errorf("range specifier code %d not valid", code)
	}

	constructor, ok := rangeFieldConstructors[code]
	if !ok {
		return nil, fmt.Errorf("unknown range specifier code %d", code)
	}

	return constructor, nil
}

// StartStopRangeField represents range spec codes 0-5: start/stop index pairs
// with configurable byte width (1, 2, or 4) and optional virtual addressing.
type StartStopRangeField struct {
	Start     uint32 `json:"start"`
	Stop      uint32 `json:"stop"`
	byteWidth int
	virtual   bool
	code      RangeSpecCode
}

func (rf *StartStopRangeField) SerializeTo() ([]byte, error) {
	var encoded []byte

	switch rf.byteWidth {
	case 1:
		if rf.Start > 0xFF || rf.Stop > 0xFF {
			return nil, fmt.Errorf(
				"values exceed 1-byte range: start=%d, stop=%d",
				rf.Start,
				rf.Stop,
			)
		}

		encoded = append(encoded, byte(rf.Start), byte(rf.Stop))
	case 2:
		if rf.Start > 0xFFFF || rf.Stop > 0xFFFF {
			return nil, fmt.Errorf(
				"values exceed 2-byte range: start=%d, stop=%d",
				rf.Start,
				rf.Stop,
			)
		}

		encoded = binary.LittleEndian.AppendUint16(encoded, uint16(rf.Start))
		encoded = binary.LittleEndian.AppendUint16(encoded, uint16(rf.Stop))
	case 4:
		encoded = binary.LittleEndian.AppendUint32(encoded, rf.Start)
		encoded = binary.LittleEndian.AppendUint32(encoded, rf.Stop)
	default:
		return nil, fmt.Errorf("invalid byte width %d", rf.byteWidth)
	}

	return encoded, nil
}

func (rf *StartStopRangeField) DecodeFromBytes(data []byte) error {
	expected := rf.byteWidth * 2
	if len(data) != expected {
		return fmt.Errorf("requires %d bytes, got %d", expected, len(data))
	}

	switch rf.byteWidth {
	case 1:
		rf.Start = uint32(data[0])
		rf.Stop = uint32(data[1])
	case 2:
		rf.Start = uint32(binary.LittleEndian.Uint16(data[0:2]))
		rf.Stop = uint32(binary.LittleEndian.Uint16(data[2:4]))
	case 4:
		rf.Start = binary.LittleEndian.Uint32(data[0:4])
		rf.Stop = binary.LittleEndian.Uint32(data[4:8])
	default:
		return fmt.Errorf("invalid byte width %d", rf.byteWidth)
	}

	return nil
}

func (rf *StartStopRangeField) String() string {
	label := "start and stop indexes"
	suffix := ""

	if rf.virtual {
		label = "virtual start and stop indexes"
		suffix = " (virtual)"
	}

	return fmt.Sprintf("Range Field: (%X) %d-octet %s\n\tStart: %d%s\n\tStop : %d%s",
		uint8(rf.code), rf.byteWidth, label, rf.Start, suffix, rf.Stop, suffix)
}

func (rf *StartStopRangeField) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

func (rf *StartStopRangeField) Size() int { return rf.byteWidth * 2 }

// ByteWidth returns the number of bytes used for each index value on the wire.
func (rf *StartStopRangeField) ByteWidth() int { return rf.byteWidth }

// IsVirtual reports whether this range uses virtual addressing.
func (rf *StartStopRangeField) IsVirtual() bool { return rf.virtual }

// Code returns the original RangeSpecCode this field was created from.
func (rf *StartStopRangeField) Code() RangeSpecCode { return rf.code }

// CountRangeField represents range spec codes 7-9 and B: a count of objects
// with configurable byte width (1, 2, or 4) and optional variable format.
type CountRangeField struct {
	Count     uint32 `json:"count"`
	byteWidth int
	variable  bool
	code      RangeSpecCode
}

func (rf *CountRangeField) SerializeTo() ([]byte, error) {
	var encoded []byte

	switch rf.byteWidth {
	case 1:
		if rf.Count > 0xFF {
			return nil, fmt.Errorf("count exceeds 1-byte range: %d", rf.Count)
		}

		encoded = append(encoded, byte(rf.Count))
	case 2:
		if rf.Count > 0xFFFF {
			return nil, fmt.Errorf("count exceeds 2-byte range: %d", rf.Count)
		}

		encoded = binary.LittleEndian.AppendUint16(encoded, uint16(rf.Count))
	case 4:
		encoded = binary.LittleEndian.AppendUint32(encoded, rf.Count)
	default:
		return nil, fmt.Errorf("invalid byte width %d", rf.byteWidth)
	}

	return encoded, nil
}

func (rf *CountRangeField) DecodeFromBytes(data []byte) error {
	if len(data) != rf.byteWidth {
		return fmt.Errorf("requires %d byte(s), got %d", rf.byteWidth, len(data))
	}

	switch rf.byteWidth {
	case 1:
		rf.Count = uint32(data[0])
	case 2:
		rf.Count = uint32(binary.LittleEndian.Uint16(data))
	case 4:
		rf.Count = binary.LittleEndian.Uint32(data[0:4])
	default:
		return fmt.Errorf("invalid byte width %d", rf.byteWidth)
	}

	return nil
}

func (rf *CountRangeField) String() string {
	label := "count of objects"
	if rf.variable {
		label = "count of objects with variable format"
	}

	return fmt.Sprintf("Range Field: (%X) %d-octet %s\n\tCount: %d",
		uint8(rf.code), rf.byteWidth, label, rf.Count)
}

func (rf *CountRangeField) NumObjects() int {
	return int(rf.Count)
}

func (rf *CountRangeField) Size() int { return rf.byteWidth }

// ByteWidth returns the number of bytes used for the count value on the wire.
func (rf *CountRangeField) ByteWidth() int { return rf.byteWidth }

// IsVariable reports whether this count uses variable object format.
func (rf *CountRangeField) IsVariable() bool { return rf.variable }

// Code returns the original RangeSpecCode this field was created from.
func (rf *CountRangeField) Code() RangeSpecCode { return rf.code }

// AllRangeField represents range spec code 6: no range data, implies all values.
type AllRangeField struct{}

func (rf *AllRangeField) SerializeTo() ([]byte, error) {
	return nil, nil
}

func (rf *AllRangeField) DecodeFromBytes(data []byte) error {
	if len(data) > 0 {
		return errors.New("AllRangeField is an empty range field")
	}

	return nil
}

func (rf *AllRangeField) String() string { return "" }

func (rf *AllRangeField) NumObjects() int { return 0 }

func (rf *AllRangeField) Size() int { return 0 }

// Code returns the RangeSpecCode for this field type.
func (rf *AllRangeField) Code() RangeSpecCode { return NoRangeField }
