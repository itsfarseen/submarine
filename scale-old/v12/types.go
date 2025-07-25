package v12

import (
	"fmt"
	. "submarine/scale"
	"submarine/scale/v11"
)

// v12 is a direct alias of v11.

type Metadata v11.Metadata
type ExtrinsicMetadataV12 v11.ExtrinsicMetadataV11
type ModuleMetadata v11.ModuleMetadata
type StorageMetadata v11.StorageMetadata
type StorageEntryMetadata v11.StorageEntryMetadata
type StorageEntryModifier v11.StorageEntryModifier
type StorageHasher v11.StorageHasher
type StorageEntryType v11.StorageEntryType
type MapType v11.MapType
type DoubleMapType v11.DoubleMapType
type FunctionMetadata v11.FunctionMetadata
type FunctionArgumentMetadata v11.FunctionArgumentMetadata
type EventMetadata v11.EventMetadata
type ErrorMetadata v11.ErrorMetadata
type ModuleConstantMetadata v11.ModuleConstantMetadata

// Decoders

func DecodeMetadata(r *Reader) (Metadata, error) {
	v11Meta, err := v11.DecodeMetadata(r)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to decode v11 metadata for v12: %w", err)
	}
	// This is a safe conversion because the struct layouts are identical.
	return Metadata(v11Meta), nil
}