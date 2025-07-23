package v14

import (
	"fmt"
	. "submarine/scale"
)

// Metadata is the top-level structure
type Metadata struct {
	Lookup    PortableRegistry
	Pallets   []PalletMetadata
	Extrinsic ExtrinsicMetadata
	Type      SiLookupTypeId
}

type PortableRegistry struct {
	Types []PortableType
}

type StorageEntryModifier byte
type StorageHasher byte

type PortableType struct {
	Id   SiLookupTypeId
	Type Si1Type
}

type ErrorMetadata struct {
	Si1Variant
	Args []Text
}

type EventMetadata struct {
	Si1Variant
	Args []Text
}

type FunctionArgumentMetadata struct {
	Name     Text
	Type     Text
	TypeName Option[Text]
}

type FunctionMetadata struct {
	Si1Variant
	Args []FunctionArgumentMetadata
}

type SignedExtensionMetadata struct {
	Identifier       Text
	Type             SiLookupTypeId
	AdditionalSigned SiLookupTypeId
}

type ExtrinsicMetadata struct {
	Type             SiLookupTypeId
	Version          uint8
	SignedExtensions []SignedExtensionMetadata
}

type PalletCallMetadata struct {
	Type SiLookupTypeId
}

type PalletConstantMetadata struct {
	Name  Text
	Type  SiLookupTypeId
	Value Bytes
	Docs  []Text
}

type PalletErrorMetadata struct {
	Type SiLookupTypeId
}

type PalletEventMetadata struct {
	Type SiLookupTypeId
}

type StorageEntryTypeKind byte

const (
	KindPlain StorageEntryTypeKind = iota
	KindMap
)

// StorageEntryType is a tagged union for different storage entry types.
type StorageEntryType struct {
	Kind  StorageEntryTypeKind
	Plain StorageEntryTypePlain
	Map   StorageEntryTypeMap
}

type StorageEntryTypePlain struct {
	Value SiLookupTypeId
}

type StorageEntryTypeMap struct {
	Hashers []StorageHasher
	Key     SiLookupTypeId
	Value   SiLookupTypeId
}

type StorageEntryMetadata struct {
	Name     Text
	Modifier StorageEntryModifier
	Type     StorageEntryType
	Fallback Bytes
	Docs     []Text
}

type PalletStorageMetadata struct {
	Prefix Text
	Items  []StorageEntryMetadata
}

type PalletMetadata struct {
	Name      Text
	Storage   Option[PalletStorageMetadata]
	Calls     Option[PalletCallMetadata]
	Events    Option[PalletEventMetadata]
	Constants []PalletConstantMetadata
	Errors    Option[PalletErrorMetadata]
	Index     uint8
}

// Decoders

func DecodeStorageEntryModifier(r *Reader) (StorageEntryModifier, error) {
	b, err := r.ReadByte()
	return StorageEntryModifier(b), err
}

func DecodeStorageHasher(r *Reader) (StorageHasher, error) {
	b, err := r.ReadByte()
	return StorageHasher(b), err
}

func DecodePortableType(r *Reader) (PortableType, error) {
	var result PortableType
	var err error
	result.Id, err = DecodeSiLookupTypeId(r)
	if err != nil {
		return result, fmt.Errorf("id: %w", err)
	}
	result.Type, err = DecodeSi1Type(r)
	if err != nil {
		return result, fmt.Errorf("type: %w", err)
	}
	return result, nil

}

