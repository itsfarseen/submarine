package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var rootModulePath string = "submarine"

func Generate(allModules *AllModules, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for moduleName, module := range allModules.Modules {
		if err := generateModule(moduleName, &module, allModules, outputDir); err != nil {
			return fmt.Errorf("generating module %s: %w", moduleName, err)
		}
	}

	return nil
}

func generateModule(moduleName string, module *Module, allModules *AllModules, outputDir string) error {
	moduleDir := filepath.Join(outputDir, moduleName)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("creating module directory %s: %w", moduleDir, err)
	}

	imports := make(map[string]string)
	imports[fmt.Sprintf("%s/scale", rootModulePath)] = "scale"

	funcs := template.FuncMap{
		"getGoType":    getGoType,
		"resolve":      resolve,
		"toPascalCase": toPascalCase,
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}

	tmpl, err := template.New("").Funcs(funcs).Parse(templates)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	// First pass to populate imports
	for name, t := range module.Types {
		getGoType(name, t, moduleName)
	}

	var buf bytes.Buffer
	templateData := map[string]interface{}{
		"PackageName": moduleName,
		"ModuleName":  moduleName,
		"Types":       module.Types,
		"Imports":     imports,
	}
	if err := tmpl.ExecuteTemplate(&buf, "file", templateData); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		debugPath := filepath.Join(moduleDir, "types.debug.go")
		os.WriteFile(debugPath, buf.Bytes(), 0644)
		return fmt.Errorf("formatting generated code for module %s: %w", moduleName, err)
	}

	outPath := filepath.Join(moduleDir, "types.go")
	if err := os.WriteFile(outPath, formatted, 0644); err != nil {
		return fmt.Errorf("writing generated code to %s: %w", outPath, err)
	}

	return nil
}

type ResolvedInfo struct {
	Type       *Type
	ModuleName string
}

func resolveImport(t *Type, moduleName string, modules map[string]Module) ResolvedInfo {
	if t.Kind == KindImport {
		targetModuleName := t.Import.Module
		targetTypeName := t.Import.Item
		targetModule := modules[targetModuleName]
		targetType := targetModule.Types[targetTypeName]
		return resolveImport(targetType, targetModuleName, modules)
	}
	return ResolvedInfo{
		Type:       t,
		ModuleName: moduleName,
	}
}

func getGoType(name string, t *Type, moduleName string, modules map[string]Module, imports []string) string {
	info := resolveImport(t, moduleName, modules)
	finalType, finalModule := info.Type, info.ModuleName
	if finalModule != moduleName {
		imports = append(imports, fmt.Sprintf("%s/generated/%s", rootModulePath, finalModule))
		return fmt.Sprintf("%s.%s", finalModule, finalType.Name)
	}

	var goTypeName string
	switch finalType.Kind {
	case KindRef:
		switch finalType.Ref.Name {
		case "text":
			goTypeName = "string"
		case "bytes":
			goTypeName = "[]byte"
		case "u8":
			goTypeName = "uint8"
		case "u32":
			goTypeName = "uint32"
		case "u64":
			goTypeName = "uint64"
		case "bool":
			goTypeName = "bool"
		default:
			goTypeName = toPascalCase(finalType.Ref.Name)
		}
	case KindStruct, KindEnumSimple, KindEnumComplex:
		goTypeName = toPascalCase(name)
	case KindOption:
		goTypeName = "*" + getGoType(name, finalType.Option.Type, moduleName, modules, imports)
	case KindVec:
		goTypeName = "[]" + getGoType(name, finalType.Vec.Type, moduleName, modules, imports)
	case KindImport:
		panic("unreachable: should be caught by resolveImport")
	default:
		panic("unreachable: we should exhaustively handle all kinds")
	}

	return goTypeName
}

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
