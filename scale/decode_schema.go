package scale

import (
	"fmt"
	"math/big"
	. "submarine/errorspan"
)

func DecodeWithSchema(r *Reader, schema *Type) (Value, *ErrorSpan) {
	switch schema.Kind {
	case KindStruct:
		return decodeStruct(r, schema.Struct)
	case KindTuple:
		return decodeTuple(r, schema.Tuple)
	case KindEnumSimple:
		return decodeEnumSimple(r, schema.EnumSimple)
	case KindEnumComplex:
		return decodeEnumComplex(r, schema.EnumComplex)
	case KindVec:
		return decodeVec(r, schema.Vec)
	case KindOption:
		return decodeOption(r, schema.Option)
	case KindArray:
		return decodeArray(r, schema.Array)
	case KindRef:
		return decodeRef(r, *schema.Ref)
	case KindBitFlags:
		return decodeBitFlags(r, schema.BitFlags)
	case KindImport:
		return Value{}, NewErrorSpan(fmt.Sprintf("import types not supported: module: %s item: %s", schema.Import.Module, schema.Import.Item))
	default:
		return Value{}, NewErrorSpan(fmt.Sprintf("unknown type kind: %s", schema.Kind))
	}
}

func decodeRef(r *Reader, refType string) (Value, *ErrorSpan) {
	switch refType {
	case "u8":
		val, err := DecodeU8(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VIntFromInt64(int64(val)), nil
	case "u16":
		val, err := DecodeU16(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VIntFromInt64(int64(val)), nil
	case "u32":
		val, err := DecodeU32(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VIntFromInt64(int64(val)), nil
	case "u64":
		val, err := DecodeU64(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VInt(new(big.Int).SetUint64(val)), nil
	case "u128":
		val, err := DecodeU128(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VInt(val), nil
	case "u256":
		val, err := DecodeU256(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VInt(val), nil
	case "i8":
		val, err := DecodeI8(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VIntFromInt64(int64(val)), nil
	case "i16":
		val, err := DecodeI16(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VIntFromInt64(int64(val)), nil
	case "i32":
		val, err := DecodeI32(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VIntFromInt64(int64(val)), nil
	case "i64":
		val, err := DecodeI64(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VIntFromInt64(val), nil
	case "i128":
		val, err := DecodeI128(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VInt(val), nil
	case "i256":
		val, err := DecodeI256(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VInt(val), nil
	case "bool":
		val, err := DecodeBool(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VBool(val), nil
	case "text":
		val, err := DecodeText(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VText(val), nil
	case "bytes":
		val, err := DecodeBytes(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VBytes(val), nil
	case "compact":
		val, err := DecodeCompact(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VInt(val), nil
	case "empty": // Unit type
		return VNull(), nil
	default:
		return Value{}, NewErrorSpan(fmt.Sprintf("unknown primitive type: %s", refType))
	}
}

func decodeStruct(r *Reader, s *Struct) (Value, *ErrorSpan) {
	result := make(map[string]Value)
	for _, field := range s.Fields {
		value, err := DecodeWithSchema(r, field.Type)
		if err != nil {
			return Value{}, err.WithPath(field.Name)
		}
		result[field.Name] = value
	}
	return VStruct(result), nil
}

func decodeTuple(r *Reader, t *Tuple) (Value, *ErrorSpan) {
	result := make([]Value, len(t.Fields))
	for i, fieldType := range t.Fields {
		value, err := DecodeWithSchema(r, &fieldType)
		if err != nil {
			return Value{}, err.WithPathInt(i)
		}
		result[i] = value
	}
	return VList(result), nil
}

func decodeEnumSimple(r *Reader, e *EnumSimple) (Value, *ErrorSpan) {
	index, err := DecodeU8(r)
	if err != nil {
		return Value{}, NewErrorSpan(err.Error()).WithPath("index")
	}
	if int(index) >= len(e.Variants) {
		return Value{}, NewErrorSpan(fmt.Sprintf("enum index %d out of bounds (max %d)", index, len(e.Variants)-1)).WithPath("index")
	}
	return VText(e.Variants[index]), nil
}

func decodeEnumComplex(r *Reader, e *EnumComplex) (Value, *ErrorSpan) {
	index, err := DecodeU8(r)
	if err != nil {
		return Value{}, NewErrorSpan(err.Error()).WithPath("index")
	}
	if int(index) >= len(e.Variants) {
		return Value{}, NewErrorSpan(fmt.Sprintf("enum index %d out of bounds (max %d)", index, len(e.Variants)-1)).WithPath("index")
	}

	variant := e.Variants[index]
	value, err2 := DecodeWithSchema(r, variant.Type)
	if err2 != nil {
		return Value{}, err2.WithPath(variant.Name)
	}

	result := make(map[string]Value)
	result[variant.Name] = value
	return VStruct(result), nil
}

func decodeVec(r *Reader, v *Vec) (Value, *ErrorSpan) {
	// Optimization for Vec<u8>
	if v.Type.Kind == KindRef && v.Type.Ref != nil && *v.Type.Ref == "u8" {
		bytes, err := DecodeBytes(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VBytes(bytes), nil
	}

	length, err := DecodeCompact(r)
	if err != nil {
		return Value{}, NewErrorSpan(err.Error()).WithPath("length")
	}

	result := make([]Value, length.Int64())
	for i := range length.Int64() {
		value, err2 := DecodeWithSchema(r, v.Type)
		if err2 != nil {
			return Value{}, err2.WithPathInt(int(i))
		}
		result[i] = value
	}
	return VList(result), nil
}

func decodeOption(r *Reader, o *Option) (Value, *ErrorSpan) {
	hasValue, err := DecodeBool(r)
	if err != nil {
		return Value{}, NewErrorSpan(err.Error()).WithPath("flag")
	}

	if !hasValue {
		return VStruct(make(map[string]Value)), nil // Empty struct for None
	}

	return DecodeWithSchema(r, o.Type)
}

func decodeArray(r *Reader, a *Array) (Value, *ErrorSpan) {
	// Optimization for [u8; N]
	if a.Type.Kind == KindRef && a.Type.Ref != nil && *a.Type.Ref == "u8" {
		bytes, err := r.ReadBytes(a.Len)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
		return VBytes(bytes), nil
	}

	result := make([]Value, a.Len)
	for i := 0; i < a.Len; i++ {
		value, err := DecodeWithSchema(r, a.Type)
		if err != nil {
			return Value{}, err.WithPathInt(i)
		}
		result[i] = value
	}
	return VList(result), nil
}

func decodeBitFlags(r *Reader, bf *BitFlags) (Value, *ErrorSpan) {
	// Determine the appropriate integer type based on bit length
	var rawValue *big.Int
	var err error

	switch {
	case bf.BitLength <= 8:
		val, decodeErr := DecodeU8(r)
		if decodeErr != nil {
			return Value{}, NewErrorSpan(decodeErr.Error())
		}
		rawValue = big.NewInt(int64(val))
	case bf.BitLength <= 16:
		val, decodeErr := DecodeU16(r)
		if decodeErr != nil {
			return Value{}, NewErrorSpan(decodeErr.Error())
		}
		rawValue = big.NewInt(int64(val))
	case bf.BitLength <= 32:
		val, decodeErr := DecodeU32(r)
		if decodeErr != nil {
			return Value{}, NewErrorSpan(decodeErr.Error())
		}
		rawValue = big.NewInt(int64(val))
	case bf.BitLength <= 64:
		val, decodeErr := DecodeU64(r)
		if decodeErr != nil {
			return Value{}, NewErrorSpan(decodeErr.Error())
		}
		rawValue = new(big.Int).SetUint64(val)
	case bf.BitLength <= 128:
		rawValue, err = DecodeU128(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
	case bf.BitLength <= 256:
		rawValue, err = DecodeU256(r)
		if err != nil {
			return Value{}, NewErrorSpan(err.Error())
		}
	default:
		return Value{}, NewErrorSpan(fmt.Sprintf("unsupported bit length: %d", bf.BitLength))
	}

	// Create a struct with boolean fields for each flag
	result := make(map[string]Value)
	for _, flag := range bf.Flags {
		flagBig := new(big.Int).SetUint64(flag.Value)
		isSet := new(big.Int).And(rawValue, flagBig).Cmp(big.NewInt(0)) != 0
		result[flag.Name] = VBool(isSet)
	}

	return VStruct(result), nil
}
