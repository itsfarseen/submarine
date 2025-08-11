package parser_test

import (
	"fmt"
	"testing"

	. "submarine/rust_types"
	. "submarine/rust_types/parser"
)

// deepEqual compares two RustType values for structural equality
// Returns nil if equal, or a string describing the first difference found
func deepEqual(a, b RustType) *string {
	return deepEqualPath(a, b, "")
}

func deepEqualPath(a, b RustType, path string) *string {
	if a.Kind != b.Kind {
		diff := fmt.Sprintf("%s: kind mismatch %d vs %d", path, a.Kind, b.Kind)
		return &diff
	}

	switch a.Kind {
	case KindBase:
		if a.Base == nil && b.Base == nil {
			return nil
		}
		if a.Base == nil {
			diff := fmt.Sprintf("%s.Base: nil vs non-nil", path)
			return &diff
		}
		if b.Base == nil {
			diff := fmt.Sprintf("%s.Base: non-nil vs nil", path)
			return &diff
		}

		if len(a.Base.Path) != len(b.Base.Path) {
			diff := fmt.Sprintf("%s.Base.Path: length %d vs %d", path, len(a.Base.Path), len(b.Base.Path))
			return &diff
		}
		for i := range a.Base.Path {
			if a.Base.Path[i] != b.Base.Path[i] {
				diff := fmt.Sprintf("%s.Base.Path[%d]: %q vs %q", path, i, a.Base.Path[i], b.Base.Path[i])
				return &diff
			}
		}

		if len(a.Base.Generics) != len(b.Base.Generics) {
			diff := fmt.Sprintf("%s.Base.Generics: length %d vs %d", path, len(a.Base.Generics), len(b.Base.Generics))
			return &diff
		}
		for i := range a.Base.Generics {
			genericPath := fmt.Sprintf("%s.Base.Generics[%d]", path, i)
			if diff := deepEqualPath(a.Base.Generics[i], b.Base.Generics[i], genericPath); diff != nil {
				return diff
			}
		}
		return nil

	case KindTuple:
		if a.Tuple == nil && b.Tuple == nil {
			return nil
		}
		if a.Tuple == nil {
			diff := fmt.Sprintf("%s.Tuple: nil vs non-nil", path)
			return &diff
		}
		if b.Tuple == nil {
			diff := fmt.Sprintf("%s.Tuple: non-nil vs nil", path)
			return &diff
		}

		if len(*a.Tuple) != len(*b.Tuple) {
			diff := fmt.Sprintf("%s.Tuple: length %d vs %d", path, len(*a.Tuple), len(*b.Tuple))
			return &diff
		}
		for i := range *a.Tuple {
			tuplePath := fmt.Sprintf("%s.Tuple[%d]", path, i)
			if diff := deepEqualPath((*a.Tuple)[i], (*b.Tuple)[i], tuplePath); diff != nil {
				return diff
			}
		}
		return nil

	case KindArray:
		if a.Array == nil && b.Array == nil {
			return nil
		}
		if a.Array == nil {
			diff := fmt.Sprintf("%s.Array: nil vs non-nil", path)
			return &diff
		}
		if b.Array == nil {
			diff := fmt.Sprintf("%s.Array: non-nil vs nil", path)
			return &diff
		}

		if a.Array.Len != b.Array.Len {
			diff := fmt.Sprintf("%s.Array.Len: %d vs %d", path, a.Array.Len, b.Array.Len)
			return &diff
		}

		basePath := fmt.Sprintf("%s.Array.Base", path)
		if diff := deepEqualPath(a.Array.Base, b.Array.Base, basePath); diff != nil {
			return diff
		}
		return nil

	default:
		diff := fmt.Sprintf("%s: unknown kind %d", path, a.Kind)
		return &diff
	}
}

