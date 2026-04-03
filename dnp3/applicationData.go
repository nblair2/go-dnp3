package dnp3

import (
	"errors"
	"fmt"
	"strings"
)

// ApplicationData holds an array of Data objects.
type ApplicationData struct {
	Objects []DataObject `json:"objects"`
	// in case we get in to trouble unrolling the objects just store the rest
	// of the data in here. Or can use this to set all data like "raw"
	extra []byte
}

func (ad *ApplicationData) FromBytes(data []byte) error {
	ad.Objects = nil // in case there was already stuff here

	for readOffset := 0; readOffset < len(data); {
		var object DataObject

		err := object.FromBytes(data[readOffset:])
		if err != nil {
			ad.extra = data[readOffset:]

			return fmt.Errorf("could not decode object: 0x % X, err: %w",
				data[readOffset:], err)
		}

		ad.Objects = append(ad.Objects, object)
		readOffset += object.SizeOf()
	}

	return nil
}

func (ad *ApplicationData) ToBytes() ([]byte, error) {
	var encoded []byte

	for _, object := range ad.Objects {
		bytesOut, err := object.ToBytes()
		if err != nil {
			return encoded, fmt.Errorf("could not encode object: %w", err)
		}

		encoded = append(encoded, bytesOut...)
	}

	encoded = append(encoded, ad.extra...)

	return encoded, nil
}

func (ad *ApplicationData) String() string {
	output := ""
	header := "Data Objects:"
	headerAdded := false

	if len(ad.Objects) > 0 {
		output += header
		headerAdded = true

		var stringBuilder strings.Builder
		for _, obj := range ad.Objects {
			stringBuilder.WriteString("\n" + indent("- "+obj.String(), "\t"))
		}

		output += stringBuilder.String()
	}

	if len(ad.extra) > 0 {
		if !headerAdded {
			output += header
		}

		output += fmt.Sprintf("\n\t- Extra: 0x % X", ad.extra)
	}

	return output
}

func (ad *ApplicationData) HasExtra() bool {
	return len(ad.extra) > 0
}

type DataObject struct {
	Header    ObjectHeader `json:"header"`
	Points    []Point      `json:"points"`
	Extra     []byte       `json:"extra,omitempty"`
	totalSize int
	indexes   []int
}

func (do *DataObject) FromBytes(data []byte) error {
	err := do.Header.FromBytes(data)
	if err != nil {
		return fmt.Errorf("can't create Data Object Header: %w", err)
	}

	headSize := do.Header.SizeOf()
	do.totalSize = headSize

	if do.Header.objectType == nil || do.Header.objectType.Constructor == nil {
		do.Extra = data[headSize:]
		do.totalSize += len(do.Extra)

		return fmt.Errorf("unsupported group/variation: %d/%d",
			do.Header.Group, do.Header.Variation)
	}

	numPoints := do.Header.RangeField.NumObjects()
	if numPoints == 0 {
		return nil
	}

	var size int

	do.Points, size, err = do.Header.objectType.Constructor(
		data[headSize:],
		numPoints,
		do.Header.PointPrefixCode.GetPointPrefixSize(),
		do.Header.PointPrefixCode,
	)
	if err != nil {
		return fmt.Errorf("can't create points: %w", err)
	}

	do.totalSize += size

	err = do.updateIndexes()
	if err != nil {
		return fmt.Errorf("failed to update indexes: %w", err)
	}

	return nil
}

func (do *DataObject) ToBytes() ([]byte, error) {
	// TODO get this to be more elegant.
	var encoded []byte

	headerBytes, err := do.Header.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to encode object header: %w", err)
	}

	encoded = append(encoded, headerBytes...)

	if len(do.Points) > 0 {
		var packer PointsPacker
		if do.Header.objectType != nil {
			packer = do.Header.objectType.Packer
		} else if def, ok := objectTypes[groupVariation{do.Header.Group, do.Header.Variation}]; ok {
			// Try to look it up if it wasn't set (e.g. manual construction)
			do.Header.objectType = def
			packer = def.Packer
		}

		if packer == nil {
			encoded = append(encoded, do.Extra...)

			return encoded, fmt.Errorf("no packer for Group %d, Var %d",
				do.Header.Group, do.Header.Variation)
		}

		packedPoints, err := packer(do.Points)
		if err != nil {
			encoded = append(encoded, do.Extra...)

			return encoded, fmt.Errorf("could not pack points: %w", err)
		}

		encoded = append(encoded, packedPoints...)
	}

	encoded = append(encoded, do.Extra...)

	return encoded, nil
}

func (do *DataObject) String() string {
	output := do.Header.String()

	if len(do.Points) == 0 {
		return output
	}

	output += "\n  Objects:"

	var stringBuilder strings.Builder

	for _, point := range do.Points {
		lines := strings.Split(point.String(), "\n")
		if len(lines) > 0 {
			lines[0] = "- " + lines[0]
			for i := 1; i < len(lines); i++ {
				lines[i] = "  " + lines[i]
			}

			stringBuilder.WriteString("\n" + indent(strings.Join(lines, "\n"), "\t"))
		}
	}

	output += stringBuilder.String()

	return output
}

func (do *DataObject) SizeOf() int {
	return do.totalSize
}

func (do *DataObject) Indexes() []int {
	return do.indexes
}

func (do *DataObject) updateIndexes() error {
	switch rangeField := do.Header.RangeField.(type) {
	case *StartStopRangeField:
		for i := rangeField.Start; i <= rangeField.Stop; i++ {
			do.indexes = append(do.indexes, int(i))
		}

		return nil
	case *AllRangeField:
		for i := range len(do.Points) {
			do.indexes = append(do.indexes, i)
		}

		return nil
	case *CountRangeField:
		return do.updateIndexesFromPrefix()
	default:
		return fmt.Errorf("unexpected range field type %T", do.Header.RangeField)
	}
}

// updateIndexesFromPrefix resolves point indexes for count-based range
// fields by inspecting the object header's PointPrefixCode.
func (do *DataObject) updateIndexesFromPrefix() error {
	switch do.Header.PointPrefixCode {
	case Index1Octet, Index2Octet, Index4Octet:
		for _, point := range do.Points {
			index, err := point.GetIndex()
			if err != nil {
				return fmt.Errorf("failed to get index from point: %w", err)
			}

			do.indexes = append(do.indexes, index)
		}

		return nil
	case NoPrefix:
		for i := range do.Points {
			do.indexes = append(do.indexes, i)
		}

		return nil
	case Size1Octet, Size2Octet, Size4Octet:
		return fmt.Errorf(
			"point prefix code %s does not determine indexes",
			do.Header.PointPrefixCode,
		)
	case Reserved:
		return errors.New("reserved point prefix code cannot be used to determine indexes")
	default:
		return fmt.Errorf("unexpected point prefix code %d", do.Header.PointPrefixCode)
	}
}
