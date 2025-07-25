package v9

import (
	"fmt"
	"submarine/scale"
)

type EventMetadata struct {
	Name string
	Args []string
	Docs []string
}

func DecodeEventMetadata(reader *scale.Reader) (EventMetadata, error) {
	var t EventMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Args, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Args: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	return t, nil
}

type ModuleMetadata struct {
	Name      string
	Storage   *StorageMetadata
	Calls     *[]FunctionMetadata
	Events    *[]EventMetadata
	Constants []ModuleConstantMetadata
	Errors    []ErrorMetadata
}

func DecodeModuleMetadata(reader *scale.Reader) (ModuleMetadata, error) {
	var t ModuleMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Storage, err = scale.DecodeOption(reader, func(reader *scale.Reader) (StorageMetadata, error) { return DecodeStorageMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Storage: %w", err)
	}

	t.Calls, err = scale.DecodeOption(reader, func(reader *scale.Reader) ([]FunctionMetadata, error) {
		return scale.DecodeVec(reader, func(reader *scale.Reader) (FunctionMetadata, error) { return DecodeFunctionMetadata(reader) })
	})
	if err != nil {
		return t, fmt.Errorf("field Calls: %w", err)
	}

	t.Events, err = scale.DecodeOption(reader, func(reader *scale.Reader) ([]EventMetadata, error) {
		return scale.DecodeVec(reader, func(reader *scale.Reader) (EventMetadata, error) { return DecodeEventMetadata(reader) })
	})
	if err != nil {
		return t, fmt.Errorf("field Events: %w", err)
	}

	t.Constants, err = scale.DecodeVec(reader, func(reader *scale.Reader) (ModuleConstantMetadata, error) {
		return DecodeModuleConstantMetadata(reader)
	})
	if err != nil {
		return t, fmt.Errorf("field Constants: %w", err)
	}

	t.Errors, err = scale.DecodeVec(reader, func(reader *scale.Reader) (ErrorMetadata, error) { return DecodeErrorMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Errors: %w", err)
	}

	return t, nil
}

type StorageEntryMetadata struct {
	Name     string
	Modifier StorageEntryModifier
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

type StorageEntryModifier int

const (
	StorageEntryModifierOptional StorageEntryModifier = 0
	StorageEntryModifierDefault  StorageEntryModifier = 1
	StorageEntryModifierRequired StorageEntryModifier = 2
)

func DecodeStorageEntryModifier(reader *scale.Reader) (StorageEntryModifier, error) {

	tag, err := reader.ReadByte()
	if err != nil {
		var t StorageEntryModifier
		return t, fmt.Errorf("enum tag: %w", err)
	}

	switch tag {

	case 0:
		return StorageEntryModifierOptional, nil

	case 1:
		return StorageEntryModifierDefault, nil

	case 2:
		return StorageEntryModifierRequired, nil

	default:
		var t StorageEntryModifier
		return t, fmt.Errorf("unknown tag: %d", tag)
	}
}

type StorageMetadata struct {
	Prefix string
	Items  []StorageEntryMetadata
}

func DecodeStorageMetadata(reader *scale.Reader) (StorageMetadata, error) {
	var t StorageMetadata
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

type ErrorMetadata struct {
	Name string
	Docs []string
}

func DecodeErrorMetadata(reader *scale.Reader) (ErrorMetadata, error) {
	var t ErrorMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	return t, nil
}

type FunctionArgumentMetadata struct {
	Name string
	Type string
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

	return t, nil
}

type FunctionMetadata struct {
	Name string
	Args []FunctionArgumentMetadata
	Docs []string
}

func DecodeFunctionMetadata(reader *scale.Reader) (FunctionMetadata, error) {
	var t FunctionMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Args, err = scale.DecodeVec(reader, func(reader *scale.Reader) (FunctionArgumentMetadata, error) {
		return DecodeFunctionArgumentMetadata(reader)
	})
	if err != nil {
		return t, fmt.Errorf("field Args: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	return t, nil
}

type ModuleConstantMetadata struct {
	Name  string
	Type  string
	Value []byte
	Docs  []string
}

func DecodeModuleConstantMetadata(reader *scale.Reader) (ModuleConstantMetadata, error) {
	var t ModuleConstantMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Type, err = scale.DecodeText(reader)
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

type StorageEntryTypeKind byte

const (
	StorageEntryTypeKindPlain     StorageEntryTypeKind = 0
	StorageEntryTypeKindMap       StorageEntryTypeKind = 1
	StorageEntryTypeKindDoubleMap StorageEntryTypeKind = 2
)

type StorageEntryType struct {
	Kind      StorageEntryTypeKind
	Plain     *string
	Map       *StorageEntryMap
	DoubleMap *StorageEntryDoubleMap
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
		value, err := scale.DecodeText(reader)
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

	case StorageEntryTypeKindDoubleMap:
		value, err := DecodeStorageEntryDoubleMap(reader)
		if err != nil {
			return t, fmt.Errorf("field DoubleMap: %w", err)
		}
		t.DoubleMap = &value
		return t, nil

	default:
		return t, fmt.Errorf("unknown tag: %d", tag)
	}
}

type StorageEntryDoubleMap struct {
	Hasher     StorageHasher
	Key1       string
	Key2       string
	Value      string
	Key2Hasher StorageHasher
}

func DecodeStorageEntryDoubleMap(reader *scale.Reader) (StorageEntryDoubleMap, error) {
	var t StorageEntryDoubleMap
	var err error

	t.Hasher, err = DecodeStorageHasher(reader)
	if err != nil {
		return t, fmt.Errorf("field Hasher: %w", err)
	}

	t.Key1, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Key1: %w", err)
	}

	t.Key2, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Key2: %w", err)
	}

	t.Value, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Value: %w", err)
	}

	t.Key2Hasher, err = DecodeStorageHasher(reader)
	if err != nil {
		return t, fmt.Errorf("field Key2Hasher: %w", err)
	}

	return t, nil
}

type StorageHasher int

const (
	StorageHasherBlake2_128   StorageHasher = 0
	StorageHasherBlake2_256   StorageHasher = 1
	StorageHasherTwox128      StorageHasher = 2
	StorageHasherTwox256      StorageHasher = 3
	StorageHasherTwox64Concat StorageHasher = 4
)

func DecodeStorageHasher(reader *scale.Reader) (StorageHasher, error) {

	tag, err := reader.ReadByte()
	if err != nil {
		var t StorageHasher
		return t, fmt.Errorf("enum tag: %w", err)
	}

	switch tag {

	case 0:
		return StorageHasherBlake2_128, nil

	case 1:
		return StorageHasherBlake2_256, nil

	case 2:
		return StorageHasherTwox128, nil

	case 3:
		return StorageHasherTwox256, nil

	case 4:
		return StorageHasherTwox64Concat, nil

	default:
		var t StorageHasher
		return t, fmt.Errorf("unknown tag: %d", tag)
	}
}

type StorageEntryMap struct {
	Hasher StorageHasher
	Key    string
	Value  string
	Linked bool
}

func DecodeStorageEntryMap(reader *scale.Reader) (StorageEntryMap, error) {
	var t StorageEntryMap
	var err error

	t.Hasher, err = DecodeStorageHasher(reader)
	if err != nil {
		return t, fmt.Errorf("field Hasher: %w", err)
	}

	t.Key, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Key: %w", err)
	}

	t.Value, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Value: %w", err)
	}

	t.Linked, err = scale.DecodeBool(reader)
	if err != nil {
		return t, fmt.Errorf("field Linked: %w", err)
	}

	return t, nil
}

type Metadata struct {
	Modules []ModuleMetadata
}

func DecodeMetadata(reader *scale.Reader) (Metadata, error) {
	var t Metadata
	var err error

	t.Modules, err = scale.DecodeVec(reader, func(reader *scale.Reader) (ModuleMetadata, error) { return DecodeModuleMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Modules: %w", err)
	}

	return t, nil
}
