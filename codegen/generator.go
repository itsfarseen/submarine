package codegen

import (
	"bytes"
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	. "submarine/scale_schema"
	"text/template"
)

func Generate(allModules *AllModules, rootModulePath string, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	templatesParsed, err := template.New("templates").Parse(templates)
	if err != nil {
		return fmt.Errorf("templates: %w", err)
	}

	codegen := Codegen{
		RootModulePath: rootModulePath,
		ModuleNames:    allModules.ModuleNames,
		Modules:        allModules.Modules,
		Template:       templatesParsed,
		Generated:      make(map[string]*ModuleCodegen),
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
	RootModulePath string
	ModuleNames    []string // for preserving order
	Modules        map[string]Module
	Template       *template.Template
	Generated      map[string]*ModuleCodegen
}

func (c *Codegen) Generate() {
	for _, moduleName := range c.ModuleNames {
		fmt.Printf("Generating %s\n", moduleName)
		err := c.generateModule(moduleName)
		if err != nil {
			slog.Warn("failed to codegen module", "module", moduleName, "error", err)
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
		slog.Warn("formatting error", "error", err)
		return buf.Bytes()
	}
	return formatted
}

func (c *Codegen) generateModule(moduleName string) error {
	var moduleCodegen ModuleCodegen
	c.Generated[moduleName] = &moduleCodegen

	moduleCodegen.Path = filepath.Join(moduleName, "types.go")
	moduleCodegen.appendImport("fmt")
	moduleCodegen.appendImport("submarine/scale")

	module := c.Modules[moduleName]

	for _, typeName := range module.TypeNames {
		err := c.generateType(moduleName, typeName)
		if err != nil {
			slog.Warn("generating type", "module", moduleName, "type", typeName, "error", err)
		}
	}

	templateData := FileHeaderTemplate{
		PackageName: moduleName,
		Imports:     moduleCodegen.Imports,
	}
	if err := c.renderTemplate(&moduleCodegen.Header, "file", templateData); err != nil {
		slog.Warn("generating header", "module", moduleName, "error", err)
	}

	return nil
}

func (c *Codegen) generateType(moduleName string, typeName string) error {
	var err error
	fmt.Printf("  %s\n", typeName)

	moduleCodegen := c.Generated[moduleName]

	var templateName string
	var templateData any

	module := c.Modules[moduleName]
	type_ := module.Types[typeName]

	switch type_.Kind {
	case KindRef:
		innerTypeName, err := c.getGoTypeForType(moduleName, type_)
		if err != nil {
			return err
		}
		moduleCodegen.Body.WriteString(fmt.Sprintf("type %s = %s\n", typeName, innerTypeName))
		const funcTemplate string = `
		func Decode%s(reader *scale.Reader) (%s, error) {
			return %s
		}
		`
		innerDecodeFunc, err := c.getDecodeFuncForTypeName(moduleName, *type_.Ref)
		if err != nil {
			return err
		}
		moduleCodegen.Body.WriteString(fmt.Sprintf(funcTemplate, typeName, typeName, innerDecodeFunc))
		return nil
	case KindStruct:
		struct_ := type_.Struct
		fields := make([]FieldOrVariant, len(struct_.Fields))
		for i, field := range struct_.Fields {
			fieldName := toPascalCase(field.Name)
			fieldType, err := c.getGoTypeForType(moduleName, field.Type)
			if err != nil {
				return fmt.Errorf("struct field %s: %w", fieldName, err)
			}
			decodeFunc, err := c.getDecodeFuncForType(moduleName, field.Type)
			if err != nil {
				return fmt.Errorf("struct field %s: %w", fieldName, err)
			}
			fields[i] = FieldOrVariant{
				Name:       fieldName,
				Type:       fieldType,
				DecodeFunc: decodeFunc,
			}
		}
		if err := c.renderTemplate(&moduleCodegen.Body, "struct", StructTemplate{Name: typeName, Fields: fields}); err != nil {
			return fmt.Errorf("executing template struct: %w", err)
		}
		return nil
	case KindEnumSimple:
		enumSimple := type_.EnumSimple
		templateName = "enum_simple"
		templateData = EnumSimpleTemplate{Name: typeName, Variants: enumSimple.Variants}
	case KindEnumComplex:
		enumComplex := type_.EnumComplex
		variants := make([]FieldOrVariant, len(enumComplex.Variants))
		for i, variant := range enumComplex.Variants {
			variantName := toPascalCase(variant.Name)
			var variantType, decodeFunc string
			if variant.Type != nil {
				variantType, err = c.getGoTypeForType(moduleName, variant.Type)
				if err != nil {
					return fmt.Errorf("enum variant %s: %w", variantName, err)
				}
				decodeFunc, err = c.getDecodeFuncForType(moduleName, variant.Type)
				if err != nil {
					return fmt.Errorf("struct field %s: %w", variantName, err)
				}
			}
			variants[i] = FieldOrVariant{
				Name:       variantName,
				Type:       variantType,
				DecodeFunc: decodeFunc,
			}
		}
		templateName = "enum_complex"
		templateData = EnumComplexTemplate{Name: typeName, Variants: variants}
	case KindVec, KindOption:
		innerType, err := c.getGoTypeForType(moduleName, type_)
		if err != nil {
			return err
		}
		moduleCodegen.Body.WriteString(fmt.Sprintf("type %s = %s\n", typeName, innerType))
		decodeFunc, err := c.getDecodeFuncForType(moduleName, type_)
		if err != nil {
			return err
		}
		const funcTemplate string = `
		func Decode%s(reader *scale.Reader) (%s, error) {
			return %s
		}
		`
		moduleCodegen.Body.WriteString(fmt.Sprintf(funcTemplate, typeName, typeName, decodeFunc))
		return nil
	case KindImport:
		// Don't generate any code for imports.
		// Instead resolve it to the source type at use sites.
		return nil
	default:
		return fmt.Errorf("unknown type kind: %s", type_.Kind)
	}

	if err := c.renderTemplate(&moduleCodegen.Body, templateName, templateData); err != nil {
		return fmt.Errorf("executing template %s: %w", templateName, err)
	}

	return nil
}

type ResolvedInfo struct {
	ModuleName string
	TypeName   string
	Type       *Type
}

func (c *Codegen) resolveImport(moduleName, typeName string) (ResolvedInfo, error) {
	module := c.Modules[moduleName]
	type_ := module.Types[typeName]
	if type_.Kind == KindImport {
		return c.resolveImport(type_.Import.Module, type_.Import.Item)
	}
	return ResolvedInfo{
		ModuleName: moduleName,
		TypeName:   typeName,
		Type:       type_,
	}, nil
}

func (c *Codegen) getDecodeFuncForTypeName(moduleName string, typeName string) (string, error) {
	moduleCodegen := c.Generated[moduleName]

	// handle primitives
	switch typeName {
	case "text", "type":
		return "scale.DecodeText(reader)", nil
	case "bytes":
		return "scale.DecodeBytes(reader)", nil
	case "u8":
		return "scale.DecodeU8(reader)", nil
	case "u32":
		return "scale.DecodeU32(reader)", nil
	case "u64":
		return "scale.DecodeU64(reader)", nil
	case "bool":
		return "scale.DecodeBool(reader)", nil
	case "compact":
		return "scale.DecodeCompact(reader)", nil
	}

	module := c.Modules[moduleName]
	type_, ok := module.Types[typeName]
	if !ok {
		return "", fmt.Errorf("Type %s not found in module %s", typeName, moduleName)
	}
	if type_.Kind == KindImport {
		resolvedType, err := c.resolveImport(type_.Import.Module, type_.Import.Item)
		if err != nil {
			return "", err
		}
		importLine := fmt.Sprintf("%s/%s", c.RootModulePath, resolvedType.ModuleName)
		moduleCodegen.appendImport(importLine)
		return fmt.Sprintf("%s.Decode%s(reader)", resolvedType.ModuleName, resolvedType.TypeName), nil
	}
	return fmt.Sprintf("Decode%s(reader)", typeName), nil
}

func (c *Codegen) getDecodeFuncForType(moduleName string, type_ *Type) (string, error) {
	switch type_.Kind {
	case KindRef:
		return c.getDecodeFuncForTypeName(moduleName, *type_.Ref)
	case KindOption:
		option := type_.Option
		itemTypeName, err := c.getGoTypeForType(moduleName, option.Type)
		if err != nil {
			return "", fmt.Errorf("option item name: %w", err)
		}
		itemDecodeFunc, err := c.getDecodeFuncForType(moduleName, option.Type)
		if err != nil {
			return "", fmt.Errorf("option item type: %w", err)
		}
		return fmt.Sprintf("scale.DecodeOption(reader, func(reader *scale.Reader) (%s, error) { return %s })", itemTypeName, itemDecodeFunc), nil
	case KindVec:
		vec := type_.Vec
		itemTypeName, err := c.getGoTypeForType(moduleName, vec.Type)
		if err != nil {
			return "", fmt.Errorf("vec item name: %w", err)
		}
		itemDecodeFunc, err := c.getDecodeFuncForType(moduleName, vec.Type)
		if err != nil {
			return "", fmt.Errorf("vec item: %w", err)
		}
		return fmt.Sprintf("scale.DecodeVec(reader, func(reader *scale.Reader) (%s, error) { return %s })", itemTypeName, itemDecodeFunc), nil
	default:
		return "", fmt.Errorf("unknown type kind: %s", type_.Kind)
	}
}

func (c *Codegen) getGoTypeForTypename(moduleName string, typeName string) (string, error) {
	moduleCodegen := c.Generated[moduleName]

	module := c.Modules[moduleName]

	// handle primitives
	switch typeName {
	case "text", "type":
		return "string", nil
	case "bytes":
		return "[]byte", nil
	case "u8":
		return "uint8", nil
	case "u32":
		return "uint32", nil
	case "u64":
		return "uint64", nil
	case "bool":
		return "bool", nil
	case "compact":
		moduleCodegen.appendImport("math/big")
		return "big.Int", nil
	default:
		type_, ok := module.Types[typeName]
		if !ok {
			return "", fmt.Errorf("Type %s not found in module %s", typeName, moduleName)
		}
		if type_.Kind == KindImport {
			resolvedType, err := c.resolveImport(type_.Import.Module, type_.Import.Item)
			if err != nil {
				return "", err
			}
			importLine := fmt.Sprintf("%s/%s", c.RootModulePath, resolvedType.ModuleName)
			moduleCodegen.appendImport(importLine)
			return fmt.Sprintf("%s.%s", resolvedType.ModuleName, resolvedType.TypeName), nil
		}
		return typeName, nil
	}
}

func (c *Codegen) getGoTypeForType(moduleName string, type_ *Type) (string, error) {
	moduleCodegen := c.Generated[moduleName]

	switch type_.Kind {
	case KindRef:
		return c.getGoTypeForTypename(moduleName, *type_.Ref)
	case KindStruct, KindEnumComplex, KindEnumSimple:
		return "", fmt.Errorf("Illegal complex nested type")
	case KindOption:
		option := type_.Option
		itemGoType, err := c.getGoTypeForType(moduleName, option.Type)
		if err != nil {
			return "", fmt.Errorf("option: %w", err)
		}
		return "*" + itemGoType, nil
	case KindVec:
		vec := type_.Vec
		itemGoType, err := c.getGoTypeForType(moduleName, vec.Type)
		if err != nil {
			return "", fmt.Errorf("vec: %w", err)
		}
		return "[]" + itemGoType, nil
	case KindImport:
		resolvedType, err := c.resolveImport(type_.Import.Module, type_.Import.Item)
		if err != nil {
			return "", err
		}
		importLine := fmt.Sprintf("%s/%s", c.RootModulePath, resolvedType.ModuleName)
		moduleCodegen.appendImport(importLine)
		return fmt.Sprintf("%s.%s", resolvedType.ModuleName, resolvedType.TypeName), nil
	default:
		panic("unreachable: we should exhaustively handle all kinds")
	}
}
