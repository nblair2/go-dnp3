package dnp3

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// pointField identifies a field in a PointBytes data layout.
type pointField int

const (
	pointFieldFlags   pointField = iota // 1 byte
	pointFieldAbsTime                   // 6 bytes
	pointFieldRelTime                   // 2 bytes
	pointFieldValue                     // variable width
)

var pointFieldWidths = map[pointField]int{
	pointFieldFlags:   1,
	pointFieldAbsTime: 6,
	pointFieldRelTime: 2,
}

// pointBytesLayout describes the field order for a PointBytes instance.
// The fields slice determines the parse/encode order. Value width is
// computed dynamically as the remaining bytes after all fixed-width fields.
type pointBytesLayout struct {
	fields []pointField
}

// suffixWidthAfter returns the total byte width of all fixed-width
// fields that appear after position idx in the field list.
func (l *pointBytesLayout) suffixWidthAfter(idx int) int {
	total := 0

	for _, field := range l.fields[idx+1:] {
		if width, ok := pointFieldWidths[field]; ok {
			total += width
		}
	}

	return total
}

// hasField reports whether the layout includes the given field type.
func (l *pointBytesLayout) hasField(field pointField) bool {
	return slices.Contains(l.fields, field)
}

// PointBytes is a general-purpose Point implementation. Optional fields
// (Flags, AbsoluteTime, RelativeTime) are nil when absent. The presence
// of each field is determined by the layout, which is set at construction
// time per DNP3 group/variation.
type PointBytes struct {
	Index             *int `json:"index,omitempty"`
	indexSize         int
	Size              int `json:"size,omitempty"`
	sizeSize          int
	Flags             *PointFlags   `json:"flags,omitempty"`
	Value             []byte        `json:"value,omitempty"`
	AbsoluteTime      *AbsoluteTime `json:"absolute_time,omitempty"`
	RelativeTime      *RelativeTime `json:"relative_time,omitempty"`
	layout            pointBytesLayout
	expectedValueSize int
}

func (p *PointBytes) DataType() PointDataType { return PointDataTypeBytes }

func (p *PointBytes) FromBytes(data []byte, prefSize int) error {
	offset := 0

	if prefSize > 0 {
		prefixValue, err := prefixToInt(data[0:prefSize])
		if err != nil {
			return fmt.Errorf("could not decode prefix: %w", err)
		}

		p.setPrefixValue(prefixValue, prefSize)
		offset = prefSize
	}

	remaining := data[offset:]

	for fieldIdx, field := range p.layout.fields {
		var err error

		remaining, err = p.parseField(field, fieldIdx, remaining)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PointBytes) ToBytes() ([]byte, error) {
	var output []byte

	indexBytes, err := p.prefixBytes()
	if err != nil {
		return nil, err
	}

	output = append(output, indexBytes...)

	for _, field := range p.layout.fields {
		encoded, err := p.encodeField(field)
		if err != nil {
			return nil, err
		}

		output = append(output, encoded...)
	}

	return output, nil
}

func (p *PointBytes) String() string {
	var parts []string

	if p.indexSize > 0 {
		parts = append(parts, fmt.Sprintf("Index: %d", *p.Index))
	}

	if p.sizeSize > 0 {
		parts = append(parts, fmt.Sprintf("Size: %d", p.Size))
	}

	if p.Flags != nil {
		parts = append(parts, p.Flags.String())
	}

	if len(p.Value) > 0 {
		parts = append(parts, fmt.Sprintf("Value: 0x % X", p.Value))
	}

	if p.AbsoluteTime != nil {
		parts = append(parts, fmt.Sprintf("Timestamp: %s", p.AbsoluteTime.Time().UTC()))
	}

	if p.RelativeTime != nil {
		parts = append(parts, fmt.Sprintf("Timestamp offset: %s", *p.RelativeTime))
	}

	return strings.Join(parts, "\n")
}

// --- Get/Set methods ---

func (p *PointBytes) GetIndex() (int, error) {
	if p.indexSize == 0 {
		return 0, ErrNoIndex
	}

	return *p.Index, nil
}

func (p *PointBytes) SetIndex(value int) error {
	return setIndex(&p.Index, &p.indexSize, value)
}

func (p *PointBytes) GetFlags() (PointFlags, error) {
	if p.Flags == nil {
		return PointFlags{}, ErrNoFlags
	}

	return *p.Flags, nil
}

func (p *PointBytes) SetFlags(flags PointFlags) error {
	if !p.layout.hasField(pointFieldFlags) {
		return ErrNoFlags
	}

	p.Flags = &flags

	return nil
}

func (p *PointBytes) GetAbsTime() (AbsoluteTime, error) {
	if p.AbsoluteTime == nil {
		return AbsoluteTime{}, ErrNoAbsTime
	}

	return *p.AbsoluteTime, nil
}

func (p *PointBytes) SetAbsTime(absTime AbsoluteTime) error {
	if !p.layout.hasField(pointFieldAbsTime) {
		return ErrNoAbsTime
	}

	p.AbsoluteTime = &absTime

	return nil
}

func (p *PointBytes) GetRelTime() (RelativeTime, error) {
	if p.RelativeTime == nil {
		return 0, ErrNoRelTime
	}

	return *p.RelativeTime, nil
}

func (p *PointBytes) SetRelTime(relTime RelativeTime) error {
	if !p.layout.hasField(pointFieldRelTime) {
		return ErrNoRelTime
	}

	p.RelativeTime = &relTime

	return nil
}

func (p *PointBytes) GetValue() any {
	if len(p.Value) > 0 {
		if p.expectedValueSize > 0 && len(p.Value) < p.expectedValueSize {
			return paddedBytes(p.Value, p.expectedValueSize)
		}

		return p.Value
	}

	// For time-only points, return the time as the value.
	if p.AbsoluteTime != nil {
		return *p.AbsoluteTime
	}

	if p.RelativeTime != nil {
		return *p.RelativeTime
	}

	return nil
}

func (p *PointBytes) SetValue(value any) error {
	val, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("PointBytes value must be byte slice, got %T", value)
	}

	if p.expectedValueSize > 0 && len(val) != p.expectedValueSize {
		return fmt.Errorf(
			"PointBytes value incorrect size: expected %d, got %d (use ExpectedValueSize())",
			p.expectedValueSize,
			len(val),
		)
	}

	p.Value = val

	return nil
}

