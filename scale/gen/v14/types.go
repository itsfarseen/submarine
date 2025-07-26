package v14

import (
	"fmt"
	"submarine/scale"
	"submarine/scale/gen/scaleInfo"
	"submarine/scale/gen/v13"
)

type ErrorMetadata struct {
	Name   string
	Fields []scaleInfo.Si1Field
	Index  uint8
	Docs   []string
	Args   []string
}

func DecodeErrorMetadata(reader *scale.Reader) (ErrorMetadata, error) {
	var t ErrorMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Fields, err = scale.DecodeVec(reader, func(reader *scale.Reader) (scaleInfo.Si1Field, error) { return DecodeSi1Field(reader) })
	if err != nil {
		return t, fmt.Errorf("field Fields: %w", err)
	}

	t.Index, err = scale.DecodeU8(reader)
	if err != nil {
		return t, fmt.Errorf("field Index: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	t.Args, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Args: %w", err)
	}

	return t, nil
}

type EventMetadata struct {
	Name   string
	Fields []scaleInfo.Si1Field
	Index  uint8
	Docs   []string
	Args   []string
}

func DecodeEventMetadata(reader *scale.Reader) (EventMetadata, error) {
	var t EventMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Fields, err = scale.DecodeVec(reader, func(reader *scale.Reader) (scaleInfo.Si1Field, error) { return DecodeSi1Field(reader) })
	if err != nil {
		return t, fmt.Errorf("field Fields: %w", err)
	}

	t.Index, err = scale.DecodeU8(reader)
	if err != nil {
		return t, fmt.Errorf("field Index: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	t.Args, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Args: %w", err)
	}

	return t, nil
}

type ExtrinsicMetadata struct {
	Type             scaleInfo.Si1LookupTypeId
	Version          uint8
	SignedExtensions []SignedExtensionMetadata
}

func DecodeExtrinsicMetadata(reader *scale.Reader) (ExtrinsicMetadata, error) {
	var t ExtrinsicMetadata
	var err error

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	t.Version, err = scale.DecodeU8(reader)
	if err != nil {
		return t, fmt.Errorf("field Version: %w", err)
	}

	t.SignedExtensions, err = scale.DecodeVec(reader, func(reader *scale.Reader) (SignedExtensionMetadata, error) {
		return DecodeSignedExtensionMetadata(reader)
	})
	if err != nil {
		return t, fmt.Errorf("field SignedExtensions: %w", err)
	}

	return t, nil
}

type FunctionArgumentMetadata struct {
	Name     string
	Type     string
	TypeName *string
}

func DecodeFunctionArgumentMetadata(reader *scale.Reader) (FunctionArgumentMetadata, error) {
	var t FunctionArgumentMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Type, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	t.TypeName, err = scale.DecodeOption(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field TypeName: %w", err)
	}

	return t, nil
}

type FunctionMetadata struct {
	Name   string
	Fields []scaleInfo.Si1Field
	Index  uint8
	Docs   []string
	Args   []FunctionArgumentMetadata
}

func DecodeFunctionMetadata(reader *scale.Reader) (FunctionMetadata, error) {
	var t FunctionMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Fields, err = scale.DecodeVec(reader, func(reader *scale.Reader) (scaleInfo.Si1Field, error) { return DecodeSi1Field(reader) })
	if err != nil {
		return t, fmt.Errorf("field Fields: %w", err)
	}

	t.Index, err = scale.DecodeU8(reader)
	if err != nil {
		return t, fmt.Errorf("field Index: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	t.Args, err = scale.DecodeVec(reader, func(reader *scale.Reader) (FunctionArgumentMetadata, error) {
		return DecodeFunctionArgumentMetadata(reader)
	})
	if err != nil {
		return t, fmt.Errorf("field Args: %w", err)
	}

	return t, nil
}

type Metadata struct {
	Lookup    PortableRegistry
	Pallets   []PalletMetadata
	Extrinsic ExtrinsicMetadata
	Type      scaleInfo.Si1LookupTypeId
}

