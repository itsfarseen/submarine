package v13

import (
	"fmt"
	. "submarine/scale"
	"submarine/scale/v13"
)

// DecodeArgFromString is a placeholder. It needs to be implemented.
// This is the hard part. I'll need to add support for common types.
func DecodeArgFromString(metadata *v13.Metadata, r *Reader, typeName string) (any, error) {
	// This is a simplified version. A real implementation would need a lot more types.
	// It also needs to handle complex types like Vec<T>, Option<T>, etc.
	// For now, I'll just implement a few primitives to get started.
	switch typeName {
	case "u8":
		return DecodeU8(r)
	case "u16":
		return r.ReadBytes(2) // Simplified
	case "u32":
		return DecodeU32(r)
	case "u64":
		return r.ReadBytes(8) // Simplified
	case "u128":
		return r.ReadBytes(16) // Simplified
	case "bool":
		return DecodeBool(r)
	case "Bytes":
		return DecodeBytes(r)
	case "AccountId": // Assuming AccountId is 32 bytes
		return r.ReadBytes(32)
	default:
		// This is where it gets tricky. We might have `Compact<Balance>` or `Vec<u8>`.
		// A proper implementation needs to parse these strings.
		// For now, we'll return an error for unsupported types.
		return nil, fmt.Errorf("unsupported type string '%s'", typeName)
	}
}
