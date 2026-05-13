package dnp3

import (
	"errors"
	"fmt"
)

// Point2Bits is a 2-bit Point implementation for packed double-bit binary
// inputs. Points are bit-packed with 4 points per byte.
type Point2Bits struct {
	Value [2]bool `json:"value"`
}

func (p *Point2Bits) DataType() PointDataType { return PointDataType2Bits }

// DecodeFromBytes should not be used directly.
func (p *Point2Bits) DecodeFromBytes(data []byte, prefSize int) error {
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

// SerializeTo should not be used directly.
func (p *Point2Bits) SerializeTo() ([]byte, error) {
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

func (p *Point2Bits) Fields() PointFields {
	return PointFields{Value: true}
}

// --- Get/Set methods ---

func (p *Point2Bits) GetIndex() (int, error)            { return 0, ErrNoIndex }
func (p *Point2Bits) GetFlags() (PointFlags, error)     { return PointFlags{}, ErrNoFlags }
func (p *Point2Bits) GetAbsTime() (AbsoluteTime, error) { return AbsoluteTime{}, ErrNoAbsTime }
func (p *Point2Bits) GetRelTime() (RelativeTime, error) { return 0, ErrNoRelTime }
func (p *Point2Bits) GetValue() any                     { return p.Value }
func (p *Point2Bits) SetIndex(int) error                { return ErrNoIndex }
func (p *Point2Bits) SetFlags(PointFlags) error         { return ErrNoFlags }
func (p *Point2Bits) SetAbsTime(AbsoluteTime) error     { return ErrNoAbsTime }
func (p *Point2Bits) SetRelTime(RelativeTime) error     { return ErrNoRelTime }

func (p *Point2Bits) SetValue(v any) error {
	val, ok := v.([2]bool)
	if !ok {
		return fmt.Errorf("Point2Bits value must be [2]bool, got %T", v)
	}

	p.Value = val

	return nil
}

// --- Constructor and packer functions ---

func newPoints2Bits(data []byte, num, prefSize int, _ PointPrefixCode) ([]Point, int, error) {
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
