package dnp3

import (
	"fmt"
)

// ApplicationData holds an array of Data objects
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
		if err := obj.FromBytes(d[i:]); err != nil {
			ad.extra = d[i:]
			return fmt.Errorf("could not decode object: %# X, err: %v",
				d[i:], err)
		}
		ad.OBJS = append(ad.OBJS, obj)
		i += obj.SizeOf()
	}

	return nil
}

func (ad *ApplicationData) ToBytes() []byte {
	var o []byte
	for _, do := range ad.OBJS {
		b := do.ToBytes()
		o = append(o, b...)
	}
	o = append(o, ad.extra...)
	return o
}

func (ad *ApplicationData) String() string {
	o := ""
	header := "Data Objects:"
	header_added := false
	if len(ad.OBJS) > 0 {
		o += header
		header_added = true
		for _, obj := range ad.OBJS {
			o += "\n\t\t  - " + obj.String()
		}
	}

	if len(ad.extra) > 0 {
		if !header_added {
			o += header
		}
		o += fmt.Sprintf("\n\t\t  - Extra: 0x % X", ad.extra)
	}
	return o
}

type DataObject struct {
	Header ObjectHeader
	Points []Point
	size   int
}

func (do *DataObject) FromBytes(d []byte) error {
	if err := do.Header.FromBytes(d); err != nil {
		return fmt.Errorf("can't create Data Object Header: %w", err)
	}
	offset := do.Header.SizeOf()

	psBits := do.calcPointBitSize()
	if psBits == 0 {
		return fmt.Errorf("can't parse this Group/Var")
	} else if psBits%8 == 0 {
		ps := psBits / 8
		ps += do.Header.calcPrefixSize()
		for range do.Header.RangeField.NumObjects() {
			p := &PointBytes{}
			p.FromBytes(d[offset : offset+ps])
			do.Points = append(do.Points, p)
			offset += ps
		}
		do.size = offset
	} else if psBits == 1 || psBits == 2 { // packed bits
		num := do.Header.RangeField.NumObjects()
		ps := ((psBits * num) + 7) / 8
		p, err := unpackBitPoints(d[offset:offset+ps], num, psBits)
		if err != nil {
			return fmt.Errorf("could not unpack bits, got error: %w", err)
		}
		do.Points = p
		do.size = offset + ps
	} else {
		return fmt.Errorf("point size not 1, 2, mod 8, got: %d", psBits)
	}
	return nil
}

func (do *DataObject) ToBytes() []byte {
	var o []byte
	o = append(o, do.Header.ToBytes()...)

	if len(do.Points) > 0 {
		// TODO these should all be the same type. But what if not?
		switch do.Points[0].(type) {
		case *PointBytes:
			for _, p := range do.Points {
				o = append(o, p.ToBytes()...)
			}
		case *Point1Bit:
			points := make([]*Point1Bit, len(do.Points))
			for i, p := range do.Points {
				points[i] = p.(*Point1Bit)
			}
			o = append(o, pack1BitPoints(points)...)
		case *Point2Bit:
			points := make([]*Point2Bit, len(do.Points))
			for i, p := range do.Points {
				points[i] = p.(*Point2Bit)
			}
			o = append(o, pack2BitPoints(points)...)
		}
	}

	return o

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
	return do.size
}

func (do *DataObject) calcPointBitSize() int {
	type gv struct{ g, v int }
	groupVar := gv{int(do.Header.Group), int(do.Header.Variation)}
	switch groupVar {
	// Bits
	case gv{1, 1}, gv{10, 1}:
		return 1
	case gv{3, 1}:
		return 2
	// Full bytes
	case gv{1, 2}, gv{2, 1}, gv{3, 2}, gv{4, 1}, gv{10, 2}, gv{11, 1},
		gv{13, 1}, gv{80, 1}:
		return 8
	case gv{20, 6}, gv{21, 10}, gv{30, 4}, gv{31, 6}, gv{34, 1}, gv{52, 1},
		gv{52, 2}:
		return 16
	case gv{2, 3}, gv{4, 3}, gv{20, 2}, gv{21, 2}, gv{22, 2}, gv{23, 2},
		gv{30, 2}, gv{31, 2}, gv{32, 2}, gv{33, 2}, gv{40, 2}, gv{41, 2},
		gv{42, 2}, gv{43, 2}:
		return 24
	case gv{20, 5}, gv{21, 9}, gv{30, 3}, gv{31, 5}, gv{34, 2}, gv{34, 3}:
		return 32
	case gv{20, 1}, gv{21, 1}, gv{22, 1}, gv{23, 1}, gv{30, 1}, gv{30, 5},
		gv{31, 1}, gv{31, 7}, gv{32, 1}, gv{32, 5}, gv{33, 1}, gv{33, 5},
		gv{40, 1}, gv{40, 3}, gv{41, 1}, gv{41, 3}, gv{42, 1}, gv{42, 5},
		gv{43, 1}, gv{43, 5}:
		return 40
	case gv{50, 1}, gv{50, 3}, gv{51, 1}, gv{51, 2}:
		return 48
	case gv{4, 2}, gv{11, 2}, gv{13, 2}, gv{51, 2}:
		return 56
	case gv{21, 6}, gv{22, 6}, gv{23, 6}, gv{30, 6}, gv{31, 4}, gv{31, 8},
		gv{32, 9}, gv{32, 6}, gv{33, 4}, gv{33, 6}, gv{40, 4}, gv{41, 4},
		gv{42, 4}, gv{42, 6}, gv{43, 4}, gv{43, 6}:
		return 72
	case gv{50, 2}:
		return 80
	case gv{12, 1}, gv{12, 2}, gv{21, 5}, gv{22, 5}, gv{23, 5}, gv{31, 3},
		gv{32, 3}, gv{32, 7}, gv{33, 3}, gv{33, 7}, gv{42, 3}, gv{42, 7},
		gv{43, 3}, gv{43, 7}, gv{50, 4}:
		return 88
	case gv{32, 8}, gv{33, 8}, gv{42, 8}, gv{43, 8}:
		return 120
	}
	return 0
}

