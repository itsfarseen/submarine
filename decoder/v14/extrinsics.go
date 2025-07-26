package v14

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	. "submarine/decoder/models"
	. "submarine/scale"
	"submarine/scale/base"
	"submarine/scale/gen/scaleInfo"
	"submarine/scale/gen/v14"
)

// DecodeExtrinsic is the main entry point for decoding an extrinsic.
// It uses the pre-decoded metadata to understand the structure of the bytes.
func DecodeExtrinsic(metadata *v14.Metadata, extrinsicBytes []byte) (*DecodedExtrinsic, error) {
	r := NewReader(extrinsicBytes)

	// An extrinsic is length-prefixed. We must decode this first to advance
	// the reader, even if we don't use the length value itself.
	n, err := DecodeCompact(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode extrinsic length prefix: %w", err)
	}

	// --- 1. Decode Extrinsic Wrapper ---
	// An extrinsic starts with a compact-encoded length of the payload.
	// We can skip this as we have the full byte slice.
	// The next byte describes the transaction format (version and signature presence).
	// For example, 0x84 means signed transaction with protocol version 4.
	txFormat, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read transaction format byte: %w", err)
	}

	log.Printf("extrinsic %d %x", n, txFormat)

	isSigned := (txFormat & 0b10000000) != 0
	var signatureData base.Signature

	if isSigned {
		// --- Correctly decode the extrinsic wrapper ---
		// The correct order is: Address, Signature, then Extra (all signed extensions).

		// 1. Decode the sender's Address. The type is given by metadata.
		_, err := base.DecodeAddress(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode sender address: %w", err)
		}

		// 3. Decode the Signature. This is a MultiSignature enum.
		signatureData, err = base.DecodeSignature(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode signature: %w", err)
		}

		// 2. Decode the data for ALL signed extensions (Era, Nonce, Tip, etc.).
		// This makes up the "SignedExtra" payload.
		for _, extension := range metadata.Extrinsic.SignedExtensions {
			_, err := DecodeArg(metadata, r, extension.Type)
			if err != nil {
				return nil, fmt.Errorf("failed to decode signed extension '%s': %w", extension.Identifier, err)
			}
		}

	}

	// --- 2. Decode the Call ---
	// The call is the actual payload we want to understand.
	// call, err := DecodeCall(metadata, r)
	call, err := DecodePalletVariant(metadata, r, "calls")

	if err != nil {
		return nil, fmt.Errorf("failed to decode call ext: %w", err)
	}

	return &DecodedExtrinsic{
		Signature: signatureData,
		Call:      *call,
	}, nil
}

// DecodeCall decodes the pallet index, call index, and the corresponding arguments.
func DecodeCall(metadata *v14.Metadata, r *Reader) (*DecodedCall, error) {
	// The call starts with the pallet index.
	palletIndex, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read pallet index: %w", err)
	}

	// The next byte is the call index within that pallet.
	callIndex, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read call index: %w", err)
	}

	// --- Find the Call Definition in Metadata ---
	var pallet v14.PalletMetadata
	foundPallet := false
	for _, p := range metadata.Pallets {
		if p.Index == palletIndex {
			pallet = p
			foundPallet = true
			break
		}
	}
	if !foundPallet {
		return nil, fmt.Errorf("pallet with index %d not found in metadata", palletIndex)
	}

	if pallet.Calls == nil {
		return nil, fmt.Errorf("pallet '%s' has no calls defined in metadata", pallet.Name)
	}

	// The `pallet.Calls.Type` is a SiLookupTypeId that points to a Variant type
	// in the lookup table, where each variant represents a call.
	callType, ok := findType(metadata, pallet.Calls.Type)
	if !ok {
		return nil, fmt.Errorf("call type definition for pallet '%s' not found", pallet.Name)
	}

	if callType.Def.Kind != scaleInfo.Si1TypeDefKindVariant {
		return nil, fmt.Errorf("expected call type to be a variant, but got kind %v", callType.Def.Kind)
	}
	callTypeDef := callType.Def.Variant

	var callVariant scaleInfo.Si1Variant
	foundCall := false
	for _, v := range callTypeDef.Variants {
		if v.Index == callIndex {
			callVariant = v
			foundCall = true
			break
		}
	}
	if !foundCall {
		return nil, fmt.Errorf("call with index %d not found in pallet '%s'", callIndex, pallet.Name)
	}

	// --- Decode Arguments ---
	decodedArgs := make([]DecodedArg, len(callVariant.Fields))
	for i, field := range callVariant.Fields {
		argName := "unnamed"
		if field.Name != nil {
			argName = *field.Name
		}

		// Decode the argument value based on its type ID.
		argValue, err := DecodeArg(metadata, r, field.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to decode arg '%s' for call '%s.%s': %w", argName, pallet.Name, callVariant.Name, err)
		}

		decodedArgs[i] = DecodedArg{
			Name:  argName,
			Value: argValue,
		}
	}

	return &DecodedCall{
		PalletName: pallet.Name,
		CallName:   callVariant.Name,
		Args:       decodedArgs,
	}, nil
}

