package scale_test

import (
	"reflect"
	. "submarine/scale"
	"testing"
)

type M = map[string]any
type A = []any

func ref(name string) *Type {
	return &Type{Kind: KindRef, Ref: &name}
}

func TestDecodeWithSchema(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		schema   *Type
		expected any
		wantErr  bool
	}{
		{
			name: "simple struct",
			data: []byte{0x08, 0x10},
			schema: &Type{
				Kind: KindStruct,
				Struct: &Struct{
					Fields: []NamedMember{
						{
							Name: "a",
							Type: ref("u8"),
						},
						{
							Name: "b",
							Type: ref("u8"),
						},
					},
				},
			},
			expected: M{"a": uint8(8), "b": uint8(16)},
		},
		{
			name: "tuple",
			data: []byte{0x0C, 0x14},
			schema: &Type{
				Kind: KindTuple,
				Tuple: &Tuple{
					Fields: []Type{
						*ref("u8"),
						*ref("u8"),
					},
				},
			},
			expected: A{uint8(0x0C), uint8(0x14)},
		},
		{
			name: "simple enum - first variant",
			data: []byte{0x00}, // variant index 0
			schema: &Type{
				Kind: KindEnumSimple,
				EnumSimple: &EnumSimple{
					Variants: []string{"Red", "Green", "Blue"},
				},
			},
			expected: "Red",
		},
		{
			name: "simple enum - second variant",
			data: []byte{0x01}, // variant index 1
			schema: &Type{
				Kind: KindEnumSimple,
				EnumSimple: &EnumSimple{
					Variants: []string{"Red", "Green", "Blue"},
				},
			},
			expected: "Green",
		},
		{
			name: "complex enum",
			data: []byte{0x01, 0x08}, // variant index 1, u8=2
			schema: &Type{
				Kind: KindEnumComplex,
				EnumComplex: &EnumComplex{
					Variants: []NamedMember{
						{Name: "None", Type: ref("unit")},
						{Name: "Some", Type: ref("u8")},
					},
				},
			},
			expected: M{"Some": uint8(0x08)},
		},
		{
			name: "vec of u8",
			data: []byte{0x08, 0x01, 0x02}, // length=2, [1, 2]
			schema: &Type{
				Kind: KindVec,
				Vec: &Vec{
					Type: ref("u8"),
				},
			},
			expected: []uint8{uint8(1), uint8(2)},
		},
		{
			name: "empty vec",
			data: []byte{0x00}, // length=0
			schema: &Type{
				Kind: KindVec,
				Vec: &Vec{
					Type: ref("u8"),
				},
			},
			expected: []uint8{},
		},
		{
			name: "option none",
			data: []byte{0x00}, // has_value=false
			schema: &Type{
				Kind: KindOption,
				Option: &Option{
					Type: ref("u8"),
				},
			},
			expected: nil,
		},
		{
			name: "option some",
			data: []byte{0x01, 0x2A},
			schema: &Type{
				Kind: KindOption,
				Option: &Option{
					Type: ref("u8"),
				},
			},
			expected: uint8(42),
		},
		{
			name: "array",
			data: []byte{0x01, 0x02, 0x03},
			schema: &Type{
				Kind: KindArray,
				Array: &Array{
					Type: ref("u8"),
					Len:  3,
				},
			},
			expected: []uint8{1, 2, 3},
		},
		{
			name: "nested struct",
			data: []byte{0x08, 0x04, 0x10, 0x20},
			schema: &Type{
				Kind: KindStruct,
				Struct: &Struct{
					Fields: []NamedMember{
						{
							Name: "inner1",
							Type: &Type{
								Kind: KindStruct,
								Struct: &Struct{
									Fields: []NamedMember{
										{Name: "a", Type: ref("u8")},
										{Name: "b", Type: ref("u8")},
									},
								},
							},
						},
						{
							Name: "inner2",
							Type: &Type{
								Kind: KindStruct,
								Struct: &Struct{
									Fields: []NamedMember{
										{Name: "x", Type: ref("u8")},
										{Name: "y", Type: ref("u8")},
									},
								},
							},
						},
					},
				},
			},
			expected: M{
				"inner1": M{"a": uint8(0x08), "b": uint8(0x04)},
				"inner2": M{"x": uint8(0x10), "y": uint8(0x20)},
			},
		},
		{
			name: "enum index out of bounds",
			data: []byte{0x03}, // index 3, but only 3 variants (0,1,2)
			schema: &Type{
				Kind: KindEnumSimple,
				EnumSimple: &EnumSimple{
					Variants: []string{"A", "B", "C"},
				},
			},
			wantErr: true,
		},
		{
			name:    "ref type error",
			data:    []byte{},
			schema:  ref("SomeType"),
			wantErr: true,
		},
		{
			name: "import type error",
			data: []byte{},
			schema: &Type{
				Kind: KindImport,
				Import: &Import{
					Module: "std",
					Item:   "Vec",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			result, err := DecodeWithSchema(r, tt.schema)

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
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestDecodeWithSchema_PrimitiveTypes(t *testing.T) {
	primitiveTests := []struct {
		name     string
		data     []byte
		refType  string
		expected any
	}{
		{"u8", []byte{0x42}, "u8", uint8(0x42)},
		{"u16", []byte{0x34, 0x12}, "u16", uint16(0x1234)},
		{"u32", []byte{0x78, 0x56, 0x34, 0x12}, "u32", uint32(0x12345678)},
		{"bool_true", []byte{0x01}, "bool", true},
		{"bool_false", []byte{0x00}, "bool", false},
	}

	for _, tt := range primitiveTests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(tt.data)
			schema := ref(tt.refType)
			result, err := DecodeWithSchema(r, schema)
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
