package v13

import (
	"fmt"
	. "submarine/decoder/models"
	. "submarine/scale"
	"submarine/scale/gen/v13"
)

// DecodeExtrinsic is the main entry point for decoding an extrinsic.
func DecodeExtrinsic(metadata *v13.Metadata, extrinsicBytes []byte) (*DecodedExtrinsic, error) {
	r := NewReader(extrinsicBytes)

	// Skip the compact-encoded length of the extrinsic
	_, err := DecodeCompact(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode extrinsic length prefix: %w", err)
	}

	// The next byte describes the transaction format.
	txFormat, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read transaction format byte: %w", err)
	}

	isSigned := (txFormat & 0b10000000) != 0
	var signatureData MultiSignature

	if isSigned {
		// 1. Decode the sender's Address.
		_, err := DecodeMultiAddress(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode sender address: %w", err)
		}

		// 2. Decode the Signature.
		signatureData, err = DecodeMultiSignature(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode signature: %w", err)
		}

		// // 3. Decode the signed extensions.
		// for _, extension := range metadata.Extrinsic.SignedExtensions {
		// 	_, err := DecodeArgFromString(metadata, r, string(extension.Type))
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to decode signed extension '%s': %w", extension.Identifier, err)
		// 	}
		// }
	}

	// --- Decode the Call ---
	call, err := DecodeCall(metadata, r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode call: %w", err)
	}

	return &DecodedExtrinsic{
		Signature: signatureData,
		Call:      *call,
	}, nil
}

// DecodeCall decodes the pallet index, call index, and the corresponding arguments.
func DecodeCall(metadata *v13.Metadata, r *Reader) (*DecodedPalletVariant, error) {
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
	var pallet v13.ModuleMetadata
	foundPallet := false
	for _, p := range metadata.Modules {
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
