package dnp3

import (
	"fmt"
)

// ApplicationData holds an array of Data objects.
type ApplicationData struct {
	OBJS []DataObject
	// in case we get in to trouble unrolling the objects just store the rest
	// of the data in here. Or can use this to set all data like "raw"
	extra []byte
}

func (ad *ApplicationData) FromBytes(d []byte) error {
	ad.OBJS = nil // in case there was already stuff here

	for i := 0; i < len(d); {
		var obj DataObject

		err := obj.FromBytes(d[i:])
		if err != nil {
			ad.extra = d[i:]

			return fmt.Errorf("could not decode object: 0x % X, err: %w",
				d[i:], err)
		}

		ad.OBJS = append(ad.OBJS, obj)
		i += obj.SizeOf()
	}

	return nil
}

func (ad *ApplicationData) ToBytes() ([]byte, error) {
	var o []byte

	for _, do := range ad.OBJS {
		b, err := do.ToBytes()
		if err != nil {
			return o, fmt.Errorf("could not encode object: %w", err)
		}

		o = append(o, b...)
	}

	o = append(o, ad.extra...)

	return o, nil
}

func (ad *ApplicationData) String() string {
	o := ""
	header := "Data Objects:"
	headerAdded := false

	if len(ad.OBJS) > 0 {
		o += header
		headerAdded = true

		for _, obj := range ad.OBJS {
			o += "\n\t\t      - " + obj.String()
		}
	}

	if len(ad.extra) > 0 {
		if !headerAdded {
			o += header
		}

		o += fmt.Sprintf("\n\t\t      - Extra: 0x % X", ad.extra)
	}

	return o
}

type DataObject struct {
	Header      ObjectHeader
	Points      []Point
	Extra       []byte
	totalSize   int
	constructor PointsConstructor
	packer      PointsPacker
}

func (do *DataObject) FromBytes(d []byte) error {
	err := do.Header.FromBytes(d)
	if err != nil {
		return fmt.Errorf("can't create Data Object Header: %w", err)
	}

	headSize := do.Header.SizeOf()
	do.totalSize = headSize

	do.constructor = do.getPointsConstructor()

	do.packer = do.getPointsPacker()
	if do.constructor == nil {
		do.Extra = d[headSize:]
		do.totalSize += len(do.Extra)

		return fmt.Errorf("unsupported group/variation: %d/%d",
			do.Header.Group, do.Header.Variation)
	}

	numPoints := do.Header.RangeField.NumObjects()
	if numPoints == 0 {
		return nil
	}

	var sizeAllPoints int

	do.Points, sizeAllPoints, err = do.constructor(
		d[headSize:],
		numPoints,
		do.Header.PointPrefixCode.GetPointPrefixSize(),
	)
	do.totalSize += sizeAllPoints

	return err
}

func (do *DataObject) ToBytes() ([]byte, error) {
	// TODO get this to be more elegant.
	var o []byte

	o = append(o, do.Header.ToBytes()...)

	if len(do.Points) > 0 {
		if do.packer == nil {
			// try to get a packer in case it was not set
			do.packer = do.getPointsPacker()
		}

		if do.packer == nil {
			o = append(o, do.Extra...)

			return o, fmt.Errorf("no packer for Group %d, Var %d",
				do.Header.Group, do.Header.Variation)
		}

		b, err := do.packer(do.Points)
		if err != nil {
			o = append(o, do.Extra...)

			return o, fmt.Errorf("could not pack points: %w", err)
		}

		o = append(o, b...)
	}

	o = append(o, do.Extra...)

	return o, nil
}

func (do *DataObject) String() string {
	o := do.Header.String()

	if len(do.Points) == 0 {
		return o
	}

	o += "\n\t\t\tObjects:"
	for _, p := range do.Points {
		o += p.String()
	}

	return o
}

func (do *DataObject) SizeOf() int {
	return do.totalSize
}

