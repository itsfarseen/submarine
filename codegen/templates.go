package codegen

type FileHeaderTemplate struct {
	PackageName string
	Imports     []string
}

type StructTemplate struct {
	Name   string
	Fields []FieldOrVariant
}

type EnumSimpleTemplate struct {
	Name     string
	Variants []string
}

type EnumComplexTemplate struct {
	Name     string
	Variants []FieldOrVariant
}

type FieldOrVariant = struct {
	Name       string
	Type       string
	DecodeFunc string
}

const templates = `
{{define "file"}}
package {{.PackageName}}

import (
	{{range $import := .Imports}}"{{$import}}"
	{{end}}
)
{{end}}

{{define "struct"}}
type {{.Name}} struct {
	{{range .Fields}}{{.Name}} {{.Type}}
	{{end}}
}

func Decode{{.Name}}(reader *scale.Reader) ({{.Name}}, error) {
	var t {{.Name}}
	var err error
	{{range .Fields}}
	t.{{.Name}}, err = {{.DecodeFunc}}
	if err != nil {
		return t, fmt.Errorf("field {{.Name}}: %w", err)
	}
	{{end}}
	return t, nil
}
{{end}}

{{define "enum_simple"}}
type {{.Name}} int

const (
	{{range $i, $v := .Variants}}{{$.Name}}{{$v}} {{$.Name}} = {{$i}}
	{{end}}
)

func Decode{{.Name}}(reader *scale.Reader) ({{.Name}}, error) {

	tag, err := reader.ReadByte()
	if err != nil {
		var t {{.Name}}
		return t, fmt.Errorf("enum tag: %w", err)
	}

	switch tag {
	{{range $i, $v := .Variants}}
	case {{$i}}:
		return {{$.Name}}{{$v}}, nil
	{{end}}
	default:
		var t {{.Name}}
		return t, fmt.Errorf("unknown tag: %d", tag)
	}
}
{{end}}

{{define "enum_complex"}}
type {{.Name}}Kind byte

const (
	{{range $i, $v := .Variants}} {{$.Name}}Kind{{$v.Name}} {{$.Name}}Kind = {{$i}}
	{{end}}
)

type {{.Name}} struct {
	Kind {{.Name}}Kind
	{{range .Variants}}{{.Name}} *{{.Type}}
	{{end}}
}

func Decode{{.Name}}(reader *scale.Reader) ({{.Name}}, error) {
	var t {{.Name}}

	tag, err := reader.ReadByte()
	if err != nil {
		return t, fmt.Errorf("enum tag: %w", err)
	}

	t.Kind = {{.Name}}Kind(tag)
	switch t.Kind {
	{{range .Variants}}
	case {{$.Name}}Kind{{.Name}}:
		value, err := {{.DecodeFunc}}
		if err != nil {
			return t, fmt.Errorf("field {{.Name}}: %w", err)
		}
		t.{{.Name}} = &value
		return t, nil
	{{end}}
	default:
		return t, fmt.Errorf("unknown tag: %d", tag)
	}
}
{{end}}
`
