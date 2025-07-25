package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type AllModules struct {
	ModuleNames []string // for preserving order
	Modules     map[string]Module
}

type Module struct {
	Types map[string]*Type
}

func Parse(files []string) (*AllModules, error) {
	allModules := &AllModules{
		ModuleNames: make([]string, 0),
		Modules:     make(map[string]Module),
	}

	for _, file := range files {
		moduleName := strings.TrimSuffix(filepath.Base(file), ".yaml")
		module, err := parseModule(file)
		if err != nil {
			return nil, err
		}
		allModules.ModuleNames = append(allModules.ModuleNames, moduleName)
		allModules.Modules[moduleName] = module
	}

	return allModules, nil
}

func parseModule(file string) (Module, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Module{}, fmt.Errorf("reading file %s: %w", file, err)
	}

	var rawModuleDefs map[string]any
	if err := yaml.Unmarshal(data, &rawModuleDefs); err != nil {
		return Module{}, fmt.Errorf("unmarshaling yaml from %s: %w", file, err)
	}

	module := Module{
		Types: make(map[string]*Type, len(rawModuleDefs)),
	}

	for typeName, rawDef := range rawModuleDefs {
		parsedType, err := parseType(rawDef)
		if err != nil {
			return Module{}, fmt.Errorf("parsing type '%s' in file '%s': %w", typeName, file, err)
		}
		module.Types[typeName] = parsedType
	}

	return module, nil
}

func parseType(rawDef any) (*Type, error) {
	switch v := rawDef.(type) {
	case string:
		return &Type{Kind: KindRef, Ref: &Ref{Name: v}}, nil
	case map[string]any:
		return parseTypeFromMap(v)
	default:
		return nil, fmt.Errorf("unexpected type definition format: %T", rawDef)
	}
}

func parseTypeFromMap(def map[string]any) (*Type, error) {
	rawType, ok := def["type"].(string)
	if !ok {
		return nil, fmt.Errorf("type definition must contain a 'type' field as a string")
	}

	t := &Type{}

	switch rawType {
	case "struct":
		t.Kind = KindStruct
		rawFields, ok := def["fields"].([]any)
		if !ok {
			return nil, fmt.Errorf("struct definition missing 'fields' or not a list")
		}

		members, err := parseNamedMembersFromList(rawFields)
		if err != nil {
			return nil, err
		}
		t.Struct = &Struct{Fields: members}

	case "enum_simple":
		t.Kind = KindEnumSimple
		rawVariants, _ := def["variants"].([]any)
		variants := make([]string, len(rawVariants))
		for i, v := range rawVariants {
			variants[i], _ = v.(string)
		}
		t.EnumSimple = &EnumSimple{Variants: variants}

	case "enum_complex":
		t.Kind = KindEnumComplex
		rawVariants, _ := def["variants"].([]any)
		variants, err := parseNamedMembersFromList(rawVariants)
		if err != nil {
			return nil, fmt.Errorf("parsing complex enum variants: %w", err)
		}
		t.EnumComplex = &EnumComplex{Variants: variants}

	case "import":
		t.Kind = KindImport
		module, _ := def["module"].(string)
		item, _ := def["item"].(string)
		t.Import = &Import{Module: module, Item: item}

	case "vec", "option":
		itemDef, ok := def["item"]
		if !ok {
			return nil, fmt.Errorf("'%s' definition missing 'item'", rawType)
		}
		itemType, err := parseType(itemDef)
		if err != nil {
			return nil, err
		}
		if rawType == "vec" {
			t.Kind = KindVec
			t.Vec = &Vec{Type: itemType}
		} else {
			t.Kind = KindOption
			t.Option = &Option{Type: itemType}
		}

	default:
		t.Kind = KindRef
		t.Ref = &Ref{Name: rawType}
	}

	return t, nil
}

func parseNamedMembersFromList(l []any) ([]NamedMember, error) {
	members := make([]NamedMember, len(l))
	for i, item := range l {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("list item is not a map")
		}
		name, _ := itemMap["name"].(string)

		// Pass the whole map to parseType. It can handle nested structs, vecs, etc.
		memberType, err := parseType(itemMap)
		if err != nil {
			return nil, fmt.Errorf("parsing member '%s': %w", name, err)
		}
		members[i] = NamedMember{Name: name, Type: memberType}
	}
	return members, nil
}
