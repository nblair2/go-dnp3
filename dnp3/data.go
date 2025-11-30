package dnp3

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// ApplicationData holds an array of Data objects.
type ApplicationData struct {
	Objects []DataObject `json:"objects"`
	// in case we get in to trouble unrolling the objects just store the rest
	// of the data in here. Or can use this to set all data like "raw"
	extra []byte
}

func (ad *ApplicationData) FromBytes(data []byte) error {
	ad.Objects = nil // in case there was already stuff here

	for readOffset := 0; readOffset < len(data); {
		var object DataObject

		err := object.FromBytes(data[readOffset:])
		if err != nil {
			ad.extra = data[readOffset:]

			return fmt.Errorf("could not decode object: 0x % X, err: %w",
				data[readOffset:], err)
		}

		ad.Objects = append(ad.Objects, object)
		readOffset += object.SizeOf()
	}

	return nil
}

func (ad *ApplicationData) ToBytes() ([]byte, error) {
	var encoded []byte

	for _, object := range ad.Objects {
		bytesOut, err := object.ToBytes()
		if err != nil {
			return encoded, fmt.Errorf("could not encode object: %w", err)
		}

		encoded = append(encoded, bytesOut...)
	}

	encoded = append(encoded, ad.extra...)

	return encoded, nil
}

func (ad *ApplicationData) String() string {
	output := ""
	header := "Data Objects:"
	headerAdded := false

	if len(ad.Objects) > 0 {
		output += header
		headerAdded = true

		for _, obj := range ad.Objects {
			output += "\n" + indent("- "+obj.String(), "\t")
		}
	}

	if len(ad.extra) > 0 {
		if !headerAdded {
			output += header
		}

		output += fmt.Sprintf("\n\t- Extra: 0x % X", ad.extra)
	}

	return output
}

func (ad *ApplicationData) HasExtra() bool {
	return len(ad.extra) > 0
}

type DataObject struct {
	Header    ObjectHeader `json:"header"`
	Points    []Point      `json:"points"`
	Extra     []byte       `json:"extra,omitempty"`
	totalSize int
}

func (do *DataObject) FromBytes(data []byte) error {
	err := do.Header.FromBytes(data)
	if err != nil {
		return fmt.Errorf("can't create Data Object Header: %w", err)
	}

	headSize := do.Header.SizeOf()
	do.totalSize = headSize

	if do.Header.objectType == nil || do.Header.objectType.Constructor == nil {
		do.Extra = data[headSize:]
		do.totalSize += len(do.Extra)

		return fmt.Errorf("unsupported group/variation: %d/%d",
			do.Header.Group, do.Header.Variation)
	}

	numPoints := do.Header.RangeField.NumObjects()
	if numPoints == 0 {
		return nil
	}

	var sizeAllPoints int

	do.Points, sizeAllPoints, err = do.Header.objectType.Constructor(
		data[headSize:],
		numPoints,
		do.Header.PointPrefixCode.GetPointPrefixSize(),
	)
	do.totalSize += sizeAllPoints

	return err
}

func (do *DataObject) ToBytes() ([]byte, error) {
	// TODO get this to be more elegant.
	var encoded []byte

	headerBytes, err := do.Header.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to encode object header: %w", err)
	}

	encoded = append(encoded, headerBytes...)

	if len(do.Points) > 0 {
		var packer PointsPacker
		if do.Header.objectType != nil {
			packer = do.Header.objectType.Packer
		} else if def, ok := objectTypes[groupVariation{do.Header.Group, do.Header.Variation}]; ok {
			// Try to look it up if it wasn't set (e.g. manual construction)
			do.Header.objectType = def
			packer = def.Packer
		}

		if packer == nil {
			encoded = append(encoded, do.Extra...)

			return encoded, fmt.Errorf("no packer for Group %d, Var %d",
				do.Header.Group, do.Header.Variation)
		}

		packedPoints, err := packer(do.Points)
		if err != nil {
			encoded = append(encoded, do.Extra...)

			return encoded, fmt.Errorf("could not pack points: %w", err)
		}

		encoded = append(encoded, packedPoints...)
	}

	encoded = append(encoded, do.Extra...)

	return encoded, nil
}

func (do *DataObject) String() string {
	output := do.Header.String()

	if len(do.Points) == 0 {
		return output
	}

	output += "\n  Objects:"

	for _, point := range do.Points {
		lines := strings.Split(point.String(), "\n")
		if len(lines) > 0 {
			lines[0] = "- " + lines[0]
			for i := 1; i < len(lines); i++ {
				lines[i] = "  " + lines[i]
			}

			output += "\n" + indent(strings.Join(lines, "\n"), "\t")
		}
	}

	return output
}

func (do *DataObject) SizeOf() int {
	return do.totalSize
}

