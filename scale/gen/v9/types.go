package v9

import ()

type StorageEntryMap struct {
	Hasher StorageHasher
	Key    string
	Value  string
	Linked bool
}

type ErrorMetadata struct {
	Name string
	Docs []string
}

type StorageMetadata struct {
	Prefix string
	Items  []StorageEntryMetadata
}

type FunctionArgumentMetadata struct {
	Name string
	Type string
}

type ModuleConstantMetadata struct {
	Name  string
	Type  string
	Value []byte
	Docs  []string
}

type ModuleMetadata struct {
	Name      string
	Storage   *StorageMetadata
	Calls     *[]FunctionMetadata
	Events    *[]EventMetadata
	Constants []ModuleConstantMetadata
	Errors    []ErrorMetadata
}

type StorageEntryModifier int

const (
	StorageEntryModifierOptional StorageEntryModifier = 0
	StorageEntryModifierDefault  StorageEntryModifier = 1
	StorageEntryModifierRequired StorageEntryModifier = 2
)

type EventMetadata struct {
	Name string
	Args []string
	Docs []string
}

type Metadata struct {
	Modules []ModuleMetadata
}

type StorageEntryDoubleMap struct {
	Hasher     StorageHasher
	Key1       string
	Key2       string
	Value      string
	Key2Hasher StorageHasher
}

type StorageHasher int

const (
	StorageHasherBlake2_128   StorageHasher = 0
	StorageHasherBlake2_256   StorageHasher = 1
	StorageHasherTwox128      StorageHasher = 2
	StorageHasherTwox256      StorageHasher = 3
	StorageHasherTwox64Concat StorageHasher = 4
)

type FunctionMetadata struct {
	Name string
	Args []FunctionArgumentMetadata
	Docs []string
}

type StorageEntryMetadata struct {
	Name     string
	Modifier StorageEntryModifier
	Type     StorageEntryType
	Fallback []byte
	Docs     []string
}

type StorageEntryType struct {
	Kind      string
	Plain     *string
	Map       *StorageEntryMap
	DoubleMap *StorageEntryDoubleMap
}
