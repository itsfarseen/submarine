package v12

import (
	"fmt"
	"submarine/scale"
	"submarine/scale/gen/v11"
)

type ErrorMetadata = v11.ErrorMetadata

func DecodeErrorMetadata(reader *scale.Reader) (ErrorMetadata, error) {
	return v11.DecodeErrorMetadata(reader)
}

type EventMetadata = v11.EventMetadata

func DecodeEventMetadata(reader *scale.Reader) (EventMetadata, error) {
	return v11.DecodeEventMetadata(reader)
}

type ExtrinsicMetadata = v11.ExtrinsicMetadata

func DecodeExtrinsicMetadata(reader *scale.Reader) (ExtrinsicMetadata, error) {
	return v11.DecodeExtrinsicMetadata(reader)
}

type FunctionArgumentMetadata = v11.FunctionArgumentMetadata

func DecodeFunctionArgumentMetadata(reader *scale.Reader) (FunctionArgumentMetadata, error) {
	return v11.DecodeFunctionArgumentMetadata(reader)
}

type FunctionMetadata = v11.FunctionMetadata

func DecodeFunctionMetadata(reader *scale.Reader) (FunctionMetadata, error) {
	return v11.DecodeFunctionMetadata(reader)
}

type Metadata struct {
	Modules   []ModuleMetadata
	Extrinsic v11.ExtrinsicMetadata
}

func DecodeMetadata(reader *scale.Reader) (Metadata, error) {
	var t Metadata
	var err error

	t.Modules, err = scale.DecodeVec(reader, func(reader *scale.Reader) (ModuleMetadata, error) { return DecodeModuleMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Modules: %w", err)
	}

	t.Extrinsic, err = DecodeExtrinsicMetadata(reader)
	if err != nil {
		return t, fmt.Errorf("field Extrinsic: %w", err)
	}

	return t, nil
}

type ModuleConstantMetadata = v11.ModuleConstantMetadata

func DecodeModuleConstantMetadata(reader *scale.Reader) (ModuleConstantMetadata, error) {
	return v11.DecodeModuleConstantMetadata(reader)
}

type ModuleMetadata struct {
	Name      string
	Storage   *v11.StorageMetadata
	Calls     *[]v11.FunctionMetadata
	Events    *[]v11.EventMetadata
	Constants []v11.ModuleConstantMetadata
	Errors    []v11.ErrorMetadata
	Index     uint8
}

func DecodeModuleMetadata(reader *scale.Reader) (ModuleMetadata, error) {
	var t ModuleMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Storage, err = scale.DecodeOption(reader, func(reader *scale.Reader) (v11.StorageMetadata, error) { return DecodeStorageMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Storage: %w", err)
	}

	t.Calls, err = scale.DecodeOption(reader, func(reader *scale.Reader) ([]v11.FunctionMetadata, error) {
		return scale.DecodeVec(reader, func(reader *scale.Reader) (v11.FunctionMetadata, error) { return DecodeFunctionMetadata(reader) })
	})
	if err != nil {
		return t, fmt.Errorf("field Calls: %w", err)
	}

	t.Events, err = scale.DecodeOption(reader, func(reader *scale.Reader) ([]v11.EventMetadata, error) {
		return scale.DecodeVec(reader, func(reader *scale.Reader) (v11.EventMetadata, error) { return DecodeEventMetadata(reader) })
	})
	if err != nil {
		return t, fmt.Errorf("field Events: %w", err)
	}

	t.Constants, err = scale.DecodeVec(reader, func(reader *scale.Reader) (v11.ModuleConstantMetadata, error) {
		return DecodeModuleConstantMetadata(reader)
	})
	if err != nil {
		return t, fmt.Errorf("field Constants: %w", err)
	}

	t.Errors, err = scale.DecodeVec(reader, func(reader *scale.Reader) (v11.ErrorMetadata, error) { return DecodeErrorMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Errors: %w", err)
	}

	t.Index, err = scale.DecodeU8(reader)
	if err != nil {
		return t, fmt.Errorf("field Index: %w", err)
	}

	return t, nil
}

type StorageEntryMetadata = v11.StorageEntryMetadata

func DecodeStorageEntryMetadata(reader *scale.Reader) (StorageEntryMetadata, error) {
	return v11.DecodeStorageEntryMetadata(reader)
}

type StorageEntryModifier = v11.StorageEntryModifier

func DecodeStorageEntryModifier(reader *scale.Reader) (StorageEntryModifier, error) {
	return v11.DecodeStorageEntryModifier(reader)
}

type StorageEntryType = v11.StorageEntryType

func DecodeStorageEntryType(reader *scale.Reader) (StorageEntryType, error) {
	return v11.DecodeStorageEntryType(reader)
}

type StorageHasher = v11.StorageHasher

func DecodeStorageHasher(reader *scale.Reader) (StorageHasher, error) {
	return v11.DecodeStorageHasher(reader)
}

type StorageMetadata = v11.StorageMetadata

func DecodeStorageMetadata(reader *scale.Reader) (StorageMetadata, error) {
	return v11.DecodeStorageMetadata(reader)
}
