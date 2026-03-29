package dnp3

import (
	"errors"
	"fmt"
)

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
