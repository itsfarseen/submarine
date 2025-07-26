package v12

import (
	"fmt"
	. "submarine/decoder/models"
	. "submarine/scale"
	"submarine/scale/gen/v12"
)

// DecodeExtrinsic is the main entry point for decoding an extrinsic.
func DecodeExtrinsic(metadata *v12.Metadata, extrinsicBytes []byte) (*DecodedExtrinsic, error) {
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
