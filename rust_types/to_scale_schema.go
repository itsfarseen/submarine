package rust_types

import (
	"strings"
	. "submarine/errorspan"
	s "submarine/scale"
)

func ToScaleSchema(rust_type *RustType) (s.Type, *ErrorSpan) {
	if rust_type == nil {
		return s.Type{}, NewErrorSpan("rust_type cannot be nil")
	}

	var ty s.Type

	switch rust_type.Kind {
	case KindArray:
		baseType, err := ToScaleSchema(&rust_type.Array.Base)
		if err != nil {
			return s.Type{}, err.WithPath("element")
		}
		ty = s.Type{
			Kind: s.KindArray,
			Array: &s.Array{
				Type: &baseType,
				Len:  rust_type.Array.Len,
			},
		}

	case KindBase:
		var err *ErrorSpan
		ty, err = convertBaseType(rust_type.Base)
		if err != nil {
			return s.Type{}, err
		}

	case KindTuple:
		if rust_type.Tuple == nil {
			// Empty tuple
			ty = s.Type{
				Kind: s.KindTuple,
				Tuple: &s.Tuple{
					Fields: []s.Type{},
				},
			}
		} else {
			fields := make([]s.Type, len(*rust_type.Tuple))
			for i, field := range *rust_type.Tuple {
				fieldType, err := ToScaleSchema(&field)
				if err != nil {
					return s.Type{}, err.WithPathInt(i)
				}
				fields[i] = fieldType
			}
			ty = s.Type{
				Kind: s.KindTuple,
				Tuple: &s.Tuple{
					Fields: fields,
				},
			}
		}

	default:
		return s.Type{}, NewErrorSpan("unexpected rust_types.RustTypeKind").
			WithPathf("kind=%d", int(rust_type.Kind))
	}

	return ty, nil
}

func convertBaseType(base *RustTypeBase) (s.Type, *ErrorSpan) {
	// Handle common built-in types and generic containers
	path := base.Path

	// Check for Vec<T>
	if len(path) == 1 && path[0] == "Vec" {
		if len(base.Generics) != 1 {
			return s.Type{}, NewErrorSpan("Vec requires exactly one generic parameter").
				WithPathf("generics_count=%d", len(base.Generics)).
				WithPath("Vec")
		}
		innerType, err := ToScaleSchema(&base.Generics[0])
		if err != nil {
			return s.Type{}, err.WithPath("generic_param").WithPath("Vec")
		}
		return s.Type{
			Kind: s.KindVec,
			Vec: &s.Vec{
				Type: &innerType,
			},
		}, nil
	}

	// Check for Option<T>
	if len(path) == 1 && path[0] == "Option" {
		if len(base.Generics) != 1 {
			return s.Type{}, NewErrorSpan("Option requires exactly one generic parameter").
				WithPathf("generics_count=%d", len(base.Generics)).
				WithPath("Option")
		}
		innerType, err := ToScaleSchema(&base.Generics[0])
		if err != nil {
			return s.Type{}, err.WithPath("generic_param").WithPath("Option")
		}
		return s.Type{
			Kind: s.KindOption,
			Option: &s.Option{
				Type: &innerType,
			},
		}, nil
	}

	// Handle generic types with parameters (but not Vec/Option)
	if len(base.Generics) > 0 {
		return s.Type{}, NewErrorSpan("generic types other than Vec and Option are not supported").
			WithPathf("type=%s", path[0]).
			WithPathf("generics_count=%d", len(base.Generics))
	}

	// All other types, treat as reference
	typeName := strings.Join(path, "::")
	return s.Type{
		Kind: s.KindRef,
		Ref:  &typeName,
	}, nil
}
