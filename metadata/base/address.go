package base

import (
	"fmt"
	"submarine/scale"
)

type AddressKind int

const (
	KindAddressId AddressKind = iota
	KindAddressIndex
	KindAddressRaw
	KindAddress32
	KindAddress20
)

type AddressId [32]byte
type AddressIndex uint64
type AddressRaw []byte
type Address32 [32]byte
type Address20 [20]byte

func (a *AddressId) Decode(r *scale.Reader) error {
	bytes, err := r.ReadBytes(32)
	if err != nil {
		return fmt.Errorf("failed to decode AddressId: %w", err)
	}
	*a = AddressId(bytes)
	return nil
}

func (a *AddressIndex) Decode(r *scale.Reader) error {
	compactIndex, err := scale.DecodeCompact(r)
	if err != nil {
		return fmt.Errorf("failed to decode AddressIndex: %w", err)
	}
	*a = AddressIndex(compactIndex.Uint64())
	return nil
}

func (a *AddressRaw) Decode(r *scale.Reader) error {
	bytes, err := scale.DecodeBytes(r)
	if err != nil {
		return fmt.Errorf("failed to decode AddressRaw: %w", err)
	}
	*a = AddressRaw(bytes)
	return nil
}

func (a *Address32) Decode(r *scale.Reader) error {
	bytes, err := r.ReadBytes(32)
	if err != nil {
		return fmt.Errorf("failed to decode Address32: %w", err)
	}
	*a = Address32(bytes)
	return nil
}

func (a *Address20) Decode(r *scale.Reader) error {
	bytes, err := r.ReadBytes(20)
	if err != nil {
		return fmt.Errorf("failed to decode Address20: %w", err)
	}
	*a = Address20(bytes)
	return nil
}

type Address struct {
	Kind   AddressKind
	Id     *AddressId
	Index  *AddressIndex
	Raw    *AddressRaw
	Addr32 *Address32
	Addr20 *Address20
}

func DecodeAddress(r *scale.Reader) (Address, error) {
	variant, err := r.ReadByte()
	if err != nil {
		return Address{}, fmt.Errorf("failed to read Address variant: %w", err)
	}

	var addr Address
	switch AddressKind(variant) {
	case KindAddressId:
		var id AddressId
		if err := id.Decode(r); err != nil {
			return Address{}, err
		}
		addr.Id = &id
		addr.Kind = KindAddressId
	case KindAddressIndex:
		var index AddressIndex
		if err := index.Decode(r); err != nil {
			return Address{}, err
		}
		addr.Index = &index
		addr.Kind = KindAddressIndex
	case KindAddressRaw:
		var raw AddressRaw
		if err := raw.Decode(r); err != nil {
			return Address{}, err
		}
		addr.Raw = &raw
		addr.Kind = KindAddressRaw
	case KindAddress32:
		var addr32 Address32
		if err := addr32.Decode(r); err != nil {
			return Address{}, err
		}
		addr.Addr32 = &addr32
		addr.Kind = KindAddress32
	case KindAddress20:
		var addr20 Address20
		if err := addr20.Decode(r); err != nil {
			return Address{}, err
		}
		addr.Addr20 = &addr20
		addr.Kind = KindAddress20
	default:
		return Address{}, fmt.Errorf("unsupported Address variant: %d", variant)
	}
	return addr, nil
}
