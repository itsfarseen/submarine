package rust_types

import (
	"strings"
	. "submarine/errorspan"
	"submarine/scale_schema"
)

func ToScaleSchema(rust_type *RustType) (scale_schema.Type, *ErrorSpan) {
	if rust_type == nil {
		return scale_schema.Type{}, NewErrorSpan("rust_type cannot be nil")
	}

	var ty scale_schema.Type

	switch rust_type.Kind {
	case KindArray:
		baseType, err := ToScaleSchema(&rust_type.Array.Base)
		if err != nil {
			return scale_schema.Type{}, err.WithPath("element")
		}
		ty = scale_schema.Type{
			Kind: scale_schema.KindArray,
			Array: &scale_schema.Array{
				Type: &baseType,
				Len:  rust_type.Array.Len,
			},
		}

	case KindBase:
		var err *ErrorSpan
		ty, err = convertBaseType(rust_type.Base)
		if err != nil {
			return scale_schema.Type{}, err
		}

	case KindTuple:
		if rust_type.Tuple == nil {
			// Empty tuple
			ty = scale_schema.Type{
				Kind: scale_schema.KindTuple,
				Tuple: &scale_schema.Tuple{
					Fields: []scale_schema.Type{},
				},
			}
		} else {
			fields := make([]scale_schema.Type, len(*rust_type.Tuple))
			for i, field := range *rust_type.Tuple {
				fieldType, err := ToScaleSchema(&field)
				if err != nil {
					return scale_schema.Type{}, err.WithPathInt(i)
				}
				fields[i] = fieldType
			}
			ty = scale_schema.Type{
				Kind: scale_schema.KindTuple,
				Tuple: &scale_schema.Tuple{
					Fields: fields,
				},
			}
		}

	default:
		return scale_schema.Type{}, NewErrorSpan("unexpected rust_types.RustTypeKind").
			WithPathf("kind=%d", int(rust_type.Kind))
	}

	return ty, nil
}

func convertBaseType(base *RustTypeBase) (scale_schema.Type, *ErrorSpan) {
	// Handle common built-in types and generic containers
	path := base.Path

	// Check for Vec<T>
	if len(path) == 1 && path[0] == "Vec" {
		if len(base.Generics) != 1 {
			return scale_schema.Type{}, NewErrorSpan("Vec requires exactly one generic parameter").
				WithPathf("generics_count=%d", len(base.Generics)).
				WithPath("Vec")
		}
		innerType, err := ToScaleSchema(&base.Generics[0])
		if err != nil {
			return scale_schema.Type{}, err.WithPath("generic_param").WithPath("Vec")
		}
		return scale_schema.Type{
			Kind: scale_schema.KindVec,
			Vec: &scale_schema.Vec{
				Type: &innerType,
			},
		}, nil
	}

	// Check for Option<T>
	if len(path) == 1 && path[0] == "Option" {
		if len(base.Generics) != 1 {
			return scale_schema.Type{}, NewErrorSpan("Option requires exactly one generic parameter").
				WithPathf("generics_count=%d", len(base.Generics)).
				WithPath("Option")
		}
		innerType, err := ToScaleSchema(&base.Generics[0])
		if err != nil {
			return scale_schema.Type{}, err.WithPath("generic_param").WithPath("Option")
		}
		return scale_schema.Type{
			Kind: scale_schema.KindOption,
			Option: &scale_schema.Option{
				Type: &innerType,
			},
		}, nil
	}

	// Handle generic types with parameters (but not Vec/Option)
	if len(base.Generics) > 0 {
		return scale_schema.Type{}, NewErrorSpan("generic types other than Vec and Option are not supported").
			WithPathf("type=%s", path[0]).
			WithPathf("generics_count=%d", len(base.Generics))
	}

	// All other types, treat as reference
	typeName := strings.Join(path, "::")
	return scale_schema.Type{
		Kind: scale_schema.KindRef,
		Ref:  &typeName,
	}, nil
}
