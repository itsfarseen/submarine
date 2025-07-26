package v9

import (
	"fmt"
	. "submarine/decoder/models"
	"submarine/scale"
	"submarine/scale/base"
	"submarine/scale/gen/v9"
)

// DecodeExtrinsic is the main entry point for decoding an extrinsic.
func DecodeExtrinsic(metadata *v9.Metadata, extrinsicBytes []byte) (*DecodedExtrinsic, error) {
	r := scale.NewReader(extrinsicBytes)

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
		Signature: base.Signature{},
		Call:      *call,
	}, nil
}

