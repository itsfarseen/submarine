package scale_test

import (
	"reflect"
	. "submarine/scale"
	"testing"
)

func TestReader_ReadByte(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected []byte
		wantErr  bool
	}{
		{"single byte", []byte{0x42}, []byte{0x42}, false},
		{"multiple bytes", []byte{0x01, 0x02, 0x03}, []byte{0x01, 0x02, 0x03}, false},
		{"empty data", []byte{}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			var results []byte

			for i := 0; i < len(tt.data)+1; i++ {
				b, err := r.ReadByte()
				if err != nil {
					if !tt.wantErr && i < len(tt.data) {
						t.Errorf("unexpected error at position %d: %v", i, err)
					}
					break
				}
				results = append(results, b)
			}

			if !tt.wantErr && !reflect.DeepEqual(results, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, results)
			}
		})
	}
}

func TestReader_ReadBytes(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		n        int
		expected []byte
		wantErr  bool
	}{
		{"read partial", []byte{0x01, 0x02, 0x03, 0x04}, 2, []byte{0x01, 0x02}, false},
		{"read all", []byte{0x01, 0x02}, 2, []byte{0x01, 0x02}, false},
		{"read zero", []byte{0x01, 0x02}, 0, []byte{}, false},
		{"read too many", []byte{0x01, 0x02}, 3, nil, true},
		{"read from empty", []byte{}, 1, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := r.ReadBytes(tt.n)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
