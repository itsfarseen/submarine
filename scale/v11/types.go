package v11

import (
	"fmt"
	. "submarine/scale"
	"submarine/scale/v10"
)

// v11 reuses most of v10, but adds an Extrinsic field to Metadata.

// Reused types from v10
type ModuleMetadata v10.ModuleMetadata
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

func DecodeMetadata(r *Reader) (Metadata, error) {
	var result Metadata
	var err error

	// Since ModuleMetadata is an alias, we can't directly use v10.DecodeModuleMetadata
	// because the return type won't match. We must cast it.
	// A generic DecodeVecWithCast would be ideal, but for now, we'll do it manually.
	numModules, err := DecodeCompact(r)
	if err != nil {
		return result, fmt.Errorf("modules vec len: %w", err)
	}

	result.Modules = make([]ModuleMetadata, numModules.Int64())
	for i := int64(0); i < numModules.Int64(); i++ {
		// We need to decode a v10 module and cast it to a v11 module.
		// This is safe because they are aliases with identical structure.
		v10Module, err := v10.DecodeModuleMetadata(r)
		if err != nil {
			return result, fmt.Errorf("modules[%d]: %w", i, err)
		}
		result.Modules[i] = ModuleMetadata(v10Module)
	}

	result.Extrinsic, err = DecodeExtrinsicMetadataV11(r)
	if err != nil {
		return result, fmt.Errorf("extrinsic: %w", err)
	}

	return result, nil
}