package rust_types

import (
	"regexp"
	"strings"
	"submarine/rust_types/parser"
)

func ParseAndSanitize(typeName string) parser.RustType {
	// We need to normalize spaces here, even though the Parser does normalize spaces, for RemoveAsTrait() to work.
	typeName = NormalizeSpaces(typeName)

	// This is done outside of the AST because the as trait syntax makes parsing way more complicated.
	// If we make as trait a type, we'd have to change RustTypeBase.Path from []string to []RustType
	// This makes parsing ambiguous. We'd have to introduce associativity to fix it, but then it makes the code more complex.
	// The current parser is very simple, and the parser output is really easy to manipulate.
	typeName = RemoveAsTrait(typeName)

	rust_type := ParseRustType(typeName)
	rust_type = SanitizeRustType(rust_type)
	return rust_type
}

// 'Foo\nBar  Baz' -> 'Foo Bar Baz'
func NormalizeSpaces(s string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllLiteralString(s, " ")
}

// <Foo as Trait>::Bar -> Foo::Bar
func RemoveAsTrait(s string) string {
	for {
		i := strings.Index(s, ">::")
		if i <= 0 {
			break
		}
		j := i
		open_count := 0
		for j = i; j >= 0; j -= 1 { // iterate backwards from i
			switch s[j] {
			case '>':
				open_count += 1
			case '<':
				open_count -= 1
			}
			if open_count == 0 {
				break
			}
		}
		before := s[:j]             // until '<'
		insideAsTrait := s[j+1 : i] // inside '<' '>', excluding
		after := s[i+1:]            // after '>'

		asTraitSrc, _, _ := strings.Cut(insideAsTrait, " as ")
		s = before + asTraitSrc + after
	}
	return s
}

func ParseRustType(s string) parser.RustType {
	rust_type, err := parser.NewRustTypesParser(s).Parse()
	if err != nil {
		panic(err)
	}
	return rust_type

}

func SanitizeRustType(rust_type parser.RustType) parser.RustType {
	switch rust_type.Kind {
	case parser.KindBase:
		base := rust_type.Base
		if len(base.Path) == 1 {
			name := base.Path[0]
			switch name {
			case "Box":
				return SanitizeRustType(rust_type.Base.Generics[0])
			case "Compact":
				return parser.Base([]string{"compact"}, nil)
			}

			if name == "PairOf" {
				return parser.Tuple(base.Generics)
			}

			if fixed, found := strings.CutSuffix(name, "Of"); found {
				return parser.Base([]string{fixed}, nil)
			}

			if fixed, found := strings.CutPrefix(name, "Bounded"); found {
				generics := base.Generics[0 : len(base.Generics)-1] // skip Size in BoundedVec<...,Size>
				return parser.Base([]string{fixed}, generics)
			}

			if fixed, found := strings.CutPrefix(name, "Weak"); found {
				return parser.Base([]string{fixed}, base.Generics)
			}

			if rust_type.String() == "Vec<u8>" {
				return parser.Base([]string{"bytes"}, nil)
			}

			if rust_type.String() == "String" {
				return parser.Base([]string{"text"}, nil)
			}

			var newName string = name

			if newName == "VecDeque" {
				newName = "Vec"
			}

			// Handle Vec<T> and Option<T>, Foo<T, U, V> becomes Foo
			if newName == "Vec" || newName == "Option" {
				var newGenerics []parser.RustType
				newGenerics = make([]parser.RustType, len(base.Generics))
				for i := range base.Generics {
					newGenerics[i] = SanitizeRustType(base.Generics[i])
				}
				return parser.Base([]string{newName}, newGenerics)
			} else {
				return parser.Base([]string{newName}, nil)
			}
		}

		if base.Path[0] == "T" {
			return SanitizeRustType(parser.Base(base.Path[1:], base.Generics))
		}

		if base.Path[len(base.Path)-1] == "PhantomData" {
			return parser.Base([]string{"null"}, nil)
		}

		// foo::bar::Baz<T, U, V> becomes foo::bar::Baz
		return parser.Base(base.Path, nil)

	case parser.KindArray:
		array := rust_type.Array
		return parser.Array(SanitizeRustType(array.Base), array.Len)
	case parser.KindTuple:
		tuple := rust_type.Tuple
		newTuple := make([]parser.RustType, len(*tuple))
		for i := range *tuple {
			newTuple[i] = SanitizeRustType((*tuple)[i])
		}
		return parser.Tuple(newTuple)

	default:
		panic("exhaustive")
	}
}
