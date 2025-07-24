package codegen

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
	Kind TypeKind
	*Struct
	*EnumSimple
	*EnumComplex
	*Import
	*Vec
	*Option
	*Ref
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

type Ref struct {
	Name string
}