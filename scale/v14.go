package scale

import "fmt"

// MetadataV14 is the top-level structure
type MetadataV14 struct {
	Lookup    PortableRegistry
	Pallets   []PalletMetadataV14
	Extrinsic ExtrinsicMetadataV14
	Type      SiLookupTypeId
}

type PortableRegistry struct {
	Types []PortableTypeV14
}

type StorageEntryModifierV14 byte
type StorageHasherV14 byte

type PortableTypeV14 struct {
	Id   SiLookupTypeId
	Type Si1Type
}

type ErrorMetadataV14 struct {
	Si1Variant
	Args []Text
}

type EventMetadataV14 struct {
	Si1Variant
	Args []Text
}

type FunctionArgumentMetadataV14 struct {
	Name     Text
	Type     Text
	TypeName Option[Text]
}

type FunctionMetadataV14 struct {
	Si1Variant
	Args []FunctionArgumentMetadataV14
}

type SignedExtensionMetadataV14 struct {
	Identifier       Text
	Type             SiLookupTypeId
	AdditionalSigned SiLookupTypeId
}

type ExtrinsicMetadataV14 struct {
	Type             SiLookupTypeId
	Version          uint8
	SignedExtensions []SignedExtensionMetadataV14
}

type PalletCallMetadataV14 struct {
	Type SiLookupTypeId
}

type PalletConstantMetadataV14 struct {
	Name  Text
	Type  SiLookupTypeId
	Value Bytes
	Docs  []Text
}

type PalletErrorMetadataV14 struct {
	Type SiLookupTypeId
}

type PalletEventMetadataV14 struct {
	Type SiLookupTypeId
}

type StorageEntryTypeV14 interface {
	isStorageEntryTypeV14()
}

type StorageEntryTypeV14Plain struct {
	Value SiLookupTypeId
}

func (s StorageEntryTypeV14Plain) isStorageEntryTypeV14() {}

type StorageEntryTypeV14Map struct {
	Hashers []StorageHasherV14
	Key     SiLookupTypeId
	Value   SiLookupTypeId
}

func (s StorageEntryTypeV14Map) isStorageEntryTypeV14() {}

type StorageEntryMetadataV14 struct {
	Name     Text
	Modifier StorageEntryModifierV14
	Type     StorageEntryTypeV14
	Fallback Bytes
	Docs     []Text
}

type PalletStorageMetadataV14 struct {
	Prefix Text
	Items  []StorageEntryMetadataV14
}

type PalletMetadataV14 struct {
	Name      Text
	Storage   Option[PalletStorageMetadataV14]
	Calls     Option[PalletCallMetadataV14]
	Events    Option[PalletEventMetadataV14]
	Constants []PalletConstantMetadataV14
	Errors    Option[PalletErrorMetadataV14]
	Index     uint8
}

// Decoders

func DecodeStorageEntryModifierV14(r *Reader) (StorageEntryModifierV14, error) {
	b, err := r.ReadByte()
	return StorageEntryModifierV14(b), err
}

func DecodeStorageHasherV14(r *Reader) (StorageHasherV14, error) {
	b, err := r.ReadByte()
	return StorageHasherV14(b), err
}

func DecodePortableTypeV14(r *Reader) (PortableTypeV14, error) {
	var result PortableTypeV14
	var err error
	result.Id, err = DecodeSiLookupTypeId(r)
	if err != nil {
		return result, fmt.Errorf("id: %w", err)
	}
	// The correction happens here.
	result.Type, err = DecodeSi1Type(r)
	if err != nil {
		return result, fmt.Errorf("type: %w", err)
	}
	return result, nil

}

