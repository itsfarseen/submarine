package v14

import (
	"submarine/scale/gen/scaleInfo"
	"submarine/scale/gen/v14"
)

// findType is a helper to safely access the type from the lookup table.
func findType(metadata *v14.Metadata, typeID scaleInfo.Si1LookupTypeId) (scaleInfo.Si1Type, bool) {
	if int(typeID) > len(metadata.Lookup.Types) {
		return scaleInfo.Si1Type{}, false
	}
	// The ID in the PortableType struct is the actual ID. We need to find it.
	for _, pType := range metadata.Lookup.Types {
		if pType.Id == typeID {
			return pType.Type, true
		}
	}

	return scaleInfo.Si1Type{}, false
}