// ObjectHeader is used to describe the structure of application data.
type ObjectHeader struct {
	// Object Type Field
	Group      uint8 `json:"group"`
	Variation  uint8 `json:"variation"`
	objectType *objectType
	// Qualifier Field
	Reserved        bool            `json:"reserved"` // Should always be set to 0
	PointPrefixCode PointPrefixCode `json:"point_prefix_code"`
	RangeSpecCode   RangeSpecCode   `json:"range_spec_code"`
	RangeField      RangeField      `json:"range_field"`
	size            int
}

type rangeFieldConstructor func() RangeField

var rangeFieldConstructors = map[RangeSpecCode]rangeFieldConstructor{
	StartStop1:        func() RangeField { return &RangeField0{} },
	VirtualStartStop1: func() RangeField { return &RangeField0{} },
	StartStop2:        func() RangeField { return &RangeField1{} },
	VirtualStartStop2: func() RangeField { return &RangeField1{} },
	StartStop4:        func() RangeField { return &RangeField2{} },
	VirtualStartStop4: func() RangeField { return &RangeField2{} },
	NoRangeField:      func() RangeField { return &RangeField6{} },
	Count1:            func() RangeField { return &RangeField7{} },
	Count1Variable:    func() RangeField { return &RangeField7{} },
	Count2:            func() RangeField { return &RangeField8{} },
	Count4:            func() RangeField { return &RangeField9{} },
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

func (oh *ObjectHeader) FromBytes(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("object headers are at 3 - 11 bytes, got %d", len(data))
	}

	oh.Group = data[0]

	oh.Variation = data[1]
	if def, ok := objectTypes[groupVariation{oh.Group, oh.Variation}]; ok {
		oh.objectType = def
	}

	oh.Reserved = (data[2] & 0b10000000) != 0
	oh.PointPrefixCode = PointPrefixCode((data[2] & 0b01110000) >> 4)
	oh.RangeSpecCode = RangeSpecCode(data[2] & 0b00001111)

	ctor, err := rangeFieldConstructorFor(oh.RangeSpecCode)
	if err != nil {
		return err
	}

	rangeField := ctor()
	rangeFieldBytes := rangeField.Size()
	consumed := 3 + rangeFieldBytes

	if len(data) < consumed {
		return fmt.Errorf(
			"can't create range field: need %d bytes, got %d",
			consumed,
			len(data),
		)
	}

	err = rangeField.FromBytes(data[3:consumed])
	if err != nil {
		return fmt.Errorf("can't create range field: %w", err)
	}

	oh.RangeField = rangeField
	oh.size = consumed

	if oh.Reserved {
		return errors.New("first qualifier octet bit must be 0")
	}

	return nil
}

func (oh *ObjectHeader) ToBytes() ([]byte, error) {
	var encoded []byte

	encoded = append(encoded, oh.Group)
	encoded = append(encoded, oh.Variation)

	var qualifierByte byte
	if oh.Reserved {
		qualifierByte |= 0b10000000
	}

	qualifierByte |= uint8((oh.PointPrefixCode << 4) & 0b01110000)
	qualifierByte |= uint8(oh.RangeSpecCode & 0b00001111)
	encoded = append(encoded, qualifierByte)

	if oh.RangeField == nil {
		return nil, errors.New("range field is nil")
	}

	rangeBytes, err := oh.RangeField.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to encode range field: %w", err)
	}

	encoded = append(encoded, rangeBytes...)

	return encoded, nil
}

func (oh *ObjectHeader) String() string {
	desc := "Unknown Group/Variation"
	if oh.objectType != nil {
		desc = oh.objectType.Description
	} else if def, ok := objectTypes[groupVariation{oh.Group, oh.Variation}]; ok {
		// Try to look it up if it wasn't set (e.g. manual construction)
		oh.objectType = def
		desc = def.Description
	}

	headerString := "Object Header:\n" + indent(fmt.Sprintf(`Grp, Var :  (%02d, %02d) - %s
Qualifier:
%s`,
		oh.Group, oh.Variation, desc,
		indent(fmt.Sprintf("Obj Prefix Code: (%d) %s\nRange Spec Code: (%d) %s",
			oh.PointPrefixCode, oh.PointPrefixCode.String(),
			oh.RangeSpecCode, oh.RangeSpecCode.String()), "\t"),
	), "\t")

	rf := oh.RangeField.String()
	if rf != "" {
		headerString += "\n" + indent(rf, "\t\t")
	}

	return headerString
}

func (oh *ObjectHeader) SizeOf() int {
	return oh.size
}

// PointPrefixCode is a 4 bit description of how objects are packed.
type PointPrefixCode uint8 // only 3 bits

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