func DecodeErrorMetadataV14(r *Reader) (ErrorMetadataV14, error) {
	var result ErrorMetadataV14
	var err error

	siVariant, err := DecodeSi1Variant(r)
	if err != nil {
		return result, err
	}
	result.Si1Variant = siVariant

	result.Args, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodeEventMetadataV14(r *Reader) (EventMetadataV14, error) {
	var result EventMetadataV14
	var err error

	siVariant, err := DecodeSi1Variant(r)
	if err != nil {
		return result, err
	}
	result.Si1Variant = siVariant

	result.Args, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodeFunctionArgumentMetadataV14(r *Reader) (FunctionArgumentMetadataV14, error) {
	var result FunctionArgumentMetadataV14
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

func DecodeFunctionMetadataV14(r *Reader) (FunctionMetadataV14, error) {
	var result FunctionMetadataV14
	var err error

	siVariant, err := DecodeSi1Variant(r)
	if err != nil {
		return result, err
	}
	result.Si1Variant = siVariant

	result.Args, err = DecodeVec(r, DecodeFunctionArgumentMetadataV14)
	return result, err
}

func DecodeSignedExtensionMetadataV14(r *Reader) (SignedExtensionMetadataV14, error) {
	var result SignedExtensionMetadataV14
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

func DecodePalletCallMetadataV14(r *Reader) (PalletCallMetadataV14, error) {
	var result PalletCallMetadataV14
	var err error
	result.Type, err = DecodeSiLookupTypeId(r)
	return result, err
}

func DecodePalletConstantMetadataV14(r *Reader) (PalletConstantMetadataV14, error) {
	var result PalletConstantMetadataV14
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

func DecodePalletErrorMetadataV14(r *Reader) (PalletErrorMetadataV14, error) {
	var result PalletErrorMetadataV14
	var err error
	result.Type, err = DecodeSiLookupTypeId(r)
	return result, err
}

func DecodePalletEventMetadataV14(r *Reader) (PalletEventMetadataV14, error) {
	var result PalletEventMetadataV14
	var err error
	result.Type, err = DecodeSiLookupTypeId(r)
	return result, err
}

func DecodeStorageEntryTypeV14(r *Reader) (StorageEntryTypeV14, error) {
	variant, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch variant {
	case 0: // Plain
		var plain StorageEntryTypeV14Plain
		plain.Value, err = DecodeSiLookupTypeId(r)
		return plain, err
	case 1: // Map
		var m StorageEntryTypeV14Map
		m.Hashers, err = DecodeVec(r, DecodeStorageHasherV14)
		if err != nil {
			return nil, err
		}
		m.Key, err = DecodeSiLookupTypeId(r)
		if err != nil {
			return nil, err
		}
		m.Value, err = DecodeSiLookupTypeId(r)
		return m, err
	default:
		return nil, fmt.Errorf("unknown variant for StorageEntryTypeV14: %d", variant)
	}
}

func DecodeStorageEntryMetadataV14(r *Reader) (StorageEntryMetadataV14, error) {
	var result StorageEntryMetadataV14
	var err error
	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Modifier, err = DecodeStorageEntryModifierV14(r)
	if err != nil {
		return result, err
	}
	result.Type, err = DecodeStorageEntryTypeV14(r)
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

func DecodePalletStorageMetadataV14(r *Reader) (PalletStorageMetadataV14, error) {
	var result PalletStorageMetadataV14
	var err error
	result.Prefix, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Items, err = DecodeVec(r, DecodeStorageEntryMetadataV14)
	return result, err
}

func DecodePalletMetadataV14(r *Reader) (PalletMetadataV14, error) {
	var result PalletMetadataV14
	var err error

	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}

	result.Storage, err = DecodeOption(r, DecodePalletStorageMetadataV14)
	if err != nil {
		return result, err
	}

	result.Calls, err = DecodeOption(r, DecodePalletCallMetadataV14)
	if err != nil {
		return result, err
	}

	result.Events, err = DecodeOption(r, DecodePalletEventMetadataV14)
	if err != nil {
		return result, err
	}

	result.Constants, err = DecodeVec(r, DecodePalletConstantMetadataV14)
	if err != nil {
		return result, err
	}

	result.Errors, err = DecodeOption(r, DecodePalletErrorMetadataV14)
	if err != nil {
		return result, err
	}

	result.Index, err = DecodeU8(r)
	return result, err
}

func DecodeExtrinsicMetadataV14(r *Reader) (ExtrinsicMetadataV14, error) {
	var result ExtrinsicMetadataV14
	var err error
	result.Type, err = DecodeSiLookupTypeId(r)
	if err != nil {
		return result, err
	}
	result.Version, err = DecodeU8(r)
	if err != nil {
		return result, err
	}
	result.SignedExtensions, err = DecodeVec(r, DecodeSignedExtensionMetadataV14)
	return result, err
}

func DecodePortableRegistry(r *Reader) (PortableRegistry, error) {
	types, err := DecodeVec(r, DecodePortableTypeV14)
	if err != nil {
		return PortableRegistry{}, fmt.Errorf("types: %w", err)

	}
	return PortableRegistry{Types: types}, nil
}

// DecodeMetadataV14 is the top-level decoder function.
func DecodeMetadataV14(r *Reader) (MetadataV14, error) {
	var result MetadataV14
	var err error

	result.Lookup, err = DecodePortableRegistry(r)
	if err != nil {
		return result, fmt.Errorf("lookup: %w", err)
	}

	result.Pallets, err = DecodeVec(r, DecodePalletMetadataV14)
	if err != nil {
		return result, fmt.Errorf("pallets: %w", err)

	}

	result.Extrinsic, err = DecodeExtrinsicMetadataV14(r)
	if err != nil {
		return result, fmt.Errorf("extrinsics: %w", err)

	}

	result.Type, err = DecodeSiLookupTypeId(r)
	if err != nil {
		return result, fmt.Errorf("type: %w", err)
	}
	return result, nil

}