func (p *PointBytes) Fields() PointFields {
	return PointFields{
		Index:        p.indexSize > 0,
		Size:         p.sizeSize > 0,
		Flags:        p.layout.hasField(pointFieldFlags),
		Value:        p.layout.hasField(pointFieldValue),
		AbsoluteTime: p.layout.hasField(pointFieldAbsTime),
		RelativeTime: p.layout.hasField(pointFieldRelTime),
	}
}

// ExpectedValueSize returns the expected byte size of the Value field,
// or 0 if there is no fixed expected size.
func (p *PointBytes) ExpectedValueSize() int {
	return p.expectedValueSize
}

// paddedBytes returns a byte slice padded with zeros to the given size.
func paddedBytes(data []byte, size int) []byte {
	if len(data) >= size {
		return data
	}

	padded := make([]byte, size)
	copy(padded, data)

	return padded
}

// setPrefixValue routes a decoded prefix value to either the Index or
// Size field based on the prefix type pre-set by the constructor.
func (p *PointBytes) setPrefixValue(value, prefSize int) {
	if p.sizeSize > 0 {
		p.Size = value
		p.sizeSize = prefSize
	} else {
		p.Index = &value
		p.indexSize = prefSize
	}
}

// prefixBytes encodes whichever prefix (index or size) is active.
func (p *PointBytes) prefixBytes() ([]byte, error) {
	if p.indexSize > 0 {
		return intToPrefix(*p.Index, p.indexSize)
	}

	if p.sizeSize > 0 {
		return intToPrefix(p.Size, p.sizeSize)
	}

	return nil, nil
}

// parseField decodes a single field from remaining and returns
// the unconsumed tail.
//
//nolint:cyclop,funlen // straightforward switches
func (p *PointBytes) parseField(
	field pointField,
	fieldIdx int,
	remaining []byte,
) ([]byte, error) {
	switch field {
	case pointFieldFlags:
		if len(remaining) < 1 {
			return nil, fmt.Errorf("not enough data for flags: need 1, have %d", len(remaining))
		}

		flags := PointFlags{}

		err := flags.FromByte(remaining[0])
		if err != nil {
			return nil, fmt.Errorf("couldn't decode flags byte: 0x%02X, err: %w", remaining[0], err)
		}

		p.Flags = &flags

		return remaining[1:], nil

	case pointFieldAbsTime:
		if len(remaining) < 6 {
			return nil, fmt.Errorf(
				"not enough data for absolute time: need 6, have %d",
				len(remaining),
			)
		}

		absTime, err := BytesToDNP3TimeAbsolute(remaining[:6])
		if err != nil {
			return nil, fmt.Errorf("couldn't decode absolute timestamp: %w", err)
		}

		p.AbsoluteTime = &absTime

		return remaining[6:], nil

	case pointFieldRelTime:
		if len(remaining) < 2 {
			return nil, fmt.Errorf(
				"not enough data for relative time: need 2, have %d",
				len(remaining),
			)
		}

		relTime, err := BytesToDNP3TimeRelative(remaining[:2])
		if err != nil {
			return nil, fmt.Errorf("couldn't decode relative timestamp: %w", err)
		}

		p.RelativeTime = &relTime

		return remaining[2:], nil

	case pointFieldValue:
		suffixWidth := p.layout.suffixWidthAfter(fieldIdx)
		valueWidth := len(remaining) - suffixWidth

		if valueWidth < 0 {
			return nil, fmt.Errorf(
				"not enough data for value: have %d, need at least %d for trailing fields",
				len(remaining), suffixWidth,
			)
		}

		p.Value = remaining[:valueWidth]

		return remaining[valueWidth:], nil

	default:
		return remaining, nil
	}
}

