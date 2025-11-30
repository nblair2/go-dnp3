package dnp3

type groupVariation struct {
	Group     uint8
	Variation uint8
}

type objectType struct {
	Description string
	Constructor PointsConstructor `json:"-"`
	Packer      PointsPacker      `json:"-"`
}

var objectTypes = map[groupVariation]*objectType{
	// Binary Input
	{1, 0}: {Description: "(Static) Binary Input - Any Variations"},
	{1, 1}: {
		Description: "(Static) Binary Input - Packed Format",
		Constructor: newPoints1Bit,
		Packer:      packerPoints1Bit,
	},
	{1, 2}: {
		Description: "(Static) Binary Input - Status with Flags",
		Constructor: newPoints1BitFlags,
		Packer:      packPointsBytes,
	},

	// Binary Input Event
	{2, 0}: {
		Description: "(Event) Binary Input Event - Any Variations",
		Constructor: constructorNoPoints,
		Packer:      packNoPoints,
	},
	{2, 1}: {
		Description: "(Event) Binary Input Event",
		Constructor: makeGenericConstructor(newPointNBytes, 1),
		Packer:      packPointsBytes,
	},
	{2, 2}: {
		Description: "(Event) Binary Input Event - with Absolute Time",
		Constructor: makeGenericConstructor(newPointNBytesAbsTime, 7),
		Packer:      packPointsBytes,
	},
	{2, 3}: {
		Description: "(Event) Binary Input Event - with Relative Time",
		Constructor: makeGenericConstructor(newPointNBytesRelTime, 3),
		Packer:      packPointsBytes,
	},

	// Double-bit Binary Input
	{3, 0}: {Description: "(Static) Double-bit Binary Input - Any Variations"},
	{3, 1}: {
		Description: "(Static) Double-bit Binary Input - Packed Format",
		Constructor: newPoints2Bits,
		Packer:      packerPoints2Bits,
	},
	{3, 2}: {
		Description: "(Static) Double-bit Binary Input - Status with Flags",
		Constructor: makeGenericConstructor(newPointNBytes, 1),
		Packer:      packPointsBytes,
	},

	// Double-bit Binary Input Event
	{4, 0}: {Description: "(Event) Double-bit Binary Input Event - Any Variations"},
	{4, 1}: {
		Description: "(Event) Double-bit Binary Input Event",
		Constructor: makeGenericConstructor(newPointNBytes, 1),
		Packer:      packPointsBytes,
	},
	{4, 2}: {
		Description: "(Event) Double-bit Binary Input Event with Absolute Time",
		Constructor: makeGenericConstructor(newPointNBytesAbsTime, 7),
		Packer:      packPointsBytes,
	},
	{4, 3}: {
		Description: "(Event) Double-bit Binary Input Event with Relative Time",
		Constructor: makeGenericConstructor(newPointNBytesRelTime, 3),
		Packer:      packPointsBytes,
	},

	// Binary Output
	{10, 0}: {Description: "(Static) Binary Output - Any Variations"},
	{10, 1}: {
		Description: "(Static) Binary Output - Packed Format",
		Constructor: newPoints1Bit,
		Packer:      packerPoints1Bit,
	},
	{10, 2}: {
		Description: "(Static) Binary Output - Status with Flags",
		Constructor: newPoints1BitFlags,
		Packer:      packPointsBytes,
	},

	// Binary Output Event
	{11, 0}: {Description: "(Event) Binary Output Event - Any Variations"},
	{11, 1}: {
		Description: "(Event) Binary Output Event - Status",
		Constructor: makeGenericConstructor(newPointNBytes, 1),
		Packer:      packPointsBytes,
	},
	{11, 2}: {
		Description: "(Event) Binary Output Event - Status with Time",
		Constructor: makeGenericConstructor(newPointNBytesAbsTime, 7),
		Packer:      packPointsBytes,
	},

	// Binary Output Command
	{12, 0}: {Description: "(Command) Binary Output Command - Any Variations"},
	{12, 1}: {
		Description: "(Command) Binary Output Command - Control Relay Output Block",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{12, 2}: {
		Description: "(Command) Binary Output Command - Pattern Control Block",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{12, 3}: {Description: "(Command) Binary Output Command - Pattern Mask"},

	// Binary Output Command Event
	{13, 0}: {Description: "(Event) Binary Output Command Event - Any Variations"},
	{13, 1}: {
		Description: "(Event) Binary Output Command Event - Command Status",
		Constructor: makeGenericConstructor(newPointNBytes, 1),
		Packer:      packPointsBytes,
	},
	{13, 2}: {
		Description: "(Event) Binary Output Command Event - Command Status with Time",
		Constructor: makeGenericConstructor(newPointNBytesAbsTime, 7),
		Packer:      packPointsBytes,
	},

	// Counter
	{20, 0}: {Description: "(Static) Counter - Any Variations"},
	{20, 1}: {
		Description: "(Static) Counter - 32-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{20, 2}: {
		Description: "(Static) Counter - 16-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 3),
		Packer:      packPointsBytes,
	},
	{20, 5}: {
		Description: "(Static) Counter - 32-bit w/o Flag",
		Constructor: makeGenericConstructor(newPointNBytes, 4),
		Packer:      packPointsBytes,
	},
	{20, 6}: {
		Description: "(Static) Counter - 16-bit w/o Flag",
		Constructor: makeGenericConstructor(newPointNBytes, 2),
		Packer:      packPointsBytes,
	},

	// Frozen Counter
	{21, 0}: {Description: "(Static) Frozen Counter - Any Variations"},
	{21, 1}: {
		Description: "(Static) Frozen Counter - 32-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{21, 2}: {
		Description: "(Static) Frozen Counter - 16-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 3),
		Packer:      packPointsBytes,
	},
	{21, 5}: {
		Description: "(Static) Frozen Counter - 32-bit with Flag and Time",
		Constructor: makeGenericConstructor(newPointNBytesFlagsAbsTime, 11),
		Packer:      packPointsBytes,
	},
	{21, 6}: {
		Description: "(Static) Frozen Counter - 16-bit with Flag and Time",
		Constructor: makeGenericConstructor(newPointNBytesFlagsAbsTime, 9),
		Packer:      packPointsBytes,
	},
	{21, 9}: {
		Description: "(Static) Frozen Counter - 32-bit w/o Flag",
		Constructor: makeGenericConstructor(newPointNBytes, 4),
		Packer:      packPointsBytes,
	},
	{21, 10}: {
		Description: "(Static) Frozen Counter - 16-bit w/o Flag",
		Constructor: makeGenericConstructor(newPointNBytes, 2),
		Packer:      packPointsBytes,
	},

	// Counter Event
	{22, 0}: {Description: "(Event) Counter Event - Any Variations"},
	{22, 1}: {
		Description: "(Event) Counter Event - 32-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{22, 2}: {
		Description: "(Event) Counter Event - 16-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 3),
		Packer:      packPointsBytes,
	},
	{22, 5}: {
		Description: "(Event) Counter Event - 32-bit with Flag and Time",
		Constructor: makeGenericConstructor(newPointNBytesFlagsAbsTime, 11),
		Packer:      packPointsBytes,
	},
	{22, 6}: {
		Description: "(Event) Counter Event - 16-bit with Flag and Time",
		Constructor: makeGenericConstructor(newPointNBytesFlagsAbsTime, 9),
		Packer:      packPointsBytes,
	},

	// Frozen Counter Event
	{23, 0}: {Description: "(Event) Frozen Counter Event - Any Variations"},
	{23, 1}: {
		Description: "(Event) Frozen Counter Event - 32-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{23, 2}: {
		Description: "(Event) Frozen Counter Event - 16-bit with Flag",
		Constructor: makeGenericConstructor(
			newPointNBytesFlags,
			2,
		), // Wait, newPoints1ByteFlags is width 2.
		Packer: packPointsBytes,
	},
	{23, 5}: {
		Description: "(Event) Frozen Counter Event - 32-bit with Flag and Time",
		Constructor: makeGenericConstructor(newPointNBytesFlagsAbsTime, 11),
		Packer:      packPointsBytes,
	},
	{23, 6}: {
		Description: "(Event) Frozen Counter Event - 16-bit with Flag and Time",
		Constructor: makeGenericConstructor(newPointNBytesFlagsAbsTime, 9),
		Packer:      packPointsBytes,
	},

	// Analog Input
	{30, 0}: {Description: "(Static) Analog Input - Any Variations"},
	{30, 1}: {
		Description: "(Static) Analog Input - 32-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{30, 2}: {
		Description: "(Static) Analog Input - 16-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 3),
		Packer:      packPointsBytes,
	},
	{30, 3}: {
		Description: "(Static) Analog Input - 32-bit w/o Flag",
		Constructor: makeGenericConstructor(newPointNBytes, 4),
		Packer:      packPointsBytes,
	},
	{30, 4}: {
		Description: "(Static) Analog Input - 16-bit w/o Flag",
		Constructor: makeGenericConstructor(newPointNBytes, 2),
		Packer:      packPointsBytes,
	},
	{30, 5}: {
		Description: "(Static) Analog Input - Single-prec. FP with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{30, 6}: {
		Description: "(Static) Analog Input - Double-prec. FP with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 9),
		Packer:      packPointsBytes,
	},

	// Frozen Analog Input
	{31, 0}: {Description: "(Static) Frozen Analog Input - Any Variations"},
	{31, 1}: {
		Description: "(Static) Frozen Analog Input - 32-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{31, 2}: {
		Description: "(Static) Frozen Analog Input - 16-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 3),
		Packer:      packPointsBytes,
	},
	{31, 3}: {
		Description: "(Static) Frozen Analog Input - 32-bit with Time-of-Freeze",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{31, 4}: {
		Description: "(Static) Frozen Analog Input - 16-bit with Time-of-Freeze",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},
	{31, 5}: {
		Description: "(Static) Frozen Analog Input - 32-bit w/o Flag",
		Constructor: makeGenericConstructor(newPointNBytes, 4),
		Packer:      packPointsBytes,
	},
	{31, 6}: {
		Description: "(Static) Frozen Analog Input - 16-bit w/o Flag",
		Constructor: makeGenericConstructor(newPointNBytes, 2),
		Packer:      packPointsBytes,
	},
	{31, 7}: {
		Description: "(Static) Frozen Analog Input - Single-prec. FP with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{31, 8}: {
		Description: "(Static) Frozen Analog Input - Double-prec. FP with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 9),
		Packer:      packPointsBytes,
	},

	// Analog Input Event
	{32, 0}: {Description: "(Event) Analog Input Event - Any Variations"},
	{32, 1}: {
		Description: "(Event) Analog Input Event - 32-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{32, 2}: {
		Description: "(Event) Analog Input Event - 16-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 3),
		Packer:      packPointsBytes,
	},
	{32, 3}: {
		Description: "(Event) Analog Input Event - 32-bit with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{32, 4}: {
		Description: "(Event) Analog Input Event - 16-bit with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},
	{32, 5}: {
		Description: "(Event) Analog Input Event - Single-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{32, 6}: {
		Description: "(Event) Analog Input Event - Double-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},
	{32, 7}: {
		Description: "(Event) Analog Input Event - Single-prec. FP with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{32, 8}: {
		Description: "(Event) Analog Input Event - Double-prec. FP with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 15),
		Packer:      packPointsBytes,
	},

	// Frozen Analog Input Event
	{33, 0}: {Description: "(Event) Frozen Analog Input Event - Any Variations"},
	{33, 1}: {
		Description: "(Event) Frozen Analog Input Event - 32-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{33, 2}: {
		Description: "(Event) Frozen Analog Input Event - 16-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 3),
		Packer:      packPointsBytes,
	},
	{33, 3}: {
		Description: "(Event) Frozen Analog Input Event - 32-bit with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{33, 4}: {
		Description: "(Event) Frozen Analog Input Event - 16-bit with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},
	{33, 5}: {
		Description: "(Event) Frozen Analog Input Event - Single-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{33, 6}: {
		Description: "(Event) Frozen Analog Input Event - Double-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},
	{33, 7}: {
		Description: "(Event) Frozen Analog Input Event - Single-prec. FP with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{33, 8}: {
		Description: "(Event) Frozen Analog Input Event - Double-prec. FP with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 15),
		Packer:      packPointsBytes,
	},

	// Analog Input Deadband
	{34, 0}: {Description: "(Static) Analog Input Deadband - Any Variations"},
	{34, 1}: {
		Description: "(Static) Analog Input Deadband - 16-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 2),
		Packer:      packPointsBytes,
	},
	{34, 2}: {
		Description: "(Static) Analog Input Deadband - 32-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 4),
		Packer:      packPointsBytes,
	},
	{34, 3}: {
		Description: "(Static) Analog Input Deadband - Single-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 4),
		Packer:      packPointsBytes,
	},

	// Analog Output Status
	{40, 0}: {Description: "(Static) Analog Output Status - Any Variations"},
	{40, 1}: {
		Description: "(Static) Analog Output Status - 32-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{40, 2}: {
		Description: "(Static) Analog Output Status - 16-bit with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 3),
		Packer:      packPointsBytes,
	},
	{40, 3}: {
		Description: "(Static) Analog Output Status - Single-prec. FP with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 5),
		Packer:      packPointsBytes,
	},
	{40, 4}: {
		Description: "(Static) Analog Output Status - Double-prec. FP with Flag",
		Constructor: makeGenericConstructor(newPointNBytesFlags, 9),
		Packer:      packPointsBytes,
	},

	// Analog Output Command
	{41, 0}: {Description: "(Command) Analog Output Command - Any Variations"},
	{41, 1}: {
		Description: "(Command) Analog Output Command - 32-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{41, 2}: {
		Description: "(Command) Analog Output Command - 16-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 3),
		Packer:      packPointsBytes,
	},
	{41, 3}: {
		Description: "(Command) Analog Output Command - Single-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{41, 4}: {
		Description: "(Command) Analog Output Command - Double-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},

	// Analog Output Event
	{42, 0}: {Description: "(Event) Analog Output Event - Any Variations"},
	{42, 1}: {
		Description: "(Event) Analog Output Event - 32-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{42, 2}: {
		Description: "(Event) Analog Output Event - 16-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 3),
		Packer:      packPointsBytes,
	},
	{42, 3}: {
		Description: "(Event) Analog Output Event - 32-bit with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{42, 4}: {
		Description: "(Event) Analog Output Event - 16-bit with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},
	{42, 5}: {
		Description: "(Event) Analog Output Event - Single-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{42, 6}: {
		Description: "(Event) Analog Output Event - Double-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},
	{42, 7}: {
		Description: "(Event) Analog Output Event - Single-prec. FP with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{42, 8}: {
		Description: "(Event) Analog Output Event - Double-prec. FP with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 15),
		Packer:      packPointsBytes,
	},

	// Analog Output Command Event
	{43, 0}: {Description: "(Event) Analog Output Command Event - Any Variations"},
	{43, 1}: {
		Description: "(Event) Analog Output Command Event - 32-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{43, 2}: {
		Description: "(Event) Analog Output Command Event - 16-bit",
		Constructor: makeGenericConstructor(newPointNBytes, 3),
		Packer:      packPointsBytes,
	},
	{43, 3}: {
		Description: "(Event) Analog Output Command Event - 32-bit with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{43, 4}: {
		Description: "(Event) Analog Output Command Event - 16-bit with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},
	{43, 5}: {
		Description: "(Event) Analog Output Command Event - Single-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 5),
		Packer:      packPointsBytes,
	},
	{43, 6}: {
		Description: "(Event) Analog Output Command Event - Double-prec. FP",
		Constructor: makeGenericConstructor(newPointNBytes, 9),
		Packer:      packPointsBytes,
	},
	{43, 7}: {
		Description: "(Event) Analog Output Command Event - Single-prec. FP with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},
	{43, 8}: {
		Description: "(Event) Analog Output Command Event - Double-prec. FP with Time",
		Constructor: makeGenericConstructor(newPointNBytes, 15),
		Packer:      packPointsBytes,
	},

	// Time and Date
	{50, 1}: {
		Description: "(Info) Time and Date - Absolute Time",
		Constructor: makeGenericConstructor(newPointAbsTime, 6),
		Packer:      packPointsBytes,
	},
	{50, 2}: {
		Description: "(Info) Time and Date - Absolute Time and Interval",
		Constructor: makeGenericConstructor(newPointNBytesAbsTime, 10),
		Packer:      packPointsBytes,
	},
	{50, 3}: {
		Description: "(Info) Time and Date - Absolute Time at Last Recorded Time",
		Constructor: makeGenericConstructor(newPointAbsTime, 6),
		Packer:      packPointsBytes,
	},
	{50, 4}: {
		Description: "(Info) Time and Date - Indexed Absolute Time and Long Interval",
		Constructor: makeGenericConstructor(newPointNBytes, 11),
		Packer:      packPointsBytes,
	},

	// Time and Date CTO
	{51, 1}: {
		Description: "(Info) CTO - Absolute Time, Synchronized",
		Constructor: makeGenericConstructor(newPointAbsTime, 6),
		Packer:      packPointsBytes,
	},
	{51, 2}: {
		Description: "(Info) CTO - Absolute Time, Unsynchronized",
		Constructor: makeGenericConstructor(newPointAbsTime, 6),
		Packer:      packPointsBytes,
	},

	// Time Delay
	{52, 1}: {
		Description: "(Info) Time Delay Coarse",
		Constructor: makeGenericConstructor(newPointRelTime, 2),
		Packer:      packPointsBytes,
	},
	{52, 2}: {
		Description: "(Info) Time Delay Fine",
		Constructor: makeGenericConstructor(newPointRelTime, 2),
		Packer:      packPointsBytes,
	},

	// Read
	{60, 1}: {
		Description: "(Command) Class 0 Data",
		Constructor: constructorNoPoints,
		Packer:      packNoPoints,
	},
	{60, 2}: {
		Description: "(Command) Class 1 Data",
		Constructor: constructorNoPoints,
		Packer:      packNoPoints,
	},
	{60, 3}: {
		Description: "(Command) Class 2 Data",
		Constructor: constructorNoPoints,
		Packer:      packNoPoints,
	},
	{60, 4}: {
		Description: "(Command) Class 3 Data",
		Constructor: constructorNoPoints,
		Packer:      packNoPoints,
	},

	// Internal Indications
	{80, 1}: {
		Description: "(Info) Internal Indications - Packed Format",
		Constructor: makeGenericConstructor(newPointNBytes, 1),
		Packer:      packPointsBytes,
	},
}
