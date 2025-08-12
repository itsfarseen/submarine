package decoder

import (
	"fmt"
	. "submarine/errorspan"
	"submarine/scale"
	s "submarine/scale/schema"
)

func DecodeWithSchema(r *scale.Reader, schema *s.Type) (any, *ErrorSpan) {
	switch schema.Kind {
	case s.KindStruct:
		return decodeStruct(r, schema.Struct)
	case s.KindTuple:
		return decodeTuple(r, schema.Tuple)
	case s.KindEnumSimple:
		return decodeEnumSimple(r, schema.EnumSimple)
	case s.KindEnumComplex:
		return decodeEnumComplex(r, schema.EnumComplex)
	case s.KindVec:
		return decodeVec(r, schema.Vec)
	case s.KindOption:
		return decodeOption(r, schema.Option)
	case s.KindArray:
		return decodeArray(r, schema.Array)
	case s.KindRef:
		return decodeRef(r, *schema.Ref)
	case s.KindImport:
		return nil, NewErrorSpan(fmt.Sprintf("import types not supported: %s.%s", schema.Import.Module, schema.Import.Item))
	default:
		return nil, NewErrorSpan(fmt.Sprintf("unknown type kind: %s", schema.Kind))
	}
}

func decodeRef(r *scale.Reader, refType string) (any, *ErrorSpan) {
	var val any
	var err error

	switch refType {
	case "u8":
		val, err = scale.DecodeU8(r)
	case "u16":
		val, err = scale.DecodeU16(r)
	case "u32":
		val, err = scale.DecodeU32(r)
	case "u64":
		val, err = scale.DecodeU64(r)
	case "u128":
		val, err = scale.DecodeU128(r)
	case "u256":
		val, err = scale.DecodeU256(r)
	case "i8":
		val, err = scale.DecodeI8(r)
	case "i16":
		val, err = scale.DecodeI16(r)
	case "i32":
		val, err = scale.DecodeI32(r)
	case "i64":
		val, err = scale.DecodeI64(r)
	case "i128":
		val, err = scale.DecodeI128(r)
	case "i256":
		val, err = scale.DecodeI256(r)
	case "bool":
		val, err = scale.DecodeBool(r)
	case "text":
		val, err = scale.DecodeText(r)
	case "bytes":
		val, err = scale.DecodeBytes(r)
	case "compact":
		val, err = scale.DecodeCompact(r)
	case "empty": // Unit type
		return nil, nil
	default:
		return nil, NewErrorSpan(fmt.Sprintf("unknown primitive type: %s", refType))
	}

	if err != nil {
		return nil, NewErrorSpan(err.Error())
	}

	return val, nil
}

func decodeStruct(r *scale.Reader, s *s.Struct) (map[string]any, *ErrorSpan) {
	result := make(map[string]any)
	for _, field := range s.Fields {
		value, err := DecodeWithSchema(r, field.Type)
		if err != nil {
			return nil, err.WithPath(field.Name)
		}
		result[field.Name] = value
	}
	return result, nil
}

func decodeTuple(r *scale.Reader, t *s.Tuple) ([]any, *ErrorSpan) {
	result := make([]any, len(t.Fields))
	for i, fieldType := range t.Fields {
		value, err := DecodeWithSchema(r, &fieldType)
		if err != nil {
			return nil, err.WithPathInt(i)
		}
		result[i] = value
	}
	return result, nil
}

func decodeEnumSimple(r *scale.Reader, e *s.EnumSimple) (string, *ErrorSpan) {
	index, err := scale.DecodeU8(r)
	if err != nil {
		return "", NewErrorSpan(err.Error()).WithPath("index")
	}
	if int(index) >= len(e.Variants) {
		return "", NewErrorSpan(fmt.Sprintf("enum index %d out of bounds (max %d)", index, len(e.Variants)-1)).WithPath("index")
	}
	return e.Variants[index], nil
}

func decodeEnumComplex(r *scale.Reader, e *s.EnumComplex) (map[string]any, *ErrorSpan) {
	index, err := scale.DecodeU8(r)
	if err != nil {
		return nil, NewErrorSpan(err.Error()).WithPath("index")
	}
	if int(index) >= len(e.Variants) {
		return nil, NewErrorSpan(fmt.Sprintf("enum index %d out of bounds (max %d)", index, len(e.Variants)-1)).WithPath("index")
	}

	variant := e.Variants[index]
	value, err2 := DecodeWithSchema(r, variant.Type)
	if err2 != nil {
		return nil, err2.WithPath(variant.Name)
	}

	return map[string]any{variant.Name: value}, nil
}

func decodeVec(r *scale.Reader, v *s.Vec) (any, *ErrorSpan) {
	// Optimization for Vec<u8>
	if v.Type.Kind == s.KindRef && v.Type.Ref != nil && *v.Type.Ref == "u8" {
		bytes, err := scale.DecodeBytes(r)
		if err != nil {
			return nil, NewErrorSpan(err.Error())
		}
		return bytes, nil
	}

	length, err := scale.DecodeCompact(r)
	if err != nil {
		return nil, NewErrorSpan(err.Error()).WithPath("length")
	}

	result := make([]any, length.Int64())
	for i := range length.Int64() {
		value, err2 := DecodeWithSchema(r, v.Type)
		if err2 != nil {
			return nil, err2.WithPathInt(int(i))
		}
		result[i] = value
	}
	return result, nil
}

func decodeOption(r *scale.Reader, o *s.Option) (any, *ErrorSpan) {
	hasValue, err := scale.DecodeBool(r)
	if err != nil {
		return nil, NewErrorSpan(err.Error()).WithPath("flag")
	}

	if !hasValue {
		return nil, nil
	}

	return DecodeWithSchema(r, o.Type)
}

func decodeArray(r *scale.Reader, a *s.Array) (any, *ErrorSpan) {
	// Optimization for [u8; N]
	if a.Type.Kind == s.KindRef && a.Type.Ref != nil && *a.Type.Ref == "u8" {
		bytes, err := r.ReadBytes(a.Len)
		if err != nil {
			return nil, NewErrorSpan(err.Error())
		}
		return bytes, nil
	}

	result := make([]any, a.Len)
	for i := 0; i < a.Len; i++ {
		value, err := DecodeWithSchema(r, a.Type)
		if err != nil {
			return nil, err.WithPathInt(i)
		}
		result[i] = value
	}
	return result, nil
}
