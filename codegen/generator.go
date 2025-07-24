package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	. "submarine/codegen/parser"
	"text/template"
)

var rootModulePath string = "submarine/generated"

var primitives = map[string]string{
	"text":    "string",
	"bytes":   "[]byte",
	"u8":      "uint8",
	"u32":     "uint32",
	"u64":     "uint64",
	"bool":    "bool",
	"type":    "string", // 'type' is used as a generic type placeholder in some definitions
	"compact": "big.Int",
}

type Foo = int

func Generate(allModules *AllModules, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	templatesParsed, err := template.New("templates").Parse(templates)
	if err != nil {
		return fmt.Errorf("templates: %w", err)
	}

	codegen := Codegen{
		ModuleNames: allModules.ModuleNames,
		Modules:     allModules.Modules,
		Template:    templatesParsed,
		Generated:   make(map[string]*ModuleCodegen),
	}

	codegen.Generate()

	for _, moduleName := range codegen.ModuleNames {
		module := codegen.Generated[moduleName]
		path := module.Path
		path = filepath.Join(outputDir, path)

		fmt.Printf("path: %s\n", path)

		dirName := filepath.Dir(path)
		if err := os.MkdirAll(dirName, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
		fmt.Printf("dir: %s\n", dirName)

		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("open file %s: %w", path, err)
		}
		defer file.Close()

		file.Write(module.getOutput())
	}

	return nil
}

type Codegen struct {
	ModuleNames []string // for preserving order
	Modules     map[string]Module
	Template    *template.Template
	Generated   map[string]*ModuleCodegen
}

func (c *Codegen) Generate() {
	for _, moduleName := range c.ModuleNames {
		fmt.Printf("Generating %s\n", moduleName)
		err := c.generateModule(moduleName)
		if err != nil {
			slog.Warn("failed to codegen module %s: %w", moduleName, err)
			continue
		}
	}
}

func (c *Codegen) renderTemplate(buffer *bytes.Buffer, templateName string, templateData any) error {
	err := c.Template.ExecuteTemplate(buffer, templateName, templateData)
	if err != nil {
		return fmt.Errorf("executing template %s: %w", templateName, err)
	}
	return nil
}

type ModuleCodegen struct {
	Path    string
	Imports []string
	Header  bytes.Buffer
	Body    bytes.Buffer
}

func (m *ModuleCodegen) appendImport(line string) {
	if !slices.Contains(m.Imports, line) {
		m.Imports = append(m.Imports, line)
	}

}

func (m *ModuleCodegen) getOutput() []byte {
	var buf bytes.Buffer
	buf.Write(m.Header.Bytes())
	buf.Write(m.Body.Bytes())
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		slog.Warn("formatting error", err)
		return formatted
	}
	return buf.Bytes()
}

func (c *Codegen) generateModule(moduleName string) error {
	var moduleCodegen ModuleCodegen
	c.Generated[moduleName] = &moduleCodegen

	moduleCodegen.Path = filepath.Join(moduleName, "types.go")

	module := c.Modules[moduleName]

	for typeName := range module.Types {
		err := c.generateType(moduleName, typeName)
		if err != nil {
			slog.Warn("generating type %s/%s: %w", moduleName, typeName, err)
		}
	}

	templateData := FileHeaderTemplate{
		PackageName: moduleName,
		Imports:     moduleCodegen.Imports,
	}
	if err := c.renderTemplate(&moduleCodegen.Header, "file", templateData); err != nil {
		slog.Warn("generating header %s: %w", moduleName, err)
	}

	return nil
}

