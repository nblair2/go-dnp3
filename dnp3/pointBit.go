package dnp3

import (
	"errors"
	"fmt"
	"strings"
)

// PointBit is a 1-bit Point implementation. It handles both packed binary
// (8 points per byte, no flags) and flags variants (1 byte per point with
// bit 7 as value and bits 0-6 as flags). The hasFlags field controls which
// mode is active, set at construction time per DNP3 group/variation.
type PointBit struct {
	Index     *int `json:"index,omitempty"`
	indexSize int
	Size      int `json:"size,omitempty"`
	sizeSize  int
	Value     bool        `json:"value"`
	Flags     *PointFlags `json:"flags,omitempty"`
	hasFlags  bool
}

func (p *PointBit) DataType() PointDataType { return PointDataTypeBit }

// should not be used directly for packed (non-flags) points.
func (p *PointBit) FromBytes(data []byte, prefSize int) error {
	if p.hasFlags {
		return p.fromBytesFlags(data, prefSize)
	}

	if len(data) > 1 {
		return errors.New("can't construct 1 bit point from multiple bytes")
	} else if prefSize != 0 {
		return errors.New("can't have a prefix on 1 bit packed points")
	}

	p.Value = data[0]&0b00000001 != 0

	return nil
}

// should not be used directly for packed (non-flags) points.
func (p *PointBit) ToBytes() ([]byte, error) {
	if p.hasFlags {
		return p.toBytesFlags()
	}

	if p.Value {
		return []byte{0b00000001}, nil
	}

	return []byte{0b00000000}, nil
}

func (p *PointBit) String() string {
	var parts []string

	if p.indexSize > 0 {
		parts = append(parts, fmt.Sprintf("Index: %d", *p.Index))
	}

	parts = append(parts, fmt.Sprintf("Value: %t", p.Value))

	if p.Flags != nil {
		parts = append(parts, p.Flags.String())
	}

	return strings.Join(parts, "\n")
}

func (p *PointBit) Fields() PointFields {
	return PointFields{
		Index: p.indexSize > 0,
		Size:  p.sizeSize > 0,
		Flags: p.hasFlags,
		Value: true,
	}
}

// --- Get/Set methods ---

func (p *PointBit) GetIndex() (int, error) {
	if p.indexSize == 0 {
		return 0, ErrNoIndex
	}

	return *p.Index, nil
}

func (p *PointBit) SetIndex(v int) error {
	if !p.hasFlags {
		return ErrNoIndex
	}

	return setIndex(&p.Index, &p.indexSize, v)
}

func (p *PointBit) GetFlags() (PointFlags, error) {
	if p.Flags == nil {
		return PointFlags{}, ErrNoFlags
	}

	return *p.Flags, nil
}

func (p *PointBit) SetFlags(f PointFlags) error {
	if !p.hasFlags {
		return ErrNoFlags
	}

	p.Flags = &f

	return nil
}

func (p *PointBit) GetAbsTime() (AbsoluteTime, error) { return AbsoluteTime{}, ErrNoAbsTime }
func (p *PointBit) SetAbsTime(AbsoluteTime) error     { return ErrNoAbsTime }
func (p *PointBit) GetRelTime() (RelativeTime, error) { return 0, ErrNoRelTime }
func (p *PointBit) SetRelTime(RelativeTime) error     { return ErrNoRelTime }

func (p *PointBit) GetValue() any { return p.Value }

func (p *PointBit) SetValue(v any) error {
	val, ok := v.(bool)
	if !ok {
		return fmt.Errorf("PointBit value must be bool, got %T", v)
	}

	p.Value = val

	return nil
}

// --- unexported helpers ---

func (p *PointBit) fromBytesFlags(data []byte, prefSize int) error {
	if len(data) > 1+prefSize {
		return errors.New("can't construct 1 bit point with flags from multiple bytes")
	}

	if prefSize > 0 {
		index, err := prefixToInt(data[0:prefSize])
		if err != nil {
			return fmt.Errorf("could not decode index prefix: %w", err)
		}

		p.Index = &index
		p.indexSize = prefSize
	}

	p.Value = data[prefSize]&0b10000000 != 0
	flagsByte := data[prefSize] & 0b01111111

	flags := PointFlags{}

	err := flags.FromByte(flagsByte)
	if err != nil {
		return fmt.Errorf("couldn't decode flags byte: 0x%02X, err: %w", flagsByte, err)
	}

	p.Flags = &flags

	return nil
}

func (p *PointBit) toBytesFlags() ([]byte, error) {
	var output []byte

	if p.indexSize > 0 {
		indexBytes, err := intToPrefix(*p.Index, p.indexSize)
		if err != nil {
			return nil, fmt.Errorf("failed to encode index: %w", err)
		}

		output = append(output, indexBytes...)
	}

	flagsByte := byte(0)
	if p.Flags != nil {
		flagsByte = p.Flags.ToByte()
	}

	if p.Value {
		flagsByte |= 0b10000000
	}

	return append(output, flagsByte), nil
}

// --- Constructor and packer functions ---

func newPointsBit(data []byte, num, prefSize int, _ PointPrefixCode) ([]Point, int, error) {
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
		point := &PointBit{Value: (sourceByte & mask) != 0}
		pointsOut = append(pointsOut, point)
	}

	size := (num + 7) / 8 // round up to nearest byte

	return pointsOut, size, nil
}

func packerPointsBit(points []Point) ([]byte, error) {
	var packed []byte

	for pointOffset := 0; pointOffset < len(points); pointOffset += 8 {
		var packedByte byte

		for bitOffset := 0; bitOffset < 8 && pointOffset+bitOffset < len(points); bitOffset++ {
			point, ok := points[pointOffset+bitOffset].(*PointBit)
			if !ok {
				return packed, fmt.Errorf("element %d is not *PointBit, got %T",
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

func newPointsBitFlags(data []byte, num, prefSize int, _ PointPrefixCode) ([]Point, int, error) {
	if num > len(data) {
		return nil, 0, fmt.Errorf("not enough bytes for %d 1-bit points with flags", num)
	}

	pointsOut := make([]Point, 0, num)

	for pointIndex := range num {
		point := &PointBit{hasFlags: true}

		err := point.FromBytes([]byte{data[pointIndex]}, prefSize)
		if err != nil {
			return pointsOut, num, fmt.Errorf("could not decode point: 0x % X, err: %w",
				data[pointIndex], err)
		}

		pointsOut = append(pointsOut, point)
	}

	return pointsOut, num, nil
}
