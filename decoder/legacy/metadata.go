package legacy

import (
	"fmt"
	v10 "submarine/scale/gen/v10"
	v11 "submarine/scale/gen/v11"
	v12 "submarine/scale/gen/v12"
	v9 "submarine/scale/gen/v9"
)

type MetadataIndexKind int

const (
	MetadataIndexKindV9  = 9
	MetadataIndexKindV12 = 12
)

type Metadata struct {
	Modules   []Module
	IndexKind MetadataIndexKind
	IndexV9   *MetadataIndexV9
	IndexV12  *MetadataIndexV12
}

type MetadataIndexV9 struct {
	EventfulModules []int
	CallfulModules  []int
}

type MetadataIndexV12 struct {
	IndexedModules map[int]int
}

type Module struct {
	Name   string
	Index  int // only available in v12+
	Calls  []Call
	Events []Event
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

type Event struct {
	Name string
	Args []string
	Docs []string
}

func MakeMetadataFromAny(m any) (Metadata, error) {
	var modules []Module
	var indexKind MetadataIndexKind

	switch v := m.(type) {
	case v9.Metadata:
		indexKind = MetadataIndexKindV9
		modules := make([]Module, len(v.Modules))
		for i, module := range v.Modules {
			modules[i] = MakeModuleFromV9Parts(module.Name, module.Calls, module.Events)
		}
	case v10.Metadata:
		indexKind = MetadataIndexKindV9
		modules := make([]Module, len(v.Modules))
		for i, module := range v.Modules {
			modules[i] = MakeModuleFromV9Parts(module.Name, module.Calls, module.Events)
		}
	case v11.Metadata:
		indexKind = MetadataIndexKindV9
		modules := make([]Module, len(v.Modules))
		for i, module := range v.Modules {
			modules[i] = MakeModuleFromV9Parts(module.Name, module.Calls, module.Events)
		}
	case v12.Metadata:
		indexKind = MetadataIndexKindV12
		modules := make([]Module, len(v.Modules))
		for i, module := range v.Modules {
			modules[i] = MakeModuleFromV9Parts(module.Name, module.Calls, module.Events)
			modules[i].Index = int(module.Index)
		}
	default:
		return Metadata{}, fmt.Errorf("not a valid v9-v12 metadata struct")
	}
	return MakeMetadataFromModules(modules, indexKind), nil
}

func MakeMetadataFromModules(modules []Module, indexKind MetadataIndexKind) Metadata {
	var metadata Metadata
	metadata.Modules = modules
	metadata.IndexKind = indexKind

	switch indexKind {
	case MetadataIndexKindV9:
		var index MetadataIndexV9
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
	case MetadataIndexKindV12:
		var index MetadataIndexV12
		metadata.IndexV12 = &index
		for i, module := range metadata.Modules {
			metadata.Modules = append(metadata.Modules, module)
			index.IndexedModules[module.Index] = i
		}
	default:
		panic("must be exhaustive")
	}

	return metadata
}

func MakeModuleFromV9Parts(name string, calls *[]v9.FunctionMetadata, events *[]v9.EventMetadata) Module {
	var module Module

	module.Name = name

	for _, call := range *calls {
		call_ := MakeCallFromV9(call)
		module.Calls = append(module.Calls, call_)
	}

	for _, event := range *events {
		event_ := MakeEventFromV9(event)
		module.Events = append(module.Events, event_)
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

func MakeEventFromV9(m v9.EventMetadata) Event {
	return Event(m)
}