func (do *DataObject) getPointsConstructor() PointsConstructor {
	type gv struct{ g, v int }

	groupVar := gv{int(do.Header.Group), int(do.Header.Variation)}
	switch groupVar {
	// Binary Input
	case gv{1, 1}: // Binary Input - Packed Format
		return newPoints1Bit
	case gv{1, 2}: // Binary Input - Status with Flags
		return newPoints1BitFlags

	// Binary Input Event
	case gv{2, 0}: // Binary Input Change
		return constructorNoPoints
	case gv{2, 1}: // Binary Input Event - 1 octet
		return newPoints1Byte
	case gv{2, 2}: // Binary Input Event - with Absolute Time (7 octets)
		return newPoints1ByteAbsTime
	case gv{2, 3}: // Binary Input Event - with Relative Time (3 octets)
		return newPoints1ByteRelTime

	// Double-bit Binary Input
	case gv{3, 1}: // Double-bit Binary Input - Packed Format
		return newPoints2Bits
	case gv{3, 2}: // Double-bit Binary Input - Status with Flags
		return newPoints1Byte

	// Double-bit Binary Input Event
	case gv{4, 1}: // Double-bit Binary Input Event - 1 octet
		return newPoints1Byte
	case gv{4, 2}: // Double-bit Binary Input Event with Absolute Time (7 octets)
		return newPoints1ByteAbsTime
	case gv{4, 3}: // Double-bit Binary Input Event with Relative Time (3 octets)
		return newPoints1ByteRelTime

	// Binary Output
	case gv{10, 1}: // Binary Output - Packed Format
		return newPoints1Bit
	case gv{10, 2}: // Binary Output - Status with Flags
		return newPoints1BitFlags

	// Binary Output Event
	case gv{11, 1}: // Binary Output Event - Status (1 octet)
		return newPoints1Byte
	case gv{11, 2}: // Binary Output Event - Status with Time (7 octets)
		return newPoints1ByteAbsTime

	// Binary Output Command
	// TODO CROB
	case gv{12, 1}: // Binary Output Command - Control Relay Output Block (11 octets)
		return newPoints11Bytes
	case gv{12, 2}: // Binary Output Command - Pattern Control Block (11 octets)
		return newPoints11Bytes

	// Binary Output Command Event
	// TODO command
	case gv{13, 1}: // Binary Output Command Event - Command Status (1 octet)
		return newPoints1Byte
	case gv{13, 2}: // Binary Output Command Event - Command Status with Time (7 octets)
		return newPoints1ByteAbsTime

	// Counter
	case gv{20, 1}: // Counter - 32-bit with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{20, 2}: // Counter - 16-bit with Flag (3 octets)
		return newPoints2BytesFlags
	case gv{20, 5}: // Counter - 32-bit w/o Flag (4 octets)
		return newPoints4Bytes
	case gv{20, 6}: // Counter - 16-bit w/o Flag (2 octets)
		return newPoints2Bytes

	// Frozen Counter
	case gv{21, 1}: // Frozen Counter - 32-bit with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{21, 2}: // Frozen Counter - 16-bit with Flag (3 octets)
		return newPoints2BytesFlags
	case gv{21, 5}: // Frozen Counter - 32-bit with Flag and Time (11 octets)
		return newPoints4BytesFlagsAbsTime
	case gv{21, 6}: // Frozen Counter - 16-bit with Flag and Time (9 octets)
		return newPoints2BytesFlagsAbsTime
	case gv{21, 9}: // Frozen Counter - 32-bit w/o Flag (4 octets)
		return newPoints4Bytes
	case gv{21, 10}: // Frozen Counter - 16-bit w/o Flag (2 octets)
		return newPoints2Bytes

	// Counter Event
	case gv{22, 1}: // Counter Event - 32-bit with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{22, 2}: // Counter Event - 16-bit with Flag (3 octets)
		return newPoints2BytesFlags
	case gv{22, 5}: // Counter Event - 32-bit with Flag and Time (11 octets)
		return newPoints4BytesFlagsAbsTime
	case gv{22, 6}: // Counter Event - 16-bit with Flag and Time (9 octets)
		return newPoints2BytesFlagsAbsTime

	// Frozen Counter Event
	case gv{23, 1}: // Frozen Counter Event - 32-bit with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{23, 2}: // Frozen Counter Event - 16-bit with Flag (3 octets)
		return newPoints1ByteFlags
	case gv{23, 5}: // Frozen Counter Event - 32-bit with Flag and Time (11 octets)
		return newPoints4BytesFlagsAbsTime
	case gv{23, 6}: // Frozen Counter Event - 16-bit with Flag and Time (9 octets)
		return newPoints2BytesFlagsAbsTime

	// Analog Input
	case gv{30, 1}: // Analog Input - 32-bit with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{30, 2}: // Analog Input - 16-bit with Flag (3 octets)
		return newPoints2BytesFlags
	case gv{30, 3}: // Analog Input - 32-bit w/o Flag (4 octets)
		return newPoints4Bytes
	case gv{30, 4}: // Analog Input - 16-bit w/o Flag (2 octets)
		return newPoints2Bytes
	case gv{30, 5}: // Analog Input - Single-prec. FP with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{30, 6}: // Analog Input - Double-prec. FP with Flag (9 octets)
		return newPoints8BytesFlags

	// Frozen Analog Input
	case gv{31, 1}: // Frozen Analog Input - 32-bit with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{31, 2}: // Frozen Analog Input - 16-bit with Flag (3 octets)
		return newPoints2BytesFlags
	case gv{31, 3}: // Frozen Analog Input - 32-bit with Time-of-Freeze (11 octets)
		return newPoints11Bytes
	case gv{31, 4}: // Frozen Analog Input - 16-bit with Time-of-Freeze (9 octets)
		return newPoints9Bytes
	case gv{31, 5}: // Frozen Analog Input - 32-bit w/o Flag (4 octets)
		return newPoints4Bytes
	case gv{31, 6}: // Frozen Analog Input - 16-bit w/o Flag (2 octets)
		return newPoints2Bytes
	case gv{31, 7}: // Frozen Analog Input - Single-prec. FP with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{31, 8}: // Frozen Analog Input - Double-prec. FP with Flag (9 octets)
		return newPoints8BytesFlags

	// Analog Input Event
	// TODO event fields
	case gv{32, 1}: // Analog Input Event - 32-bit (5 octets)
		return newPoints5Bytes
	case gv{32, 2}: // Analog Input Event - 16-bit (3 octets)
		return newPoints3Bytes
	case gv{32, 3}: // Analog Input Event - 32-bit with Time (11 octets)
		return newPoints11Bytes
	case gv{32, 4}: // Analog Input Event - 16-bit with Time (9 octets)
		return newPoints9Bytes
	case gv{32, 5}: // Analog Input Event - Single-prec. FP (5 octets)
		return newPoints5Bytes
	case gv{32, 6}: // Analog Input Event - Double-prec. FP (9 octets)
		return newPoints9Bytes
	case gv{32, 7}: // Analog Input Event - Single-prec. FP with Time (11 octets)
		return newPoints11Bytes
	case gv{32, 8}: // Analog Input Event - Double-prec. FP with Time (15 octets)
		return newPoints15Bytes

	// Frozen Analog Input Event
	// TODO event
	case gv{33, 1}: // Frozen Analog Input Event - 32-bit (5 octets)
		return newPoints5Bytes
	case gv{33, 2}: // Frozen Analog Input Event - 16-bit (3 octets)
		return newPoints3Bytes
	case gv{33, 3}: // Frozen Analog Input Event - 32-bit with Time (11 octets)
		return newPoints11Bytes
	case gv{33, 4}: // Frozen Analog Input Event - 16-bit with Time (9 octets)
		return newPoints9Bytes
	case gv{33, 5}: // Frozen Analog Input Event - Single-prec. FP (5 octets)
		return newPoints5Bytes
	case gv{33, 6}: // Frozen Analog Input Event - Double-prec. FP (9 octets)
		return newPoints9Bytes
	case gv{33, 7}: // Frozen Analog Input Event - Single-prec. FP with Time (11 octets)
		return newPoints11Bytes
	case gv{33, 8}: // Frozen Analog Input Event - Double-prec. FP with Time (15 octets)
		return newPoints15Bytes

	// Analog Input Deadband
	case gv{34, 1}: // Analog Input Deadband - 16-bit (2 octets)
		return newPoints2Bytes
	case gv{34, 2}: // Analog Input Deadband - 32-bit (4 octets)
		return newPoints4Bytes
	case gv{34, 3}: // Analog Input Deadband - Single-prec. FP (4 octets)
		return newPoints4Bytes

	// Analog Output Status
	case gv{40, 1}: // Analog Output Status - 32-bit with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{40, 2}: // Analog Output Status - 16-bit with Flag (3 octets)
		return newPoints2BytesFlags
	case gv{40, 3}: // Analog Output Status - Single-prec. FP with Flag (5 octets)
		return newPoints4BytesFlags
	case gv{40, 4}: // Analog Output Status - Double-prec. FP with Flag (9 octets)
		return newPoints8BytesFlags

	// Analog Output Command
	// TODO command
	case gv{41, 1}: // Analog Output Command - 32-bit (5 octets)
		return newPoints5Bytes
	case gv{41, 2}: // Analog Output Command - 16-bit (3 octets)
		return newPoints3Bytes
	case gv{41, 3}: // Analog Output Command - Single-prec. FP (5 octets)
		return newPoints5Bytes
	case gv{41, 4}: // Analog Output Command - Double-prec. FP (9 octets)
		return newPoints9Bytes

	// Analog Output Event
	// TODO event
	case gv{42, 1}: // Analog Output Event - 32-bit (5 octets)
		return newPoints5Bytes
	case gv{42, 2}: // Analog Output Event - 16-bit (3 octets)
		return newPoints3Bytes
	case gv{42, 3}: // Analog Output Event - 32-bit with Time (11 octets)
		return newPoints11Bytes
	case gv{42, 4}: // Analog Output Event - 16-bit with Time (9 octets)
		return newPoints9Bytes
	case gv{42, 5}: // Analog Output Event - Single-prec. FP (5 octets)
		return newPoints5Bytes
	case gv{42, 6}: // Analog Output Event - Double-prec. FP (9 octets)
		return newPoints9Bytes
	case gv{42, 7}: // Analog Output Event - Single-prec. FP with Time (11 octets)
		return newPoints11Bytes
	case gv{42, 8}: // Analog Output Event - Double-prec. FP with Time (15 octets)
		return newPoints15Bytes

	// Analog Output Command Event
	// TODO command / event
	case gv{43, 1}: // Analog Output Command Event - 32-bit (5 octets)
		return newPoints5Bytes
	case gv{43, 2}: // Analog Output Command Event - 16-bit (3 octets)
		return newPoints3Bytes
	case gv{43, 3}: // Analog Output Command Event - 32-bit with Time (11 octets)
		return newPoints11Bytes
	case gv{43, 4}: // Analog Output Command Event - 16-bit with Time (9 octets)
		return newPoints9Bytes
	case gv{43, 5}: // Analog Output Command Event - Single-prec. FP (5 octets)
		return newPoints5Bytes
	case gv{43, 6}: // Analog Output Command Event - Double-prec. FP (9 octets)
		return newPoints9Bytes
	case gv{43, 7}: // Analog Output Command Event - Single-prec. FP with Time (11 octets)
		return newPoints11Bytes
	case gv{43, 8}: // Analog Output Command Event - Double-prec. FP with Time (15 octets)
		return newPoints15Bytes

	// Time and Date
	// TODO interval
	case gv{50, 1}: // Time and Date - Absolute Time (6 octets)
		return newPointsAbsTime
	case gv{50, 2}: // Time and Date - Absolute Time and Interval (10 octets)
		return newPoints4BytesAbsTime
	case gv{50, 3}: // Time and Date - Absolute Time at Last Recorded Time (6 octets)
		return newPointsAbsTime
	case gv{50, 4}: // Time and Date - Indexed Absolute Time and Long Interval (11 octets)
		return newPoints11Bytes

	// Time and Date CTO
	case gv{51, 1}: // CTO - Absolute Time, Synchronized (6 octets)
		return newPointsAbsTime
	case gv{51, 2}: // CTO - Absolute Time, Unsynchronized (6 octets)
		return newPointsAbsTime

	// Time Delay
	// TODO fine vs coarse
	case gv{52, 1}: // Time Delay Coarse (2 octets)
		return newPointsRelTime
	case gv{52, 2}: // Time Delay Fine (2 octets)
		return newPointsRelTime

	// Read
	case gv{60, 1}: // Class 0 Data
		return constructorNoPoints
	case gv{60, 2}: // Class 1 Data
		return constructorNoPoints
	case gv{60, 3}: // Class 2 Data
		return constructorNoPoints
	case gv{60, 4}: // Class 3 Data
		return constructorNoPoints

	// Internal Indications
	case gv{80, 1}: // Internal Indications - Packed Format (1 octet)
		return newPoints1Byte
	}

	return nil
}

