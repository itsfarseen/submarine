package v11

import (
	"fmt"
	. "submarine/scale"
	"submarine/scale/v10"
)

// v11 reuses most of v10, but adds an Extrinsic field to Metadata and Index to ModuleMetadata.

// Reused types from v10
type StorageMetadata v10.StorageMetadata
type StorageEntryMetadata v10.StorageEntryMetadata
type StorageEntryModifier v10.StorageEntryModifier
type StorageHasher v10.StorageHasher
type StorageEntryType v10.StorageEntryType
type MapType v10.MapType
type DoubleMapType v10.DoubleMapType
type FunctionMetadata v10.FunctionMetadata
type FunctionArgumentMetadata v10.FunctionArgumentMetadata
type EventMetadata v10.EventMetadata
type ErrorMetadata v10.ErrorMetadata
type ModuleConstantMetadata v10.ModuleConstantMetadata

// v11 specific types
type Metadata struct {
	Modules   []ModuleMetadata
	Extrinsic ExtrinsicMetadataV11
}

type ModuleMetadata struct {
	Name      Text
	Storage   Option[StorageMetadata]
	Calls     Option[[]FunctionMetadata]
	Events    Option[[]EventMetadata]
	Constants []ModuleConstantMetadata
	Errors    []ErrorMetadata
	Index     uint8
}

type ExtrinsicMetadataV11 struct {
	Version          uint8
	SignedExtensions []Text
}

// Decoders

func DecodeExtrinsicMetadataV11(r *Reader) (ExtrinsicMetadataV11, error) {
	var result ExtrinsicMetadataV11
	var err error
	result.Version, err = DecodeU8(r)
	if err != nil {
		return result, err
	}
	result.SignedExtensions, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodeModuleMetadata(r *Reader) (ModuleMetadata, error) {
	var result ModuleMetadata
	var err error

	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}

	result.Storage, err = DecodeOption(r, func(r *Reader) (StorageMetadata, error) {
		v10val, err := v10.DecodeStorageMetadata(r)
		return StorageMetadata(v10val), err
	})
	if err != nil {
		return result, err
	}

	result.Calls, err = DecodeOption(r, func(r *Reader) ([]FunctionMetadata, error) {
		return DecodeVec(r, func(r *Reader) (FunctionMetadata, error) {
			v10val, err := v10.DecodeFunctionMetadata(r)
			return FunctionMetadata(v10val), err
		})
	})
	if err != nil {
		return result, err
	}

	result.Events, err = DecodeOption(r, func(r *Reader) ([]EventMetadata, error) {
		return DecodeVec(r, func(r *Reader) (EventMetadata, error) {
			v10val, err := v10.DecodeEventMetadata(r)
			return EventMetadata(v10val), err
		})
	})
	if err != nil {
		return result, err
	}

	result.Constants, err = DecodeVec(r, func(r *Reader) (ModuleConstantMetadata, error) {
		v10val, err := v10.DecodeModuleConstantMetadata(r)
		return ModuleConstantMetadata(v10val), err
	})
	if err != nil {
		return result, err
	}

	result.Errors, err = DecodeVec(r, func(r *Reader) (ErrorMetadata, error) {
		v10val, err := v10.DecodeErrorMetadata(r)
		return ErrorMetadata(v10val), err
	})
	if err != nil {
		return result, err
	}

	result.Index, err = DecodeU8(r)
	return result, err
}

func DecodeMetadata(r *Reader) (Metadata, error) {
	var result Metadata
	var err error

	result.Modules, err = DecodeVec(r, DecodeModuleMetadata)
	if err != nil {
		return result, fmt.Errorf("modules: %w", err)
	}

	result.Extrinsic, err = DecodeExtrinsicMetadataV11(r)
	if err != nil {
		return result, fmt.Errorf("extrinsic: %w", err)
	}

	return result, nil
}
