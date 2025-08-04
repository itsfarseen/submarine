package rust_types_test

import (
	"fmt"
	. "submarine/rust_types"
	"testing"
)

// deepEqual compares two RustType values for structural equality
// Returns nil if equal, or a string describing the first difference found
func deepEqual(a, b RustType) *string {
	return deepEqualPath(a, b, "")
}

func deepEqualPath(a, b RustType, path string) *string {
	if a == nil && b == nil {
		return nil
	}
	if a == nil {
		diff := fmt.Sprintf("%s: nil vs %T", path, b)
		return &diff
	}
	if b == nil {
		diff := fmt.Sprintf("%s: %T vs nil", path, a)
		return &diff
	}

	switch at := a.(type) {
	case Ident:
		if bt, ok := b.(Ident); ok {
			if at.Name != bt.Name {
				diff := fmt.Sprintf("%s.Name: %q vs %q", path, at.Name, bt.Name)
				return &diff
			}
			return nil
		}
		diff := fmt.Sprintf("%s: type mismatch Ident vs %T", path, b)
		return &diff

	case Generic:
		if bt, ok := b.(Generic); ok {
			if at.Outer != bt.Outer {
				diff := fmt.Sprintf("%s.Outer: %q vs %q", path, at.Outer, bt.Outer)
				return &diff
			}
			if len(at.Params) != len(bt.Params) {
				diff := fmt.Sprintf("%s.Params: length %d vs %d", path, len(at.Params), len(bt.Params))
				return &diff
			}
			for i := range at.Params {
				paramPath := fmt.Sprintf("%s.Params[%d]", path, i)
				if diff := deepEqualPath(at.Params[i], bt.Params[i], paramPath); diff != nil {
					return diff
				}
			}
			return nil
		}
		diff := fmt.Sprintf("%s: type mismatch Generic vs %T", path, b)
		return &diff

	case Assoc:
		if bt, ok := b.(Assoc); ok {
			if diff := deepEqualPath(at.Outer, bt.Outer, path+".Outer"); diff != nil {
				return diff
			}
			if diff := deepEqualPath(at.Next, bt.Next, path+".Next"); diff != nil {
				return diff
			}
			return nil
		}
		diff := fmt.Sprintf("%s: type mismatch Assoc vs %T", path, b)
		return &diff

	case AsTrait:
		if bt, ok := b.(AsTrait); ok {
			if diff := deepEqualPath(at.Src, bt.Src, path+".Src"); diff != nil {
				return diff
			}
			if diff := deepEqualPath(at.Target, bt.Target, path+".Target"); diff != nil {
				return diff
			}
			return nil
		}
		diff := fmt.Sprintf("%s: type mismatch AsTrait vs %T", path, b)
		return &diff
	}

	diff := fmt.Sprintf("%s: unknown type %T", path, a)
	return &diff
}

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
