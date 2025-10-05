package dnp3

import (
	"errors"
	"fmt"
	"time"
)

type Point interface {
	FromBytes(data []byte, prefSize int) error
	ToBytes() []byte
	String() string
}

var pad = "\n\t\t\t      - "

type PointsConstructor func([]byte, int, int) ([]Point, int, error)

type PointsPacker func([]Point) ([]byte, error)

type Point1Bit struct {
	Value bool
}

// should not be used directly.
func (p *Point1Bit) FromBytes(data []byte, prefSize int) error {
	if len(data) > 1 {
		return errors.New("can't construct 1 bit point from multiple bytes")
	} else if prefSize != 0 {
		return errors.New("can't have a prefix on 1 bit packed points")
	}

	p.Value = data[0]&0b00000001 != 0

	return nil
}

// should not be used directly.
func (p *Point1Bit) ToBytes() []byte {
	if p.Value {
		return []byte{0b00000001}
	}

	return []byte{0b00000000}
}

func (p *Point1Bit) String() string {
	return fmt.Sprintf(pad+"Value: %t", p.Value)
}

func newPoints1Bit(d []byte, num, prefSize int) ([]Point, int, error) {
	if num > (8 * len(d)) {
		return nil, 0, fmt.Errorf("not enough bytes for %d bit points", num)
	} else if prefSize != 0 {
		return nil, 0, errors.New("prefix size must be 0 for packed bits")
	}

	var mask uint8

	o := make([]Point, 0, num)

	for i := range num {
		mask = 0b00000001 << (i % 8)
		b := d[i/8]
		p := &Point1Bit{}
		p.Value = (b & mask) != 0
		o = append(o, p)
	}

	size := (num + 7) / 8 // round up to nearest byte

	return o, size, nil
}

func packerPoints1Bit(points []Point) ([]byte, error) {
	var o []byte
	for i := 0; i < len(points); i += 8 {
		var b byte

		for j := 0; j < 8 && i+j < len(points); j++ {
			pt, ok := points[i+j].(*Point1Bit)
			if !ok {
				return o, fmt.Errorf("element %d is not *Point1Bit, got %T",
					i+j, points[i+j])
			}

			if pt.Value {
				b |= 1 << j
			}
		}

		o = append(o, b)
	}

	return o, nil
}

type Point1BitFlags struct {
	Prefix []byte
	Value  bool
	Flags  PointFlags
}

func (p *Point1BitFlags) FromBytes(data []byte, prefSize int) error {
	if len(data) > 1+prefSize {
		return errors.New("can't construct 1 bit point with flags from multiple bytes")
	}

	if prefSize > 0 {
		p.Prefix = data[0:prefSize]
	}

	p.Value = data[prefSize]&0b10000000 != 0
	b := data[prefSize] & 0b01111111 // clear bit 7

	err := p.Flags.FromByte(b)
	if err != nil {
		return fmt.Errorf("couldn't decode flags byte: 0x % X, err: %w", b, err)
	}

	return nil
}

func (p *Point1BitFlags) ToBytes() []byte {
	o := p.Prefix

	b := p.Flags.ToByte()
	if p.Value {
		b |= 0b10000000
	}

	return append(o, b)
}

func (p *Point1BitFlags) String() string {
	o := pad + ""
	if len(p.Prefix) > 0 {
		o += fmt.Sprintf("Prefix: 0x % X\n\t\t\t\t", p.Prefix)
	}

	o += fmt.Sprintf("Value : %t", p.Value)
	o += p.Flags.String()

	return o
}

func newPoints1BitFlags(d []byte, num, prefSize int) ([]Point, int, error) {
	if num > len(d) {
		return nil, 0, fmt.Errorf("not enough bytes for %d 1-bit points with flags", num)
	}

	o := make([]Point, 0, num)

	for i := range num {
		p := &Point1BitFlags{}

		err := p.FromBytes([]byte{d[i]}, prefSize)
		if err != nil {
			return o, num, fmt.Errorf("could not decode point: 0x % X, err: %w",
				d[i], err)
		}

		o = append(o, p)
	}

	return o, num, nil
}

