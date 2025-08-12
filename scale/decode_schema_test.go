package scale_test

import (
	"math/big"
	"reflect"
	. "submarine/scale"
	"testing"
)

func ref(name string) *Type {
	return &Type{Kind: KindRef, Ref: &name}
}

func TestDecodeWithSchema(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		schema   *Type
		expected Value
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
			expected: VStruct(map[string]Value{
				"a": VIntFromInt64(8),
				"b": VIntFromInt64(16),
			}),
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
			expected: VList([]Value{
				VIntFromInt64(12),
				VIntFromInt64(20),
			}),
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
			expected: VText("Red"),
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
			expected: VText("Green"),
		},
		{
			name: "complex enum",
			data: []byte{0x01, 0x08}, // variant index 1, u8=8
			schema: &Type{
				Kind: KindEnumComplex,
				EnumComplex: &EnumComplex{
					Variants: []NamedMember{
						{Name: "None", Type: ref("empty")},
						{Name: "Some", Type: ref("u8")},
					},
				},
			},
			expected: VStruct(map[string]Value{
				"Some": VIntFromInt64(8),
			}),
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
			expected: VBytes([]byte{1, 2}),
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
			expected: VBytes([]byte{}),
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
			expected: VStruct(make(map[string]Value)),
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
			expected: VIntFromInt64(42),
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
			expected: VBytes([]byte{1, 2, 3}),
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
			expected: VStruct(map[string]Value{
				"inner1": VStruct(map[string]Value{
					"a": VIntFromInt64(8),
					"b": VIntFromInt64(4),
				}),
				"inner2": VStruct(map[string]Value{
					"x": VIntFromInt64(16),
					"y": VIntFromInt64(32),
				}),
			}),
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
			name: "bit flags - u8",
			data: []byte{0x95}, // binary: 10010101 = 149
			schema: &Type{
				Kind: KindBitFlags,
				BitFlags: &BitFlags{
					BitLength: 8,
					Flags: []BitFlag{
						{Name: "Display", Value: 1},         // bit 0: set
						{Name: "Legal", Value: 2},           // bit 1: not set
						{Name: "Web", Value: 4},             // bit 2: set
						{Name: "Riot", Value: 8},            // bit 3: not set
						{Name: "Email", Value: 16},          // bit 4: set
						{Name: "PgpFingerprint", Value: 32}, // bit 5: not set
						{Name: "Image", Value: 64},          // bit 6: not set
						{Name: "Twitter", Value: 128},       // bit 7: set
					},
				},
			},
			expected: VStruct(map[string]Value{
				"Display":        VBool(true),
				"Legal":          VBool(false),
				"Web":            VBool(true),
				"Riot":           VBool(false),
				"Email":          VBool(true),
				"PgpFingerprint": VBool(false),
				"Image":          VBool(false),
				"Twitter":        VBool(true),
			}),
		},
		{
			name: "bit flags - u64",
			data: []byte{0x91, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // 145 in little-endian
			schema: &Type{
				Kind: KindBitFlags,
				BitFlags: &BitFlags{
					BitLength: 64,
					Flags: []BitFlag{
						{Name: "Display", Value: 1},   // bit 0: set
						{Name: "Email", Value: 16},    // bit 4: set
						{Name: "Twitter", Value: 128}, // bit 7: set
					},
				},
			},
			expected: VStruct(map[string]Value{
				"Display": VBool(true),
				"Email":   VBool(true),
				"Twitter": VBool(true),
			}),
		},
		{
			name: "bit flags - all false",
			data: []byte{0x00},
			schema: &Type{
				Kind: KindBitFlags,
				BitFlags: &BitFlags{
					BitLength: 8,
					Flags: []BitFlag{
						{Name: "Flag1", Value: 1},
						{Name: "Flag2", Value: 2},
					},
				},
			},
			expected: VStruct(map[string]Value{
				"Flag1": VBool(false),
				"Flag2": VBool(false),
			}),
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
		expected Value
	}{
		{"u8", []byte{0x42}, "u8", VIntFromInt64(0x42)},
		{"u16", []byte{0x34, 0x12}, "u16", VIntFromInt64(0x1234)},
		{"u32", []byte{0x78, 0x56, 0x34, 0x12}, "u32", VIntFromInt64(0x12345678)},
		{"u64", []byte{0x78, 0x56, 0x34, 0x12, 0x00, 0x00, 0x00, 0x00}, "u64", VInt(big.NewInt(0x12345678))},
		{"bool_true", []byte{0x01}, "bool", VBool(true)},
		{"bool_false", []byte{0x00}, "bool", VBool(false)},
		{"text", []byte{0x14, 0x48, 0x65, 0x6C, 0x6C, 0x6F}, "text", VText("Hello")},
		{"bytes", []byte{0x0C, 0x01, 0x02, 0x03}, "bytes", VBytes([]byte{1, 2, 3})},
		{"empty", []byte{}, "empty", VNull()},
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
