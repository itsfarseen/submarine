package scale

import (
	"fmt"
)

type MultiAddressKind int

const (
	KindAddressId MultiAddressKind = iota
	KindAddressIndex
	KindAddressRaw
	KindAddress32
	KindAddress20
)

type MultiAddress struct {
	Kind         MultiAddressKind
	AddressId    [32]byte
	AddressIndex uint64
	AddressRaw   []byte
	Address32    [32]byte
	Address20    [20]byte
}

func DecodeMultiAddress(r *Reader) (MultiAddress, error) {
	// Read the variant index, which is the first byte of a MultiAddress encoding.
	variant, err := r.ReadByte()
	if err != nil {
		return MultiAddress{}, fmt.Errorf("failed to read MultiAddress variant: %w", err)
	}

	switch variant {
	case 0: // Id variant: a 32-byte account ID.
		var addr [32]byte
		bytes, err := r.ReadBytes(32)
		if err != nil {
			return MultiAddress{}, fmt.Errorf("failed to decode AddressId: %w", err)
		}
		addr = [32]byte(bytes)
		return MultiAddress{Kind: KindAddressId, AddressId: addr}, nil

	case 1: // Index variant: a compact-encoded integer.
		compactIndex, err := DecodeCompact(r)
		if err != nil {
			return MultiAddress{}, fmt.Errorf("failed to decode AddressIndex: %w", err)
		}
		return MultiAddress{Kind: KindAddressIndex, AddressIndex: compactIndex.Uint64()}, nil

	case 2: // Raw variant: a SCALE encoded vector of bytes.
		bytes, err := DecodeBytes(r)
		if err != nil {
			return MultiAddress{}, fmt.Errorf("failed to decode AddressRaw: %w", err)
		}
		return MultiAddress{Kind: KindAddressRaw, AddressRaw: bytes}, nil

	case 3: // Address32 variant: a 32-byte array.
		var addr [32]byte
		bytes, err := r.ReadBytes(32)
		if err != nil {
			return MultiAddress{}, fmt.Errorf("failed to decode Address32: %w", err)
		}
		addr = [32]byte(bytes)
		return MultiAddress{Kind: KindAddress32, Address32: addr}, nil

	case 4: // Address20 variant: a 20-byte array.
		var addr [20]byte
		bytes, err := r.ReadBytes(20)
		if err != nil {
			return MultiAddress{}, fmt.Errorf("failed to decode Address20: %w", err)
		}
		addr = [20]byte(bytes)
		return MultiAddress{Kind: KindAddress20, Address20: addr}, nil

	default:
		// If the variant byte is not one of the known values, return an error.
		return MultiAddress{}, fmt.Errorf("unsupported MultiAddress variant: %d", variant)
	}
}