func DecodeMetadata(reader *scale.Reader) (Metadata, error) {
	var t Metadata
	var err error

	t.Lookup, err = DecodePortableRegistry(reader)
	if err != nil {
		return t, fmt.Errorf("field Lookup: %w", err)
	}

	t.Pallets, err = scale.DecodeVec(reader, func(reader *scale.Reader) (PalletMetadata, error) { return DecodePalletMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Pallets: %w", err)
	}

	t.Extrinsic, err = DecodeExtrinsicMetadata(reader)
	if err != nil {
		return t, fmt.Errorf("field Extrinsic: %w", err)
	}

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	return t, nil
}

type PalletCallMetadata struct {
	Type scaleInfo.Si1LookupTypeId
}

func DecodePalletCallMetadata(reader *scale.Reader) (PalletCallMetadata, error) {
	var t PalletCallMetadata
	var err error

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	return t, nil
}

type PalletConstantMetadata struct {
	Name  string
	Type  scaleInfo.Si1LookupTypeId
	Value []byte
	Docs  []string
}

func DecodePalletConstantMetadata(reader *scale.Reader) (PalletConstantMetadata, error) {
	var t PalletConstantMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	t.Value, err = scale.DecodeBytes(reader)
	if err != nil {
		return t, fmt.Errorf("field Value: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	return t, nil
}

type PalletErrorMetadata struct {
	Type scaleInfo.Si1LookupTypeId
}

func DecodePalletErrorMetadata(reader *scale.Reader) (PalletErrorMetadata, error) {
	var t PalletErrorMetadata
	var err error

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	return t, nil
}

type PalletEventMetadata struct {
	Type scaleInfo.Si1LookupTypeId
}

func DecodePalletEventMetadata(reader *scale.Reader) (PalletEventMetadata, error) {
	var t PalletEventMetadata
	var err error

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	return t, nil
}

type PalletMetadata struct {
	Name      string
	Storage   *PalletStorageMetadata
	Calls     *PalletCallMetadata
	Events    *PalletEventMetadata
	Constants []PalletConstantMetadata
	Errors    *PalletErrorMetadata
	Index     uint8
}

func DecodePalletMetadata(reader *scale.Reader) (PalletMetadata, error) {
	var t PalletMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Storage, err = scale.DecodeOption(reader, func(reader *scale.Reader) (PalletStorageMetadata, error) { return DecodePalletStorageMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Storage: %w", err)
	}

	t.Calls, err = scale.DecodeOption(reader, func(reader *scale.Reader) (PalletCallMetadata, error) { return DecodePalletCallMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Calls: %w", err)
	}

	t.Events, err = scale.DecodeOption(reader, func(reader *scale.Reader) (PalletEventMetadata, error) { return DecodePalletEventMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Events: %w", err)
	}

	t.Constants, err = scale.DecodeVec(reader, func(reader *scale.Reader) (PalletConstantMetadata, error) {
		return DecodePalletConstantMetadata(reader)
	})
	if err != nil {
		return t, fmt.Errorf("field Constants: %w", err)
	}

	t.Errors, err = scale.DecodeOption(reader, func(reader *scale.Reader) (PalletErrorMetadata, error) { return DecodePalletErrorMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Errors: %w", err)
	}

	t.Index, err = scale.DecodeU8(reader)
	if err != nil {
		return t, fmt.Errorf("field Index: %w", err)
	}

	return t, nil
}

type PalletStorageMetadata struct {
	Prefix string
	Items  []StorageEntryMetadata
}

func DecodePalletStorageMetadata(reader *scale.Reader) (PalletStorageMetadata, error) {
	var t PalletStorageMetadata
	var err error

	t.Prefix, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Prefix: %w", err)
	}

	t.Items, err = scale.DecodeVec(reader, func(reader *scale.Reader) (StorageEntryMetadata, error) { return DecodeStorageEntryMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Items: %w", err)
	}

	return t, nil
}

type PortableRegistry struct {
	Types []PortableType
}

func DecodePortableRegistry(reader *scale.Reader) (PortableRegistry, error) {
	var t PortableRegistry
	var err error

	t.Types, err = scale.DecodeVec(reader, func(reader *scale.Reader) (PortableType, error) { return DecodePortableType(reader) })
	if err != nil {
		return t, fmt.Errorf("field Types: %w", err)
	}

	return t, nil
}

type PortableType struct {
	Id   scaleInfo.Si1LookupTypeId
	Type scaleInfo.Si1Type
}

func DecodePortableType(reader *scale.Reader) (PortableType, error) {
	var t PortableType
	var err error

	t.Id, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Id: %w", err)
	}

	t.Type, err = DecodeSi1Type(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	return t, nil
}

type Si1Field = scaleInfo.Si1Field

func DecodeSi1Field(reader *scale.Reader) (Si1Field, error) {
	return scaleInfo.DecodeSi1Field(reader)
}

type Si1LookupTypeId = scaleInfo.Si1LookupTypeId

func DecodeSi1LookupTypeId(reader *scale.Reader) (Si1LookupTypeId, error) {
	return scaleInfo.DecodeSi1LookupTypeId(reader)
}

type Si1Type = scaleInfo.Si1Type

func DecodeSi1Type(reader *scale.Reader) (Si1Type, error) {
	return scaleInfo.DecodeSi1Type(reader)
}

type SignedExtensionMetadata struct {
	Identifier       string
	Type             scaleInfo.Si1LookupTypeId
	AdditionalSigned scaleInfo.Si1LookupTypeId
}

func DecodeSignedExtensionMetadata(reader *scale.Reader) (SignedExtensionMetadata, error) {
	var t SignedExtensionMetadata
	var err error

	t.Identifier, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Identifier: %w", err)
	}

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	t.AdditionalSigned, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field AdditionalSigned: %w", err)
	}

	return t, nil
}

type StorageEntryMap struct {
	Hashers []v13.StorageHasher
	Key     scaleInfo.Si1LookupTypeId
	Value   scaleInfo.Si1LookupTypeId
}

func DecodeStorageEntryMap(reader *scale.Reader) (StorageEntryMap, error) {
	var t StorageEntryMap
	var err error

	t.Hashers, err = scale.DecodeVec(reader, func(reader *scale.Reader) (v13.StorageHasher, error) { return DecodeStorageHasher(reader) })
	if err != nil {
		return t, fmt.Errorf("field Hashers: %w", err)
	}

	t.Key, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Key: %w", err)
	}

	t.Value, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Value: %w", err)
	}

	return t, nil
}

type StorageEntryMetadata struct {
	Name     string
	Modifier v13.StorageEntryModifier
	Type     StorageEntryType
	Fallback []byte
	Docs     []string
}

func DecodeStorageEntryMetadata(reader *scale.Reader) (StorageEntryMetadata, error) {
	var t StorageEntryMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Modifier, err = DecodeStorageEntryModifier(reader)
	if err != nil {
		return t, fmt.Errorf("field Modifier: %w", err)
	}

	t.Type, err = DecodeStorageEntryType(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	t.Fallback, err = scale.DecodeBytes(reader)
	if err != nil {
		return t, fmt.Errorf("field Fallback: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	return t, nil
}

type StorageEntryModifier = v13.StorageEntryModifier

func DecodeStorageEntryModifier(reader *scale.Reader) (StorageEntryModifier, error) {
	return v13.DecodeStorageEntryModifier(reader)
}

type StorageEntryTypeKind byte

const (
	StorageEntryTypeKindPlain StorageEntryTypeKind = 0
	StorageEntryTypeKindMap   StorageEntryTypeKind = 1
)

type StorageEntryType struct {
	Kind  StorageEntryTypeKind
	Plain *scaleInfo.Si1LookupTypeId
	Map   *StorageEntryMap
}

func DecodeStorageEntryType(reader *scale.Reader) (StorageEntryType, error) {
	var t StorageEntryType

	tag, err := reader.ReadByte()
	if err != nil {
		return t, fmt.Errorf("enum tag: %w", err)
	}

	t.Kind = StorageEntryTypeKind(tag)
	switch t.Kind {

	case StorageEntryTypeKindPlain:
		value, err := DecodeSi1LookupTypeId(reader)
		if err != nil {
			return t, fmt.Errorf("field Plain: %w", err)
		}
		t.Plain = &value
		return t, nil

	case StorageEntryTypeKindMap:
		value, err := DecodeStorageEntryMap(reader)
		if err != nil {
			return t, fmt.Errorf("field Map: %w", err)
		}
		t.Map = &value
		return t, nil

	default:
		return t, fmt.Errorf("unknown tag: %d", tag)
	}
}

type StorageHasher = v13.StorageHasher

func DecodeStorageHasher(reader *scale.Reader) (StorageHasher, error) {
	return v13.DecodeStorageHasher(reader)
}