func (c *Codegen) generateType(moduleName string, typeName string) error {
	fmt.Printf("  %s\n", typeName)

	moduleCodegen := c.Generated[moduleName]

	var templateName string
	var templateData any

	module := c.Modules[moduleName]
	type_ := module.Types[typeName]

	switch type_.Kind {
	case KindRef:
		innerType, err := c.getGoType(moduleName, "UNKNOWN_", type_)
		if err != nil {
			return err
		}
		moduleCodegen.Body.WriteString(fmt.Sprintf("type %s = %s\n", typeName, innerType))
		return nil
	case KindStruct:
		struct_ := type_.Struct
		fields := make([]FieldOrVariant, len(struct_.Fields))
		for i, field := range struct_.Fields {
			fieldName := toPascalCase(field.Name)
			log.Print(field)
			fieldType, err := c.getGoType(moduleName, fieldName, field.Type)
			if err != nil {
				return fmt.Errorf("struct field %s: %w", fieldName, err)
			}
			fields[i] = FieldOrVariant{
				Name: fieldName,
				Type: fieldType,
			}
		}
		templateName = "struct"
		templateData = StructTemplate{Name: typeName, Fields: fields}
	case KindEnumSimple:
		enumSimple := type_.EnumSimple
		templateName = "enum_simple"
		templateData = EnumSimpleTemplate{Name: typeName, Variants: enumSimple.Variants}
	case KindEnumComplex:
		enumComplex := type_.EnumComplex
		variants := make([]FieldOrVariant, len(enumComplex.Variants))
		for i, variant := range enumComplex.Variants {
			variantName := toPascalCase(variant.Name)
			variantType, err := c.getGoType(moduleName, variantName, variant.Type)
			if err != nil {
				return fmt.Errorf("enum variant %s: %w", variantName, err)
			}
			variants[i] = FieldOrVariant{
				Name: variantName,
				Type: variantType,
			}
		}
		templateName = "enum_complex"
		templateData = EnumComplexTemplate{Name: typeName, Variants: variants}
	case KindImport:
		return nil // handled by imports
	default:
		return fmt.Errorf("unknown type kind: %s", type_.Kind)
	}

	if err := c.renderTemplate(&moduleCodegen.Body, templateName, templateData); err != nil {
		return fmt.Errorf("executing template %s: %w", templateName, err)
	}

	return nil
}

type ResolvedInfo struct {
	Type       *Type
	ModuleName string
	TypeName   string
}

func (c *Codegen) resolveImport(importType *Import) ResolvedInfo {
	targetModule := c.Modules[importType.Module]
	targetType := targetModule.Types[importType.Item]
	if targetType.Kind != KindImport {
		return ResolvedInfo{Type: targetType, ModuleName: importType.Module, TypeName: importType.Item}
	}

	return c.resolveImport(targetType.Import)
}

func (c *Codegen) getGoType(moduleName string, typeName string, type_ *Type) (string, error) {
	moduleCodegen := c.Generated[moduleName]

	module := c.Modules[moduleName]

	var goTypeName string
	switch type_.Kind {
	case KindRef:
		switch type_.Ref.Name {
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
		case "compact":
			moduleCodegen.appendImport("math/big")
			goTypeName = "big.Int"
		default:
			refType_ := module.Types[type_.Ref.Name]
			if refType_ == nil {
				return "MISSING_" + type_.Ref.Name, nil
			}
			return c.getGoType(moduleName, type_.Ref.Name, refType_)
		}
	case KindStruct, KindEnumComplex, KindEnumSimple:
		return typeName, nil
	case KindOption:
		option := type_.Option
		itemGoType, err := c.getGoType(moduleName, "UNKNOWN_", option.Type)
		if err != nil {
			return "", fmt.Errorf("option item: %w", err)
		}
		goTypeName = "*" + itemGoType
	case KindVec:
		vec := type_.Vec
		itemGoType, err := c.getGoType(moduleName, "UNKNOWN_", vec.Type)
		if err != nil {
			return "", fmt.Errorf("vec item: %w", err)
		}
		goTypeName = "[]" + itemGoType
	case KindImport:
		importType := type_.Import
		resolved := c.resolveImport(importType)
		importLine := fmt.Sprintf("%s/%s", rootModulePath, resolved.ModuleName)
		moduleCodegen.appendImport(importLine)
		innerGoType, err := c.getGoType(resolved.ModuleName, resolved.TypeName, resolved.Type)
		if err != nil {
			return "", err
		}
		goTypeName = fmt.Sprintf("%s.%s", resolved.ModuleName, innerGoType)
	default:
		fmt.Printf("Kind ERR %s\n", type_.Kind)
		panic("unreachable: we should exhaustively handle all kinds")
	}
	return goTypeName, nil
}
