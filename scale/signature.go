package scale

import (
	"fmt"
)

type MultiSignature interface {
	isMultiSignature()
}

type SignatureEd25519 [64]byte

func (s SignatureEd25519) isMultiSignature() {}

type SignatureSr25519 [64]byte

func (s SignatureSr25519) isMultiSignature() {}

type SignatureEcdsa [65]byte

func (s SignatureEcdsa) isMultiSignature() {}

func DecodeMultiSignature(r *Reader) (MultiSignature, error) {
	// Read the variant index, which is the first byte of a MultiSignature encoding.
	variant, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read MultiSignature variant: %w", err)
	}

	switch variant {
	case 0: // Ed25519 variant: a 64-byte signature.
		var sig SignatureEd25519
		bytes, err := r.ReadBytes(64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode SignatureEd25519: %w", err)
		}
		copy(sig[:], bytes)
		return sig, nil

	case 1: // Sr25519 variant: a 64-byte signature.
		var sig SignatureSr25519
		bytes, err := r.ReadBytes(64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode SignatureSr25519: %w", err)
		}
		copy(sig[:], bytes)
		return sig, nil

	case 2: // Ecdsa variant: a 65-byte signature.
		var sig SignatureEcdsa
		bytes, err := r.ReadBytes(65)
		if err != nil {
			return nil, fmt.Errorf("failed to decode SignatureEcdsa: %w", err)
		}
		copy(sig[:], bytes)
		return sig, nil

	default:
		// If the variant byte is not one of the known values, return an error.
		return nil, fmt.Errorf("unsupported MultiSignature variant: %d", variant)
	}
}
