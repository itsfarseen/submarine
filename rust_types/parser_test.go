package rust_types_test

import (
	"reflect"
	"submarine/rust_types"
	"testing"
)

func TestRustTypesParser(t *testing.T) {
	tests := []struct {
		input       string
		expectedAST rust_types.RustType
		expectedStr string
	}{
		{
			"Box<T>",
			rust_types.Generic{Outer: "Box", Inner: rust_types.Ident{Name: "T"}},
			"Box<T>",
		},
		{
			"Option<Vec<u32>>",
			rust_types.Generic{
				Outer: "Option",
				Inner: rust_types.Generic{
					Outer: "Vec",
					Inner: rust_types.Ident{Name: "u32"},
				},
			},
			"Option<Vec<u32>>",
		},
		{
			"Foo::Bar",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "Foo"},
				Next:  rust_types.Ident{Name: "Bar"},
			},
			"Foo::Bar",
		},
		{
			"mod::Foo::Bar",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "mod"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "Foo"},
					Next:  rust_types.Ident{Name: "Bar"},
				},
			},
			"mod::Foo::Bar",
		},
		{
			"<Foo as Bar>",
			rust_types.AsTrait{
				Src:    rust_types.Ident{Name: "Foo"},
				Target: rust_types.Ident{Name: "Bar"},
			},
			"<Foo as Bar>",
		},
		{
			"<Foo as Bar::Baz>",
			rust_types.AsTrait{
				Src: rust_types.Ident{Name: "Foo"},
				Target: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "Bar"},
					Next:  rust_types.Ident{Name: "Baz"},
				},
			},
			"<Foo as Bar::Baz>",
		},
		{
			"String",
			rust_types.Ident{Name: "String"},
			"String",
		},
		{
			"Vec<String>",
			rust_types.Generic{Outer: "Vec", Inner: rust_types.Ident{Name: "String"}},
			"Vec<String>",
		},
		{
			"HashMap<String, i32>",
			rust_types.Generic{
				Outer: "HashMap",
				Inner: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "String"},
					Next:  rust_types.Ident{Name: "i32"},
				},
			},
			"HashMap<String::i32>",
		},
		{
			"std::collections::HashMap",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "std"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "collections"},
					Next:  rust_types.Ident{Name: "HashMap"},
				},
			},
			"std::collections::HashMap",
		},
		{
			"Arc<Mutex<Vec<u8>>>",
			rust_types.Generic{
				Outer: "Arc",
				Inner: rust_types.Generic{
					Outer: "Mutex",
					Inner: rust_types.Generic{
						Outer: "Vec",
						Inner: rust_types.Ident{Name: "u8"},
					},
				},
			},
			"Arc<Mutex<Vec<u8>>>",
		},
		{
			"Result<T, E>",
			rust_types.Generic{
				Outer: "Result",
				Inner: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "T"},
					Next:  rust_types.Ident{Name: "E"},
				},
			},
			"Result<T::E>",
		},
		{
			"<T as Clone>",
			rust_types.AsTrait{
				Src:    rust_types.Ident{Name: "T"},
				Target: rust_types.Ident{Name: "Clone"},
			},
			"<T as Clone>",
		},
		{
			"<Self as Iterator>",
			rust_types.AsTrait{
				Src:    rust_types.Ident{Name: "Self"},
				Target: rust_types.Ident{Name: "Iterator"},
			},
			"<Self as Iterator>",
		},
		{
			"Option<Box<dyn Error>>",
			rust_types.Generic{
				Outer: "Option",
				Inner: rust_types.Generic{
					Outer: "Box",
					Inner: rust_types.Assoc{
						Outer: rust_types.Ident{Name: "dyn"},
						Next:  rust_types.Ident{Name: "Error"},
					},
				},
			},
			"Option<Box<dyn::Error>>",
		},
		{
			"std::fmt::Display",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "std"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "fmt"},
					Next:  rust_types.Ident{Name: "Display"},
				},
			},
			"std::fmt::Display",
		},
		{
			"<Vec<T> as IntoIterator>",
			rust_types.AsTrait{
				Src: rust_types.Generic{
					Outer: "Vec",
					Inner: rust_types.Ident{Name: "T"},
				},
				Target: rust_types.Ident{Name: "IntoIterator"},
			},
			"<Vec<T> as IntoIterator>",
		},
		{
			"Rc<RefCell<T>>",
			rust_types.Generic{
				Outer: "Rc",
				Inner: rust_types.Generic{
					Outer: "RefCell",
					Inner: rust_types.Ident{Name: "T"},
				},
			},
			"Rc<RefCell<T>>",
		},
		{
			"tokio::sync::Mutex",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "tokio"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "sync"},
					Next:  rust_types.Ident{Name: "Mutex"},
				},
			},
			"tokio::sync::Mutex",
		},
		{
			"Pin<Box<Future>>",
			rust_types.Generic{
				Outer: "Pin",
				Inner: rust_types.Generic{
					Outer: "Box",
					Inner: rust_types.Ident{Name: "Future"},
				},
			},
			"Pin<Box<Future>>",
		},
		{
			"<dyn Fn as CallOnce>",
			rust_types.AsTrait{
				Src: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "dyn"},
					Next:  rust_types.Ident{Name: "Fn"},
				},
				Target: rust_types.Ident{Name: "CallOnce"},
			},
			"<dyn::Fn as CallOnce>",
		},
		{
			"BTreeMap<K, V>",
			rust_types.Generic{
				Outer: "BTreeMap",
				Inner: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "K"},
					Next:  rust_types.Ident{Name: "V"},
				},
			},
			"BTreeMap<K::V>",
		},
		{
			"core::mem::ManuallyDrop",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "core"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "mem"},
					Next:  rust_types.Ident{Name: "ManuallyDrop"},
				},
			},
			"core::mem::ManuallyDrop",
		},
		{
			"<str as ToOwned>",
			rust_types.AsTrait{
				Src:    rust_types.Ident{Name: "str"},
				Target: rust_types.Ident{Name: "ToOwned"},
			},
			"<str as ToOwned>",
		},
		{
			"PhantomData<T>",
			rust_types.Generic{Outer: "PhantomData", Inner: rust_types.Ident{Name: "T"}},
			"PhantomData<T>",
		},
		// Whitespace normalization test cases
		{
			" Box < T > ",
			rust_types.Generic{Outer: "Box", Inner: rust_types.Ident{Name: "T"}},
			"Box<T>",
		},
		{
			"Option\t<\tVec\t<\tu32\t>\t>",
			rust_types.Generic{
				Outer: "Option",
				Inner: rust_types.Generic{
					Outer: "Vec",
					Inner: rust_types.Ident{Name: "u32"},
				},
			},
			"Option<Vec<u32>>",
		},
		{
			"Foo\n::\nBar",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "Foo"},
				Next:  rust_types.Ident{Name: "Bar"},
			},
			"Foo::Bar",
		},
		{
			"  mod  ::  Foo  ::  Bar  ",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "mod"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "Foo"},
					Next:  rust_types.Ident{Name: "Bar"},
				},
			},
			"mod::Foo::Bar",
		},
		{
			"<\tFoo\tas\tBar\t>",
			rust_types.AsTrait{
				Src:    rust_types.Ident{Name: "Foo"},
				Target: rust_types.Ident{Name: "Bar"},
			},
			"<Foo as Bar>",
		},
		{
			"< Foo as Bar :: Baz >",
			rust_types.AsTrait{
				Src: rust_types.Ident{Name: "Foo"},
				Target: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "Bar"},
					Next:  rust_types.Ident{Name: "Baz"},
				},
			},
			"<Foo as Bar::Baz>",
		},
		{
			"\n\tString\n\t",
			rust_types.Ident{Name: "String"},
			"String",
		},
		{
			"Vec\n<\nString\n>",
			rust_types.Generic{Outer: "Vec", Inner: rust_types.Ident{Name: "String"}},
			"Vec<String>",
		},
		{
			" HashMap < String ,  i32 > ",
			rust_types.Generic{
				Outer: "HashMap",
				Inner: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "String"},
					Next:  rust_types.Ident{Name: "i32"},
				},
			},
			"HashMap<String::i32>",
		},
		{
			"std\t::\tcollections\t::\tHashMap",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "std"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "collections"},
					Next:  rust_types.Ident{Name: "HashMap"},
				},
			},
			"std::collections::HashMap",
		},
		{
			"Arc\n<\nMutex\n<\nVec\n<\nu8\n>\n>\n>",
			rust_types.Generic{
				Outer: "Arc",
				Inner: rust_types.Generic{
					Outer: "Mutex",
					Inner: rust_types.Generic{
						Outer: "Vec",
						Inner: rust_types.Ident{Name: "u8"},
					},
				},
			},
			"Arc<Mutex<Vec<u8>>>",
		},
		{
			" Result < T ,  E > ",
			rust_types.Generic{
				Outer: "Result",
				Inner: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "T"},
					Next:  rust_types.Ident{Name: "E"},
				},
			},
			"Result<T::E>",
		},
		{
			"<\nT\nas\nClone\n>",
			rust_types.AsTrait{
				Src:    rust_types.Ident{Name: "T"},
				Target: rust_types.Ident{Name: "Clone"},
			},
			"<T as Clone>",
		},
		{
			"< Self\tas\t Iterator >",
			rust_types.AsTrait{
				Src:    rust_types.Ident{Name: "Self"},
				Target: rust_types.Ident{Name: "Iterator"},
			},
			"<Self as Iterator>",
		},
		{
			"\tOption\t<\tBox\t<\tdyn\t::\tError\t>\t>",
			rust_types.Generic{
				Outer: "Option",
				Inner: rust_types.Generic{
					Outer: "Box",
					Inner: rust_types.Assoc{
						Outer: rust_types.Ident{Name: "dyn"},
						Next:  rust_types.Ident{Name: "Error"},
					},
				},
			},
			"Option<Box<dyn::Error>>",
		},
		{
			"  std  \n::  fmt  \n::  Display  ",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "std"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "fmt"},
					Next:  rust_types.Ident{Name: "Display"},
				},
			},
			"std::fmt::Display",
		},
		{
			"<\tVec\t<\tT\t>\tas\tIntoIterator\t>",
			rust_types.AsTrait{
				Src: rust_types.Generic{
					Outer: "Vec",
					Inner: rust_types.Ident{Name: "T"},
				},
				Target: rust_types.Ident{Name: "IntoIterator"},
			},
			"<Vec<T> as IntoIterator>",
		},
		{
			"Rc\n\t<\n\tRefCell\n\t<\n\tT\n\t>\n\t>",
			rust_types.Generic{
				Outer: "Rc",
				Inner: rust_types.Generic{
					Outer: "RefCell",
					Inner: rust_types.Ident{Name: "T"},
				},
			},
			"Rc<RefCell<T>>",
		},
		{
			" tokio  ::  sync  ::  Mutex ",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "tokio"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "sync"},
					Next:  rust_types.Ident{Name: "Mutex"},
				},
			},
			"tokio::sync::Mutex",
		},
		{
			"Pin\t<\nBox\t<\nFuture\t>\n>",
			rust_types.Generic{
				Outer: "Pin",
				Inner: rust_types.Generic{
					Outer: "Box",
					Inner: rust_types.Ident{Name: "Future"},
				},
			},
			"Pin<Box<Future>>",
		},
		{
			"<\n\tdyn\n\t::\n\tFn\n\tas\n\tCallOnce\n>",
			rust_types.AsTrait{
				Src: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "dyn"},
					Next:  rust_types.Ident{Name: "Fn"},
				},
				Target: rust_types.Ident{Name: "CallOnce"},
			},
			"<dyn::Fn as CallOnce>",
		},
		{
			"  BTreeMap  <  K  ,  V  >  ",
			rust_types.Generic{
				Outer: "BTreeMap",
				Inner: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "K"},
					Next:  rust_types.Ident{Name: "V"},
				},
			},
			"BTreeMap<K::V>",
		},
		{
			"\tcore\t::\n\tmem\t::\n\tManuallyDrop",
			rust_types.Assoc{
				Outer: rust_types.Ident{Name: "core"},
				Next: rust_types.Assoc{
					Outer: rust_types.Ident{Name: "mem"},
					Next:  rust_types.Ident{Name: "ManuallyDrop"},
				},
			},
			"core::mem::ManuallyDrop",
		},
		{
			"<  str  as  ToOwned  >",
			rust_types.AsTrait{
				Src:    rust_types.Ident{Name: "str"},
				Target: rust_types.Ident{Name: "ToOwned"},
			},
			"<str as ToOwned>",
		},
		{
			"\n\nPhantomData\n\n<\n\nT\n\n>\n\n",
			rust_types.Generic{Outer: "PhantomData", Inner: rust_types.Ident{Name: "T"}},
			"PhantomData<T>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parser := rust_types.NewRustTypesParser(tt.input)
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