var PointPrefixCodeNames = map[PointPrefixCode]string{
	NoPrefix:    "NO_PREFIX",
	OctetIndex1: "1_OCTET_INDEX",
	OctetIndex2: "2_OCTET_INDEX",
	OctetIndex4: "4_OCTET_INDEX",
	OctetSize1:  "1_OCTET_SIZE",
	OctetSize2:  "2_OCTET_SIZE",
	OctetSize4:  "4_OCTET_SIZE",
	Reserved:    "RESERVED",
}

var PointPrefixCodeSize = map[PointPrefixCode]int{
	NoPrefix:    0,
	OctetIndex1: 1,
	OctetIndex2: 2,
	OctetIndex4: 4,
	OctetSize1:  1,
	OctetSize2:  2,
	OctetSize4:  3,
	Reserved:    0,
}

func (ppc PointPrefixCode) String() string {
	if name, ok := PointPrefixCodeNames[ppc]; ok {
		return name
	}

	return fmt.Sprintf("unknown object prefix code %d", ppc)
}

func (ppc PointPrefixCode) GetPointPrefixSize() int {
	if size, ok := PointPrefixCodeSize[ppc]; ok {
		return size
	}

	return 0
}

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
	ToBytes() ([]byte, error)
	FromBytes(data []byte) error
	String() string
	NumObjects() int
	Size() int
}

// RangeField0 1 byte start and stop values.
type RangeField0 struct {
	Start uint8 `json:"start"`
	Stop  uint8 `json:"stop"`
}

func (rf *RangeField0) ToBytes() ([]byte, error) {
	return []byte{rf.Start, rf.Stop}, nil
}

