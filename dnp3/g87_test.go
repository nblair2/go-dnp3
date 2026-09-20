package dnp3_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/nblair2/go-dnp3/v4/dnp3"
)

func TestGroup87Variation1RequestRoundTrip(t *testing.T) {
	t.Parallel()

	raw := []byte{
		0xc0, byte(dnp3.Write),
		0x57, 0x01, 0x5b, 0x01,
		0x0a, 0x00,
		0x01, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	request := mustDecodeRequest(t, raw)
	data := request.GetData()
	object := requireSingleObject(t, data.Objects)
	assertGroup87Header(t, object.Header)

	if len(object.Points) != 1 {
		t.Fatalf("Points: got %d, want 1", len(object.Points))
	}

	wantValue := []byte{0x01, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	assertPointBytes(t, object.Points[0], wantValue, "point 0")

	if len(object.Indexes()) != 0 {
		t.Fatalf("Indexes: got %v, want none", object.Indexes())
	}

	if data.HasExtra() {
		t.Fatalf("unexpected extra: 0x % X", data.GetExtra())
	}

	assertRequestRoundTrip(t, request, raw)
}

func TestGroup87Variation1VariableSizes(t *testing.T) {
	t.Parallel()

	raw := []byte{
		0xc0, byte(dnp3.Write),
		0x57, 0x01, 0x5b, 0x02,
		0x03, 0x00, 0xaa, 0xbb, 0xcc,
		0x01, 0x00, 0xdd,
		0x32, 0x01, 0x07, 0x01,
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
	}

	request := mustDecodeRequest(t, raw)

	objects := request.GetData().Objects
	if len(objects) != 2 {
		t.Fatalf("Objects: got %d, want 2", len(objects))
	}

	if objects[1].Header.Group != 50 || objects[1].Header.Variation != 1 {
		t.Fatalf(
			"following group/variation: got %d/%d, want 50/1",
			objects[1].Header.Group,
			objects[1].Header.Variation,
		)
	}

	wantValues := [][]byte{{0xaa, 0xbb, 0xcc}, {0xdd}}
	if len(objects[0].Points) != len(wantValues) {
		t.Fatalf("Points: got %d, want %d", len(objects[0].Points), len(wantValues))
	}

	for pointIndex, point := range objects[0].Points {
		assertPointBytes(t, point, wantValues[pointIndex], "variable-size point")
	}

	assertRequestRoundTrip(t, request, raw)
}

func mustDecodeRequest(t *testing.T, raw []byte) *dnp3.ApplicationRequest {
	t.Helper()

	request, err := dnp3.NewApplicationRequestFromBytes(raw)
	if err != nil {
		t.Fatalf("NewApplicationRequestFromBytes: %v", err)
	}

	return request
}

func requireSingleObject(t *testing.T, objects []dnp3.DataObject) dnp3.DataObject {
	t.Helper()

	if len(objects) != 1 {
		t.Fatalf("Objects: got %d, want 1", len(objects))
	}

	return objects[0]
}

func assertGroup87Header(t *testing.T, header dnp3.ObjectHeader) {
	t.Helper()

	if header.Group != 87 || header.Variation != 1 {
		t.Fatalf("group/variation: got %d/%d, want 87/1", header.Group, header.Variation)
	}

	if header.PointPrefixCode != dnp3.Size2Octet {
		t.Fatalf("PointPrefixCode: got %s, want %s", header.PointPrefixCode, dnp3.Size2Octet)
	}

	if header.RangeSpecCode != dnp3.Count1Variable {
		t.Fatalf("RangeSpecCode: got %s, want %s", header.RangeSpecCode, dnp3.Count1Variable)
	}
}

func assertPointBytes(t *testing.T, point dnp3.Point, wantValue []byte, label string) {
	t.Helper()

	pointBytes, ok := point.(*dnp3.PointBytes)
	if !ok {
		t.Fatalf("%s type: got %T, want *dnp3.PointBytes", label, point)
	}

	if pointBytes.Size != len(wantValue) {
		t.Fatalf("%s Size: got %d, want %d", label, pointBytes.Size, len(wantValue))
	}

	if !slices.Equal(pointBytes.Value, wantValue) {
		t.Fatalf("%s Value: got 0x % X, want 0x % X", label, pointBytes.Value, wantValue)
	}
}

func assertRequestRoundTrip(t *testing.T, request *dnp3.ApplicationRequest, raw []byte) {
	t.Helper()

	encoded, err := request.SerializeTo()
	if err != nil {
		t.Fatalf("SerializeTo: %v", err)
	}

	if !slices.Equal(encoded, raw) {
		t.Fatalf("round trip mismatch: got 0x % X, want 0x % X", encoded, raw)
	}
}

func TestGroup87Variation1RejectsMalformedValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		raw     []byte
		wantErr error
	}{
		{
			name:    "missing size prefix byte",
			raw:     []byte{0x57, 0x01, 0x5b, 0x01, 0x0a},
			wantErr: dnp3.ErrInsufficientData,
		},
		{
			name:    "declared value exceeds input",
			raw:     []byte{0x57, 0x01, 0x5b, 0x01, 0x0a, 0x00, 0x01},
			wantErr: dnp3.ErrInsufficientData,
		},
		{
			name: "second value exceeds input",
			raw: []byte{
				0x57, 0x01, 0x5b, 0x02,
				0x01, 0x00, 0xaa,
				0x02, 0x00, 0xbb,
			},
			wantErr: dnp3.ErrInsufficientData,
		},
		{
			name:    "indexless variable value",
			raw:     []byte{0x57, 0x01, 0x07, 0x01},
			wantErr: dnp3.ErrUnsupportedObject,
		},
		{
			name:    "descriptor-dependent indexed value",
			raw:     []byte{0x57, 0x01, 0x17, 0x01, 0x00, 0x01},
			wantErr: dnp3.ErrUnsupportedObject,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := dnp3.NewDataObjectFromBytes(testCase.raw)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("NewDataObjectFromBytes: got %v, want %v", err, testCase.wantErr)
			}
		})
	}
}
