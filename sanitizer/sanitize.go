package sanitizer

import (
	"regexp"
	"strings"
	"submarine/sanitizer/rust_types"
)

var identRe *regexp.Regexp = regexp.MustCompile("[a-zA-Z0-9_]")

func Sanitize(typeName string) string {
	typeName = NormalizeSpaces(typeName)
	typeName = RemoveAsTrait(typeName)
	return SanitizeByAST(typeName)
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

func SanitizeByAST(s string) string {
	rust_type, err := rust_types.NewRustTypesParser(s).Parse()
	if err != nil {
		panic(err)
	}

	rust_type = SanitizeRustType(rust_type)
	return rust_type.String()
}

func SanitizeRustType(rust_type rust_types.RustType) rust_types.RustType {
	switch rust_type.Kind {
	case rust_types.KindBase:
		base := rust_type.Base
		if len(base.Path) == 1 {
			name := base.Path[0]
			switch name {
			case "Box":
				return SanitizeRustType(rust_type.Base.Generics[0])
			case "Compact":
				return rust_types.Base([]string{"compact"}, nil)
			}

			if name == "PairOf" {
				return rust_types.Tuple(base.Generics)
			}

			if fixed, found := strings.CutSuffix(name, "Of"); found {
				return rust_types.Base([]string{fixed}, nil)
			}

			if fixed, found := strings.CutPrefix(name, "Bounded"); found {
				generics := base.Generics[0 : len(base.Generics)-1] // skip Size in BoundedVec<...,Size>
				return rust_types.Base([]string{fixed}, generics)
			}

			if fixed, found := strings.CutPrefix(name, "Weak"); found {
				return rust_types.Base([]string{fixed}, base.Generics)
			}

			if rust_type.String() == "Vec<u8>" {
				return rust_types.Base([]string{"bytes"}, nil)
			}

			if rust_type.String() == "String" {
				return rust_types.Base([]string{"text"}, nil)
			}

			var newName string = name

			if newName == "VecDeque" {
				newName = "Vec"
			}

			// Handle Vec<T> and Option<T>, Foo<T, U, V> becomes Foo
			if newName == "Vec" || newName == "Option" {
				var newGenerics []rust_types.RustType
				newGenerics = make([]rust_types.RustType, len(base.Generics))
				for i := range base.Generics {
					newGenerics[i] = SanitizeRustType(base.Generics[i])
				}
				return rust_types.Base([]string{newName}, newGenerics)
			} else {
				return rust_types.Base([]string{newName}, nil)
			}
		}

		if base.Path[0] == "T" {
			return SanitizeRustType(rust_types.Base(base.Path[1:], base.Generics))
		}

		if base.Path[len(base.Path)-1] == "PhantomData" {
			return rust_types.Base([]string{"null"}, nil)
		}

		// foo::bar::Baz<T, U, V> becomes foo::bar::Baz
		return rust_types.Base(base.Path, nil)

	case rust_types.KindArray:
		array := rust_type.Array
		return rust_types.Array(SanitizeRustType(array.Base), array.Len)
	case rust_types.KindTuple:
		tuple := rust_type.Tuple
		newTuple := make([]rust_types.RustType, len(*tuple))
		for i := range *tuple {
			newTuple[i] = SanitizeRustType((*tuple)[i])
		}
		return rust_types.Tuple(newTuple)

	default:
		panic("exhaustive")
	}
}
