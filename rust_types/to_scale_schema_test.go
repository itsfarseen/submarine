package rust_types_test

import (
	r "submarine/rust_types"
	s "submarine/scale/schema"
	"testing"
)

func TestToScaleSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    *r.RustType
		expected s.Type
		wantErr  bool
	}{
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name:  "simple ref type",
			input: &r.RustType{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"MyType"}}},
			expected: s.Type{
				Kind: s.KindRef,
				Ref:  stringPtr("MyType"),
			},
		},
		{
			name:  "module path ref type",
			input: &r.RustType{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"foo", "bar", "Baz"}}},
			expected: s.Type{
				Kind: s.KindRef,
				Ref:  stringPtr("foo::bar::Baz"),
			},
		},
		{
			name: "Vec type",
			input: &r.RustType{
				Kind: r.KindBase,
				Base: &r.RustTypeBase{
					Path: []string{"Vec"},
					Generics: []r.RustType{
						{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"u8"}}},
					},
				},
			},
			expected: s.Type{
				Kind: s.KindVec,
				Vec: &s.Vec{
					Type: &s.Type{
						Kind: s.KindRef,
						Ref:  stringPtr("u8"),
					},
				},
			},
		},
		{
			name: "Option type",
			input: &r.RustType{
				Kind: r.KindBase,
				Base: &r.RustTypeBase{
					Path: []string{"Option"},
					Generics: []r.RustType{
						{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"String"}}},
					},
				},
			},
			expected: s.Type{
				Kind: s.KindOption,
				Option: &s.Option{
					Type: &s.Type{
						Kind: s.KindRef,
						Ref:  stringPtr("String"),
					},
				},
			},
		},
		{
			name: "Array type",
			input: &r.RustType{
				Kind: r.KindArray,
				Array: &r.RustTypeArray{
					Base: r.RustType{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"u8"}}},
					Len:  32,
				},
			},
			expected: s.Type{
				Kind: s.KindArray,
				Array: &s.Array{
					Type: &s.Type{
						Kind: s.KindRef,
						Ref:  stringPtr("u8"),
					},
					Len: 32,
				},
			},
		},
		{
			name: "empty tuple",
			input: &r.RustType{
				Kind:  r.KindTuple,
				Tuple: nil,
			},
			expected: s.Type{
				Kind: s.KindTuple,
				Tuple: &s.Tuple{
					Fields: []s.Type{},
				},
			},
		},
		{
			name: "tuple with fields",
			input: &r.RustType{
				Kind: r.KindTuple,
				Tuple: &[]r.RustType{
					{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"u32"}}},
					{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"String"}}},
				},
			},
			expected: s.Type{
				Kind: s.KindTuple,
				Tuple: &s.Tuple{
					Fields: []s.Type{
						{
							Kind: s.KindRef,
							Ref:  stringPtr("u32"),
						},
						{
							Kind: s.KindRef,
							Ref:  stringPtr("String"),
						},
					},
				},
			},
		},
		{
			name: "Vec with wrong generic count",
			input: &r.RustType{
				Kind: r.KindBase,
				Base: &r.RustTypeBase{
					Path:     []string{"Vec"},
					Generics: []r.RustType{},
				},
			},
			wantErr: true,
		},
		{
			name: "Option with wrong generic count",
			input: &r.RustType{
				Kind: r.KindBase,
				Base: &r.RustTypeBase{
					Path: []string{"Option"},
					Generics: []r.RustType{
						{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"u8"}}},
						{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"u16"}}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported generic type",
			input: &r.RustType{
				Kind: r.KindBase,
				Base: &r.RustTypeBase{
					Path: []string{"HashMap"},
					Generics: []r.RustType{
						{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"String"}}},
						{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"u32"}}},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "nested Vec of Option",
			input: &r.RustType{
				Kind: r.KindBase,
				Base: &r.RustTypeBase{
					Path: []string{"Vec"},
					Generics: []r.RustType{
						{
							Kind: r.KindBase,
							Base: &r.RustTypeBase{
								Path: []string{"Option"},
								Generics: []r.RustType{
									{Kind: r.KindBase, Base: &r.RustTypeBase{Path: []string{"u64"}}},
								},
							},
						},
					},
				},
			},
			expected: s.Type{
				Kind: s.KindVec,
				Vec: &s.Vec{
					Type: &s.Type{
						Kind: s.KindOption,
						Option: &s.Option{
							Type: &s.Type{
								Kind: s.KindRef,
								Ref:  stringPtr("u64"),
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := r.ToScaleSchema(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !deepEqualType(result, tt.expected) {
				t.Errorf("expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

// deepEqualType compares two s.Type values for equality
func deepEqualType(a, b s.Type) bool {
	if a.Kind != b.Kind {
		return false
	}

	switch a.Kind {
	case s.KindRef:
		return (a.Ref == nil && b.Ref == nil) || (a.Ref != nil && b.Ref != nil && *a.Ref == *b.Ref)
	case s.KindVec:
		return (a.Vec == nil && b.Vec == nil) ||
			(a.Vec != nil && b.Vec != nil && deepEqualType(*a.Vec.Type, *b.Vec.Type))
	case s.KindOption:
		return (a.Option == nil && b.Option == nil) ||
			(a.Option != nil && b.Option != nil && deepEqualType(*a.Option.Type, *b.Option.Type))
	case s.KindArray:
		return (a.Array == nil && b.Array == nil) ||
			(a.Array != nil && b.Array != nil &&
				a.Array.Len == b.Array.Len &&
				deepEqualType(*a.Array.Type, *b.Array.Type))
	case s.KindTuple:
		if a.Tuple == nil && b.Tuple == nil {
			return true
		}
		if a.Tuple == nil || b.Tuple == nil {
			return false
		}
		if len(a.Tuple.Fields) != len(b.Tuple.Fields) {
			return false
		}
		for i, field := range a.Tuple.Fields {
			if !deepEqualType(field, b.Tuple.Fields[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
