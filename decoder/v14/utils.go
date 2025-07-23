package v14

import (
	. "submarine/scale"
	"submarine/scale/v14"
)

// findType is a helper to safely access the type from the lookup table.
func findType(metadata *v14.Metadata, typeID SiLookupTypeId) (Si1Type, bool) {
	if int(typeID) > len(metadata.Lookup.Types) {
		return Si1Type{}, false
	}
	// The ID in the PortableType struct is the actual ID. We need to find it.
	for _, pType := range metadata.Lookup.Types {
		if pType.Id == typeID {
			return pType.Type, true
		}
	}

	return Si1Type{}, false
}
