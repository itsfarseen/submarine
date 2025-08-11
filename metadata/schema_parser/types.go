package schema_parser

import (
	. "submarine/scale_schema"
)

type AllModules struct {
	ModuleNames []string // for preserving order
	Modules     map[string]Module
}

type Module struct {
	Types     map[string]*Type
	TypeNames []string
}
