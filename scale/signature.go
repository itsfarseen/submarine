package scale

import (
	"fmt"
)

type MultiSignatureKind int

const (
	KindSignatureEd25519 MultiSignatureKind = iota
	KindSignatureSr25519
	KindSignatureEcdsa
)

type MultiSignature struct {
	Kind    MultiSignatureKind
	Ed25519 [64]byte
	Sr25519 [64]byte
	Ecdsa   [65]byte
}

func DecodeMultiSignature(r *Reader) (MultiSignature, error) {
	// Read the variant index, which is the first byte of a MultiSignature encoding.
	variant, err := r.ReadByte()
	if err != nil {
		return MultiSignature{}, fmt.Errorf("failed to read MultiSignature variant: %w", err)
	}

	switch variant {
	case 0: // Ed25519 variant: a 64-byte signature.
		var sig [64]byte
		bytes, err := r.ReadBytes(64)
		if err != nil {
			return MultiSignature{}, fmt.Errorf("failed to decode SignatureEd25519: %w", err)
		}
		copy(sig[:], bytes)
		return MultiSignature{Kind: KindSignatureEd25519, Ed25519: sig}, nil

	case 1: // Sr25519 variant: a 64-byte signature.
		var sig [64]byte
		bytes, err := r.ReadBytes(64)
		if err != nil {
			return MultiSignature{}, fmt.Errorf("failed to decode SignatureSr25519: %w", err)
		}
		copy(sig[:], bytes)
		return MultiSignature{Kind: KindSignatureSr25519, Sr25519: sig}, nil

	case 2: // Ecdsa variant: a 65-byte signature.
		var sig [65]byte
		bytes, err := r.ReadBytes(65)
		if err != nil {
			return MultiSignature{}, fmt.Errorf("failed to decode SignatureEcdsa: %w", err)
		}
		copy(sig[:], bytes)
		return MultiSignature{Kind: KindSignatureEcdsa, Ecdsa: sig}, nil

	default:
		// If the variant byte is not one of the known values, return an error.
		return MultiSignature{}, fmt.Errorf("unsupported MultiSignature variant: %d", variant)
	}
}