type Point2Bits struct {
	Value [2]bool
}

// should not be used directly.
func (p *Point2Bits) FromBytes(data []byte, prefSize int) error {
	if len(data) > 1 {
		return errors.New("can't construct 2 bit point from multiple bytes")
	} else if prefSize != 0 {
		return errors.New("can't have prefix on 2 bit packed points")
	}
	// assume bit is the lowest order
	p.Value = [2]bool{
		data[0]&0b00000001 != 0,
		data[0]&0b00000010 != 0,
	}

	return nil
}

// should not be used directly.
func (p *Point2Bits) ToBytes() []byte {
	var o byte
	if p.Value[0] {
		o += 0b00000001
	}

	if p.Value[1] {
		o += 0b00000010
	}

	return []byte{o}
}

func (p *Point2Bits) String() string {
	return fmt.Sprintf(pad+"%t\n\t\t\t\t%t", p.Value[0], p.Value[1])
}

func newPoints2Bits(d []byte, num, prefSize int) ([]Point, int, error) {
	if num > (8 * len(d) / 2) {
		return nil, 0, fmt.Errorf("not enough bytes for %d bit points", num)
	} else if prefSize != 0 {
		return nil, 0, errors.New("can't have a prefix for 2 bit packed")
	}

	var (
		mask1 uint8
		mask2 uint8
	)

	o := make([]Point, 0, num)

	for i := range num {
		mask1 = 0b00000001 << (i % 8)
		mask2 = 0b00000010 << (i % 8)

		b := d[i%8]
		p := &Point2Bits{}
		p.Value = [2]bool{
			(b & mask1) != 0,
			(b & mask2) != 0,
		}
		o = append(o, p)
	}

	size := (num + 3) / 4 // round up to nearest byte

	return o, size, nil
}

func packerPoints2Bits(points []Point) ([]byte, error) {
	var o []byte
	for i := 0; i < len(points); i += 4 {
		var b byte

		for j := 0; j < 4 && i+j < len(points); j++ {
			pt, ok := points[i+j].(*Point2Bits)
			if !ok {
				return o, fmt.Errorf("element %d is not *Point2Bits, got %T", i+j, points[i+j])
			}

			if pt.Value[0] {
				b |= 0b00000001 << (j * 2)
			}

			if pt.Value[1] {
				b |= 0b00000010 << (j * 2)
			}
		}

		o = append(o, b)
	}

	return o, nil
}

type PointNBytes struct {
	Prefix []byte
	Value  []byte
}

func (p *PointNBytes) FromBytes(d []byte, prefSize int) error {
	if prefSize != 0 {
		p.Prefix = d[0:prefSize]
	}

	p.Value = d[prefSize:]

	return nil
}

func (p *PointNBytes) ToBytes() []byte {
	return append(p.Prefix, p.Value...)
}

func (p *PointNBytes) String() string {
	o := pad + ""
	if len(p.Prefix) > 0 {
		o += fmt.Sprintf("Prefix: 0x % X\n\t\t\t\t", p.Prefix)
	}

	return o + fmt.Sprintf("Value : 0x % X", p.Value)
}

func newPointNBytes() Point {
	return &PointNBytes{}
}

type PointNBytesFlags struct {
	Prefix []byte
	Value  []byte
	Flags  PointFlags
}

func (p *PointNBytesFlags) FromBytes(d []byte, prefSize int) error {
	if len(d) != 2+prefSize {
		return fmt.Errorf("need %d bytes, got %d", 2+prefSize, len(d))
	}

	if prefSize != 0 {
		p.Prefix = d[0:prefSize]
	}

	err := p.Flags.FromByte(d[prefSize])
	if err != nil {
		return fmt.Errorf("couldn't decode flags byte: 0x % X, err: %w", d[prefSize], err)
	}

	p.Value = d[prefSize+1:]

	return nil
}

func (p *PointNBytesFlags) ToBytes() []byte {
	var o []byte

	o = append(o, p.Prefix...)
	o = append(o, p.Flags.ToByte())

	return append(o, p.Value...)
}

