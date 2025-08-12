package schema_parser_test

import (
	. "submarine/metadata/schema_parser"
	. "submarine/scale"
	"testing"
)

// Test helper types to reduce noise
type M = map[string]any
type A = []any

func TestParseType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected *Type
		wantErr  bool
	}{
		{
			name:  "simple string reference",
			input: "u32",
			expected: &Type{
				Kind: KindRef,
				Ref:  stringPtr("u32"),
			},
		},
		{
			name: "unnecessary nesting of string references",
			input: M{
				"type": "u32",
			},
			wantErr: true,
		},
		{
			name: "struct type",
			input: M{
				"type": "struct",
				"fields": A{
					M{"name": "id", "type": "u32"},
					M{"name": "name", "type": "string"},
				},
			},
			expected: &Type{
				Kind: KindStruct,
				Struct: &Struct{
					Fields: []NamedMember{
						{Name: "id", Type: &Type{Kind: KindRef, Ref: stringPtr("u32")}},
						{Name: "name", Type: &Type{Kind: KindRef, Ref: stringPtr("string")}},
					},
				},
			},
		},
		{
			name: "tuple type",
			input: M{
				"type":   "tuple",
				"fields": A{"u32", "string"},
			},
			expected: &Type{
				Kind: KindTuple,
				Tuple: &Tuple{
					Fields: []Type{
						{Kind: KindRef, Ref: stringPtr("u32")},
						{Kind: KindRef, Ref: stringPtr("string")},
					},
				},
			},
		},
		{
			name: "simple enum",
			input: M{
				"type":     "enum_simple",
				"variants": A{"Red", "Green", "Blue"},
			},
			expected: &Type{
				Kind: KindEnumSimple,
				EnumSimple: &EnumSimple{
					Variants: []string{"Red", "Green", "Blue"},
				},
			},
		},
		{
			name: "complex enum",
			input: M{
				"type": "enum_complex",
				"variants": A{
					M{"name": "Success", "type": "u32"},
					M{"name": "Error", "type": "string"},
				},
			},
			expected: &Type{
				Kind: KindEnumComplex,
				EnumComplex: &EnumComplex{
					Variants: []NamedMember{
						{Name: "Success", Type: &Type{Kind: KindRef, Ref: stringPtr("u32")}},
						{Name: "Error", Type: &Type{Kind: KindRef, Ref: stringPtr("string")}},
					},
				},
			},
		},
		{
			name: "import type",
			input: M{
				"type":   "import",
				"module": "std",
				"item":   "Vec",
			},
			expected: &Type{
				Kind: KindImport,
				Import: &Import{
					Module: "std",
					Item:   "Vec",
				},
			},
		},
		{
			name: "vec type",
			input: M{
				"type": "vec",
				"item": "u32",
			},
			expected: &Type{
				Kind: KindVec,
				Vec: &Vec{
					Type: &Type{Kind: KindRef, Ref: stringPtr("u32")},
				},
			},
		},
		{
			name: "option type",
			input: M{
				"type": "option",
				"item": "string",
			},
			expected: &Type{
				Kind: KindOption,
				Option: &Option{
					Type: &Type{Kind: KindRef, Ref: stringPtr("string")},
				},
			},
		},
		{
			name: "complex nested type",
			input: M{
				"type": "option",
				"item": M{
					"type": "vec",
					"item": M{
						"type": "tuple",
						"fields": A{
							M{
								"type": "struct",
								"fields": A{
									M{"name": "field1", "type": "string"},
									M{"name": "field2", "type": M{
										"type": "vec",
										"item": "string",
									}},
								},
							},
						},
					},
				},
			},
			expected: &Type{
				Kind: KindOption,
				Option: &Option{
					Type: &Type{
						Kind: KindVec,
						Vec: &Vec{
							Type: &Type{
								Kind: KindTuple,
								Tuple: &Tuple{
									Fields: []Type{
										{
											Kind: KindStruct,
											Struct: &Struct{
												Fields: []NamedMember{
													{Name: "field1", Type: &Type{Kind: KindRef, Ref: stringPtr("string")}},
													{Name: "field2", Type: &Type{
														Kind: KindVec,
														Vec: &Vec{
															Type: &Type{Kind: KindRef, Ref: stringPtr("string")},
														},
													}},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "array type",
			input: M{
				"type": "array",
				"item": "u8",
				"len":  32,
			},
			expected: &Type{
				Kind: KindArray,
				Array: &Array{
					Type: &Type{Kind: KindRef, Ref: stringPtr("u8")},
					Len:  32,
				},
			},
		},
		{
			name:    "invalid type - number",
			input:   42,
			wantErr: true,
		},
		{
			name:    "invalid type - missing type field",
			input:   M{"fields": A{}},
			wantErr: true,
		},
		{
			name:    "invalid struct - missing fields",
			input:   M{"type": "struct"},
			wantErr: true,
		},
		{
			name:    "invalid enum_simple - non-string variant",
			input:   M{"type": "enum_simple", "variants": A{"Red", 42}},
			wantErr: true,
		},
		{
			name:    "invalid vec - missing item",
			input:   M{"type": "vec"},
			wantErr: true,
		},
		{
			name:    "invalid array - missing len",
			input:   M{"type": "array", "item": "u32"},
			wantErr: true,
		},
		{
			name:    "unknown type",
			input:   M{"type": "unknown_type"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseType(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseType() expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseType() unexpected error: %v", err)
				return
			}

			if !compareTypes(result, tt.expected) {
				t.Errorf("ParseType() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseNamedMembers(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected []NamedMember
		wantErr  bool
	}{
		{
			name: "valid named members",
			input: A{
				M{"name": "field1", "type": "u32"},
				M{"name": "field2", "type": "string"},
			},
			expected: []NamedMember{
				{Name: "field1", Type: &Type{Kind: KindRef, Ref: stringPtr("u32")}},
				{Name: "field2", Type: &Type{Kind: KindRef, Ref: stringPtr("string")}},
			},
			wantErr: false,
		},
		{
			name:     "invalid member - not a map",
			input:    A{"not_a_map"},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid member - missing name",
			input:    A{M{"type": "u32"}},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid member - missing type",
			input:    A{M{"name": "field1"}},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNamedMembers(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseNamedMembers() expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseNamedMembers() unexpected error: %v", err)
				return
			}

			if !compareNamedMembers(result, tt.expected) {
				t.Errorf("ParseNamedMembers() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseTupleMembers(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected []Type
		wantErr  bool
	}{
		{
			name:  "valid tuple members",
			input: A{"u32", "string"},
			expected: []Type{
				{Kind: KindRef, Ref: stringPtr("u32")},
				{Kind: KindRef, Ref: stringPtr("string")},
			},
			wantErr: false,
		},
		{
			name: "complex tuple members",
			input: A{
				M{"type": "vec", "item": "u8"},
			},
			expected: []Type{
				{
					Kind: KindVec,
					Vec: &Vec{
						Type: &Type{Kind: KindRef, Ref: stringPtr("u8")},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "invalid tuple member",
			input:    A{42},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTupleMembers(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTupleMembers() expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTupleMembers() unexpected error: %v", err)
				return
			}

			if !compareTypeSlices(result, tt.expected) {
				t.Errorf("ParseTupleMembers() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func compareTypes(a, b *Type) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.Kind != b.Kind {
		return false
	}

	switch a.Kind {
	case KindRef:
		return stringPtrEqual(a.Ref, b.Ref)
	case KindStruct:
		return compareNamedMembers(a.Struct.Fields, b.Struct.Fields)
	case KindTuple:
		return compareTypeSlices(a.Tuple.Fields, b.Tuple.Fields)
	case KindEnumSimple:
		return compareStringSlices(a.EnumSimple.Variants, b.EnumSimple.Variants)
	case KindEnumComplex:
		return compareNamedMembers(a.EnumComplex.Variants, b.EnumComplex.Variants)
	case KindImport:
		return a.Import.Module == b.Import.Module && a.Import.Item == b.Import.Item
	case KindVec:
		return compareTypes(a.Vec.Type, b.Vec.Type)
	case KindOption:
		return compareTypes(a.Option.Type, b.Option.Type)
	case KindArray:
		return a.Array.Len == b.Array.Len && compareTypes(a.Array.Type, b.Array.Type)
	}

	return false
}

func stringPtrEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func compareNamedMembers(a, b []NamedMember) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Name != b[i].Name || !compareTypes(a[i].Type, b[i].Type) {
			return false
		}
	}

	return true
}

func compareTypeSlices(a, b []Type) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !compareTypes(&a[i], &b[i]) {
			return false
		}
	}

	return true
}

func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
