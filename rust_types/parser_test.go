package rust_types_test

import (
	"reflect"
	. "submarine/rust_types"
	"testing"
)

func TestRustTypesParser(t *testing.T) {
	tests := []struct {
		input       string
		expectedAST RustType
		expectedStr string
	}{
		{
			"String",
			Ident{Name: "String"},
			"String",
		},
		{
			"Box<T>",
			Generic{Outer: "Box", Params: []RustType{Ident{Name: "T"}}},
			"Box<T>",
		},
		{
			"Option<Vec<u32>>",
			Generic{
				Outer: "Option",
				Params: []RustType{
					Generic{
						Outer:  "Vec",
						Params: []RustType{Ident{Name: "u32"}},
					},
				},
			},
			"Option<Vec<u32>>",
		},
		{
			"HashMap<String, i32>",
			Generic{
				Outer:  "HashMap",
				Params: []RustType{Ident{Name: "String"}, Ident{Name: "i32"}},
			},
			"HashMap<String, i32>",
		},
		{
			"Result<T, E>",
			Generic{
				Outer:  "Result",
				Params: []RustType{Ident{Name: "T"}, Ident{Name: "E"}},
			},
			"Result<T, E>",
		},
		{
			"Foo::Bar",
			Assoc{
				Outer: Ident{Name: "Foo"},
				Next:  Ident{Name: "Bar"},
			},
			"Foo::Bar",
		},
		{
			"std::collections::HashMap",
			Assoc{
				Outer: Ident{Name: "std"},
				Next: Assoc{
					Outer: Ident{Name: "collections"},
					Next:  Ident{Name: "HashMap"},
				},
			},
			"std::collections::HashMap",
		},
		{
			"<Foo as Bar>",
			AsTrait{
				Src:    Ident{Name: "Foo"},
				Target: Ident{Name: "Bar"},
			},
			"<Foo as Bar>",
		},
		{
			"<Foo as Bar::Baz>",
			AsTrait{
				Src: Ident{Name: "Foo"},
				Target: Assoc{
					Outer: Ident{Name: "Bar"},
					Next:  Ident{Name: "Baz"},
				},
			},
			"<Foo as Bar::Baz>",
		},
		{
			"<Vec<T> as IntoIterator>",
			AsTrait{
				Src: Generic{
					Outer:  "Vec",
					Params: []RustType{Ident{Name: "T"}},
				},
				Target: Ident{Name: "IntoIterator"},
			},
			"<Vec<T> as IntoIterator>",
		},
		{
			"Arc<Mutex<Vec<u8>>>",
			Generic{
				Outer: "Arc",
				Params: []RustType{
					Generic{
						Outer: "Mutex",
						Params: []RustType{
							Generic{
								Outer:  "Vec",
								Params: []RustType{Ident{Name: "u8"}},
							},
						},
					},
				},
			},
			"Arc<Mutex<Vec<u8>>>",
		},
		// Whitespace normalization test cases
		{
			" Box < T > ",
			Generic{Outer: "Box", Params: []RustType{Ident{Name: "T"}}},
			"Box<T>",
		},
		{
			" HashMap < String ,  i32 > ",
			Generic{
				Outer:  "HashMap",
				Params: []RustType{Ident{Name: "String"}, Ident{Name: "i32"}},
			},
			"HashMap<String, i32>",
		},
		// Complex type demonstrating all parser features
		{
			"std::sync::Arc<Mutex<HashMap<String, <Vec<T> as IntoIterator>::Item>>>",
			Assoc{
				Outer: Ident{Name: "std"},
				Next: Assoc{
					Outer: Ident{Name: "sync"},
					Next: Generic{
						Outer: "Arc",
						Params: []RustType{
							Generic{
								Outer: "Mutex",
								Params: []RustType{
									Generic{
										Outer: "HashMap",
										Params: []RustType{
											Ident{Name: "String"},
											Assoc{
												Outer: AsTrait{
													Src: Generic{
														Outer:  "Vec",
														Params: []RustType{Ident{Name: "T"}},
													},
													Target: Ident{Name: "IntoIterator"},
												},
												Next: Ident{Name: "Item"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"std::sync::Arc<Mutex<HashMap<String, <Vec<T> as IntoIterator>::Item>>>",
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
			if !reflect.DeepEqual(result, tt.expectedAST) {
				t.Errorf("AST mismatch:\nGot:      %+v\nExpected: %+v", result, tt.expectedAST)
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