func (p *PointNBytesFlags) String() string {
	o := pad + ""
	if len(p.Prefix) > 0 {
		o += fmt.Sprintf("Prefix: 0x % X\n\t\t\t\t", p.Prefix)
	}

	o += fmt.Sprintf("Value : 0x % X", p.Value)
	o += p.Flags.String()

	return o
}

func newPointNBytesFlags() Point {
	return &PointNBytesFlags{}
}

type PointNBytesFlagsAbsTime struct {
	Prefix  []byte
	Value   []byte
	Flags   PointFlags
	AbsTime time.Time
}

func (p *PointNBytesFlagsAbsTime) FromBytes(d []byte, prefSize int) error {
	if len(d) != 8+prefSize {
		return fmt.Errorf("need %d bytes, got %d", 8+prefSize, len(d))
	}

	if prefSize != 0 {
		p.Prefix = d[0:prefSize]
	}

	err := p.Flags.FromByte(d[prefSize])
	if err != nil {
		return fmt.Errorf("couldn't decode flags byte: 0x % X, err: %w", d[prefSize], err)
	}

	p.AbsTime = bytesToDNP3TimeAbsolute(d[prefSize+1 : prefSize+7])
	p.Value = d[prefSize+7:]

	return nil
}

func (p *PointNBytesFlagsAbsTime) ToBytes() []byte {
	var o []byte

	o = append(o, p.Prefix...)
	o = append(o, p.Flags.ToByte())
	o = append(o, dnp3TimeAbsoluteToBytes(p.AbsTime)...)
	o = append(o, p.Value...)

	return o
}

func (p *PointNBytesFlagsAbsTime) String() string {
	o := pad + ""
	if len(p.Prefix) > 0 {
		o += fmt.Sprintf("Prefix: 0x % X\n\t\t\t\t", p.Prefix)
	}

	o += fmt.Sprintf("Value: 0x % X", p.Value)
	o += p.Flags.String()
	o += fmt.Sprintf("\n\t\t\t\tTimestamp: %s", p.AbsTime.UTC())

	return o
}

func newPointNBytesFlagsAbsTime() Point {
	return &PointNBytesFlagsAbsTime{}
}

type PointNBytesAbsTime struct {
	Prefix  []byte
	Value   []byte
	AbsTime time.Time
}

func (p *PointNBytesAbsTime) FromBytes(d []byte, prefSize int) error {
	if len(d) != 7+prefSize {
		return fmt.Errorf("need %d bytes, got %d", 8+prefSize, len(d))
	}

	if prefSize != 0 {
		p.Prefix = d[0:prefSize]
	}

	p.Value = d[prefSize : len(d)-6]
	p.AbsTime = bytesToDNP3TimeAbsolute(d[len(d)-6:])

	return nil
}

func (p *PointNBytesAbsTime) ToBytes() []byte {
	var o []byte

	o = append(o, p.Prefix...)
	o = append(o, p.Value...)

	return append(o, dnp3TimeAbsoluteToBytes(p.AbsTime)...)
}

func (p *PointNBytesAbsTime) String() string {
	o := pad + ""
	if len(p.Prefix) > 0 {
		o += fmt.Sprintf("Prefix: 0x % X\n\t\t\t\t", p.Prefix)
	}

	o += fmt.Sprintf("Value: 0x % X", p.Value)
	o += fmt.Sprintf("\n\t\t\t\tTimestamp: %s", p.AbsTime.UTC())

	return o
}

func newPointNBytesAbsTime() Point {
	return &PointNBytesAbsTime{}
}

type PointNBytesRelTime struct {
	Prefix  []byte
	Value   []byte
	RelTime time.Duration
}

func (p *PointNBytesRelTime) FromBytes(d []byte, prefSize int) error {
	if len(d) < 3+prefSize {
		return fmt.Errorf("need %d bytes, got %d", 3+prefSize, len(d))
	}

	if prefSize != 0 {
		p.Prefix = d[0:prefSize]
	}

	p.Value = d[prefSize : len(d)-2]
	p.RelTime = bytesTodnp3TimeRelative(d[len(d)-2:])

	return nil
}

