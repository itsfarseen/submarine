package rust_types

import (
	"fmt"
	"strings"
)

type RustTypeKind int

const (
	KindBase RustTypeKind = iota
	KindTuple
	KindArray
)

type RustType struct {
	Kind  RustTypeKind
	Base  *RustTypeBase
	Tuple *[]RustType
	Array *RustTypeArray
}

type RustTypeBase struct {
	Path     []string
	Generics []RustType // can be empty if no generics
}

// types like [u8; 32] or [foo::Foo<Bar>; 12]
type RustTypeArray struct {
	Base RustType
	Len  int
}

// Constructor functions
func Base(path []string, generics []RustType) RustType {
	return RustType{
		Kind: KindBase,
		Base: &RustTypeBase{
			Path:     path,
			Generics: generics,
		},
	}
}

func Tuple(types []RustType) RustType {
	return RustType{
		Kind:  KindTuple,
		Tuple: &types,
	}
}

func Array(base RustType, len int) RustType {
	return RustType{
		Kind: KindArray,
		Array: &RustTypeArray{
			Base: base,
			Len:  len,
		},
	}
}

func (rt RustType) String() string {
	switch rt.Kind {
	case KindBase:
		if rt.Base == nil {
			return ""
		}
		path := strings.Join(rt.Base.Path, "::")
		if len(rt.Base.Generics) == 0 {
			return path
		}
		var params []string
		for _, param := range rt.Base.Generics {
			params = append(params, param.String())
		}
		return fmt.Sprintf("%s<%s>", path, strings.Join(params, ", "))
	case KindTuple:
		if rt.Tuple == nil {
			return "()"
		}
		var params []string
		for _, param := range *rt.Tuple {
			params = append(params, param.String())
		}
		return fmt.Sprintf("(%s)", strings.Join(params, ", "))
	case KindArray:
		if rt.Array == nil {
			return "[]"
		}
		return fmt.Sprintf("[%s; %d]", rt.Array.Base.String(), rt.Array.Len)
	default:
		return ""
	}
}
