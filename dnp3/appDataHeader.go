package dnp3

import (
	"encoding/binary"
	"fmt"
)

// ObjectHeaders are used to describe the structure of application data
type ObjectHeader struct {
	// Object Type Field
	Group     uint8
	Variation uint8
	// Qualifier Field
	Reserved      bool // Should always be set to 0, not enforced
	ObjPrefixCode ObjPrefixCode
	RangeSpecCode RangeSpecCode

	RangeField RangeField

	size int
}

func (oh *ObjectHeader) FromBytes(d []byte) error {
	if len(d) < 3 {
		return fmt.Errorf("object headers are at 3 - 11 bytes, got %d", len(d))
	} else if d[2]&0b10000000 != 0 {
		return fmt.Errorf("first qualifier octet bit must be 0 (got %d)", d[2])
	}

	oh.Group = uint8(d[0])
	oh.Variation = uint8(d[1])
	oh.Reserved = (d[2] & 0b10000000) != 0
	oh.ObjPrefixCode = ObjPrefixCode((d[2] & 0b01110000) >> 4)
	oh.RangeSpecCode = RangeSpecCode(d[2] & 0b00001111)

	switch oh.RangeSpecCode {
	case StartStop1:
		oh.RangeField = &RangeField0{}
		oh.RangeField.FromBytes(d[3:5])
		oh.size = 5
	case StartStop2:
		oh.RangeField = &RangeField1{}
		oh.RangeField.FromBytes(d[3:7])
		oh.size = 7
	case StartStop4:
		oh.RangeField = &RangeField2{}
		oh.RangeField.FromBytes(d[3:11])
		oh.size = 11
	case VirtualStartStop1:
		oh.RangeField = &RangeField3{}
		oh.RangeField.FromBytes(d[3:5])
		oh.size = 5
	case VirtualStartStop2:
		oh.RangeField = &RangeField4{}
		oh.RangeField.FromBytes(d[3:7])
		oh.size = 7
	case VirtualStartStop4:
		oh.RangeField = &RangeField5{}
		oh.RangeField.FromBytes(d[3:11])
		oh.size = 11
	case NoRangeField:
		oh.RangeField = &RangeField6{}
		oh.size = 3
	case Count1:
		oh.RangeField = &RangeField7{}
		oh.RangeField.FromBytes([]byte{d[3]})
		oh.size = 4
	case Count2:
		oh.RangeField = &RangeField8{}
		oh.RangeField.FromBytes(d[3:5])
		oh.size = 5
	case Count4:
		oh.RangeField = &RangeField9{}
		oh.RangeField.FromBytes(d[3:7])
		oh.size = 7
	case Count1Variable:
		oh.RangeField = &RangeFieldB{}
		oh.RangeField.FromBytes([]byte{d[3]})
		oh.size = 4
	case ReservedA, ReservedC, ReservedD, ReservedE, ReservedF:
		oh.size = 3
		return fmt.Errorf("range specifier code %d not valid", oh.RangeSpecCode)
	}

	return nil
}

func (oh *ObjectHeader) ToBytes() []byte {
	var o []byte
	o = append(o, oh.Group)
	o = append(o, oh.Variation)

	var b byte = 0
	if oh.Reserved {
		b |= 0b10000000
	}
	b |= uint8((oh.ObjPrefixCode << 4) & 0b01110000)
	b |= uint8(oh.RangeSpecCode & 0b00001111)
	o = append(o, b)

	o = append(o, oh.RangeField.ToBytes()...)

	return o
}

func (oh *ObjectHeader) String() string {
	return fmt.Sprintf(`Object Header:
				Group:     %02d
				Variation: %02d
				Qualifier:
					Obj Prefix Code : (%d) %s
					Range Spec Code : (%d) %s
				%s`,
		oh.Group, oh.Variation, oh.ObjPrefixCode,
		oh.ObjPrefixCode.String(), oh.RangeSpecCode, oh.RangeSpecCode.String(),
		oh.RangeField.String())
}

func (oh *ObjectHeader) SizeOf() int {
	return oh.size
}

func (oh *ObjectHeader) calcPrefixSize() int {
	switch oh.ObjPrefixCode {
	case NoPrefix:
		return 0
	case OctetIndex1:
		return 1
	case OctetIndex2:
		return 2
	case OctetIndex4:
		return 4
	case OctetSize1:
		return 1
	case OctetSize2:
		return 2
	case OctetSize4:
		return 4
	}
	return 0
}