func (p *PointNBytesRelTime) ToBytes() []byte {
	var o []byte

	o = append(o, p.Prefix...)
	o = append(o, p.Value...)

	return append(o, dnp3TimeRelativeToBytes(p.RelTime)...)
}

func (p *PointNBytesRelTime) String() string {
	o := pad + ""
	if len(p.Prefix) > 0 {
		o += fmt.Sprintf("Prefix: 0x % X\n\t\t\t\t", p.Prefix)
	}

	o += fmt.Sprintf("Value: 0x % X", p.Value)
	o += fmt.Sprintf("\n\t\t\t\tTimestamp offset: %s", p.RelTime)

	return o
}

func newPointNBytesRelTime() Point {
	return &PointNBytesRelTime{}
}

type PointAbsTime struct {
	Prefix  []byte
	AbsTime time.Time
}

func (p *PointAbsTime) FromBytes(d []byte, prefSize int) error {
	if len(d) != 6+prefSize {
		return fmt.Errorf("need %d bytes, got %d", 6+prefSize, len(d))
	}

	if prefSize != 0 {
		p.Prefix = d[0:prefSize]
	}

	p.AbsTime = bytesToDNP3TimeAbsolute(d[0:6])

	return nil
}

func (p *PointAbsTime) ToBytes() []byte {
	return append(p.Prefix, dnp3TimeAbsoluteToBytes(p.AbsTime)...)
}

func (p *PointAbsTime) String() string {
	o := pad + ""
	if len(p.Prefix) > 0 {
		o += fmt.Sprintf("Prefix: 0x % X\n\t\t\t\t", p.Prefix)
	}

	return o + fmt.Sprintf("Timestamp: %s", p.AbsTime.UTC())
}

func newPointAbsTime() Point {
	return &PointAbsTime{}
}

type PointRelTime struct {
	Prefix  []byte
	RelTime time.Duration
}

func (p *PointRelTime) FromBytes(d []byte, prefSize int) error {
	if len(d) < 2+prefSize {
		return fmt.Errorf("need %d bytes, got %d", prefSize+2, len(d))
	}

	if prefSize != 0 {
		p.Prefix = d[0:prefSize]
	}

	p.RelTime = bytesTodnp3TimeRelative(d[prefSize : prefSize+2])

	return nil
}

func (p *PointRelTime) ToBytes() []byte {
	return append(p.Prefix, dnp3TimeRelativeToBytes(p.RelTime)...)
}

func (p *PointRelTime) String() string {
	o := pad + ""
	if len(p.Prefix) > 0 {
		o += fmt.Sprintf("Prefix: 0x % X\n\t\t\t\t", p.Prefix)
	}

	return o + fmt.Sprintf("Timestamp offset: %s", p.RelTime)
}

func newPointRelTime() Point {
	return &PointRelTime{}
}

func newPointsGeneric(
	newPoint func() Point,
	data []byte,
	width, num, prefSize int,
) ([]Point, int, error) {
	size := num * (prefSize + width)
	if size > len(data) {
		return nil, 0, fmt.Errorf(
			"not enough bytes for %d %d-byte points with %d-byte prefix",
			num,
			width,
			prefSize,
		)
	}

	o := make([]Point, 0, num)

	for i := range num {
		p := newPoint()

		err := p.FromBytes(data[i*(width+prefSize):(i+1)*(width+prefSize)], prefSize)
		if err != nil {
			return o, size, fmt.Errorf("could not decode point: 0x % X, err: %w", data, err)
		}

		o = append(o, p)
	}

	return o, size, nil
}

// Only Bytes.
func newPoints1Byte(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytes, data, 1, num, prefSize)
}

func newPoints2Bytes(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytes, data, 2, num, prefSize)
}

func newPoints3Bytes(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytes, data, 3, num, prefSize)
}