func (rf *RangeField0) FromBytes(data []byte) error {
	if len(data) != 2 {
		return fmt.Errorf("requires 2 bytes, got %d", len(data))
	}

	rf.Start = data[0]
	rf.Stop = data[1]

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

func (rf *RangeField0) Size() int { return 2 }

// RangeField1 2 byte start and stop values.
type RangeField1 struct {
	Start uint16 `json:"start"`
	Stop  uint16 `json:"stop"`
}

func (rf *RangeField1) ToBytes() ([]byte, error) {
	var o []byte

	o = binary.LittleEndian.AppendUint16(o, rf.Start)
	o = binary.LittleEndian.AppendUint16(o, rf.Stop)

	return o, nil
}

func (rf *RangeField1) FromBytes(data []byte) error {
	if len(data) != 4 {
		return fmt.Errorf("requires 4 bytes, got %d", len(data))
	}

	rf.Start = binary.LittleEndian.Uint16(data[0:2])
	rf.Stop = binary.LittleEndian.Uint16(data[2:4])

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

func (rf *RangeField1) Size() int { return 4 }

// RangeField2 4 byte start and stop values.
type RangeField2 struct {
	Start uint32 `json:"start"`
	Stop  uint32 `json:"stop"`
}

func (rf *RangeField2) ToBytes() ([]byte, error) {
	var o []byte

	o = binary.LittleEndian.AppendUint32(o, rf.Start)
	o = binary.LittleEndian.AppendUint32(o, rf.Stop)

	return o, nil
}

func (rf *RangeField2) FromBytes(data []byte) error {
	if len(data) != 8 {
		return fmt.Errorf("requires 8 bytes, got %d", len(data))
	}

	rf.Start = binary.LittleEndian.Uint32(data[0:4])
	rf.Stop = binary.LittleEndian.Uint32(data[4:8])

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

func (rf *RangeField2) Size() int { return 8 }

// RangeField3 1 byte VIRTUAL start and stop values.
type RangeField3 struct {
	Start uint8 `json:"start"`
	Stop  uint8 `json:"stop"`
}

func (rf *RangeField3) ToBytes() ([]byte, error) {
	return []byte{rf.Start, rf.Stop}, nil
}

func (rf *RangeField3) FromBytes(data []byte) error {
	if len(data) != 2 {
		return fmt.Errorf("requires 2 bytes, got %d", len(data))
	}

	rf.Start = data[0]
	rf.Stop = data[1]

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

func (rf *RangeField3) Size() int { return 2 }

// RangeField4 2 byte VIRTUAL start and stop values.
type RangeField4 struct {
	Start uint16 `json:"start"`
	Stop  uint16 `json:"stop"`
}

func (rf *RangeField4) ToBytes() ([]byte, error) {
	var o []byte

	o = binary.LittleEndian.AppendUint16(o, rf.Start)
	o = binary.LittleEndian.AppendUint16(o, rf.Stop)

	return o, nil
}

func (rf *RangeField4) FromBytes(data []byte) error {
	if len(data) != 4 {
		return fmt.Errorf("requires 4 bytes, got %d", len(data))
	}

	rf.Start = binary.LittleEndian.Uint16(data[0:2])
	rf.Stop = binary.LittleEndian.Uint16(data[2:4])

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

func (rf *RangeField4) Size() int { return 4 }

// RangeField5 4 byte VIRTUAL start and stop values.
type RangeField5 struct {
	Start uint32 `json:"start"`
	Stop  uint32 `json:"stop"`
}

func (rf *RangeField5) ToBytes() ([]byte, error) {
	var o []byte

	o = binary.LittleEndian.AppendUint32(o, rf.Start)
	o = binary.LittleEndian.AppendUint32(o, rf.Stop)

	return o, nil
}

func (rf *RangeField5) FromBytes(data []byte) error {
	if len(data) != 8 {
		return fmt.Errorf("requires 8 bytes, got %d", len(data))
	}

	rf.Start = binary.LittleEndian.Uint32(data[0:4])
	rf.Stop = binary.LittleEndian.Uint32(data[4:8])

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

func (rf *RangeField5) Size() int { return 8 }

// RangeField6 - No Range Field: used, Implies all values.
type RangeField6 struct{}

func (rf *RangeField6) ToBytes() ([]byte, error) {
	return nil, nil
}

func (rf *RangeField6) FromBytes(data []byte) error {
	if len(data) > 0 {
		return errors.New("error Range Field: 6 is an empty Range Field")
	}

	return nil
}

func (rf *RangeField6) String() string { return "" }

func (rf *RangeField6) NumObjects() int { return 0 }

func (rf *RangeField6) Size() int { return 0 }

// RangeField7 1 byte count of objects.
type RangeField7 struct {
	Count uint8 `json:"count"`
}

func (rf *RangeField7) ToBytes() ([]byte, error) {
	return []byte{rf.Count}, nil
}

func (rf *RangeField7) FromBytes(data []byte) error {
	if len(data) != 1 {
		return fmt.Errorf("requires 1 byte, got %d", len(data))
	}

	rf.Count = data[0]

	return nil
}

func (rf *RangeField7) String() string {
	return fmt.Sprintf(`Range Field: (7) 1-octet count of objects
	Count: %d`, rf.Count)
}

func (rf *RangeField7) NumObjects() int {
	return int(rf.Count)
}

func (rf *RangeField7) Size() int { return 1 }

// RangeField8 2 byte count of objects.
type RangeField8 struct {
	Count uint16 `json:"count"`
}

func (rf *RangeField8) ToBytes() ([]byte, error) {
	var o []byte

	o = binary.LittleEndian.AppendUint16(o, rf.Count)

	return o, nil
}

func (rf *RangeField8) FromBytes(data []byte) error {
	if len(data) != 2 {
		return fmt.Errorf("requires 2 byte, got %d", len(data))
	}

	rf.Count = binary.LittleEndian.Uint16(data)

	return nil
}

func (rf *RangeField8) String() string {
	return fmt.Sprintf(`Range Field: (8) 2-octet count of objects
	Count: %d`, rf.Count)
}

func (rf *RangeField8) NumObjects() int {
	return int(rf.Count)
}

func (rf *RangeField8) Size() int { return 2 }

// RangeField9 4 byte count of objects.
type RangeField9 struct {
	Count uint32 `json:"count"`
}

func (rf *RangeField9) ToBytes() ([]byte, error) {
	var o []byte

	o = binary.LittleEndian.AppendUint32(o, rf.Count)

	return o, nil
}

func (rf *RangeField9) FromBytes(data []byte) error {
	if len(data) != 4 {
		return fmt.Errorf("requires 4 byte, got %d", len(data))
	}

	rf.Count = binary.LittleEndian.Uint32(data[0:4])

	return nil
}

func (rf *RangeField9) String() string {
	return fmt.Sprintf(`Range Field: (9) 4-octet count of objects
	Count: %d`, rf.Count)
}

func (rf *RangeField9) NumObjects() int {
	return int(rf.Count)
}

func (rf *RangeField9) Size() int { return 4 }

// RangeFieldB 1 byte count of objects with variable format.
type RangeFieldB struct {
	Count uint8 `json:"count"`
	// Variable ?
}

func (rf *RangeFieldB) ToBytes() ([]byte, error) {
	return []byte{rf.Count}, nil
}

func (rf *RangeFieldB) FromBytes(data []byte) error {
	if len(data) != 1 {
		return fmt.Errorf("requires 1 byte, got %d", len(data))
	}

	rf.Count = data[0]

	return nil
}

func (rf *RangeFieldB) String() string {
	return fmt.Sprintf(`Range Field:: (B) 1-octet count of objects with variable format
	Count: %d`, rf.Count)
}

func (rf *RangeFieldB) NumObjects() int {
	return int(rf.Count)
}

func (rf *RangeFieldB) Size() int { return 1 }
