package scale_schema

import (
	"fmt"
	. "submarine/errorspan"
)

func ParseType(rawTypeDef any) (*Type, *ErrorSpan) {
	switch v := rawTypeDef.(type) {
	case string:
		return &Type{Kind: KindRef, Ref: &v}, nil
	case map[string]any:
		return ParseComplexType(v)
	default:
		return nil, NewErrorSpan(fmt.Sprintf("unexpected type definition format: %T", rawTypeDef))
	}
}

func ParseComplexType(def map[string]any) (*Type, *ErrorSpan) {
	rawType, ok := def["type"].(string)
	if !ok {
		return nil, NewErrorSpan("missing 'type' or not a string")
	}

	t := &Type{}

	switch rawType {
	case "struct":
		t.Kind = KindStruct
		rawFields, ok := def["fields"].([]any)
		if !ok {
			return nil, NewErrorSpan("missing 'fields' or not a list").WithPath("struct")
		}

		members, err := ParseNamedMembers(rawFields)
		if err != nil {
			return nil, err.WithPath("struct.fields")
		}
		t.Struct = &Struct{Fields: members}
	case "tuple":
		t.Kind = KindTuple
		rawFields, ok := def["fields"].([]any)
		if !ok {
			return nil, NewErrorSpan("missing 'fields' or not a list").WithPath("tuple")
		}

		members, err := ParseTupleMembers(rawFields)
		if err != nil {
			return nil, err.WithPath("tuple.fields")
		}
		t.Tuple = &Tuple{Fields: members}

	case "enum_simple":
		t.Kind = KindEnumSimple
		rawVariants, ok := def["variants"].([]any)
		if !ok {
			return nil, NewErrorSpan("missing 'variants' or not a list").WithPath("enum_simple")
		}
		variants := make([]string, len(rawVariants))
		for i, v := range rawVariants {
			variants[i], ok = v.(string)
			if !ok {
				return nil, NewErrorSpan("variant not a string").
					WithPathInt(i).
					WithPath("enum_simple.variants")
			}

		}
		t.EnumSimple = &EnumSimple{Variants: variants}

	case "enum_complex":
		t.Kind = KindEnumComplex
		rawVariants, ok := def["variants"].([]any)
		if !ok {
			return nil, NewErrorSpan("missing 'variants' or not a list").WithPath("enum_complex")
		}
		variants, err := ParseNamedMembers(rawVariants)
		if err != nil {
			return nil, err.WithPath("enum_complex.variants")
		}
		t.EnumComplex = &EnumComplex{Variants: variants}

	case "import":
		t.Kind = KindImport
		module, ok := def["module"].(string)
		if !ok {
			return nil, NewErrorSpan("missing 'module' or not a string").
				WithPath("import")
		}
		item, ok := def["item"].(string)
		if !ok {
			return nil, NewErrorSpan("missing 'item' or not a string").
				WithPath("import")
		}
		t.Import = &Import{Module: module, Item: item}

	case "vec", "option", "array":
		itemDef, ok := def["item"]
		if !ok {
			return nil, NewErrorSpan("missing 'item'").WithPath(rawType)
		}
		itemType, err := ParseType(itemDef)
		if err != nil {
			return nil, err.WithPath(rawType)
		}
		switch rawType {
		case "vec":
			t.Kind = KindVec
			t.Vec = &Vec{Type: itemType}
		case "option":
			t.Kind = KindOption
			t.Option = &Option{Type: itemType}
		case "array":
			length, ok := def["len"].(int)
			if !ok {
				return nil, NewErrorSpan("missing 'len' or not an int").WithPath(rawType)
			}
			t.Kind = KindArray
			t.Array = &Array{Type: itemType, Len: length}
		default:
			panic("unreachable")
		}

	default:
		return nil, NewErrorSpan(fmt.Sprintf("unknown type: %s", rawType))
	}

	return t, nil
}

func ParseNamedMembers(rawNamedMembers []any) ([]NamedMember, *ErrorSpan) {
	members := make([]NamedMember, len(rawNamedMembers))
	for i, member := range rawNamedMembers {
		memberMap, ok := member.(map[string]any)
		if !ok {
			return nil, NewErrorSpan("member is not a map").WithPathInt(i)
		}
		name, ok := memberMap["name"].(string)
		if !ok {
			return nil, NewErrorSpan("missing 'name'").WithPathInt(i)
		}
		type_, ok := memberMap["type"]
		if !ok {
			return nil, NewErrorSpan("missing 'type'").WithPathInt(i)
		}

		memberType, err := ParseType(type_)
		if err != nil {
			return nil, err.WithPathInt(i)
		}
		members[i] = NamedMember{Name: name, Type: memberType}
	}
	return members, nil
}

func ParseTupleMembers(rawTupleMembers []any) ([]Type, *ErrorSpan) {
	members := make([]Type, len(rawTupleMembers))
	for i, type_ := range rawTupleMembers {
		memberType, err := ParseType(type_)
		if err != nil {
			return nil, err.WithPathInt(i)
		}
		members[i] = *memberType
	}
	return members, nil
}
