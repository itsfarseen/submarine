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

func TestDecodeCompact(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int64
		wantErr  bool
	}{
		{"mode 0 - zero", []byte{0x00}, 0, false},
		{"mode 0 - max", []byte{0xFC}, 63, false},          // 0b11111100 >> 2 = 63
		{"mode 1 - 64", []byte{0x01, 0x01}, 64, false},     // (0x01 >> 2) | (0x01 << 6) = 0 | 64 = 64
		{"mode 1 - max", []byte{0xFD, 0xFF}, 16383, false}, // 0x3FFF
		{"mode 2 - 16384", []byte{0x02, 0x00, 0x01, 0x00}, 16384, false},
		{"mode 2 - max", []byte{0xFE, 0xFF, 0xFF, 0xFF}, 1073741823, false},           // 0x3FFFFFFF
		{"mode 3 - 4 bytes", []byte{0x03, 0x00, 0x00, 0x00, 0x40}, 1073741824, false}, // 2^30
		{"insufficient data mode 1", []byte{0x01}, 0, true},
		{"insufficient data mode 2", []byte{0x02, 0x00}, 0, true},
		{"insufficient data mode 3", []byte{0x03, 0x00}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeCompact(r)

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

			if result.Int64() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.Int64())
			}
		})
	}
}

