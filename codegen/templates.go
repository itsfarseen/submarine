package codegen

const templates = `
{{define "file"}}
package {{.PackageName}}

import (
	{{range $path, $alias := .Imports}}
	"{{$path}}"
	{{end}}
)

{{range $name, $type := .Types}}
	{{if eq $type.Kind "struct"}}
		{{template "struct" (dict "name" $name "type" $type "moduleName" $.ModuleName)}}
	{{else if eq $type.Kind "enum_simple"}}
		{{template "enum_simple" (dict "name" $name "type" $type "moduleName" $.ModuleName)}}
	{{else if eq $type.Kind "enum_complex"}}
		{{template "enum_complex" (dict "name" $name "type" $type "moduleName" $.ModuleName)}}
	{{else if eq $type.Kind "import"}}
		type {{toPascalCase $name}} = {{getGoType $name $type $.ModuleName}}
	{{end}}
{{end}}
{{end}}

{{define "struct"}}
type {{toPascalCase .name}} struct {
	{{range .type.Struct.Fields}}
		{{toPascalCase .Name}} {{getGoType .Name .Type $.moduleName}}
	{{end}}
}
{{end}}

{{define "enum_simple"}}
type {{toPascalCase .name}} int

const (
	{{range $i, $v := .type.EnumSimple.Variants}}
		{{toPascalCase $.name}}{{toPascalCase $v}} {{toPascalCase $.name}} = {{$i}}
	{{end}}
)
{{end}}

{{define "enum_complex"}}
type {{toPascalCase .name}} struct {
	Kind string
	{{range .type.EnumComplex.Variants}}
		{{toPascalCase .Name}} *{{getGoType .Name .Type $.moduleName}}
	{{end}}
}
{{end}}
`