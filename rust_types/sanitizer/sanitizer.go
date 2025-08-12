package sanitizer

import (
	"regexp"
	"strings"

	. "submarine/rust_types"
	"submarine/rust_types/parser"
)

func ParseAndSanitize(typeName string) RustType {
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

func ParseRustType(s string) RustType {
	rust_type, err := parser.NewRustTypesParser(s).Parse()
	if err != nil {
		panic(err)
	}
	return rust_type

}

func SanitizeRustType(rust_type RustType) RustType {
	switch rust_type.Kind {
	case KindBase:
		base := rust_type.Base
		if len(base.Path) == 1 {
			name := base.Path[0]
			switch name {
			case "Box":
				return SanitizeRustType(rust_type.Base.Generics[0])
			case "Compact":
				return Base([]string{"compact"}, nil)
			}

			if name == "PairOf" {
				return Tuple(base.Generics)
			}

			if fixed, found := strings.CutSuffix(name, "Of"); found {
				return Base([]string{fixed}, nil)
			}

			if fixed, found := strings.CutPrefix(name, "Bounded"); found {
				generics := base.Generics[0 : len(base.Generics)-1] // skip Size in BoundedVec<...,Size>
				return Base([]string{fixed}, generics)
			}

			if fixed, found := strings.CutPrefix(name, "Weak"); found {
				return Base([]string{fixed}, base.Generics)
			}

			if rust_type.String() == "String" {
				return Base([]string{"text"}, nil)
			}

			var newName string = name

			if newName == "VecDeque" {
				newName = "Vec"
			}

			// Handle Vec<T> and Option<T>, Foo<T, U, V> becomes Foo
			if newName == "Vec" || newName == "Option" {
				var newGenerics []RustType
				newGenerics = make([]RustType, len(base.Generics))
				for i := range base.Generics {
					newGenerics[i] = SanitizeRustType(base.Generics[i])
				}
				return Base([]string{newName}, newGenerics)
			} else {
				return Base([]string{newName}, nil)
			}
		}

		if base.Path[0] == "T" {
			return SanitizeRustType(Base(base.Path[1:], base.Generics))
		}

		if base.Path[len(base.Path)-1] == "PhantomData" {
			return Base([]string{"empty"}, nil)
		}

		// foo::bar::Baz<T, U, V> becomes foo::bar::Baz
		return Base(base.Path, nil)

	case KindArray:
		array := rust_type.Array
		return Array(SanitizeRustType(array.Base), array.Len)
	case KindTuple:
		tuple := rust_type.Tuple
		newTuple := make([]RustType, len(*tuple))
		for i := range *tuple {
			newTuple[i] = SanitizeRustType((*tuple)[i])
		}
		return Tuple(newTuple)

	default:
		panic("exhaustive")
	}
}
