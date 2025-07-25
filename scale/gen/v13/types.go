package v13

import (
	"submarine/scale/gen/v11"
	"submarine/scale/gen/v9"
)

type StorageEntryNMap struct {
	KeyVec  []string
	Hashers []v11.StorageHasher
	Value   string
}

type StorageEntryMap struct {
	Hasher v11.StorageHasher
	Key    string
	Value  string
	Linked bool
}

type StorageEntryDoubleMap struct {
	Hasher     v11.StorageHasher
	Key1       string
	Key2       string
	Value      string
	Key2Hasher v11.StorageHasher
}

type StorageEntryMetadata struct {
	Name     string
	Modifier v9.StorageEntryModifier
	Type     StorageEntryType
	Fallback []byte
	Docs     []string
}

type StorageMetadata struct {
	Prefix string
	Items  []StorageEntryMetadata
}

type StorageEntryType struct {
	Kind      string
	Plain     *string
	Map       *StorageEntryMap
	DoubleMap *StorageEntryDoubleMap
	NMap      *StorageEntryNMap
}
