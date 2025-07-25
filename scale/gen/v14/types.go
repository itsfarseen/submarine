package v14

import (
	"submarine/scale/gen/scaleInfo"
	"submarine/scale/gen/v11"
	"submarine/scale/gen/v9"
)

type ExtrinsicMetadata struct {
	Type             scaleInfo.Si1LookupTypeId
	Version          uint8
	SignedExtensions []SignedExtensionMetadata
}

type PalletConstantMetadata struct {
	Name  string
	Type  scaleInfo.Si1LookupTypeId
	Value []byte
	Docs  []string
}

type PalletCallMetadata struct {
	Type scaleInfo.Si1LookupTypeId
}

type FunctionMetadata struct {
	Name   string
	Fields []scaleInfo.Si1Field
	Index  uint8
	Docs   []string
	Args   []FunctionArgumentMetadata
}

type PalletErrorMetadata struct {
	Type scaleInfo.Si1LookupTypeId
}

type PalletEventMetadata struct {
	Type scaleInfo.Si1LookupTypeId
}

type StorageEntryMetadata struct {
	Name     string
	Modifier v9.StorageEntryModifier
	Type     StorageEntryType
	Fallback []byte
	Docs     []string
}

type StorageEntryType struct {
	Kind  string
	Plain *scaleInfo.Si1LookupTypeId
	Map   *StorageEntryMap
}

type PortableRegistry struct {
	Types []PortableType
}

type Metadata struct {
	Lookup    PortableRegistry
	Pallets   []PalletMetadata
	Extrinsic ExtrinsicMetadata
	Type      scaleInfo.Si1LookupTypeId
}

type PortableType struct {
	Id   scaleInfo.Si1LookupTypeId
	Type scaleInfo.Si1Type
}

type PalletStorageMetadata struct {
	Prefix string
	Items  []StorageEntryMetadata
}

type FunctionArgumentMetadata struct {
	Name     string
	Type     string
	TypeName *string
}

type SignedExtensionMetadata struct {
	Identifier       string
	Type             scaleInfo.Si1LookupTypeId
	AdditionalSigned scaleInfo.Si1LookupTypeId
}

type StorageEntryMap struct {
	Hashers []v11.StorageHasher
	Key     scaleInfo.Si1LookupTypeId
	Value   scaleInfo.Si1LookupTypeId
}

type EventMetadata struct {
	Name   string
	Fields []scaleInfo.Si1Field
	Index  uint8
	Docs   []string
	Args   []string
}

type ErrorMetadata struct {
	Name   string
	Fields []scaleInfo.Si1Field
	Index  uint8
	Docs   []string
	Args   []string
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
