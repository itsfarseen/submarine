package v11

import (
	"submarine/scale/gen/v9"
)

type ModuleMetadata struct {
	Name      string
	Storage   *StorageMetadata
	Calls     *[]v9.FunctionMetadata
	Events    *[]v9.EventMetadata
	Constants []v9.ModuleConstantMetadata
	Errors    []v9.ErrorMetadata
}

type Metadata struct {
	Modules   []ModuleMetadata
	Extrinsic ExtrinsicMetadata
}

type StorageEntryType struct {
	Kind      string
	Plain     *string
	Map       *StorageEntryMap
	DoubleMap *StorageEntryDoubleMap
}

type StorageEntryMap struct {
	Hasher StorageHasher
	Key    string
	Value  string
	Linked bool
}

type StorageHasher int

const (
	StorageHasherBlake2_128       StorageHasher = 0
	StorageHasherBlake2_256       StorageHasher = 1
	StorageHasherBlake2_128Concat StorageHasher = 2
	StorageHasherTwox128          StorageHasher = 3
	StorageHasherTwox256          StorageHasher = 4
	StorageHasherTwox64Concat     StorageHasher = 5
	StorageHasherIdentity         StorageHasher = 6
)

type StorageEntryDoubleMap struct {
	Hasher     StorageHasher
	Key1       string
	Key2       string
	Value      string
	Key2Hasher StorageHasher
}

type StorageEntryMetadata struct {
	Name     string
	Modifier v9.StorageEntryModifier
	Type     StorageEntryType
	Fallback []byte
	Docs     []string
}

type ExtrinsicMetadata struct {
	Version          uint8
	SignedExtensions []string
}

type StorageMetadata struct {
	Prefix string
	Items  []StorageEntryMetadata
}