// DecodeArg is a recursive function that decodes a value of any type
// by looking up its definition in the metadata.
func DecodeArg(metadata *v14.Metadata, r *Reader, typeID scaleInfo.Si1LookupTypeId) (any, error) {
	// Find the type definition in the lookup table.
	typ, ok := findType(metadata, typeID)
	if !ok {
		return nil, fmt.Errorf("type with ID %d not found in lookup table", typeID)
	}

	// Use a switch to handle the different kinds of types.
	switch typ.Def.Kind {
	case scaleInfo.Si1TypeDefKindComposite:
		// For a struct, decode each field recursively.
		// We'll represent it as a map.
		result := make(map[string]any)
		for _, field := range typ.Def.Composite.Fields {
			fieldName := "unnamed"
			if field.Name != nil {
				fieldName = *field.Name
			}
			fieldValue, err := DecodeArg(metadata, r, field.Type)
			if err != nil {
				var typeName string
				if field.TypeName != nil {
					typeName = *field.TypeName
				}
				return nil, fmt.Errorf("composite (%s: %s): %w", fieldName, typeName, err)
			}
			result[fieldName] = fieldValue
		}
		return result, nil

	case scaleInfo.Si1TypeDefKindVariant:
		// For an enum, read the variant index and decode its fields.
		variantIndex, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("variant index: %w", err)
		}
		for _, variant := range typ.Def.Variant.Variants {
			if variant.Index == variantIndex {
				// Similar to a composite, decode the fields.
				result := make(map[string]any)
				for _, field := range variant.Fields {
					fieldName := "unnamed"
					if field.Name != nil {
						fieldName = *field.Name
					}
					fieldValue, err := DecodeArg(metadata, r, field.Type)
					if err != nil {
						return nil, fmt.Errorf("variant (%d, %s): %w", variantIndex, fieldName, err)
					}
					result[fieldName] = fieldValue
				}
				// Return as a map with the variant name as the key.
				return map[string]any{string(variant.Name): result}, nil
			}
		}
		return nil, fmt.Errorf("variant with index %d not found for type %d", variantIndex, typeID)

	case scaleInfo.Si1TypeDefKindSequence:
		// For a sequence (Vec), decode the compact length then each item.
		length, err := DecodeCompact(r)
		if err != nil {
			return nil, fmt.Errorf("sequence length: %w", err)
		}
		len64 := length.Int64()
		slice := make([]any, len64)
		for i := range len64 {
			elem, err := DecodeArg(metadata, r, typ.Def.Sequence.Type)
			if err != nil {
				return nil, fmt.Errorf("sequence (%d): %w", i, err)
			}
			slice[i] = elem
		}
		return slice, nil

	case scaleInfo.Si1TypeDefKindArray:
		// For a fixed-size array, decode each item.
		slice := make([]any, typ.Def.Array.Len)
		for i := uint32(0); i < typ.Def.Array.Len; i++ {
			elem, err := DecodeArg(metadata, r, typ.Def.Array.Type)
			if err != nil {
				return nil, err
			}
			slice[i] = elem
		}
		return slice, nil

	case scaleInfo.Si1TypeDefKindTuple:
		// For a tuple, decode each item.
		slice := make([]any, len(typ.Def.Tuple.Fields))
		for i, fieldTypeID := range typ.Def.Tuple.Fields {
			elem, err := DecodeArg(metadata, r, fieldTypeID)
			if err != nil {
				return nil, err
			}
			slice[i] = elem
		}
		return slice, nil

	case scaleInfo.Si1TypeDefKindCompact:
		// For a compact integer, use the primitive decoder.
		return DecodeCompact(r)

	case scaleInfo.Si1TypeDefKindPrimitive:
		// For a primitive type, use the corresponding decoder.
		switch typ.Def.Primitive.Kind {
		case scaleInfo.Si1TypeDefPrimitiveKindBool: // Bool
			return DecodeBool(r)
		case scaleInfo.Si1TypeDefPrimitiveKindChar: // Char (decode as string of len 1)
			b, err := r.ReadBytes(1)
			return string(b), err
		case scaleInfo.Si1TypeDefPrimitiveKindStr: // Str
			return DecodeText(r)
		case scaleInfo.Si1TypeDefPrimitiveKindU8: // U8
			return DecodeU8(r)
		case scaleInfo.Si1TypeDefPrimitiveKindU16: // U16
			b, err := r.ReadBytes(2)
			return binary.LittleEndian.Uint16(b), err
		case scaleInfo.Si1TypeDefPrimitiveKindU32: // U32
			return DecodeU32(r)
		case scaleInfo.Si1TypeDefPrimitiveKindU64: // U64
			b, err := r.ReadBytes(8)
			return binary.LittleEndian.Uint64(b), err
		case scaleInfo.Si1TypeDefPrimitiveKindU128: // U128
			b, err := r.ReadBytes(16)
			// Reverse for big.Int
			for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
				b[i], b[j] = b[j], b[i]
			}
			return new(big.Int).SetBytes(b), err
		case scaleInfo.Si1TypeDefPrimitiveKindU256: // U256
			b, err := r.ReadBytes(32)
			// Reverse for big.Int
			for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
				b[i], b[j] = b[j], b[i]
			}
			return new(big.Int).SetBytes(b), err
		// Note: Signed integers (I8, I16, etc.) would follow a similar pattern.
		default:
			return nil, fmt.Errorf("unsupported primitive type: %d", typ.Def.Primitive.Kind)
		}

	case scaleInfo.Si1TypeDefKindBitSequence:
		// A bit sequence is encoded as a compact length (number of bits)
		// followed by the packed bits.
		numBits, err := DecodeCompact(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode bit sequence length: %w", err)
		}

		// Calculate the number of bytes needed to store the bits.
		numBytes := (numBits.Int64() + 7) / 8

		// Read the packed bytes.
		return r.ReadBytes(int(numBytes))

	default:
		return nil, fmt.Errorf("unsupported type definition %T for type ID %d", typ.Def, typeID)
	}
}
