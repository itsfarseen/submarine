package v11

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	. "submarine/scale"
	"submarine/scale/system"
	"submarine/scale/gen/v11"
)

// DecodeArgFromString recursively decodes an argument based on its type string.
func DecodeArgFromString(metadata *v11.Metadata, r *Reader, typeName string) (any, error) {
	typeName = strings.TrimSpace(typeName)

	// Handle compact encoding wrapper
	if strings.HasPrefix(typeName, "Compact<") && strings.HasSuffix(typeName, ">") {
		return DecodeCompact(r)
	}

	// Handle vector wrapper
	if strings.HasPrefix(typeName, "Vec<") && strings.HasSuffix(typeName, ">") {
		innerTypeName := typeName[4 : len(typeName)-1]
		// Optimization for Vec<u8> which is decoded as Bytes
		if innerTypeName == "u8" {
			return DecodeBytes(r)
		}
		return DecodeVec(r, func(r *Reader) (any, error) {
			return DecodeArgFromString(metadata, r, innerTypeName)
		})
	}

	// Handle option wrapper
	if strings.HasPrefix(typeName, "Option<") && strings.HasSuffix(typeName, ">") {
		innerTypeName := typeName[7 : len(typeName)-1]
		return DecodeOption(r, func(r *Reader) (any, error) {
			return DecodeArgFromString(metadata, r, innerTypeName)
		})
	}

	// Handle tuple wrapper
	if strings.HasPrefix(typeName, "(") && strings.HasSuffix(typeName, ")") {
		innerTypesStr := typeName[1 : len(typeName)-1]
		// This is a simplified tuple parser. It won't handle nested complex types correctly.
		// e.g., (u32, Vec<(u8, u8)>) will fail.
		// But it should work for simple cases like (u32, bool).
		innerTypes := strings.Split(innerTypesStr, ",")
		result := make([]any, len(innerTypes))
		for i, innerType := range innerTypes {
			val, err := DecodeArgFromString(metadata, r, strings.TrimSpace(innerType))
			if err != nil {
				return nil, fmt.Errorf("failed to decode tuple element %d ('%s'): %w", i, innerType, err)
			}
			result[i] = val
		}
		return result, nil
	}

	// Handle fixed-size array
	if strings.HasPrefix(typeName, "[") && strings.HasSuffix(typeName, "]") {
		parts := strings.Split(strings.Trim(typeName, "[]"), ";")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid array type string: %s", typeName)
		}
		innerTypeName := strings.TrimSpace(parts[0])
		sizeStr := strings.TrimSpace(parts[1])
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid array size '%s': %w", sizeStr, err)
		}

		result := make([]any, size)
		for i := 0; i < size; i++ {
			val, err := DecodeArgFromString(metadata, r, innerTypeName)
			if err != nil {
				return nil, fmt.Errorf("failed to decode array element %d ('%s'): %w", i, innerTypeName, err)
			}
			result[i] = val
		}
		return result, nil
	}

	// Handle primitive and common types
	switch typeName {
	case "u8":
		return DecodeU8(r)
	case "u16":
		b, err := r.ReadBytes(2)
		if err != nil {
			return nil, err
		}
		return binary.LittleEndian.Uint16(b), nil
	case "u32", "BlockNumber", "LeasePeriod":
		return DecodeU32(r)
	case "u64":
		b, err := r.ReadBytes(8)
		if err != nil {
			return nil, err
		}
		return binary.LittleEndian.Uint64(b), nil
	case "u128", "Balance":
		return DecodeU128(r)
	case "bool":
		return DecodeBool(r)
	case "Bytes":
		return DecodeBytes(r)
	case "Text", "String", "Type":
		return DecodeText(r)
	case "AccountId", "AuthorityId": // Typically a 32-byte array
		return r.ReadBytes(32)
	case "H256", "Hash": // 32-byte hash
		return r.ReadBytes(32)
	case "AccountInfo":
		return system.DecodeAccountInfoWithTripleRefCount(r)
	case "DispatchResult":
		return system.DecodeDispatchOutcome(r)
	case "Weight":
		return system.DecodeWeight(r)
	case "Phase":
		return system.DecodePhase(r)
	case "EventRecord":
		return system.DecodeEventRecord(r)
	case "LastRuntimeUpgradeInfo":
		return system.DecodeLastRuntimeUpgradeInfo(r)
	case "BlockLength":
		return system.DecodeBlockLength(r)
	case "BlockWeights":
		return system.DecodeBlockWeights(r)
	case "DispatchInfo":
		return system.DecodeDispatchInfo(r)
	case "DispatchError":
		return system.DecodeDispatchError(r)
	case "DispatchResultOf":
		return system.DecodeDispatchOutcome(r)
	case "RawOrigin":
		return system.DecodeRawOrigin(r)
	default:
		// This is where it gets tricky. We might have `T::AccountId` or other complex types.
		// A proper implementation would need to look up these types in the runtime,
		// but v11 metadata doesn't provide a comprehensive type registry like v14.
		return nil, fmt.Errorf("unsupported type string '%s'", typeName)
	}
}