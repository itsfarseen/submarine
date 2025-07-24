package v9

import (
	"fmt"
	. "submarine/decoder/models"
	. "submarine/scale"
	"submarine/scale/v9"
)

// DecodeExtrinsic is the main entry point for decoding an extrinsic.
func DecodeExtrinsic(metadata *v9.Metadata, extrinsicBytes []byte) (*DecodedExtrinsic, error) {
	r := NewReader(extrinsicBytes)

	// In V9, the extrinsic format is simpler and doesn't have the version byte
	// or signed extensions defined in the same way as later versions.
	// It's typically just the call. This is a simplification.
	// A full implementation would need to handle the extrinsic wrapper format for v9,
	// which includes address, signature, era, and nonce.

	// For this implementation, we'll assume the extrinsic bytes start with the call.
	// This will likely fail for signed extrinsics but will work for inherent extrinsics.

	// --- Decode the Call ---
	call, err := DecodeCall(metadata, r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode call: %w", err)
	}

	return &DecodedExtrinsic{
		// Signature data is not decoded in this simplified version for v9.
		Signature: MultiSignature{},
		Call:      *call,
	}, nil
}

// DecodeCall decodes the pallet index, call index, and the corresponding arguments.
func DecodeCall(metadata *v9.Metadata, r *Reader) (*DecodedPalletVariant, error) {
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
	if int(palletIndex) >= len(metadata.Modules) {
		return nil, fmt.Errorf("pallet with index %d not found in metadata", palletIndex)
	}
	pallet := metadata.Modules[palletIndex]

	if !pallet.Calls.HasValue {
		return nil, fmt.Errorf("pallet '%s' has no calls defined in metadata", pallet.Name)
	}

	if int(callIndex) >= len(pallet.Calls.Value) {
		return nil, fmt.Errorf("call with index %d not found in pallet '%s'", callIndex, pallet.Name)
	}

	callVariant := pallet.Calls.Value[callIndex]

	// --- Decode Arguments ---
	decodedArgs := make([]DecodedArg, len(callVariant.Args))
	for i, arg := range callVariant.Args {
		argValue, err := DecodeArgFromString(metadata, r, string(arg.Type))
		if err != nil {
			return nil, fmt.Errorf("failed to decode arg '%s' for call '%s.%s': %w", arg.Name, pallet.Name, callVariant.Name, err)
		}

		decodedArgs[i] = DecodedArg{
			Name:  string(arg.Name),
			Value: argValue,
		}
	}

	return &DecodedPalletVariant{
		PalletName:  string(pallet.Name),
		VariantName: string(callVariant.Name),
		Args:        decodedArgs,
	}, nil
}
