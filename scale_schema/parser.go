package scale_schema

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	. "submarine/errorspan"
)

func ParseModuleFiles(files []string) (*AllModules, error) {
	allModules := &AllModules{
		ModuleNames: make([]string, 0),
		Modules:     make(map[string]Module),
	}

	for _, file := range files {
		moduleName := strings.TrimSuffix(filepath.Base(file), ".yaml")
		module, err := ParseModuleFile(file)
		if err != nil {
			return nil, err.WithPath(file)
		}
		allModules.ModuleNames = append(allModules.ModuleNames, moduleName)
		allModules.Modules[moduleName] = module
	}

	return allModules, nil
}

func ParseModuleFile(file string) (Module, *ErrorSpan) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Module{}, NewErrorSpan(fmt.Sprintf("read file: %v", err))
	}

	var rawModuleDef map[string]any
	if err := yaml.Unmarshal(data, &rawModuleDef); err != nil {
		return Module{}, NewErrorSpan(fmt.Sprintf("unmarshal yaml: %v", err))
	}

	module, err_ := ParseModule(rawModuleDef)
	if err_ != nil {
		return Module{}, err_
	}

	return module, nil
}

func ParseModule(rawModuleDef map[string]any) (Module, *ErrorSpan) {
	module := Module{
		Types: make(map[string]*Type, len(rawModuleDef)),
	}

	for typeName, rawTypeDef := range rawModuleDef {
		parsedType, err := ParseType(rawTypeDef)
		if err != nil {
			return module, err.WithPath(typeName)
		}
		module.Types[typeName] = parsedType
		module.TypeNames = append(module.TypeNames, typeName)
	}

	sort.Strings(module.TypeNames)

	return module, nil
}

func ParseType(rawTypeDef any) (*Type, *ErrorSpan) {
	switch v := rawTypeDef.(type) {
	case string:
		return &Type{Kind: KindRef, Ref: &v}, nil
	case map[string]any:
		return ParseComplexType(v)
	default:
		return nil, NewErrorSpan(fmt.Sprintf("unexpected type definition format: %T", rawTypeDef))
	}
}

func ParseComplexType(def map[string]any) (*Type, *ErrorSpan) {
	rawType, ok := def["type"].(string)
	if !ok {
		return nil, nil
	}

	t := &Type{}

	switch rawType {
	case "struct":
		t.Kind = KindStruct
		rawFields, ok := def["fields"].([]any)
		if !ok {
			return nil, NewErrorSpan("missing 'fields' or not a list").WithPath("struct")
		}

		members, err := ParsedNamedMembers(rawFields)
		if err != nil {
			return nil, err.WithPath("struct.fields")
		}
		t.Struct = &Struct{Fields: members}

	case "enum_simple":
		t.Kind = KindEnumSimple
		rawVariants, ok := def["variants"].([]any)
		if !ok {
			return nil, NewErrorSpan("missing 'variants' or not a list").WithPath("enum_simple")
		}
		variants := make([]string, len(rawVariants))
		for i, v := range rawVariants {
			variants[i], ok = v.(string)
			if !ok {
				return nil, NewErrorSpan("variant not a string").
					WithPathInt(i).
					WithPath("enum_simple.variants")
			}

		}
		t.EnumSimple = &EnumSimple{Variants: variants}

	case "enum_complex":
		t.Kind = KindEnumComplex
		rawVariants, ok := def["variants"].([]any)
		if !ok {
			return nil, NewErrorSpan("missing 'variants' or not a list").WithPath("enum_complex")
		}
		variants, err := ParsedNamedMembers(rawVariants)
		if err != nil {
			return nil, err.WithPath("enum_complex.variants")
		}
		t.EnumComplex = &EnumComplex{Variants: variants}

	case "import":
		t.Kind = KindImport
		module, ok := def["module"].(string)
		if !ok {
			return nil, NewErrorSpan("missing 'module' or not a string").
				WithPath("import")
		}
		item, ok := def["item"].(string)
		if !ok {
			return nil, NewErrorSpan("missing 'item' or not a string").
				WithPath("import")
		}
		t.Import = &Import{Module: module, Item: item}

	case "vec", "option":
		itemDef, ok := def["item"]
		if !ok {
			return nil, NewErrorSpan("missing 'item'").WithPath(rawType)
		}
		itemType, err := ParseType(itemDef)
		if err != nil {
			return nil, err.WithPath(rawType)
		}
		if rawType == "vec" {
			t.Kind = KindVec
			t.Vec = &Vec{Type: itemType}
		} else {
			t.Kind = KindOption
			t.Option = &Option{Type: itemType}
		}

	default:
		return nil, NewErrorSpan(fmt.Sprintf("unknown type: %s", rawType))
	}

	return t, nil
}

func ParsedNamedMembers(rawNamedMembers []any) ([]NamedMember, *ErrorSpan) {
	members := make([]NamedMember, len(rawNamedMembers))
	for i, member := range rawNamedMembers {
		memberMap, ok := member.(map[string]any)
		if !ok {
			return nil, NewErrorSpan("member is not a map").WithPathInt(i)
		}
		name, ok := memberMap["name"].(string)
		if !ok {
			return nil, NewErrorSpan("missing 'name'").WithPathInt(i)
		}
		type_, ok := memberMap["type"]
		if !ok {
			return nil, NewErrorSpan("missing 'type'").WithPathInt(i)
		}

		memberType, err := ParseType(type_)
		if err != nil {
			return nil, err.WithPathInt(i)
		}
		members[i] = NamedMember{Name: name, Type: memberType}
	}
	return members, nil
}
