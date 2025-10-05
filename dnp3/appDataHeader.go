package dnp3

import (
	"errors"
	"fmt"
)

// ObjectHeader is used to describe the structure of application data.
type ObjectHeader struct {
	// Object Type Field
	Group     uint8
	Variation uint8
	// Qualifier Field
	Reserved        bool // Should always be set to 0
	PointPrefixCode PointPrefixCode
	RangeSpecCode   RangeSpecCode
	RangeField      RangeField
	size            int
	description     string
}

func newRangeFieldFromSpec(code RangeSpecCode, d []byte) (RangeField, int, error) {
	switch code {
	case StartStop1:
		rf := &RangeField0{}

		err := rf.FromBytes(d[3:5])
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 5, nil
	case StartStop2:
		rf := &RangeField1{}

		err := rf.FromBytes(d[3:7])
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 7, nil
	case StartStop4:
		rf := &RangeField2{}

		err := rf.FromBytes(d[3:11])
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 11, nil
	case VirtualStartStop1:
		rf := &RangeField3{}

		err := rf.FromBytes(d[3:5])
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 5, nil
	case VirtualStartStop2:
		rf := &RangeField4{}

		err := rf.FromBytes(d[3:7])
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 7, nil
	case VirtualStartStop4:
		rf := &RangeField5{}

		err := rf.FromBytes(d[3:11])
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 11, nil
	case NoRangeField:
		return &RangeField6{}, 3, nil
	case Count1:
		rf := &RangeField7{}

		err := rf.FromBytes([]byte{d[3]})
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 4, nil
	case Count2:
		rf := &RangeField8{}

		err := rf.FromBytes(d[3:5])
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 5, nil
	case Count4:
		rf := &RangeField9{}

		err := rf.FromBytes(d[3:7])
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 7, nil
	case Count1Variable:
		rf := &RangeFieldB{}

		err := rf.FromBytes([]byte{d[3]})
		if err != nil {
			return nil, 0, fmt.Errorf("can't create range field: %w", err)
		}

		return rf, 4, nil
	case ReservedA, ReservedC, ReservedD, ReservedE, ReservedF:
		return nil, 3, fmt.Errorf("range specifier code %d not valid", code)
	default:
		return nil, 0, fmt.Errorf("unknown range specifier code %d", code)
	}
}

func (oh *ObjectHeader) FromBytes(d []byte) error {
	if len(d) < 3 {
		return fmt.Errorf("object headers are at 3 - 11 bytes, got %d", len(d))
	}

	oh.Group = d[0]
	oh.Variation = d[1]
	oh.description = getGroupVariationDescription(oh.Group, oh.Variation)
	oh.Reserved = (d[2] & 0b10000000) != 0
	oh.PointPrefixCode = PointPrefixCode((d[2] & 0b01110000) >> 4)
	oh.RangeSpecCode = RangeSpecCode(d[2] & 0b00001111)

	rf, size, err := newRangeFieldFromSpec(oh.RangeSpecCode, d)
	oh.RangeField = rf
	oh.size = size

	if err != nil {
		return err
	}

	if oh.Reserved {
		return errors.New("first qualifier octet bit must be 0")
	}

	return nil
}

func (oh *ObjectHeader) ToBytes() []byte {
	var o []byte

	o = append(o, oh.Group)
	o = append(o, oh.Variation)

	var b byte
	if oh.Reserved {
		b |= 0b10000000
	}

	b |= uint8((oh.PointPrefixCode << 4) & 0b01110000)
	b |= uint8(oh.RangeSpecCode & 0b00001111)
	o = append(o, b)

	o = append(o, oh.RangeField.ToBytes()...)

	return o
}