type Point interface {
	FromBytes([]byte) error
	ToBytes() []byte
	String() string
}

func unpackBitPoints(d []byte, num, size int) ([]Point, error) {
	switch size {
	case 1:
		return unpack1BitPoints(d, num)
	case 2:
		return unpack2BitPoints(d, num)
	default:
		return nil, fmt.Errorf("can't unpack bit size %d, must be 1-2", size)
	}
}

type Point1Bit struct {
	Value bool
}

// should not be used directly
func (p *Point1Bit) FromBytes(data []byte) error {
	if len(data) > 1 {
		return fmt.Errorf("can't construct 1 bit point from multiple bytes")
	}
	// assume bit is the lowest order
	p.Value = data[0]&0b00000001 != 0
	return nil
}

func (p *Point1Bit) ToBytes() []byte {
	if p.Value {
		return []byte{0b00000001}
	}
	return []byte{0b00000000}
}

func (p *Point1Bit) String() string {
	return fmt.Sprintf("\n\t\t\t      - Value: %t", p.Value)
}

func unpack1BitPoints(d []byte, num int) ([]Point, error) {
	if num > (8 * len(d)) {
		return nil, fmt.Errorf("not enough bytes for %d bit points", num)
	}

	var o []Point
	var mask uint8
	for i := range num {
		mask = 0b00000001 << (i % 8)
		b := d[i/8]
		p := &Point1Bit{}
		p.Value = (uint8(b) & mask) != 0
		o = append(o, p)
	}
	return o, nil
}

func pack1BitPoints(points []*Point1Bit) []byte {
	var o []byte
	for i := 0; i < len(points); i += 8 {
		var b byte
		for j := 0; j < 8 && i+j < len(points); j += 1 {
			if points[i+j].Value {
				b |= 1 << j
			}
		}
		o = append(o, b)
	}
	return o
}

type Point2Bit struct {
	Value [2]bool
}

// should not be used directly
func (p *Point2Bit) FromBytes(data []byte) error {
	if len(data) > 1 {
		return fmt.Errorf("can't construct 2 bit point from multiple bytes")
	}
	// assume bit is the lowest order
	p.Value = [2]bool{
		data[0]&0b00000001 != 0,
		data[0]&0b00000010 != 0,
	}
	return nil
}

func (p *Point2Bit) ToBytes() []byte {
	var o byte
	if p.Value[0] {
		o += 0b00000001
	}
	if p.Value[1] {
		o += 0b00000010
	}
	return []byte{o}
}

func (p *Point2Bit) String() string {
	return fmt.Sprintf("\n\t\t\t      - Value: %t, %t", p.Value[0], p.Value[1])
}

func unpack2BitPoints(d []byte, num int) ([]Point, error) {
	if num > (8 * len(d) / 2) {
		return nil, fmt.Errorf("not enough bytes for %d bit points", num)
	}

	var o []Point
	var mask1 uint8
	var mask2 uint8
	for i := range num {
		mask1 = 0b00000001 << (i % 8)
		mask2 = 0b00000010 << (i % 8)

		b := d[i%8]
		p := &Point2Bit{}
		p.Value = [2]bool{
			(uint8(b) & mask1) != 0,
			(uint8(b) & mask2) != 0,
		}
		o = append(o, p)
	}
	return o, nil
}

func pack2BitPoints(points []*Point2Bit) []byte {
	var o []byte
	for i := 0; i < len(points); i += 4 {
		var b byte
		for j := 0; j < 4 && i+j < len(points); j += 1 {
			if points[i+j].Value[0] {
				b |= 0b00000001 << (j * 2)
			}
			if points[i+j].Value[1] {
				b |= 0b00000010 << (j * 2)
			}
		}
		o = append(o, b)
	}
	return o
}

type PointBytes struct {
	Value []byte
}

func (point *PointBytes) FromBytes(d []byte) error {
	point.Value = d
	return nil
}

func (point *PointBytes) ToBytes() []byte {
	return point.Value
}

func (point *PointBytes) String() string {
	return fmt.Sprintf("\n\t\t\t      - Raw: 0x % X", point.Value)
}
