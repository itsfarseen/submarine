package v13

import (
	"fmt"
	"submarine/scale"
	"submarine/scale/gen/v11"
	"submarine/scale/gen/v9"
)

type StorageEntryDoubleMap struct {
	Hasher     v11.StorageHasher
	Key1       string
	Key2       string
	Value      string
	Key2Hasher v11.StorageHasher
}

func DecodeStorageEntryDoubleMap(reader *scale.Reader) (StorageEntryDoubleMap, error) {
	var t StorageEntryDoubleMap
	var err error

	t.Hasher, err = v11.DecodeStorageHasher(reader)
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

	t.Key2Hasher, err = v11.DecodeStorageHasher(reader)
	if err != nil {
		return t, fmt.Errorf("field Key2Hasher: %w", err)
	}

	return t, nil
}

type StorageEntryNMap struct {
	KeyVec  []string
	Hashers []v11.StorageHasher
	Value   string
}

func DecodeStorageEntryNMap(reader *scale.Reader) (StorageEntryNMap, error) {
	var t StorageEntryNMap
	var err error

	t.KeyVec, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field KeyVec: %w", err)
	}

	t.Hashers, err = scale.DecodeVec(reader, func(reader *scale.Reader) (v11.StorageHasher, error) { return v11.DecodeStorageHasher(reader) })
	if err != nil {
		return t, fmt.Errorf("field Hashers: %w", err)
	}

	t.Value, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Value: %w", err)
	}

	return t, nil
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

type StorageEntryMap struct {
	Hasher v11.StorageHasher
	Key    string
	Value  string
	Linked bool
}

func DecodeStorageEntryMap(reader *scale.Reader) (StorageEntryMap, error) {
	var t StorageEntryMap
	var err error

	t.Hasher, err = v11.DecodeStorageHasher(reader)
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

type StorageEntryTypeKind byte

const (
	StorageEntryTypeKindPlain     StorageEntryTypeKind = 0
	StorageEntryTypeKindMap       StorageEntryTypeKind = 1
	StorageEntryTypeKindDoubleMap StorageEntryTypeKind = 2
	StorageEntryTypeKindNMap      StorageEntryTypeKind = 3
)

type StorageEntryType struct {
	Kind      StorageEntryTypeKind
	Plain     *string
	Map       *StorageEntryMap
	DoubleMap *StorageEntryDoubleMap
	NMap      *StorageEntryNMap
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

	case StorageEntryTypeKindNMap:
		value, err := DecodeStorageEntryNMap(reader)
		if err != nil {
			return t, fmt.Errorf("field NMap: %w", err)
		}
		t.NMap = &value
		return t, nil

	default:
		return t, fmt.Errorf("unknown tag: %d", tag)
	}
}

type StorageEntryMetadata struct {
	Name     string
	Modifier v9.StorageEntryModifier
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

	t.Modifier, err = v9.DecodeStorageEntryModifier(reader)
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
