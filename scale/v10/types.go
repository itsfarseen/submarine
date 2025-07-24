package v10

import (
	"fmt"
	. "submarine/scale"
	"submarine/scale/v9"
)

// v10 is a direct extension of v9, with only a minor change to StorageHasher.
// Since our StorageHasher is a simple byte, we can reuse the v9 structures directly.

type Metadata v9.Metadata
type ModuleMetadata v9.ModuleMetadata
type StorageMetadata v9.StorageMetadata
type StorageEntryMetadata v9.StorageEntryMetadata
type StorageEntryModifier v9.StorageEntryModifier
type StorageHasher v9.StorageHasher
type StorageEntryType v9.StorageEntryType
type MapType v9.MapType
type DoubleMapType v9.DoubleMapType
type FunctionMetadata v9.FunctionMetadata
type FunctionArgumentMetadata v9.FunctionArgumentMetadata
type EventMetadata v9.EventMetadata
type ErrorMetadata v9.ErrorMetadata
type ModuleConstantMetadata v9.ModuleConstantMetadata

// Decoders

func DecodeMetadata(r *Reader) (Metadata, error) {
	v9Meta, err := v9.DecodeMetadata(r)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to decode v9 metadata for v10: %w", err)
	}
	// This is a safe conversion because the struct layouts are identical.
	return Metadata(v9Meta), nil
}

// DecodeModuleMetadata is needed for v11 to call into.
func DecodeModuleMetadata(r *Reader) (ModuleMetadata, error) {
	v9Module, err := v9.DecodeModuleMetadata(r)
	if err != nil {
		return ModuleMetadata{}, err
	}
	return ModuleMetadata(v9Module), nil
}
