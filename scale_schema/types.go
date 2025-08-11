package scale_schema

type AllModules struct {
	ModuleNames []string // for preserving order
	Modules     map[string]Module
}

type Module struct {
	Types     map[string]*Type
	TypeNames []string
}

type TypeKind string

const (
	KindStruct      TypeKind = "struct"
	KindEnumSimple  TypeKind = "enum_simple"
	KindEnumComplex TypeKind = "enum_complex"
	KindImport      TypeKind = "import"
	KindVec         TypeKind = "vec"
	KindOption      TypeKind = "option"
	KindRef         TypeKind = "ref"
)

type Type struct {
	Kind        TypeKind
	Struct      *Struct
	EnumSimple  *EnumSimple
	EnumComplex *EnumComplex
	Import      *Import
	Vec         *Vec
	Option      *Option
	Ref         *string
}

type Struct struct {
	Fields []NamedMember
}

type EnumSimple struct {
	Variants []string
}

type EnumComplex struct {
	Variants []NamedMember
}

type NamedMember struct {
	Name string
	Type *Type
}

type Import struct {
	Module string
	Item   string
}

type Vec struct {
	Type *Type
}

type Option struct {
	Type *Type
}
