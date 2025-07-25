package v12

import (
	"submarine/scale/gen/v11"
	"submarine/scale/gen/v9"
)

type Metadata struct {
	Modules   []ModuleMetadata
	Extrinsic v11.ExtrinsicMetadata
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