func DecodeErrorMetadata(r *Reader) (ErrorMetadata, error) {
	var result ErrorMetadata
	var err error

	siVariant, err := DecodeSi1Variant(r)
	if err != nil {
		return result, err
	}
	result.Si1Variant = siVariant

	result.Args, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodeEventMetadata(r *Reader) (EventMetadata, error) {
	var result EventMetadata
	var err error

	siVariant, err := DecodeSi1Variant(r)
	if err != nil {
		return result, err
	}
	result.Si1Variant = siVariant

	result.Args, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodeFunctionArgumentMetadata(r *Reader) (FunctionArgumentMetadata, error) {
	var result FunctionArgumentMetadata
	var err error
	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Type, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.TypeName, err = DecodeOption(r, DecodeText)
	return result, err
}

func DecodeFunctionMetadata(r *Reader) (FunctionMetadata, error) {
	var result FunctionMetadata
	var err error

	siVariant, err := DecodeSi1Variant(r)
	if err != nil {
		return result, err
	}
	result.Si1Variant = siVariant

	result.Args, err = DecodeVec(r, DecodeFunctionArgumentMetadata)
	return result, err
}

func DecodeSignedExtensionMetadata(r *Reader) (SignedExtensionMetadata, error) {
	var result SignedExtensionMetadata
	var err error
	result.Identifier, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Type, err = DecodeSiLookupTypeId(r)
	if err != nil {
		return result, err
	}
	result.AdditionalSigned, err = DecodeSiLookupTypeId(r)
	return result, err
}

func DecodePalletCallMetadata(r *Reader) (PalletCallMetadata, error) {
	var result PalletCallMetadata
	var err error
	result.Type, err = DecodeSiLookupTypeId(r)
	return result, err
}

func DecodePalletConstantMetadata(r *Reader) (PalletConstantMetadata, error) {
	var result PalletConstantMetadata
	var err error
	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Type, err = DecodeSiLookupTypeId(r)
	if err != nil {
		return result, err
	}
	result.Value, err = DecodeBytes(r)
	if err != nil {
		return result, err
	}
	result.Docs, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodePalletErrorMetadata(r *Reader) (PalletErrorMetadata, error) {
	var result PalletErrorMetadata
	var err error
	result.Type, err = DecodeSiLookupTypeId(r)
	return result, err
}

func DecodePalletEventMetadata(r *Reader) (PalletEventMetadata, error) {
	var result PalletEventMetadata
	var err error
	result.Type, err = DecodeSiLookupTypeId(r)
	return result, err
}

func DecodeStorageEntryType(r *Reader) (StorageEntryType, error) {
	variant, err := r.ReadByte()
	if err != nil {
		return StorageEntryType{}, err
	}
	switch StorageEntryTypeKind(variant) {
	case KindPlain:
		var plain StorageEntryTypePlain
		plain.Value, err = DecodeSiLookupTypeId(r)
		if err != nil {
			return StorageEntryType{}, err
		}
		return StorageEntryType{Kind: KindPlain, Plain: plain}, nil
	case KindMap:
		var m StorageEntryTypeMap
		m.Hashers, err = DecodeVec(r, DecodeStorageHasher)
		if err != nil {
			return StorageEntryType{}, err
		}
		m.Key, err = DecodeSiLookupTypeId(r)
		if err != nil {
			return StorageEntryType{}, err
		}
		m.Value, err = DecodeSiLookupTypeId(r)
		if err != nil {
			return StorageEntryType{}, err
		}
		return StorageEntryType{Kind: KindMap, Map: m}, nil
	default:
		return StorageEntryType{}, fmt.Errorf("unknown variant for StorageEntryType: %d", variant)
	}
}

func DecodeStorageEntryMetadata(r *Reader) (StorageEntryMetadata, error) {
	var result StorageEntryMetadata
	var err error
	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Modifier, err = DecodeStorageEntryModifier(r)
	if err != nil {
		return result, err
	}
	result.Type, err = DecodeStorageEntryType(r)
	if err != nil {
		return result, err
	}
	result.Fallback, err = DecodeBytes(r)
	if err != nil {
		return result, err
	}
	result.Docs, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodePalletStorageMetadata(r *Reader) (PalletStorageMetadata, error) {
	var result PalletStorageMetadata
	var err error
	result.Prefix, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Items, err = DecodeVec(r, DecodeStorageEntryMetadata)
	return result, err
}

func DecodePalletMetadata(r *Reader) (PalletMetadata, error) {
	var result PalletMetadata
	var err error

	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}

	result.Storage, err = DecodeOption(r, DecodePalletStorageMetadata)
	if err != nil {
		return result, err
	}

	result.Calls, err = DecodeOption(r, DecodePalletCallMetadata)
	if err != nil {
		return result, err
	}

	result.Events, err = DecodeOption(r, DecodePalletEventMetadata)
	if err != nil {
		return result, err
	}

	result.Constants, err = DecodeVec(r, DecodePalletConstantMetadata)
	if err != nil {
		return result, err
	}

	result.Errors, err = DecodeOption(r, DecodePalletErrorMetadata)
	if err != nil {
		return result, err
	}

	result.Index, err = DecodeU8(r)
	return result, err
}

func DecodeExtrinsicMetadata(r *Reader) (ExtrinsicMetadata, error) {
	var result ExtrinsicMetadata
	var err error
	result.Type, err = DecodeSiLookupTypeId(r)
	if err != nil {
		return result, err
	}
	result.Version, err = DecodeU8(r)
	if err != nil {
		return result, err
	}
	result.SignedExtensions, err = DecodeVec(r, DecodeSignedExtensionMetadata)
	return result, err
}

func DecodePortableRegistry(r *Reader) (PortableRegistry, error) {
	types, err := DecodeVec(r, DecodePortableType)
	if err != nil {
		return PortableRegistry{}, fmt.Errorf("types: %w", err)

	}
	return PortableRegistry{Types: types}, nil
}

// DecodeMetadata is the top-level decoder function.
func DecodeMetadata(r *Reader) (Metadata, error) {
	var result Metadata
	var err error

	result.Lookup, err = DecodePortableRegistry(r)
	if err != nil {
		return result, fmt.Errorf("lookup: %w", err)
	}

	result.Pallets, err = DecodeVec(r, DecodePalletMetadata)
	if err != nil {
		return result, fmt.Errorf("pallets: %w", err)

	}

	result.Extrinsic, err = DecodeExtrinsicMetadata(r)
	if err != nil {
		return result, fmt.Errorf("extrinsics: %w", err)

	}

	result.Type, err = DecodeSiLookupTypeId(r)
	if err != nil {
		return result, fmt.Errorf("type: %w", err)
	}
	return result, nil

}
