package dnp3

import "fmt"

type ApplicationData struct {
	raw  []byte
	OBJS []DataObject
}

func (ad *ApplicationData) FromBytes(d []byte) error {
	ad.raw = d

	for i := 0; i < len(d); {
		var obj DataObject
		if err := obj.FromBytes(d[i:]); err != nil {
			return fmt.Errorf("could not decode object: %s, err: %v",
				obj, err)
		}
		ad.OBJS = append(ad.OBJS, obj)
		i += obj.SizeOf()
	}

	return nil
}

func (ad *ApplicationData) ToBytes() []byte {
	return ad.raw
}

func (ad *ApplicationData) String() string {
	o := ""
	first := true
	for _, obj := range ad.OBJS {
		if first {
			o += "Objects:\n"
			first = false
		}
		o += obj.String() + "\n"
	}
	return o
}

type DataObject struct {
	Header    ObjectHeader
	PointSize int
	Points    []Point
}

func (do *DataObject) FromBytes(d []byte) error {
	if err := do.Header.FromBytes(d); err != nil {
		return fmt.Errorf("can't create Data Object Header: %w", err)
	}
	offset := do.Header.SizeOf()
	do.PointSize = do.calcPointSize()
	do.PointSize += do.Header.calcPrefixSize()
	for range do.Header.RangeField.NumObjects() {
		p := Point{}
		p.FromBytes(d[offset : offset+do.PointSize])
		do.Points = append(do.Points, p)
		offset += do.PointSize
	}

	return nil
}

func (do *DataObject) ToBytes() []byte {
	var o []byte
	o = append(o, do.Header.ToBytes()...)
	for _, p := range do.Points {
		o = append(o, p.ToBytes()...)
	}
	return o
}

func (do *DataObject) String() string {
	o := fmt.Sprintf("\t\t      - %s", do.Header.String())
	first := true
	for _, p := range do.Points {
		if first {
			o += "\n\t\t\tPoints:"
			first = false
		}
		o += p.String()
	}
	return o
}

func (do *DataObject) SizeOf() int {
	return do.Header.SizeOf() + do.PointSize*len(do.Points)
}

func (do *DataObject) calcPointSize() int {
	type gv struct{ g, v int }
	groupVar := gv{int(do.Header.Group), int(do.Header.Variation)}
	switch groupVar {
	case gv{1, 2}, gv{2, 1}, gv{3, 2}, gv{4, 1}, gv{10, 2}, gv{11, 1},
		gv{13, 1}:
		return 1
	case gv{20, 6}, gv{21, 10}, gv{30, 4}, gv{31, 6}, gv{34, 1}:
		return 2
	case gv{2, 3}, gv{4, 3}, gv{20, 2}, gv{21, 2}, gv{22, 2}, gv{23, 2},
		gv{30, 2}, gv{31, 2}, gv{32, 2}, gv{33, 2}, gv{40, 2}, gv{41, 2},
		gv{42, 2}, gv{43, 2}:
		return 3
	case gv{20, 5}, gv{21, 9}, gv{30, 3}, gv{31, 5}, gv{34, 2}, gv{34, 3}:
		return 4
	case gv{20, 1}, gv{21, 1}, gv{22, 1}, gv{23, 1}, gv{30, 1}, gv{30, 5},
		gv{31, 1}, gv{31, 7}, gv{32, 1}, gv{32, 5}, gv{33, 1}, gv{33, 5},
		gv{40, 1}, gv{40, 3}, gv{41, 1}, gv{41, 3}, gv{42, 1}, gv{42, 5},
		gv{43, 1}, gv{43, 5}:
		return 5
	case gv{50, 1}, gv{50, 3}, gv{51, 1}, gv{51, 2}:
		return 6
	case gv{4, 2}, gv{11, 2}, gv{13, 2}, gv{51, 2}:
		return 7
	case gv{21, 6}, gv{22, 6}, gv{23, 6}, gv{30, 6}, gv{31, 4}, gv{31, 8},
		gv{32, 9}, gv{32, 6}, gv{33, 4}, gv{33, 6}, gv{40, 4}, gv{41, 4},
		gv{42, 4}, gv{42, 6}, gv{43, 4}, gv{43, 6}:
		return 9
	case gv{50, 2}:
		return 10
	case gv{12, 1}, gv{12, 2}, gv{21, 5}, gv{22, 5}, gv{23, 5}, gv{31, 3},
		gv{32, 3}, gv{32, 7}, gv{33, 3}, gv{33, 7}, gv{42, 3}, gv{42, 7},
		gv{43, 3}, gv{43, 7}, gv{50, 4}:
		return 11
	case gv{32, 8}, gv{33, 8}, gv{42, 8}, gv{43, 8}:
		return 15
	}
	return 0
}

type Point struct {
	Value []byte
}

func (point *Point) FromBytes(d []byte) {
	point.Value = d
}

func (point *Point) ToBytes() []byte {
	return point.Value
}

func (point *Point) String() string {
	return fmt.Sprintf("\n\t\t\t      - Value: 0x % X", point.Value)
}
