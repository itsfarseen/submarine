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
	Name string
	Type string
}

const templates = `
{{define "file"}}
package {{.PackageName}}

import (
	{{range $import := .Imports}}
	"{{$import}}"
	{{end}}
)
{{end}}

{{define "struct"}}
type {{.Name}} struct {
	{{range .Fields}}
		{{.Name}} {{.Type}}
	{{end}}
}
{{end}}

{{define "enum_simple"}}
type {{.Name}} int

const (
	{{range $i, $v := .Variants}}
		{{$.Name}}{{$v}} {{$.Name}} = {{$i}}
	{{end}}
)
{{end}}

{{define "enum_complex"}}
type {{.Name}} struct {
	Kind string
	{{range .Variants}}
		{{.Name}} *{{.Type}}
	{{end}}
}
{{end}}
`

