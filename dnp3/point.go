package dnp3

import (
	"errors"
	"fmt"
	"time"
)

type Point interface {
	FromBytes(data []byte, prefSize int) error
	ToBytes() ([]byte, error)
	String() string
}

type PointsConstructor func([]byte, int, int) ([]Point, int, error)

type PointsPacker func([]Point) ([]byte, error)

type Point1Bit struct {
	Value bool `json:"value"`
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
func (p *Point1Bit) ToBytes() ([]byte, error) {
	if p.Value {
		return []byte{0b00000001}, nil
	}

	return []byte{0b00000000}, nil
}

func (p *Point1Bit) String() string {
	return fmt.Sprintf("Value: %t", p.Value)
}

func newPoints1Bit(data []byte, num, prefSize int) ([]Point, int, error) {
	if num > (8 * len(data)) {
		return nil, 0, fmt.Errorf("not enough bytes for %d bit points", num)
	} else if prefSize != 0 {
		return nil, 0, errors.New("prefix size must be 0 for packed bits")
	}

	var mask uint8

	pointsOut := make([]Point, 0, num)

	for pointIndex := range num {
		mask = 0b00000001 << (pointIndex % 8)
		sourceByte := data[pointIndex/8]
		point := &Point1Bit{Value: (sourceByte & mask) != 0}
		pointsOut = append(pointsOut, point)
	}

	size := (num + 7) / 8 // round up to nearest byte

	return pointsOut, size, nil
}

func packerPoints1Bit(points []Point) ([]byte, error) {
	var packed []byte

	for pointOffset := 0; pointOffset < len(points); pointOffset += 8 {
		var packedByte byte

		for bitOffset := 0; bitOffset < 8 && pointOffset+bitOffset < len(points); bitOffset++ {
			point, ok := points[pointOffset+bitOffset].(*Point1Bit)
			if !ok {
				return packed, fmt.Errorf("element %d is not *Point1Bit, got %T",
					pointOffset+bitOffset, points[pointOffset+bitOffset])
			}

			if point.Value {
				packedByte |= 1 << bitOffset
			}
		}

		packed = append(packed, packedByte)
	}

	return packed, nil
}

type Point1BitFlags struct {
	Prefix []byte     `json:"prefix,omitempty"`
	Value  bool       `json:"value"`
	Flags  PointFlags `json:"flags"`
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

func (p *Point1BitFlags) ToBytes() ([]byte, error) {
	output := p.Prefix

	b := p.Flags.ToByte()
	if p.Value {
		b |= 0b10000000
	}

	return append(output, b), nil
}

func (p *Point1BitFlags) String() string {
	output := ""
	if len(p.Prefix) > 0 {
		output += fmt.Sprintf("Prefix: 0x % X\n", p.Prefix)
	}

	output += fmt.Sprintf("Value : %t\n", p.Value)
	output += p.Flags.String()

	return output
}

func newPoints1BitFlags(data []byte, num, prefSize int) ([]Point, int, error) {
	if num > len(data) {
		return nil, 0, fmt.Errorf("not enough bytes for %d 1-bit points with flags", num)
	}

	pointsOut := make([]Point, 0, num)

	for pointIndex := range num {
		point := &Point1BitFlags{}

		err := point.FromBytes([]byte{data[pointIndex]}, prefSize)
		if err != nil {
			return pointsOut, num, fmt.Errorf("could not decode point: 0x % X, err: %w",
				data[pointIndex], err)
		}

		pointsOut = append(pointsOut, point)
	}

	return pointsOut, num, nil
}

type Point2Bits struct {
	Value [2]bool `json:"value"`
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
func (p *Point2Bits) ToBytes() ([]byte, error) {
	var packedValue byte
	if p.Value[0] {
		packedValue += 0b00000001
	}

	if p.Value[1] {
		packedValue += 0b00000010
	}

	return []byte{packedValue}, nil
}

func (p *Point2Bits) String() string {
	return fmt.Sprintf("%t\n%t", p.Value[0], p.Value[1])
}

func newPoints2Bits(data []byte, num, prefSize int) ([]Point, int, error) {
	if num > (8*len(data))/2 {
		return nil, 0, fmt.Errorf("not enough bytes for %d bit points", num)
	} else if prefSize != 0 {
		return nil, 0, errors.New("can't have a prefix for 2 bit packed")
	}

	var (
		mask1 uint8
		mask2 uint8
	)

	pointsOut := make([]Point, 0, num)

	for pointIndex := range num {
		mask1 = 0b00000001 << (pointIndex % 8)
		mask2 = 0b00000010 << (pointIndex % 8)

		sourceByte := data[pointIndex%8]
		point := &Point2Bits{
			Value: [2]bool{
				(sourceByte & mask1) != 0,
				(sourceByte & mask2) != 0,
			},
		}
		pointsOut = append(pointsOut, point)
	}

	size := (num + 3) / 4 // round up to nearest byte

	return pointsOut, size, nil
}

func packerPoints2Bits(points []Point) ([]byte, error) {
	var packed []byte

	for pointOffset := 0; pointOffset < len(points); pointOffset += 4 {
		var packedByte byte

		for pairIndex := 0; pairIndex < 4 && pointOffset+pairIndex < len(points); pairIndex++ {
			elementIndex := pointOffset + pairIndex

			point, ok := points[elementIndex].(*Point2Bits)
			if !ok {
				return packed, fmt.Errorf(
					"element %d is not *Point2Bits, got %T",
					elementIndex,
					points[elementIndex],
				)
			}

			if point.Value[0] {
				packedByte |= 0b00000001 << (pairIndex * 2)
			}

			if point.Value[1] {
				packedByte |= 0b00000010 << (pairIndex * 2)
			}
		}

		packed = append(packed, packedByte)
	}

	return packed, nil
}

type PointNBytes struct {
	Prefix []byte `json:"prefix,omitempty"`
	Value  []byte `json:"value"`
}

func (p *PointNBytes) FromBytes(data []byte, prefSize int) error {
	if prefSize != 0 {
		p.Prefix = data[0:prefSize]
	}

	p.Value = data[prefSize:]

	return nil
}

func (p *PointNBytes) ToBytes() ([]byte, error) {
	return append(p.Prefix, p.Value...), nil
}

func (p *PointNBytes) String() string {
	output := ""
	if len(p.Prefix) > 0 {
		output += fmt.Sprintf("Prefix: 0x % X\n", p.Prefix)
	}

	return output + fmt.Sprintf("Value : 0x % X", p.Value)
}

func newPointNBytes() *PointNBytes {
	return &PointNBytes{}
}

type PointNBytesFlags struct {
	Prefix []byte     `json:"prefix,omitempty"`
	Value  []byte     `json:"value"`
	Flags  PointFlags `json:"flags"`
}

func (p *PointNBytesFlags) FromBytes(data []byte, prefSize int) error {
	if prefSize != 0 {
		p.Prefix = data[0:prefSize]
	}

	err := p.Flags.FromByte(data[prefSize])
	if err != nil {
		return fmt.Errorf("couldn't decode flags byte: 0x % X, err: %w", data[prefSize], err)
	}

	p.Value = data[prefSize+1:]

	return nil
}

func (p *PointNBytesFlags) ToBytes() ([]byte, error) {
	var output []byte

	output = append(output, p.Prefix...)
	output = append(output, p.Flags.ToByte())

	return append(output, p.Value...), nil
}

func (p *PointNBytesFlags) String() string {
	output := ""
	if len(p.Prefix) > 0 {
		output += fmt.Sprintf("Prefix: 0x % X\n", p.Prefix)
	}

	output += fmt.Sprintf("Value : 0x % X\n", p.Value)
	output += p.Flags.String()

	return output
}

func newPointNBytesFlags() *PointNBytesFlags {
	return &PointNBytesFlags{}
}

type PointNBytesFlagsAbsTime struct {
	Prefix       []byte     `json:"prefix,omitempty"`
	Value        []byte     `json:"value"`
	Flags        PointFlags `json:"flags"`
	AbsoluteTime time.Time  `json:"absolute_time"`
}

func (p *PointNBytesFlagsAbsTime) FromBytes(data []byte, prefSize int) error {
	if len(data) != 8+prefSize {
		return fmt.Errorf("need %d bytes, got %d", 8+prefSize, len(data))
	}

	if prefSize != 0 {
		p.Prefix = data[0:prefSize]
	}

	err := p.Flags.FromByte(data[prefSize])
	if err != nil {
		return fmt.Errorf("couldn't decode flags byte: 0x % X, err: %w", data[prefSize], err)
	}

	absTime, err := BytesToDNP3TimeAbsolute(data[prefSize+1 : prefSize+7])
	if err != nil {
		return fmt.Errorf("couldn't decode absolute timestamp: %w", err)
	}

	p.AbsoluteTime = absTime
	p.Value = data[prefSize+7:]

	return nil
}

func (p *PointNBytesFlagsAbsTime) ToBytes() ([]byte, error) {
	var output []byte

	output = append(output, p.Prefix...)
	output = append(output, p.Flags.ToByte())

	timeBytes, err := DNP3TimeAbsoluteToBytes(p.AbsoluteTime)
	if err != nil {
		return nil, fmt.Errorf("failed to encode timestamp: %w", err)
	}

	output = append(output, timeBytes...)
	output = append(output, p.Value...)

	return output, nil
}

func (p *PointNBytesFlagsAbsTime) String() string {
	output := ""
	if len(p.Prefix) > 0 {
		output += fmt.Sprintf("Prefix: 0x % X\n", p.Prefix)
	}

	output += fmt.Sprintf("Value: 0x % X\n", p.Value)
	output += p.Flags.String()
	output += fmt.Sprintf("\nTimestamp: %s", p.AbsoluteTime.UTC())

	return output
}

func newPointNBytesFlagsAbsTime() *PointNBytesFlagsAbsTime {
	return &PointNBytesFlagsAbsTime{}
}

type PointNBytesAbsTime struct {
	Prefix       []byte    `json:"prefix,omitempty"`
	Value        []byte    `json:"value"`
	AbsoluteTime time.Time `json:"absolute_time"`
}

func (p *PointNBytesAbsTime) FromBytes(data []byte, prefSize int) error {
	if len(data) != 7+prefSize {
		return fmt.Errorf("need %d bytes, got %d", 8+prefSize, len(data))
	}

	if prefSize != 0 {
		p.Prefix = data[0:prefSize]
	}

	p.Value = data[prefSize : len(data)-6]

	absTime, err := BytesToDNP3TimeAbsolute(data[len(data)-6:])
	if err != nil {
		return fmt.Errorf("couldn't decode absolute timestamp: %w", err)
	}

	p.AbsoluteTime = absTime

	return nil
}

func (p *PointNBytesAbsTime) ToBytes() ([]byte, error) {
	var output []byte

	output = append(output, p.Prefix...)
	output = append(output, p.Value...)

	timeBytes, err := DNP3TimeAbsoluteToBytes(p.AbsoluteTime)
	if err != nil {
		return nil, fmt.Errorf("failed to encode timestamp: %w", err)
	}

	return append(output, timeBytes...), nil
}

func (p *PointNBytesAbsTime) String() string {
	output := ""
	if len(p.Prefix) > 0 {
		output += fmt.Sprintf("Prefix: 0x % X\n", p.Prefix)
	}

	output += fmt.Sprintf("Value: 0x % X", p.Value)
	output += fmt.Sprintf("\nTimestamp: %s", p.AbsoluteTime.UTC())

	return output
}

func newPointNBytesAbsTime() *PointNBytesAbsTime {
	return &PointNBytesAbsTime{}
}

type PointNBytesRelTime struct {
	Prefix       []byte       `json:"prefix,omitempty"`
	Value        []byte       `json:"value"`
	RelativeTime RelativeTime `json:"relative_time"`
}

func (p *PointNBytesRelTime) FromBytes(data []byte, prefSize int) error {
	if len(data) < 3+prefSize {
		return fmt.Errorf("need %d bytes, got %d", 3+prefSize, len(data))
	}

	if prefSize != 0 {
		p.Prefix = data[0:prefSize]
	}

	p.Value = data[prefSize : len(data)-2]

	relativeTime, err := BytesToDNP3TimeRelative(data[len(data)-2:])
	if err != nil {
		return fmt.Errorf("couldn't decode relative timestamp: %w", err)
	}

	p.RelativeTime = relativeTime

	return nil
}

func (p *PointNBytesRelTime) ToBytes() ([]byte, error) {
	var output []byte

	output = append(output, p.Prefix...)
	output = append(output, p.Value...)

	timeBytes, err := DNP3TimeRelativeToBytes(p.RelativeTime)
	if err != nil {
		return nil, fmt.Errorf("failed to encode relative timestamp: %w", err)
	}

	return append(output, timeBytes...), nil
}

func (p *PointNBytesRelTime) String() string {
	output := ""
	if len(p.Prefix) > 0 {
		output += fmt.Sprintf("Prefix: 0x % X\n", p.Prefix)
	}

	output += fmt.Sprintf("Value: 0x % X", p.Value)
	output += fmt.Sprintf("\nTimestamp offset: %s", p.RelativeTime)

	return output
}

func newPointNBytesRelTime() *PointNBytesRelTime {
	return &PointNBytesRelTime{}
}

type PointAbsTime struct {
	Prefix       []byte    `json:"prefix,omitempty"`
	AbsoluteTime time.Time `json:"absolute_time"`
}

func (p *PointAbsTime) FromBytes(data []byte, prefSize int) error {
	if len(data) != 6+prefSize {
		return fmt.Errorf("need %d bytes, got %d", 6+prefSize, len(data))
	}

	if prefSize != 0 {
		p.Prefix = data[0:prefSize]
	}

	absTime, err := BytesToDNP3TimeAbsolute(data[0:6])
	if err != nil {
		return fmt.Errorf("couldn't decode absolute timestamp: %w", err)
	}

	p.AbsoluteTime = absTime

	return nil
}

func (p *PointAbsTime) ToBytes() ([]byte, error) {
	timeBytes, err := DNP3TimeAbsoluteToBytes(p.AbsoluteTime)
	if err != nil {
		return nil, fmt.Errorf("failed to encode timestamp: %w", err)
	}

	return append(p.Prefix, timeBytes...), nil
}

func (p *PointAbsTime) String() string {
	output := ""
	if len(p.Prefix) > 0 {
		output += fmt.Sprintf("Prefix: 0x % X\n", p.Prefix)
	}

	return output + fmt.Sprintf("Timestamp: %s", p.AbsoluteTime.UTC())
}

func newPointAbsTime() *PointAbsTime {
	return &PointAbsTime{}
}

type PointRelTime struct {
	Prefix       []byte       `json:"prefix,omitempty"`
	RelativeTime RelativeTime `json:"relative_time"`
}

func (p *PointRelTime) FromBytes(data []byte, prefSize int) error {
	if len(data) < 2+prefSize {
		return fmt.Errorf("need %d bytes, got %d", prefSize+2, len(data))
	}

	if prefSize != 0 {
		p.Prefix = data[0:prefSize]
	}

	relativeTime, err := BytesToDNP3TimeRelative(data[prefSize : prefSize+2])
	if err != nil {
		return fmt.Errorf("couldn't decode relative timestamp: %w", err)
	}

	p.RelativeTime = relativeTime

	return nil
}

func (p *PointRelTime) ToBytes() ([]byte, error) {
	timeBytes, err := DNP3TimeRelativeToBytes(p.RelativeTime)
	if err != nil {
		return nil, fmt.Errorf("failed to encode relative timestamp: %w", err)
	}

	return append(p.Prefix, timeBytes...), nil
}

func (p *PointRelTime) String() string {
	output := ""
	if len(p.Prefix) > 0 {
		output += fmt.Sprintf("Prefix: 0x % X\n", p.Prefix)
	}

	return output + fmt.Sprintf("Timestamp offset: %s", p.RelativeTime)
}

func newPointRelTime() *PointRelTime {
	return &PointRelTime{}
}

func makeGenericConstructor[T Point](
	newPoint func() T,
	width int,
) PointsConstructor {
	return func(data []byte, num, prefSize int) ([]Point, int, error) {
		return newPointsGeneric[T](newPoint, data, width, num, prefSize)
	}
}

func newPointsGeneric[T Point](
	newPoint func() T,
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

	pointsOut := make([]Point, 0, num)

	for pointIndex := range num {
		point := newPoint()

		pointDataStart := pointIndex * (width + prefSize)
		pointDataEnd := (pointIndex + 1) * (width + prefSize)
		pointData := data[pointDataStart:pointDataEnd]

		err := point.FromBytes(pointData, prefSize)
		if err != nil {
			return pointsOut, size, fmt.Errorf("could not decode point: 0x % X, err: %w", data, err)
		}

		pointsOut = append(pointsOut, point)
	}

	return pointsOut, size, nil
}

func packPointsBytes(points []Point) ([]byte, error) {
	var encoded []byte

	for _, point := range points {
		b, err := point.ToBytes()
		if err != nil {
			return nil, fmt.Errorf("error packing points: %w", err)
		}

		encoded = append(encoded, b...)
	}

	return encoded, nil
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
	Reserved       bool `json:"reserved"` // should be 0
	PointValue     bool `json:"point_value"`
	ReferenceCheck bool `json:"reference_check"`
	OverRange      bool `json:"over_range"`
	LocalForce     bool `json:"local_force"`
	RemoteForce    bool `json:"remote_force"`
	CommFail       bool `json:"comm_fail"`
	Restart        bool `json:"restart"`
	Online         bool `json:"online"`
}

func (f *PointFlags) FromByte(data byte) error {
	f.Reserved = data&0b10000000 != 0
	if f.Reserved {
		return errors.New("reserved bit must be 0")
	}

	f.ReferenceCheck = data&0b01000000 != 0
	f.OverRange = data&0b00100000 != 0
	f.LocalForce = data&0b00010000 != 0
	f.RemoteForce = data&0b00001000 != 0
	f.CommFail = data&0b00000100 != 0
	f.Restart = data&0b00000010 != 0
	f.Online = data&0b00000001 != 0

	return nil
}

func (f *PointFlags) ToByte() byte {
	var flagByte byte
	if f.Reserved {
		flagByte |= 0b10000000
	}

	if f.ReferenceCheck {
		flagByte |= 0b01000000
	}

	if f.OverRange {
		flagByte |= 0b00100000
	}

	if f.LocalForce {
		flagByte |= 0b00010000
	}

	if f.RemoteForce {
		flagByte |= 0b00001000
	}

	if f.CommFail {
		flagByte |= 0b00000100
	}

	if f.Restart {
		flagByte |= 0b00000010
	}

	if f.Online {
		flagByte |= 0b00000001
	}

	return flagByte
}

func (f *PointFlags) String() string {
	return fmt.Sprintf(`Flags:
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
