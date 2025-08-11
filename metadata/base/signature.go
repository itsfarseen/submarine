package base

import (
	"fmt"
	"submarine/scale"
)

type SignatureKind int

const (
	KindSignatureEd25519 SignatureKind = iota
	KindSignatureSr25519
	KindSignatureEcdsa
)

type SignatureEd25519 [64]byte
type SignatureSr25519 [64]byte
type SignatureEcdsa [65]byte

func (s *SignatureEd25519) Decode(r *scale.Reader) error {
	bytes, err := r.ReadBytes(64)
	if err != nil {
		return fmt.Errorf("failed to decode SignatureEd25519: %w", err)
	}
	copy(s[:], bytes)
	return nil
}

func (s *SignatureSr25519) Decode(r *scale.Reader) error {
	bytes, err := r.ReadBytes(64)
	if err != nil {
		return fmt.Errorf("failed to decode SignatureSr25519: %w", err)
	}
	copy(s[:], bytes)
	return nil
}

func (s *SignatureEcdsa) Decode(r *scale.Reader) error {
	bytes, err := r.ReadBytes(65)
	if err != nil {
		return fmt.Errorf("failed to decode SignatureEcdsa: %w", err)
	}
	copy(s[:], bytes)
	return nil
}

type Signature struct {
	Kind    SignatureKind
	Ed25519 *SignatureEd25519
	Sr25519 *SignatureSr25519
	Ecdsa   *SignatureEcdsa
}

func DecodeSignature(r *scale.Reader) (Signature, error) {
	variant, err := r.ReadByte()
	if err != nil {
		return Signature{}, fmt.Errorf("failed to read Signature variant: %w", err)
	}

	var sig Signature
	switch SignatureKind(variant) {
	case KindSignatureEd25519:
		var ed25519 SignatureEd25519
		if err := ed25519.Decode(r); err != nil {
			return Signature{}, err
		}
		sig.Ed25519 = &ed25519
		sig.Kind = KindSignatureEd25519
	case KindSignatureSr25519:
		var sr25519 SignatureSr25519
		if err := sr25519.Decode(r); err != nil {
			return Signature{}, err
		}
		sig.Sr25519 = &sr25519
		sig.Kind = KindSignatureSr25519
	case KindSignatureEcdsa:
		var ecdsa SignatureEcdsa
		if err := ecdsa.Decode(r); err != nil {
			return Signature{}, err
		}
		sig.Ecdsa = &ecdsa
		sig.Kind = KindSignatureEcdsa
	default:
		return Signature{}, fmt.Errorf("unsupported Signature variant: %d", variant)
	}
	return sig, nil
}
