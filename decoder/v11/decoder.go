package v11

import (
	"fmt"
	. "submarine/decoder/models"
	. "submarine/scale"
	"submarine/scale/gen/v11"
)

// DecodeCall decodes the pallet index, call index, and the corresponding arguments.
func DecodeCall(metadata *v11.Metadata, r *Reader) (*DecodedPalletVariant, error) {
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
	// In metadata v11 and older, the pallet index is the index in the filtered
	// list of pallets that actually have calls.
	callableModules := make([]v11.ModuleMetadata, 0)
	for _, p := range metadata.Modules {
		if p.Calls != nil {
			callableModules = append(callableModules, p)
		}
	}

	if int(palletIndex) >= len(callableModules) {
		return nil, fmt.Errorf("pallet with index %d not found in callable modules", palletIndex)
	}
	pallet := callableModules[palletIndex]

	if pallet.Calls == nil {
		return nil, fmt.Errorf("pallet '%s' has no calls defined in metadata", pallet.Name)
	}

	if int(callIndex) >= len(*pallet.Calls) {
		return nil, fmt.Errorf("call with index %d not found in pallet '%s'", callIndex, pallet.Name)
	}

	callVariant := (*pallet.Calls)[callIndex]

	// --- Decode Arguments ---
	decodedArgs := make([]DecodedArg, len(callVariant.Args))
	for i, arg := range callVariant.Args {
		argValue, err := DecodeArgFromString(metadata, r, arg.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to decode arg '%s' for call '%s.%s': %w", arg.Name, pallet.Name, callVariant.Name, err)
		}

		decodedArgs[i] = DecodedArg{
			Name:  arg.Name,
			Value: argValue,
		}
	}

	return &DecodedPalletVariant{
		PalletName:  pallet.Name,
		VariantName: callVariant.Name,
		Args:        decodedArgs,
	}, nil
}
