package v12

import (
	"fmt"
	"submarine/scale"
	"submarine/scale/gen/v11"
	"submarine/scale/gen/v9"
)

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

	t.Extrinsic, err = v11.DecodeExtrinsicMetadata(reader)
	if err != nil {
		return t, fmt.Errorf("field Extrinsic: %w", err)
	}

	return t, nil
}

type ModuleMetadata struct {
	Name      string
	Storage   *v11.StorageMetadata
	Calls     *[]v9.FunctionMetadata
	Events    *[]v9.EventMetadata
	Constants []v9.ModuleConstantMetadata
	Errors    []v9.ErrorMetadata
	Index     uint8
}

func DecodeModuleMetadata(reader *scale.Reader) (ModuleMetadata, error) {
	var t ModuleMetadata
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Storage, err = scale.DecodeOption(reader, func(reader *scale.Reader) (v11.StorageMetadata, error) { return v11.DecodeStorageMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Storage: %w", err)
	}

	t.Calls, err = scale.DecodeOption(reader, func(reader *scale.Reader) ([]v9.FunctionMetadata, error) {
		return scale.DecodeVec(reader, func(reader *scale.Reader) (v9.FunctionMetadata, error) { return v9.DecodeFunctionMetadata(reader) })
	})
	if err != nil {
		return t, fmt.Errorf("field Calls: %w", err)
	}

	t.Events, err = scale.DecodeOption(reader, func(reader *scale.Reader) ([]v9.EventMetadata, error) {
		return scale.DecodeVec(reader, func(reader *scale.Reader) (v9.EventMetadata, error) { return v9.DecodeEventMetadata(reader) })
	})
	if err != nil {
		return t, fmt.Errorf("field Events: %w", err)
	}

	t.Constants, err = scale.DecodeVec(reader, func(reader *scale.Reader) (v9.ModuleConstantMetadata, error) {
		return v9.DecodeModuleConstantMetadata(reader)
	})
	if err != nil {
		return t, fmt.Errorf("field Constants: %w", err)
	}

	t.Errors, err = scale.DecodeVec(reader, func(reader *scale.Reader) (v9.ErrorMetadata, error) { return v9.DecodeErrorMetadata(reader) })
	if err != nil {
		return t, fmt.Errorf("field Errors: %w", err)
	}

	t.Index, err = scale.DecodeU8(reader)
	if err != nil {
		return t, fmt.Errorf("field Index: %w", err)
	}

	return t, nil
}