// ObjectPrefixCode is a 4 bit description of how objects are packed
type ObjPrefixCode uint8 // only 3 bits

const (
	NoPrefix = iota // 0
	OctetIndex1
	OctetIndex2
	OctetIndex4
	OctetSize1
	OctetSize2
	OctetSize4
	Reserved // 7
)

var ObjPrefixCodeNames = map[ObjPrefixCode]string{
	NoPrefix:    "NO_PREFIX",
	OctetIndex1: "1_OCTET_INDEX",
	OctetIndex2: "2_OCTET_INDEX",
	OctetIndex4: "4_OCTET_INDEX",
	OctetSize1:  "1_OCTET_SIZE",
	OctetSize2:  "2_OCTET_SIZE",
	OctetSize4:  "4_OCTET_SIZE",
	Reserved:    "RESERVED",
}

func (opc ObjPrefixCode) String() string {
	if name, ok := ObjPrefixCodeNames[opc]; ok {
		return name
	}
	return fmt.Sprintf("unknown object prefix code %d", opc)
}

// RangeSpec describes how big and how the rangefield is formatted
type RangeSpecCode uint8 //only 4 bits

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
	FromBytes([]byte) error
	String() string
	NumObjects() int
}

// RangeField0 1 byte start and stop values
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
	return fmt.Sprintf(`Range Field (0) 1-octet start and stop indexes:
					Start: %d
					Stop : %d`, rf.Start, rf.Stop)
}

func (rf *RangeField0) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField1 2 byte start and stop values
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
	return fmt.Sprintf(`Range Field (1) 2-octet start and stop indexes:
					Start: %d
					Stop : %d`, rf.Start, rf.Stop)
}

func (rf *RangeField1) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField2 4 byte start and stop values
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
	return fmt.Sprintf(`Range Field (2) 4-octet start and stop indexes:
					Start: %d
					Stop : %d`, rf.Start, rf.Stop)
}

func (rf *RangeField2) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField3 1 byte VIRTUAL start and stop values
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
	return fmt.Sprintf(`Range Field (3) 1-octet virtual start and stop indexes:
					Start: %d (virtual)
					Stop : %d (virtual)`, rf.Start, rf.Stop)
}

func (rf *RangeField3) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField4 2 byte VIRTUAL start and stop values
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
	return fmt.Sprintf(`Range Field (4) 2-octet virtual start and stop indexes:
					Start: %d (virtual)
					Stop : %d (virtual)`, rf.Start, rf.Stop)
}

func (rf *RangeField4) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField5 4 byte VIRTUAL start and stop values
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
	return fmt.Sprintf(`Range Field (5) 4-octet virtual start and stop indexes:
					Start: %d (virtual)
					Stop : %d (virtual)`, rf.Start, rf.Stop)
}

func (rf *RangeField5) NumObjects() int {
	return int(rf.Stop - rf.Start + 1)
}

// RangeField6 - No range field used, Implies all values
type RangeField6 struct{}

func (rf *RangeField6) ToBytes() []byte {
	return nil
}

func (rf *RangeField6) FromBytes(d []byte) error {
	if len(d) > 0 {
		return fmt.Errorf("range field 6 is an empty range field")
	}
	return nil
}

func (rf *RangeField6) String() string { return "" }

func (rf *RangeField6) NumObjects() int { return 0 }

// RangeField7 1 byte count of objects
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
	return fmt.Sprintf(`Range Field (7) 1-octet count of objects:
					Count: %d`, rf.Count)
}

func (rf *RangeField7) NumObjects() int {
	return int(rf.Count)
}

// RangeField8 2 byte count of objects
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
	return fmt.Sprintf(`Range Field (8) 2-octet count of objects:
					Count: %d`, rf.Count)
}

func (rf *RangeField8) NumObjects() int {
	return int(rf.Count)
}

// RangeField9 4 byte count of objects
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
	return fmt.Sprintf(`Range Field (9) 4-octet count of objects:
					Count: %d`, rf.Count)
}

func (rf *RangeField9) NumObjects() int {
	return int(rf.Count)
}

// RangeFieldB 1 byte count of objects with variable format
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
	return fmt.Sprintf(`Range Field (B) 1-octet count of objects with variable format:
					Count: %d`, rf.Count)
}

func (rf *RangeFieldB) NumObjects() int {
	return int(rf.Count)
}
