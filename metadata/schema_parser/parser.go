package schema_parser

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	. "submarine/errorspan"
	. "submarine/scale_schema"
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
