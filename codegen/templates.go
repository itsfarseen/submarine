package codegen

const templates = `
{{define "file"}}
package generated

import (
	"fmt"
	"io"
	"submarine/scale"
)

{{range $name, $type := .Types}}
	{{$resolved := getResolvedType $type}}
	{{if eq $resolved.Kind "struct"}}
		{{template "struct" (dict "name" $name "type" $resolved)}}
	{{else if eq $resolved.Kind "enum_simple"}}
		{{template "enum_simple" (dict "name" $name "type" $resolved)}}
	{{else if eq $resolved.Kind "enum_complex"}}
		{{template "enum_complex" (dict "name" $name "type" $resolved)}}
	{{end}}
{{end}}

{{end}}

{{define "struct"}}
type {{getGoType .name .type}} struct {
	{{range .type.Struct.Fields}}
		{{toPascalCase .Name}} {{getGoType .Name .Type}}
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
type {{getGoType .name .type}} struct {
	Kind string
	{{range .type.EnumComplex.Variants}}
		{{toPascalCase .Name}} *{{getGoType .Name .Type}}
	{{end}}
}
{{end}}
`