func (do *DataObject) getPointsPacker() PointsPacker {
	type gv struct{ g, v int }

	groupVar := gv{int(do.Header.Group), int(do.Header.Variation)}
	switch groupVar {
	// No points
	case gv{60, 1}, gv{60, 2}, gv{60, 3}, gv{60, 4}, gv{2, 0}:
		return packNoPoints
	// Bits
	case gv{1, 1}, gv{10, 1}:
		return packerPoints1Bit
	case gv{3, 1}:
		return packerPoints2Bits
	// Full bytes
	case
		gv{1, 2}, gv{2, 1}, gv{2, 2}, gv{3, 2}, gv{4, 1}, gv{10, 2}, gv{11, 1},
		gv{13, 1}, gv{80, 1},
		gv{20, 6}, gv{21, 10}, gv{30, 4}, gv{31, 6}, gv{34, 1}, gv{52, 1}, gv{52, 2},
		gv{2, 3}, gv{4, 3}, gv{20, 2}, gv{21, 2}, gv{22, 2}, gv{23, 2},
		gv{30, 2}, gv{31, 2}, gv{32, 2}, gv{33, 2}, gv{40, 2}, gv{41, 2},
		gv{42, 2}, gv{43, 2},
		gv{20, 5}, gv{21, 9}, gv{30, 3}, gv{31, 5}, gv{34, 2}, gv{34, 3},
		gv{20, 1}, gv{21, 1}, gv{22, 1}, gv{23, 1}, gv{30, 1}, gv{30, 5},
		gv{31, 1}, gv{31, 7}, gv{32, 1}, gv{32, 5}, gv{33, 1}, gv{33, 5},
		gv{40, 1}, gv{40, 3}, gv{41, 1}, gv{41, 3}, gv{42, 1}, gv{42, 5},
		gv{43, 1}, gv{43, 5},
		gv{50, 1}, gv{50, 3}, gv{51, 1}, gv{51, 2},
		gv{4, 2}, gv{11, 2}, gv{13, 2},
		gv{21, 6}, gv{22, 6}, gv{23, 6}, gv{30, 6}, gv{31, 4}, gv{31, 8},
		gv{32, 9}, gv{32, 6}, gv{33, 4}, gv{33, 6}, gv{40, 4}, gv{41, 4},
		gv{42, 4}, gv{42, 6}, gv{43, 4}, gv{43, 6},
		gv{50, 2},
		gv{12, 1}, gv{12, 2}, gv{21, 5}, gv{22, 5}, gv{23, 5}, gv{31, 3},
		gv{32, 3}, gv{32, 7}, gv{33, 3}, gv{33, 7}, gv{42, 3}, gv{42, 7},
		gv{43, 3}, gv{43, 7}, gv{50, 4},
		gv{32, 8}, gv{33, 8}, gv{42, 8}, gv{43, 8}:
		return packPointsBytes
	}

	return nil
}