// encodeField encodes a single field into bytes.
func (p *PointBytes) encodeField(field pointField) ([]byte, error) {
	switch field {
	case pointFieldFlags:
		if p.Flags == nil {
			return nil, errors.New("flags field is required by layout but is nil")
		}

		return []byte{p.Flags.ToByte()}, nil

	case pointFieldAbsTime:
		if p.AbsoluteTime == nil {
			return nil, errors.New("absolute time field is required by layout but is nil")
		}

		return DNP3TimeAbsoluteToBytes(*p.AbsoluteTime)

	case pointFieldRelTime:
		if p.RelativeTime == nil {
			return nil, errors.New("relative time field is required by layout but is nil")
		}

		return DNP3TimeRelativeToBytes(*p.RelativeTime)

	case pointFieldValue:
		return p.encodeFieldValue(), nil

	default:
		return nil, nil
	}
}

// encodeFieldValue returns the value field with padding applied if needed.
func (p *PointBytes) encodeFieldValue() []byte {
	if p.expectedValueSize > 0 && len(p.Value) < p.expectedValueSize {
		return paddedBytes(p.Value, p.expectedValueSize)
	}

	return p.Value
}

// --- Predefined layouts ---

var (
	layoutValue = pointBytesLayout{
		fields: []pointField{pointFieldValue},
	}
	layoutFlags = pointBytesLayout{
		fields: []pointField{pointFieldFlags, pointFieldValue},
	}
	layoutFlagsAbsTime = pointBytesLayout{
		fields: []pointField{pointFieldFlags, pointFieldAbsTime, pointFieldValue},
	}
	layoutValueAbsTime = pointBytesLayout{
		fields: []pointField{pointFieldValue, pointFieldAbsTime},
	}
	layoutValueRelTime = pointBytesLayout{
		fields: []pointField{pointFieldValue, pointFieldRelTime},
	}
	layoutAbsTime = pointBytesLayout{
		fields: []pointField{pointFieldAbsTime},
	}
	layoutRelTime = pointBytesLayout{
		fields: []pointField{pointFieldRelTime},
	}
)

// --- Constructor helpers ---

func newPointBytesWithLayout(layout pointBytesLayout, width int) func() *PointBytes {
	fixedFieldsWidth := 0
	if layout.hasField(pointFieldFlags) {
		fixedFieldsWidth += pointFieldWidths[pointFieldFlags]
	}

	if layout.hasField(pointFieldAbsTime) {
		fixedFieldsWidth += pointFieldWidths[pointFieldAbsTime]
	}

	if layout.hasField(pointFieldRelTime) {
		fixedFieldsWidth += pointFieldWidths[pointFieldRelTime]
	}

	calculatedExpectedValueSize := 0
	if layout.hasField(pointFieldValue) {
		calculatedExpectedValueSize = width - fixedFieldsWidth
		if calculatedExpectedValueSize < 0 {
			// This should not happen if width is correctly calculated in objectType.go
			// but as a safeguard, we can set it to 0 or return an error.
			// For now, let's assume it means no specific value size is expected.
			calculatedExpectedValueSize = 0
		}
	}

	return func() *PointBytes {
		return &PointBytes{
			layout:            layout,
			expectedValueSize: calculatedExpectedValueSize,
		}
	}
}

// makeBytesConstructor creates a PointsConstructor for PointBytes with
// the given layout and total data width (excluding prefix).
func makeBytesConstructor(layout pointBytesLayout, width int) PointsConstructor {
	newPoint := newPointBytesWithLayout(layout, width)

	return func(data []byte, num, prefSize int, prefCode PointPrefixCode) ([]Point, int, error) {
		return newPointsBytesGeneric(newPoint, data, width, num, prefSize, prefCode)
	}
}

func newPointsBytesGeneric(
	newPoint func() *PointBytes,
	data []byte,
	width, num, prefSize int,
	prefCode PointPrefixCode,
) ([]Point, int, error) {
	size := num * (prefSize + width)
	if size > len(data) {
		return nil, 0, fmt.Errorf(
			"not enough bytes for %d %d-byte points with %d-byte prefix",
			num, width, prefSize,
		)
	}

	pointsOut := make([]Point, 0, num)

	for pointIndex := range num {
		point := newPoint()

		// Tell the point whether its prefix is an index or a size
		// BEFORE calling FromBytes, so setPrefixValue routes correctly.
		if slices.Contains([]PointPrefixCode{Size1Octet, Size2Octet, Size4Octet}, prefCode) {
			point.sizeSize = prefSize
		}

		pointDataStart := pointIndex * (width + prefSize)
		pointDataEnd := (pointIndex + 1) * (width + prefSize)
		pointData := data[pointDataStart:pointDataEnd]

		err := point.FromBytes(pointData, prefSize)
		if err != nil {
			return pointsOut, size, fmt.Errorf(
				"could not decode point: 0x % X, err: %w",
				pointData,
				err,
			)
		}

		pointsOut = append(pointsOut, point)
	}

	return pointsOut, size, nil
}