func TestDecodeU8(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint8
		wantErr  bool
	}{
		{"zero", []byte{0x00}, 0, false},
		{"max", []byte{0xFF}, 255, false},
		{"middle", []byte{0x80}, 128, false},
		{"empty", []byte{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeU8(r)

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

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDecodeU16(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint16
		wantErr  bool
	}{
		{"zero", []byte{0x00, 0x00}, 0, false},
		{"little endian", []byte{0x34, 0x12}, 0x1234, false},
		{"max", []byte{0xFF, 0xFF}, 0xFFFF, false},
		{"insufficient data", []byte{0x34}, 0, true},
		{"empty", []byte{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeU16(r)

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

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDecodeU32(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint32
		wantErr  bool
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0, false},
		{"little endian", []byte{0x78, 0x56, 0x34, 0x12}, 0x12345678, false},
		{"max", []byte{0xFF, 0xFF, 0xFF, 0xFF}, 0xFFFFFFFF, false},
		{"insufficient data", []byte{0x78, 0x56, 0x34}, 0, true},
		{"empty", []byte{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeU32(r)

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

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDecodeU64(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint64
		wantErr  bool
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0, false},
		{"little endian", []byte{0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11}, 0x1122334455667788, false},
		{"insufficient data", []byte{0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22}, 0, true},
		{"empty", []byte{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeU64(r)

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

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDecodeU128(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
		wantErr  bool
	}{
		{"zero", make([]byte, 16), "0", false},
		{"one",
			[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			"1", false},
		{"max",
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			"340282366920938463463374607431768211455", false},
		{"large value",
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
			"338953138925153547590470800371487866880", false},
		{"insufficient data", make([]byte, 15), "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeU128(r)

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

			if result.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}

			if result.Sign() < 0 {
				t.Errorf("expected to be not negative. expected: %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestDecodeU256(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string // Use string representation for big.Int comparison
		wantErr  bool
	}{
		{"zero",
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			"0", false},
		{"one",
			[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			"1", false},
		{"max",
			[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
				0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			"115792089237316195423570985008687907853269984665640564039457584007913129639935", false},
		{"large value",
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
			"115339776388732929035197660848497720713218148788040405586178452820382218977280", false},
		{"insufficient data", make([]byte, 15), "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeU256(r)

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

			if result.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}

			if result.Sign() < 0 {
				t.Errorf("expected to be not negative. expected: %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestDecodeI8(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int8
		wantErr  bool
	}{
		{"zero", []byte{0x00}, 0, false},
		{"positive", []byte{0x7F}, 127, false},
		{"negative", []byte{0x80}, -128, false},
		{"minus one", []byte{0xFF}, -1, false},
		{"empty", []byte{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeI8(r)

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

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDecodeI16(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int16
		wantErr  bool
	}{
		{"zero", []byte{0x00, 0x00}, 0, false},
		{"positive", []byte{0x34, 0x12}, 0x1234, false},
		{"negative", []byte{0x00, 0x80}, -32768, false},
		{"minus one", []byte{0xFF, 0xFF}, -1, false},
		{"insufficient data", []byte{0x34}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeI16(r)

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

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDecodeI32(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int32
		wantErr  bool
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00}, 0, false},
		{"positive", []byte{0x78, 0x56, 0x34, 0x12}, 0x12345678, false},
		{"negative", []byte{0x00, 0x00, 0x00, 0x80}, -2147483648, false},
		{"minus one", []byte{0xFF, 0xFF, 0xFF, 0xFF}, -1, false},
		{"insufficient data", []byte{0x78, 0x56, 0x34}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeI32(r)

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

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDecodeI64(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int64
		wantErr  bool
	}{
		{"zero", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0, false},
		{"positive", []byte{0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11}, 0x1122334455667788, false},
		{"negative", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80}, -9223372036854775808, false},
		{"minus one", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, -1, false},
		{"insufficient data", []byte{0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeI64(r)

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

			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDecodeI128(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string // Use string representation for big.Int comparison
		wantErr  bool
	}{
		{"zero", make([]byte, 16), "0", false},
		{"one",
			[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			"1", false},
		{"minus one",
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			"-1", false},
		{"large positive",
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00},
			"1329227995784915872903807060280344575", false},
		{"large negative",
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
			"-170141183460469231731687303715884105728", false},
		{"insufficient data", make([]byte, 15), "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeI128(r)

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

			if result.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestDecodeI256(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string // Use string representation for big.Int comparison
		wantErr  bool
	}{
		{"zero",
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			"0", false},
		{"one",
			[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			"1", false},
		{"minus one",
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			"-1", false},
		{"large positive",
			[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00},
			"452312848583266388373324160190187140051835877600158453279131187530910662655", false},
		{"large negative",
			[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80},
			"-57896044618658097711785492504343953926634992332820282019728792003956564819968", false},
		{"insufficient data", make([]byte, 15), "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeI256(r)

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

			if result.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestDecodeBool(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
		wantErr  bool
	}{
		{"false", []byte{0x00}, false, false},
		{"true", []byte{0x01}, true, false},
		{"invalid", []byte{0x02}, false, true},
		{"empty", []byte{}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeBool(r)

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

			if result != tt.expected {
				t.Errorf("expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestDecodeText(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
		wantErr  bool
	}{
		{"empty string", []byte{0x00}, "", false},
		{"hello", []byte{0x14, 0x68, 0x65, 0x6C, 0x6C, 0x6F}, "hello", false}, // len=5, "hello"
		{"unicode", []byte{0x0C, 0xE2, 0x9C, 0x93}, "✓", false},               // len=3, checkmark
		{"invalid length", []byte{0x14, 0x68, 0x65}, "", true},                // len=5 but only 2 bytes
		{"empty data", []byte{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeText(r)

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

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDecodeBytes(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected []byte
		wantErr  bool
	}{
		{"empty bytes", []byte{0x00}, []byte{}, false},
		{"some bytes", []byte{0x0C, 0x01, 0x02, 0x03}, []byte{0x01, 0x02, 0x03}, false}, // len=3
		{"invalid length", []byte{0x08, 0x01}, nil, true},                               // len=2 but only 1 byte
		{"empty data", []byte{}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeBytes(r)

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

func TestDecodeVec(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected []uint8
		wantErr  bool
	}{
		{"empty vec", []byte{0x00}, []uint8{}, false},
		{"vec of u8", []byte{0x0C, 0x01, 0x02, 0x03}, []uint8{1, 2, 3}, false}, // len=3
		{"invalid length", []byte{0x08, 0x01}, nil, true},                      // len=2 but only 1 element
		{"empty data", []byte{}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeVec(r, DecodeU8)

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

func TestDecodeOption(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected *uint8
		wantErr  bool
	}{
		{"none", []byte{0x00}, nil, false},
		{"some", []byte{0x01, 0x42}, uint8Ptr(0x42), false},
		{"invalid flag", []byte{0x02}, nil, true},
		{"missing value", []byte{0x01}, nil, true},
		{"empty data", []byte{}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeOption(r, DecodeU8)

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

			if (result == nil) != (tt.expected == nil) {
				t.Errorf("expected %v, got %v", tt.expected, result)
				return
			}

			if result != nil && *result != *tt.expected {
				t.Errorf("expected %v, got %v", *tt.expected, *result)
			}
		})
	}
}

func uint8Ptr(v uint8) *uint8 {
	return &v
}

func bytes(val byte, count int) []byte {
	result := make([]byte, count)
	for i := range result {
		result[i] = val
	}
	return result
}

func bytes_padz(val []byte, count int) []byte {
	result := make([]byte, count)
	var i int
	for i = range result {
		result[i] = val[i]
	}
	for ; i < count; i += 1 {
		result[i] = 0
	}
	return result
}
