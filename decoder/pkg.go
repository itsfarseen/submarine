package decoder

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	. "submarine/scale"
)

// DecodedArg holds the name and decoded value of a single extrinsic argument.
type DecodedArg struct {
	Name  string
	Value any // Using `any` to hold various decoded types.
}

// DecodedCall represents the action part of an extrinsic.
type DecodedCall struct {
	PalletName string
	CallName   string
	Args       []DecodedArg
}

// DecodedExtrinsic represents the full decoded extrinsic.
// For this example, we focus on the Call, but a full implementation
// would also include signature, address, etc.
type DecodedExtrinsic struct {
	Signature MultiSignature
	Call      DecodedPalletVariant
}

// DecodeExtrinsic is the main entry point for decoding an extrinsic.
// It uses the pre-decoded metadata to understand the structure of the bytes.
func DecodeExtrinsic(metadata *MetadataV14, extrinsicBytes []byte) (*DecodedExtrinsic, error) {
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
	var signatureData MultiSignature

	if isSigned {
		// --- Correctly decode the extrinsic wrapper ---
		// The correct order is: Address, Signature, then Extra (all signed extensions).

		// 1. Decode the sender's Address. The type is given by metadata.
		_, err := DecodeMultiAddress(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode sender address: %w", err)
		}

		// 3. Decode the Signature. This is a MultiSignature enum.
		signatureData, err = DecodeMultiSignature(r)
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

// ===================================================================
// Call and Argument Decoding Logic
// ===================================================================

// DecodeCall decodes the pallet index, call index, and the corresponding arguments.
func DecodeCall(metadata *MetadataV14, r *Reader) (*DecodedCall, error) {
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
	var pallet PalletMetadataV14
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

	if !pallet.Calls.HasValue {
		return nil, fmt.Errorf("pallet '%s' has no calls defined in metadata", pallet.Name)
	}

	// The `pallet.Calls.Value.Type` is a SiLookupTypeId that points to a Variant type
	// in the lookup table, where each variant represents a call.
	callType, ok := findType(metadata, pallet.Calls.Value.Type)
	if !ok {
		return nil, fmt.Errorf("call type definition for pallet '%s' not found", pallet.Name)
	}

	callTypeDef, ok := callType.Def.(Si1TypeDefVariant)
	if !ok {
		return nil, fmt.Errorf("expected call type to be a variant, but got %T", callType.Def)
	}

	var callVariant Si1Variant
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
		if field.Name.HasValue {
			argName = string(field.Name.Value)
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
		PalletName: string(pallet.Name),
		CallName:   string(callVariant.Name),
		Args:       decodedArgs,
	}, nil
}

// DecodeArg is a recursive function that decodes a value of any type
// by looking up its definition in the metadata.
func DecodeArg(metadata *MetadataV14, r *Reader, typeID SiLookupTypeId) (any, error) {
	// Find the type definition in the lookup table.
	typ, ok := findType(metadata, typeID)
	if !ok {
		return nil, fmt.Errorf("type with ID %d not found in lookup table", typeID)
	}

	// Use a switch to handle the different kinds of types.
	switch def := typ.Def.(type) {
	case Si1TypeDefComposite:
		// For a struct, decode each field recursively.
		// We'll represent it as a map.
		result := make(map[string]any)
		for _, field := range def.Fields {
			fieldName := "unnamed"
			if field.Name.HasValue {
				fieldName = string(field.Name.Value)
			}
			fieldValue, err := DecodeArg(metadata, r, field.Type)
			if err != nil {
				return nil, fmt.Errorf("composite (%s: %s): %w", fieldName, field.TypeName.Value, err)
			}
			result[fieldName] = fieldValue
		}
		return result, nil

	case Si1TypeDefVariant:
		// For an enum, read the variant index and decode its fields.
		variantIndex, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("variant index: %w", err)
		}
		for _, variant := range def.Variants {
			if variant.Index == variantIndex {
				// Similar to a composite, decode the fields.
				result := make(map[string]any)
				for _, field := range variant.Fields {
					fieldName := "unnamed"
					if field.Name.HasValue {
						fieldName = string(field.Name.Value)
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

	case Si1TypeDefSequence:
		// For a sequence (Vec), decode the compact length then each item.
		length, err := DecodeCompact(r)
		if err != nil {
			return nil, fmt.Errorf("sequence length: %w", err)
		}
		len64 := length.Int64()
		slice := make([]any, len64)
		for i := range len64 {
			elem, err := DecodeArg(metadata, r, def.Type)
			if err != nil {
				return nil, fmt.Errorf("sequence (%d): %w", i, err)
			}
			slice[i] = elem
		}
		return slice, nil

	case Si1TypeDefArray:
		// For a fixed-size array, decode each item.
		slice := make([]any, def.Len)
		for i := uint32(0); i < def.Len; i++ {
			elem, err := DecodeArg(metadata, r, def.Type)
			if err != nil {
				return nil, err
			}
			slice[i] = elem
		}
		return slice, nil

	case Si1TypeDefTuple:
		// For a tuple, decode each item.
		slice := make([]any, len(def.Fields))
		for i, fieldTypeID := range def.Fields {
			elem, err := DecodeArg(metadata, r, fieldTypeID)
			if err != nil {
				return nil, err
			}
			slice[i] = elem
		}
		return slice, nil

	case Si1TypeDefCompact:
		// For a compact integer, use the primitive decoder.
		return DecodeCompact(r)

	case Si1TypeDefPrimitive:
		// For a primitive type, use the corresponding decoder.
		switch def {
		case 0: // Bool
			return DecodeBool(r)
		case 1: // Char (decode as string of len 1)
			b, err := r.ReadBytes(1)
			return string(b), err
		case 2: // Str
			return DecodeText(r)
		case 3: // U8
			return DecodeU8(r)
		case 4: // U16
			b, err := r.ReadBytes(2)
			return binary.LittleEndian.Uint16(b), err
		case 5: // U32
			return DecodeU32(r)
		case 6: // U64
			b, err := r.ReadBytes(8)
			return binary.LittleEndian.Uint64(b), err
		case 7: // U128
			b, err := r.ReadBytes(16)
			// Reverse for big.Int
			for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
				b[i], b[j] = b[j], b[i]
			}
			return new(big.Int).SetBytes(b), err
		case 8: // U256
			b, err := r.ReadBytes(32)
			// Reverse for big.Int
			for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
				b[i], b[j] = b[j], b[i]
			}
			return new(big.Int).SetBytes(b), err
		// Note: Signed integers (I8, I16, etc.) would follow a similar pattern.
		default:
			return nil, fmt.Errorf("unsupported primitive type: %d", def)
		}

	case Si1TypeDefBitSequence:
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

// findType is a helper to safely access the type from the lookup table.
func findType(metadata *MetadataV14, typeID SiLookupTypeId) (Si1Type, bool) {
	if int(typeID) > len(metadata.Lookup.Types) {
		return Si1Type{}, false
	}
	// The ID in the PortableTypeV14 struct is the actual ID. We need to find it.
	for _, pType := range metadata.Lookup.Types {
		if pType.Id == typeID {
			return pType.Type, true
		}
	}

	return Si1Type{}, false
}
