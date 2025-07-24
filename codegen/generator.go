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
	funcs := template.FuncMap{
		"getGoType": func(name string, t *Type) string {
			return getGoType(name, t, allModules)
		},
		"getResolvedType": func(t *Type) *Type {
			return resolveImport(t, allModules)
		},
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

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "file", module); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Create a new file with the unformatted code for debugging
		debugPath := filepath.Join(outputDir, moduleName+".debug.go")
		os.WriteFile(debugPath, buf.Bytes(), 0644)
		return fmt.Errorf("formatting generated code: %w", err)
	}

	// Write the formatted code to a file
	outputPath := filepath.Join(outputDir, moduleName+".go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("writing generated code to %s: %w", outputPath, err)
	}

	return nil
}

func getGoType(name string, t *Type, allModules *AllModules) string {
	resolvedType := resolveImport(t, allModules)

	switch resolvedType.Kind {
	case KindRef:
		switch resolvedType.Ref.Name {
		case "text":
			return "string"
		case "bytes":
			return "[]byte"
		case "u8":
			return "uint8"
		case "u32":
			return "uint32"
		case "u64":
			return "uint64"
		case "bool":
			return "bool"
		default:
			return toPascalCase(resolvedType.Ref.Name)
		}
	case KindStruct, KindEnumSimple, KindEnumComplex:
		return toPascalCase(name)
	case KindOption:
		return "*" + getGoType(name, resolvedType.Option.Type, allModules)
	case KindVec:
		return "[]" + getGoType(name, resolvedType.Vec.Type, allModules)
	default:
		return "any"
	}
}

func resolveImport(t *Type, allModules *AllModules) *Type {
	if t.Kind == KindImport {
		targetModule, ok := allModules.Modules[t.Import.Module]
		if !ok {
			return t // Should be caught by validation
		}
		targetType, ok := targetModule.Types[t.Import.Item]
		if !ok {
			return t // Should be caught by validation
		}
		return resolveImport(targetType, allModules)
	}
	return t
}

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
