package dnp3

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// RangeSpec describes how big and how the rangefield is formatted.
type RangeSpecCode uint8 // only 4 bits

const (
	StartStop1 = iota // 0
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

var RangeSpecCodeNames = map[RangeSpecCode]string{
	StartStop1:        "1_OCTET_START_STOP",
	StartStop2:        "2_OCTET_START_STOP",
	StartStop4:        "4_OCTET_START_STOP",
	VirtualStartStop1: "1_OCTET_VIRTUAL_START_STOP",
	VirtualStartStop2: "2_OCTET_VIRTUAL_START_STOP",
	VirtualStartStop4: "4_OCTET_VIRTUAL_START_STOP",
	NoRangeField:      "NO_RANGE_FIELD",
	Count1:            "1_OCTET_COUNT",
	Count2:            "2_OCTET_COUNT",
	Count4:            "4_OCTET_COUNT",
	ReservedA:         "RESERVED",
	Count1Variable:    "COUNT_1_VARIABLE",
	ReservedC:         "RESERVED",
	ReservedD:         "RESERVED",
	ReservedE:         "RESERVED",
	ReservedF:         "RESERVED",
}

func (rsc RangeSpecCode) String() string {
	if name, ok := RangeSpecCodeNames[rsc]; ok {
		return name
	}

	return fmt.Sprintf("unknown object prefix code %d", rsc)
}

type RangeField interface {
	ToBytes() []byte
	FromBytes(data []byte) error
	String() string
	NumObjects() int
}

// RangeField0 1 byte start and stop values.
type RangeField0 struct {
	Start uint8
	Stop  uint8
}

func (rf *RangeField0) ToBytes() []byte {
	return []byte{rf.Start, rf.Stop}
}

func (rf *RangeField0) FromBytes(d []byte) error {
	if len(d) != 2 {
		return fmt.Errorf("requires 2 bytes, got %d", len(d))
	}

	rf.Start = d[0]
	rf.Stop = d[1]

	return nil
}

func (rf *RangeField0) String() string {
	return fmt.Sprintf(`Range Field: (0) 1-octet start and stop indexes
					Start: %d
					Stop : %d`, rf.Start, rf.Stop)
}

func (rf *RangeField0) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField1 2 byte start and stop values.
type RangeField1 struct {
	Start uint16
	Stop  uint16
}

func (rf *RangeField1) ToBytes() []byte {
	var o []byte

	o = binary.LittleEndian.AppendUint16(o, rf.Start)
	o = binary.LittleEndian.AppendUint16(o, rf.Stop)

	return o
}

func (rf *RangeField1) FromBytes(d []byte) error {
	if len(d) != 4 {
		return fmt.Errorf("requires 4 bytes, got %d", len(d))
	}

	rf.Start = binary.LittleEndian.Uint16(d[0:2])
	rf.Stop = binary.LittleEndian.Uint16(d[2:4])

	return nil
}

func (rf *RangeField1) String() string {
	return fmt.Sprintf(`Range Field: (1) 2-octet start and stop indexes
					Start: %d
					Stop : %d`, rf.Start, rf.Stop)
}

func (rf *RangeField1) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField2 4 byte start and stop values.
type RangeField2 struct {
	Start uint32
	Stop  uint32
}

func (rf *RangeField2) ToBytes() []byte {
	var o []byte

	o = binary.LittleEndian.AppendUint32(o, rf.Start)
	o = binary.LittleEndian.AppendUint32(o, rf.Stop)

	return o
}

func (rf *RangeField2) FromBytes(d []byte) error {
	if len(d) != 8 {
		return fmt.Errorf("requires 8 bytes, got %d", len(d))
	}

	rf.Start = binary.LittleEndian.Uint32(d[0:4])
	rf.Stop = binary.LittleEndian.Uint32(d[4:8])

	return nil
}

func (rf *RangeField2) String() string {
	return fmt.Sprintf(`Range Field: (2) 4-octet start and stop indexes
					Start: %d
					Stop : %d`, rf.Start, rf.Stop)
}

func (rf *RangeField2) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField3 1 byte VIRTUAL start and stop values.
type RangeField3 struct {
	Start uint8
	Stop  uint8
}

func (rf *RangeField3) ToBytes() []byte {
	return []byte{rf.Start, rf.Stop}
}

func (rf *RangeField3) FromBytes(d []byte) error {
	if len(d) != 2 {
		return fmt.Errorf("requires 2 bytes, got %d", len(d))
	}

	rf.Start = d[0]
	rf.Stop = d[1]

	return nil
}

func (rf *RangeField3) String() string {
	return fmt.Sprintf(`Range Field: (3) 1-octet virtual start and stop indexes
					Start: %d (virtual)
					Stop : %d (virtual)`, rf.Start, rf.Stop)
}

func (rf *RangeField3) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField4 2 byte VIRTUAL start and stop values.
type RangeField4 struct {
	Start uint16
	Stop  uint16
}

func (rf *RangeField4) ToBytes() []byte {
	var o []byte

	o = binary.LittleEndian.AppendUint16(o, rf.Start)
	o = binary.LittleEndian.AppendUint16(o, rf.Stop)

	return o
}

func (rf *RangeField4) FromBytes(d []byte) error {
	if len(d) != 4 {
		return fmt.Errorf("requires 4 bytes, got %d", len(d))
	}

	rf.Start = binary.LittleEndian.Uint16(d[0:2])
	rf.Stop = binary.LittleEndian.Uint16(d[2:4])

	return nil
}

func (rf *RangeField4) String() string {
	return fmt.Sprintf(`Range Field: (4) 2-octet virtual start and stop indexes
					Start: %d (virtual)
					Stop : %d (virtual)`, rf.Start, rf.Stop)
}

func (rf *RangeField4) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField5 4 byte VIRTUAL start and stop values.
type RangeField5 struct {
	Start uint32
	Stop  uint32
}

func (rf *RangeField5) ToBytes() []byte {
	var o []byte

	o = binary.LittleEndian.AppendUint32(o, rf.Start)
	o = binary.LittleEndian.AppendUint32(o, rf.Stop)

	return o
}

func (rf *RangeField5) FromBytes(d []byte) error {
	if len(d) != 8 {
		return fmt.Errorf("requires 8 bytes, got %d", len(d))
	}

	rf.Start = binary.LittleEndian.Uint32(d[0:4])
	rf.Stop = binary.LittleEndian.Uint32(d[4:8])

	return nil
}

func (rf *RangeField5) String() string {
	return fmt.Sprintf(`Range Field: (5) 4-octet virtual start and stop indexes
					Start: %d (virtual)
					Stop : %d (virtual)`, rf.Start, rf.Stop)
}

func (rf *RangeField5) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField6 - No Range Field: used, Implies all values.
type RangeField6 struct{}

func (rf *RangeField6) ToBytes() []byte {
	return nil
}

func (rf *RangeField6) FromBytes(d []byte) error {
	if len(d) > 0 {
		return errors.New("error Range Field: 6 is an empty Range Field")
	}

	return nil
}

func (rf *RangeField6) String() string { return "" }

func (rf *RangeField6) NumObjects() int { return 0 }

// RangeField7 1 byte count of objects.
type RangeField7 struct {
	Count uint8
}

func (rf *RangeField7) ToBytes() []byte {
	return []byte{rf.Count}
}

func (rf *RangeField7) FromBytes(d []byte) error {
	if len(d) != 1 {
		return fmt.Errorf("requires 1 byte, got %d", len(d))
	}

	rf.Count = d[0]

	return nil
}

func (rf *RangeField7) String() string {
	return fmt.Sprintf(`Range Field: (7) 1-octet count of objects
					Count: %d`, rf.Count)
}

func (rf *RangeField7) NumObjects() int {
	return int(rf.Count)
}

// RangeField8 2 byte count of objects.
type RangeField8 struct {
	Count uint16
}

func (rf *RangeField8) ToBytes() []byte {
	var o []byte

	o = binary.LittleEndian.AppendUint16(o, rf.Count)

	return o
}

func (rf *RangeField8) FromBytes(d []byte) error {
	if len(d) != 2 {
		return fmt.Errorf("requires 2 byte, got %d", len(d))
	}

	rf.Count = binary.LittleEndian.Uint16(d)

	return nil
}

func (rf *RangeField8) String() string {
	return fmt.Sprintf(`Range Field: (8) 2-octet count of objects
					Count: %d`, rf.Count)
}

func (rf *RangeField8) NumObjects() int {
	return int(rf.Count)
}

// RangeField9 4 byte count of objects.
type RangeField9 struct {
	Count uint32
}

func (rf *RangeField9) ToBytes() []byte {
	var o []byte

	o = binary.LittleEndian.AppendUint32(o, rf.Count)

	return o
}

func (rf *RangeField9) FromBytes(d []byte) error {
	if len(d) != 4 {
		return fmt.Errorf("requires 4 byte, got %d", len(d))
	}

	rf.Count = binary.LittleEndian.Uint32(d[0:4])

	return nil
}

func (rf *RangeField9) String() string {
	return fmt.Sprintf(`Range Field: (9) 4-octet count of objects
					Count: %d`, rf.Count)
}

func (rf *RangeField9) NumObjects() int {
	return int(rf.Count)
}

// RangeFieldB 1 byte count of objects with variable format.
type RangeFieldB struct {
	Count uint8
	// Variable ?
}

func (rf *RangeFieldB) ToBytes() []byte {
	return []byte{rf.Count}
}

func (rf *RangeFieldB) FromBytes(d []byte) error {
	if len(d) != 1 {
		return fmt.Errorf("requires 1 byte, got %d", len(d))
	}

	rf.Count = d[0]

	return nil
}

func (rf *RangeFieldB) String() string {
	return fmt.Sprintf(`Range Field:: (B) 1-octet count of objects with variable format
					Count: %d`, rf.Count)
}

func (rf *RangeFieldB) NumObjects() int {
	return int(rf.Count)
}