func TestRustTypesParser(t *testing.T) {
	tests := []struct {
		input       string
		expectedAST RustType
		expectedStr string
	}{
		// Simple identifiers
		{
			"String",
			Base([]string{"String"}, nil),
			"String",
		},
		{
			"u32",
			Base([]string{"u32"}, nil),
			"u32",
		},

		// Simple paths
		{
			"std::String",
			Base([]string{"std", "String"}, nil),
			"std::String",
		},
		{
			"foo::bar::Baz",
			Base([]string{"foo", "bar", "Baz"}, nil),
			"foo::bar::Baz",
		},

		// Generics with single parameter
		{
			"Box<T>",
			Base([]string{"Box"}, []RustType{
				Base([]string{"T"}, nil),
			}),
			"Box<T>",
		},
		{
			"Vec<u32>",
			Base([]string{"Vec"}, []RustType{
				Base([]string{"u32"}, nil),
			}),
			"Vec<u32>",
		},

		// Generics with multiple parameters
		{
			"HashMap<String, i32>",
			Base([]string{"HashMap"}, []RustType{
				Base([]string{"String"}, nil),
				Base([]string{"i32"}, nil),
			}),
			"HashMap<String, i32>",
		},
		{
			"Result<T, E>",
			Base([]string{"Result"}, []RustType{
				Base([]string{"T"}, nil),
				Base([]string{"E"}, nil),
			}),
			"Result<T, E>",
		},

		// Nested generics
		{
			"Option<Vec<u32>>",
			Base([]string{"Option"}, []RustType{
				Base([]string{"Vec"}, []RustType{
					Base([]string{"u32"}, nil),
				}),
			}),
			"Option<Vec<u32>>",
		},
		{
			"Arc<Mutex<Vec<u8>>>",
			Base([]string{"Arc"}, []RustType{
				Base([]string{"Mutex"}, []RustType{
					Base([]string{"Vec"}, []RustType{
						Base([]string{"u8"}, nil),
					}),
				}),
			}),
			"Arc<Mutex<Vec<u8>>>",
		},

		// Paths with generics
		{
			"std::collections::HashMap<String, i32>",
			Base([]string{"std", "collections", "HashMap"}, []RustType{
				Base([]string{"String"}, nil),
				Base([]string{"i32"}, nil),
			}),
			"std::collections::HashMap<String, i32>",
		},
		{
			"foo::Foo<bar::baz::Baa<Boo>, Bee::Goo>",
			Base([]string{"foo", "Foo"}, []RustType{
				Base([]string{"bar", "baz", "Baa"}, []RustType{
					Base([]string{"Boo"}, nil),
				}),
				Base([]string{"Bee", "Goo"}, nil),
			}),
			"foo::Foo<bar::baz::Baa<Boo>, Bee::Goo>",
		},
		{
			"T::Faa<T>",
			Base([]string{"T", "Faa"}, []RustType{
				Base([]string{"T"}, nil),
			}),
			"T::Faa<T>",
		},

		// Tuples
		{
			"()",
			Tuple([]RustType{}),
			"()",
		},
		{
			"(u32)",
			Tuple([]RustType{
				Base([]string{"u32"}, nil),
			}),
			"(u32)",
		},
		{
			"(String, i32)",
			Tuple([]RustType{
				Base([]string{"String"}, nil),
				Base([]string{"i32"}, nil),
			}),
			"(String, i32)",
		},
		{
			"(T, U, V)",
			Tuple([]RustType{
				Base([]string{"T"}, nil),
				Base([]string{"U"}, nil),
				Base([]string{"V"}, nil),
			}),
			"(T, U, V)",
		},
		{
			"(Vec<u32>, HashMap<String, i32>)",
			Tuple([]RustType{
				Base([]string{"Vec"}, []RustType{
					Base([]string{"u32"}, nil),
				}),
				Base([]string{"HashMap"}, []RustType{
					Base([]string{"String"}, nil),
					Base([]string{"i32"}, nil),
				}),
			}),
			"(Vec<u32>, HashMap<String, i32>)",
		},

		// Arrays
		{
			"[u8; 32]",
			Array(Base([]string{"u8"}, nil), 32),
			"[u8; 32]",
		},
		{
			"[i32; 10]",
			Array(Base([]string{"i32"}, nil), 10),
			"[i32; 10]",
		},
		{
			"[String; 5]",
			Array(Base([]string{"String"}, nil), 5),
			"[String; 5]",
		},
		{
			"[Vec<u32>; 3]",
			Array(
				Base([]string{"Vec"}, []RustType{
					Base([]string{"u32"}, nil),
				}),
				3,
			),
			"[Vec<u32>; 3]",
		},
		{
			"[foo::Foo<Bar>; 12]",
			Array(
				Base([]string{"foo", "Foo"}, []RustType{
					Base([]string{"Bar"}, nil),
				}),
				12,
			),
			"[foo::Foo<Bar>; 12]",
		},

		// Complex nested combinations
		{
			"Vec<(String, i32)>",
			Base([]string{"Vec"}, []RustType{
				Tuple([]RustType{
					Base([]string{"String"}, nil),
					Base([]string{"i32"}, nil),
				}),
			}),
			"Vec<(String, i32)>",
		},
		{
			"HashMap<String, [u8; 32]>",
			Base([]string{"HashMap"}, []RustType{
				Base([]string{"String"}, nil),
				Array(Base([]string{"u8"}, nil), 32),
			}),
			"HashMap<String, [u8; 32]>",
		},
		{
			"[(String, i32); 10]",
			Array(
				Tuple([]RustType{
					Base([]string{"String"}, nil),
					Base([]string{"i32"}, nil),
				}),
				10,
			),
			"[(String, i32); 10]",
		},
		{
			"Option<[Vec<(u32, String)>; 5]>",
			Base([]string{"Option"}, []RustType{
				Array(
					Base([]string{"Vec"}, []RustType{
						Tuple([]RustType{
							Base([]string{"u32"}, nil),
							Base([]string{"String"}, nil),
						}),
					}),
					5,
				),
			}),
			"Option<[Vec<(u32, String)>; 5]>",
		},

		// Whitespace normalization
		{
			" Box < T > ",
			Base([]string{"Box"}, []RustType{
				Base([]string{"T"}, nil),
			}),
			"Box<T>",
		},
		{
			" HashMap < String ,  i32 > ",
			Base([]string{"HashMap"}, []RustType{
				Base([]string{"String"}, nil),
				Base([]string{"i32"}, nil),
			}),
			"HashMap<String, i32>",
		},
		{
			" ( String , i32 ) ",
			Tuple([]RustType{
				Base([]string{"String"}, nil),
				Base([]string{"i32"}, nil),
			}),
			"(String, i32)",
		},
		{
			" [ u8 ; 32 ] ",
			Array(Base([]string{"u8"}, nil), 32),
			"[u8; 32]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := NewRustTypesParser(tt.input)
			result, err := parser.Parse()

			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}

			// Test AST deep equality
			if diff := deepEqual(result, tt.expectedAST); diff != nil {
				t.Errorf("AST mismatch: %s", *diff)
			}

			// Test string reconstruction
			actualStr := result.String()

			// Verify round-trip capability (parse -> reconstruct)
			if actualStr != tt.expectedStr {
				t.Errorf("Round-trip failed: input %q -> output %q", tt.input, actualStr)
			}
		})
	}
}
