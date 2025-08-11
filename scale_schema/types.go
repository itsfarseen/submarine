package scale_schema

type TypeKind string

const (
	KindStruct      TypeKind = "struct"
	KindTuple       TypeKind = "tuple"
	KindEnumSimple  TypeKind = "enum_simple"
	KindEnumComplex TypeKind = "enum_complex"
	KindImport      TypeKind = "import"
	KindVec         TypeKind = "vec"
	KindOption      TypeKind = "option"
	KindArray       TypeKind = "array"
	KindRef         TypeKind = "ref"
)

type Type struct {
	Kind        TypeKind
	Struct      *Struct
	Tuple       *Tuple
	EnumSimple  *EnumSimple
	EnumComplex *EnumComplex
	Import      *Import
	Vec         *Vec
	Option      *Option
	Array       *Array
	Ref         *string
}

type Struct struct {
	Fields []NamedMember
}

type Tuple struct {
	Fields []Type
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

type Array struct {
	Type *Type
	Len  int
}
