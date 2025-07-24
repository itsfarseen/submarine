package v12

import (
	"fmt"
	"log"
	. "submarine/scale"
)

type StorageEntryModifier byte
type StorageHasher byte

type Metadata struct {
	Modules   []ModuleMetadata
	Extrinsic ExtrinsicMetadataV12
}

type ModuleMetadata struct {
	Name      Text
	Storage   Option[StorageMetadata]
	Calls     Option[[]FunctionMetadata]
	Events    Option[[]EventMetadata]
	Constants []ModuleConstantMetadata
	Errors    []ErrorMetadata
}

type StorageMetadata struct {
	Prefix Text
	Items  []StorageEntryMetadata
}

type StorageEntryMetadata struct {
	Name     Text
	Modifier StorageEntryModifier
	Type     StorageEntryType
	Fallback Bytes
	Docs     []Text
}

type StorageEntryTypeKind byte

const (
	KindPlain StorageEntryTypeKind = iota
	KindMap
	KindDoubleMap
)

type StorageEntryType struct {
	Kind      StorageEntryTypeKind
	Plain     Text
	Map       MapType
	DoubleMap DoubleMapType
}

type MapType struct {
	Hasher StorageHasher
	Key    Text
	Value  Text
	Linked bool
}

type DoubleMapType struct {
	Hasher     StorageHasher
	Key1       Text
	Key2       Text
	Value      Text
	Key2Hasher StorageHasher
}

type FunctionMetadata struct {
	Name Text
	Args []FunctionArgumentMetadata
	Docs []Text
}

type FunctionArgumentMetadata struct {
	Name Text
	Type Text
}

type EventMetadata struct {
	Name Text
	Args []Text
	Docs []Text
}

type ErrorMetadata struct {
	Name Text
	Docs []Text
}

type ModuleConstantMetadata struct {
	Name  Text
	Type  Text
	Value Bytes
	Docs  []Text
}

type ExtrinsicMetadataV12 struct {
	Version          uint8
	SignedExtensions []Text
}

// Decoders

func DecodeMetadata(r *Reader) (Metadata, error) {
	var result Metadata
	var err error

	result.Modules, err = DecodeVec(r, DecodeModuleMetadata)
	if err != nil {
		return result, fmt.Errorf("modules: %w", err)
	}

	result.Extrinsic, err = DecodeExtrinsicMetadataV12(r)
	if err != nil {
		return result, fmt.Errorf("extrinsic: %w", err)
	}

	return result, nil
}

func DecodeModuleMetadata(r *Reader) (ModuleMetadata, error) {
	var result ModuleMetadata
	var err error

	result.Name, err = DecodeText(r)
	if err != nil {
		return result, fmt.Errorf("name: %w", err)
	}
	log.Print("DBG", result)

	result.Storage, err = DecodeOption(r, DecodeStorageMetadata)
	if err != nil {
		return result, fmt.Errorf("storage: %w", err)
	}

	result.Calls, err = DecodeOption(r, func(r *Reader) ([]FunctionMetadata, error) {
		return DecodeVec(r, DecodeFunctionMetadata)
	})

	if err != nil {
		return result, fmt.Errorf("calls: %w", err)
	}

	result.Events, err = DecodeOption(r, func(r *Reader) ([]EventMetadata, error) {
		return DecodeVec(r, DecodeEventMetadata)
	})
	if err != nil {
		return result, fmt.Errorf("events: %w", err)
	}

	result.Constants, err = DecodeVec(r, DecodeModuleConstantMetadata)
	if err != nil {
		return result, fmt.Errorf("constants: %w", err)
	}

	result.Errors, err = DecodeVec(r, DecodeErrorMetadata)
	if err != nil {
		return result, fmt.Errorf("errors: %w", err)
	}

	return result, err
}

func DecodeStorageMetadata(r *Reader) (StorageMetadata, error) {
	var result StorageMetadata
	var err error
	result.Prefix, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Items, err = DecodeVec(r, DecodeStorageEntryMetadata)
	return result, err
}

func DecodeStorageEntryMetadata(r *Reader) (StorageEntryMetadata, error) {
	var result StorageEntryMetadata
	var err error
	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}

	modifier, err := DecodeU8(r)
	if err != nil {
		return result, err
	}
	result.Modifier = StorageEntryModifier(modifier)

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

func DecodeStorageEntryType(r *Reader) (StorageEntryType, error) {
	variant, err := r.ReadByte()
	if err != nil {
		return StorageEntryType{}, err
	}

	var result StorageEntryType
	result.Kind = StorageEntryTypeKind(variant)

	switch result.Kind {
	case KindPlain:
		result.Plain, err = DecodeText(r)
		if err != nil {
			return result, err
		}
	case KindMap:
		hasher, err := DecodeU8(r)
		if err != nil {
			return result, err
		}
		result.Map.Hasher = StorageHasher(hasher)

		result.Map.Key, err = DecodeText(r)
		if err != nil {
			return result, err
		}

		result.Map.Value, err = DecodeText(r)
		if err != nil {
			return result, err
		}

		result.Map.Linked, err = DecodeBool(r)
		if err != nil {
			return result, err
		}
	case KindDoubleMap:
		hasher, err := DecodeU8(r)
		if err != nil {
			return result, err
		}
		result.DoubleMap.Hasher = StorageHasher(hasher)

		result.DoubleMap.Key1, err = DecodeText(r)
		if err != nil {
			return result, err
		}

		result.DoubleMap.Key2, err = DecodeText(r)
		if err != nil {
			return result, err
		}

		result.DoubleMap.Value, err = DecodeText(r)
		if err != nil {
			return result, err
		}

		hasher, err = DecodeU8(r)
		if err != nil {
			return result, err
		}
		result.DoubleMap.Key2Hasher = StorageHasher(hasher)
	default:
		err = fmt.Errorf("unknown StorageEntryType variant: %d", variant)
	}
	return result, err
}

func DecodeFunctionMetadata(r *Reader) (FunctionMetadata, error) {
	var result FunctionMetadata
	var err error
	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Args, err = DecodeVec(r, DecodeFunctionArgumentMetadata)
	if err != nil {
		return result, err
	}
	result.Docs, err = DecodeVec(r, DecodeText)
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
	return result, err
}

func DecodeEventMetadata(r *Reader) (EventMetadata, error) {
	var result EventMetadata
	var err error
	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Args, err = DecodeVec(r, DecodeText)
	if err != nil {
		return result, err
	}
	result.Docs, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodeErrorMetadata(r *Reader) (ErrorMetadata, error) {
	var result ErrorMetadata
	var err error
	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Docs, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodeModuleConstantMetadata(r *Reader) (ModuleConstantMetadata, error) {
	var result ModuleConstantMetadata
	var err error
	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Type, err = DecodeText(r)
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

func DecodeExtrinsicMetadataV12(r *Reader) (ExtrinsicMetadataV12, error) {
	var result ExtrinsicMetadataV12
	var err error
	result.Version, err = DecodeU8(r)
	if err != nil {
		return result, err
	}
	result.SignedExtensions, err = DecodeVec(r, DecodeText)
	return result, err
}