func (oh *ObjectHeader) String() string {
	o := fmt.Sprintf(`Object Header:
				Grp, Var :  (%02d, %02d) - %s
				Qualifier:
					Obj Prefix Code: (%d) %s
					Range Spec Code: (%d) %s`,
		oh.Group, oh.Variation, oh.description, oh.PointPrefixCode,
		oh.PointPrefixCode.String(), oh.RangeSpecCode, oh.RangeSpecCode.String())

	rf := oh.RangeField.String()
	if rf != "" {
		o += "\n\t\t\t\t" + rf
	}

	return o
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

func getGroupVariationDescription(g, v uint8) string {
	type gv struct{ g, v uint8 }

	groupVar := gv{g, v}
	switch groupVar {
	case gv{1, 0}:
		return "(Static) Binary Input - Any Variations"
	case gv{1, 1}:
		return "(Static) Binary Input - Packed Format"
	case gv{1, 2}:
		return "(Static) Binary Input - Status with Flags"
	case gv{2, 0}:
		return "(Event) Binary Input Event - Any Variations"
	case gv{2, 1}:
		return "(Event) Binary Input Event"
	case gv{2, 2}:
		return "(Event) Binary Input Event - with Absolute Time"
	case gv{2, 3}:
		return "(Event) Binary Input Event - with Relative Time"
	case gv{3, 0}:
		return "(Static) Double-bit Binary Input - Any Variations"
	case gv{3, 1}:
		return "(Static) Double-bit Binary Input - Packed Format"
	case gv{3, 2}:
		return "(Static) Double-bit Binary Input - Status with Flags"
	case gv{4, 0}:
		return "(Event) Double-bit Binary Input Event - Any Variations"
	case gv{4, 1}:
		return "(Event) Double-bit Binary Input Event"
	case gv{4, 2}:
		return "(Event) Double-bit Binary Input Event with Absolute Time"
	case gv{4, 3}:
		return "(Event) Double-bit Binary Input Event with Relative Time"
	case gv{10, 0}:
		return "(Static) Binary Output - Any Variations"
	case gv{10, 1}:
		return "(Static) Binary Output - Packed Format"
	case gv{10, 2}:
		return "(Static) Binary Output - Status with Flags"
	case gv{11, 0}:
		return "(Event) Binary Output Event - Any Variations"
	case gv{11, 1}:
		return "(Event) Binary Output Event - Status"
	case gv{11, 2}:
		return "(Event) Binary Output Event - Status with Time"
	case gv{12, 0}:
		return "(Command) Binary Output Command - Any Variations"
	case gv{12, 1}:
		return "(Command) Binary Output Command - Control Relay Output Block"
	case gv{12, 2}:
		return "(Command) Binary Output Command - Pattern Control Block"
	case gv{12, 3}:
		return "(Command) Binary Output Command - Pattern Mask"
	case gv{13, 0}:
		return "(Event) Binary Output Command Event - Any Variations"
	case gv{13, 1}:
		return "(Event) Binary Output Command Event - Command Status"
	case gv{13, 2}:
		return "(Event) Binary Output Command Event - Command Status with Time"
	case gv{20, 0}:
		return "(Static) Counter - Any Variations"
	case gv{20, 1}:
		return "(Static) Counter - 32-bit with Flag"
	case gv{20, 2}:
		return "(Static) Counter - 16-bit with Flag"
	case gv{20, 5}:
		return "(Static) Counter - 32-bit w/o Flag"
	case gv{20, 6}:
		return "(Static) Counter - 16-bit w/o Flag"
	case gv{21, 0}:
		return "(Static) Frozen Counter - Any Variations"
	case gv{21, 1}:
		return "(Static) Frozen Counter - 32-bit with Flag"
	case gv{21, 2}:
		return "(Static) Frozen Counter - 16-bit with Flag"
	case gv{21, 5}:
		return "(Static) Frozen Counter - 32-bit with Flag and Time"
	case gv{21, 6}:
		return "(Static) Frozen Counter - 16-bit with Flag and Time"
	case gv{21, 9}:
		return "(Static) Frozen Counter - 32-bit w/o Flag"
	case gv{21, 10}:
		return "(Static) Frozen Counter - 16-bit w/o Flag"
	case gv{22, 0}:
		return "(Event) Counter Event - Any Variations"
	case gv{22, 1}:
		return "(Event) Counter Event - 32-bit with Flag"
	case gv{22, 2}:
		return "(Event) Counter Event - 16-bit with Flag"
	case gv{22, 5}:
		return "(Event) Counter Event - 32-bit with Flag and Time"
	case gv{22, 6}:
		return "(Event) Counter Event - 16-bit with Flag and Time"
	case gv{23, 0}:
		return "(Event) Frozen Counter Event - Any Variations"
	case gv{23, 1}:
		return "(Event) Frozen Counter Event - 32-bit with Flag"
	case gv{23, 2}:
		return "(Event) Frozen Counter Event - 16-bit with Flag"
	case gv{23, 5}:
		return "(Event) Frozen Counter Event - 32-bit with Flag and Time"
	case gv{23, 6}:
		return "(Event) Frozen Counter Event - 16-bit with Flag and Time"
	case gv{30, 0}:
		return "(Static) Analog Input - Any Variations"
	case gv{30, 1}:
		return "(Static) Analog Input - 32-bit with Flag"
	case gv{30, 2}:
		return "(Static) Analog Input - 16-bit with Flag"
	case gv{30, 3}:
		return "(Static) Analog Input - 32-bit w/o Flag"
	case gv{30, 4}:
		return "(Static) Analog Input - 16-bit w/o Flag"
	case gv{30, 5}:
		return "(Static) Analog Input - Single-prec. FP with Flag"
	case gv{30, 6}:
		return "(Static) Analog Input - Double-prec. FP with Flag"
	case gv{31, 0}:
		return "(Static) Frozen Analog Input - Any Variations"
	case gv{31, 1}:
		return "(Static) Frozen Analog Input - 32-bit with Flag"
	case gv{31, 2}:
		return "(Static) Frozen Analog Input - 16-bit with Flag"
	case gv{31, 3}:
		return "(Static) Frozen Analog Input - 32-bit with Time-of-Freeze"
	case gv{31, 4}:
		return "(Static) Frozen Analog Input - 16-bit with Time-of-Freeze"
	case gv{31, 5}:
		return "(Static) Frozen Analog Input - 32-bit w/o Flag"
	case gv{31, 6}:
		return "(Static) Frozen Analog Input - 16-bit w/o Flag"
	case gv{31, 7}:
		return "(Static) Frozen Analog Input - Single-prec. FP with Flag"
	case gv{31, 8}:
		return "(Static) Frozen Analog Input - Double-prec. FP with Flag"
	case gv{32, 0}:
		return "(Event) Analog Input Event - Any Variations"
	case gv{32, 1}:
		return "(Event) Analog Input Event - 32-bit"
	case gv{32, 2}:
		return "(Event) Analog Input Event - 16-bit"
	case gv{32, 3}:
		return "(Event) Analog Input Event - 32-bit with Time"
	case gv{32, 4}:
		return "(Event) Analog Input Event - 16-bit with Time"
	case gv{32, 5}:
		return "(Event) Analog Input Event - Single-prec. FP"
	case gv{32, 6}:
		return "(Event) Analog Input Event - Double-prec. FP"
	case gv{32, 7}:
		return "(Event) Analog Input Event - Single-prec. FP with Time"
	case gv{32, 8}:
		return "(Event) Analog Input Event - Double-prec. FP with Time"
	case gv{33, 0}:
		return "(Event) Frozen Analog Input Event - Any Variations"
	case gv{33, 1}:
		return "(Event) Frozen Analog Input Event - 32-bit"
	case gv{33, 2}:
		return "(Event) Frozen Analog Input Event - 16-bit"
	case gv{33, 3}:
		return "(Event) Frozen Analog Input Event - 32-bit with Time"
	case gv{33, 4}:
		return "(Event) Frozen Analog Input Event - 16-bit with Time"
	case gv{33, 5}:
		return "(Event) Frozen Analog Input Event - Single-prec. FP"
	case gv{33, 6}:
		return "(Event) Frozen Analog Input Event - Double-prec. FP"
	case gv{33, 7}:
		return "(Event) Frozen Analog Input Event - Single-prec. FP with Time"
	case gv{33, 8}:
		return "(Event) Frozen Analog Input Event - Double-prec. FP with Time"
	case gv{34, 0}:
		return "(Static) Analog Input Deadband - Any Variations"
	case gv{34, 1}:
		return "(Static) Analog Input Deadband - 16-bit"
	case gv{34, 2}:
		return "(Static) Analog Input Deadband - 32-bit"
	case gv{34, 3}:
		return "(Static) Analog Input Deadband - Single-prec. FP"
	case gv{40, 0}:
		return "(Static) Analog Output Status - Any Variations"
	case gv{40, 1}:
		return "(Static) Analog Output Status - 32-bit with Flag"
	case gv{40, 2}:
		return "(Static) Analog Output Status - 16-bit with Flag"
	case gv{40, 3}:
		return "(Static) Analog Output Status - Single-prec. FP with Flag"
	case gv{40, 4}:
		return "(Static) Analog Output Status - Double-prec. FP with Flag"
	case gv{41, 0}:
		return "(Command) Analog Output Command - Any Variations"
	case gv{41, 1}:
		return "(Command) Analog Output Command - 32-bit"
	case gv{41, 2}:
		return "(Command) Analog Output Command - 16-bit"
	case gv{41, 3}:
		return "(Command) Analog Output Command - Single-prec. FP"
	case gv{41, 4}:
		return "(Command) Analog Output Command - Double-prec. FP"
	case gv{42, 0}:
		return "(Event) Analog Output Event - Any Variations"
	case gv{42, 1}:
		return "(Event) Analog Output Event - 32-bit"
	case gv{42, 2}:
		return "(Event) Analog Output Event - 16-bit"
	case gv{42, 3}:
		return "(Event) Analog Output Event - 32-bit with Time"
	case gv{42, 4}:
		return "(Event) Analog Output Event - 16-bit with Time"
	case gv{42, 5}:
		return "(Event) Analog Output Event - Single-prec. FP"
	case gv{42, 6}:
		return "(Event) Analog Output Event - Double-prec. FP"
	case gv{42, 7}:
		return "(Event) Analog Output Event - Single-prec. FP with Time"
	case gv{42, 8}:
		return "(Event) Analog Output Event - Double-prec. FP with Time"
	case gv{43, 0}:
		return "(Event) Analog Output Command Event - Any Variations"
	case gv{43, 1}:
		return "(Event) Analog Output Command Event - 32-bit"
	case gv{43, 2}:
		return "(Event) Analog Output Command Event - 16-bit"
	case gv{43, 3}:
		return "(Event) Analog Output Command Event - 32-bit with Time"
	case gv{43, 4}:
		return "(Event) Analog Output Command Event - 16-bit with Time"
	case gv{43, 5}:
		return "(Event) Analog Output Command Event - Single-prec. FP"
	case gv{43, 6}:
		return "(Event) Analog Output Command Event - Double-prec. FP"
	case gv{43, 7}:
		return "(Event) Analog Output Command Event - Single-prec. FP with Time"
	case gv{43, 8}:
		return "(Event) Analog Output Command Event - Double-prec. FP with Time"
	case gv{50, 1}:
		return "(Info) Time and Date - Absolute Time"
	case gv{50, 2}:
		return "(Info) Time and Date - Absolute Time and Interval"
	case gv{50, 3}:
		return "(Info) Time and Date - Absolute Time at Last Recorded Time"
	case gv{50, 4}:
		return "(Info) Time and Date - Indexed Absolute Time and Long Interval"
	case gv{51, 1}:
		return "(Info) Time and Date CTO - Absolute Time, Synchronized"
	case gv{51, 2}:
		return "(Info) Time and Date CTO - Absolute Time, Unsynchronized"
	case gv{52, 1}:
		return "(Info) Time Delay Coarse"
	case gv{52, 2}:
		return "(Info) Time Delay Fine"
	case gv{60, 1}:
		return "(Info) Class Objects - Class 0 Data"
	case gv{60, 2}:
		return "(Info) Class Objects - Class 1 Data"
	case gv{60, 3}:
		return "(Info) Class Objects - Class 2 Data"
	case gv{60, 4}:
		return "(Info) Class Objects - Class 3 Data"
	case gv{80, 1}:
		return "(Static) Internal Indications - Packed Format"
	default:
		return "Unknown Group/Variation"
	}
}