func newPoints4Bytes(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytes, data, 4, num, prefSize)
}

func newPoints5Bytes(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytes, data, 5, num, prefSize)
}

func newPoints9Bytes(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytes, data, 9, num, prefSize)
}

func newPoints11Bytes(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytes, data, 11, num, prefSize)
}

func newPoints15Bytes(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytes, data, 15, num, prefSize)
}

// Bytes + Flags.
func newPoints1ByteFlags(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytesFlags, data, 2, num, prefSize)
}

func newPoints2BytesFlags(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytesFlags, data, 3, num, prefSize)
}

func newPoints4BytesFlags(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytesFlags, data, 5, num, prefSize)
}

func newPoints8BytesFlags(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytesFlags, data, 9, num, prefSize)
}

// Bytes + Flags + Time.
func newPoints2BytesFlagsAbsTime(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytesFlagsAbsTime, data, 9, num, prefSize)
}

func newPoints4BytesFlagsAbsTime(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytesFlagsAbsTime, data, 11, num, prefSize)
}

// Bytes + Time.
func newPoints1ByteAbsTime(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytesAbsTime, data, 7, num, prefSize)
}

func newPoints1ByteRelTime(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytesRelTime, data, 3, num, prefSize)
}

func newPoints4BytesAbsTime(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointNBytesAbsTime, data, 10, num, prefSize)
}

// Time.
func newPointsAbsTime(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointAbsTime, data, 6, num, prefSize)
}

func newPointsRelTime(data []byte, num, prefSize int) ([]Point, int, error) {
	return newPointsGeneric(newPointRelTime, data, 2, num, prefSize)
}

func packPointsBytes(points []Point) ([]byte, error) {
	var o []byte
	for _, p := range points {
		o = append(o, p.ToBytes()...)
	}

	return o, nil
}

func constructorNoPoints(_ []byte, num, _ int) ([]Point, int, error) {
	if num != 0 {
		return nil, 0, fmt.Errorf("no points expected, got %d", num)
	}

	return nil, 0, nil
}

func packNoPoints(points []Point) ([]byte, error) {
	if len(points) != 0 {
		return nil, fmt.Errorf("no points expected, got %d", len(points))
	}

	return nil, nil
}

type PointFlags struct {
	Reserved       bool // should be 0
	PointValue     bool
	ReferenceCheck bool
	OverRange      bool
	LocalForce     bool
	RemoteForce    bool
	CommFail       bool
	Restart        bool
	Online         bool
}

func (f *PointFlags) FromByte(data byte) error {
	f.Reserved = data&0b10000000 != 0
	f.ReferenceCheck = data&0b01000000 != 0
	f.OverRange = data&0b00100000 != 0
	f.LocalForce = data&0b00010000 != 0
	f.RemoteForce = data&0b00001000 != 0
	f.CommFail = data&0b00000100 != 0
	f.Restart = data&0b00000010 != 0

	f.Online = data&0b00000001 != 0
	if f.Reserved {
		return errors.New("reserved bit must be 0")
	}

	return nil
}

func (f *PointFlags) ToByte() byte {
	var b byte
	if f.Reserved {
		b |= 0b10000000
	}

	if f.ReferenceCheck {
		b |= 0b01000000
	}

	if f.OverRange {
		b |= 0b00100000
	}

	if f.LocalForce {
		b |= 0b00010000
	}

	if f.RemoteForce {
		b |= 0b00001000
	}

	if f.CommFail {
		b |= 0b00000100
	}

	if f.Restart {
		b |= 0b00000010
	}

	if f.Online {
		b |= 0b00000001
	}

	return b
}

func (f *PointFlags) String() string {
	return fmt.Sprintf(`
				Flags:
					Reference Check: %t
					Over-Range     : %t
					Local Force    : %t
					Remote Force   : %t
					Comm Fail      : %t
					Restart        : %t
					Online         : %t`,
		f.ReferenceCheck, f.OverRange, f.LocalForce, f.RemoteForce,
		f.CommFail, f.Restart, f.Online)
}
