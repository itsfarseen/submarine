package legacy

import (
	"fmt"
	"reflect"
	v10 "submarine/metadata/generated/v10"
	v11 "submarine/metadata/generated/v11"
	v12 "submarine/metadata/generated/v12"
	v9 "submarine/metadata/generated/v9"
)

type Metadata struct {
	Version  int
	Modules  []ModuleMetadata
	IndexV9  *MetadataIndexV9
	IndexV12 *MetadataIndexV12
}

func (m *Metadata) GetModuleForExtrinsic(index int) (ModuleMetadata, error) {
	var actualIndex int
	var ok bool

	if m.Version >= 12 {
		indexMap := m.IndexV12.IndexedModules
		actualIndex, ok = indexMap[index]
		if !ok {
			return ModuleMetadata{}, fmt.Errorf("index not found (IndexV12): %d", index)
		}
	} else if m.Version >= 9 {
		indexList := m.IndexV9.CallfulModules
		if index >= len(indexList) {
			return ModuleMetadata{}, fmt.Errorf("index not found (CallfulModules): %d", index)
		}
		actualIndex = indexList[index]
	} else {
		return ModuleMetadata{}, fmt.Errorf("unreachable")
	}

	return m.Modules[actualIndex], nil
}

func (m *Metadata) GetModuleForEvent(index int) (ModuleMetadata, error) {
	var actualIndex int
	var ok bool
	if m.Version >= 12 {
		indexMap := m.IndexV12.IndexedModules
		actualIndex, ok = indexMap[index]
		if !ok {
			return ModuleMetadata{}, fmt.Errorf("index not found (IndexV12): %d", index)
		}
	} else if m.Version >= 9 {
		indexList := m.IndexV9.EventfulModules
		if index >= len(indexList) {
			return ModuleMetadata{}, fmt.Errorf("index not found (EventfulModules): %d", index)
		}
		actualIndex = indexList[index]
	} else {
		return ModuleMetadata{}, fmt.Errorf("unreachable")
	}

	return m.Modules[actualIndex], nil
}

type MetadataIndexV9 struct {
	EventfulModules []int
	CallfulModules  []int
}

type MetadataIndexV12 struct {
	IndexedModules map[int]int
}

type ModuleMetadata struct {
	Name   string
	Index  int // only available in v12+
	Calls  []Call
	Events []EventMetadata
}

type Call struct {
	Name string
	Args []CallArgument
	Docs []string
}

type CallArgument struct {
	Name string
	Type string
}

type EventMetadata struct {
	Name string
	Args []string
	Docs []string
}

func MakeMetadataFromAny(m any) (Metadata, error) {
	var modules []ModuleMetadata
	var version int

	switch v := m.(type) {
	case *v9.Metadata:
		version = 9
		modules = make([]ModuleMetadata, len(v.Modules))
		for i, module := range v.Modules {
			modules[i] = MakeModuleFromV9Parts(module.Name, module.Calls, module.Events)
		}
	case *v10.Metadata:
		version = 10
		modules = make([]ModuleMetadata, len(v.Modules))
		for i, module := range v.Modules {
			modules[i] = MakeModuleFromV9Parts(module.Name, module.Calls, module.Events)
		}
	case *v11.Metadata:
		version = 11
		modules = make([]ModuleMetadata, len(v.Modules))
		for i, module := range v.Modules {
			modules[i] = MakeModuleFromV9Parts(module.Name, module.Calls, module.Events)
		}
	case *v12.Metadata:
		version = 12
		modules = make([]ModuleMetadata, len(v.Modules))
		for i, module := range v.Modules {
			modules[i] = MakeModuleFromV9Parts(module.Name, module.Calls, module.Events)
			modules[i].Index = int(module.Index)
		}
	default:
		return Metadata{}, fmt.Errorf("not a valid v9-v12 metadata struct: %v", reflect.TypeOf(m))
	}
	return MakeMetadataFromModules(modules, version), nil
}

func MakeMetadataFromModules(modules []ModuleMetadata, version int) Metadata {
	var metadata Metadata
	metadata.Modules = modules
	metadata.Version = version

	if metadata.Version >= 14 {
		panic("not a legacy version")
	} else if metadata.Version >= 12 {
		index := MetadataIndexV12{make(map[int]int)}
		metadata.IndexV12 = &index
		for i, module := range metadata.Modules {
			metadata.Modules = append(metadata.Modules, module)
			index.IndexedModules[module.Index] = i
		}
	} else if metadata.Version >= 9 {
		index := MetadataIndexV9{}
		metadata.IndexV9 = &index
		for i, module := range metadata.Modules {
			metadata.Modules = append(metadata.Modules, module)
			if len(module.Events) > 0 {
				index.EventfulModules = append(index.EventfulModules, i)
			}
			if len(module.Calls) > 0 {
				index.CallfulModules = append(index.CallfulModules, i)
			}
		}
	} else {
		panic("unsupported version")
	}

	return metadata
}

func MakeModuleFromV9Parts(name string, calls *[]v9.FunctionMetadata, events *[]v9.EventMetadata) ModuleMetadata {
	var module ModuleMetadata

	module.Name = name

	if calls != nil {
		for _, call := range *calls {
			call_ := MakeCallFromV9(call)
			module.Calls = append(module.Calls, call_)
		}
	}

	if events != nil {
		for _, event := range *events {
			event_ := MakeEventFromV9(event)
			module.Events = append(module.Events, event_)
		}
	}

	return module
}

func MakeCallFromV9(m v9.FunctionMetadata) Call {
	var call Call
	call.Name = m.Name
	for _, arg := range m.Args {
		arg_ := MakeCallArgFromV9(arg)
		call.Args = append(call.Args, arg_)
	}
	call.Docs = m.Docs
	return call
}

func MakeCallArgFromV9(m v9.FunctionArgumentMetadata) CallArgument {
	return CallArgument(m)
}

func MakeEventFromV9(m v9.EventMetadata) EventMetadata {
	return EventMetadata(m)
}
