package scaleInfo

import (
	"math/big"
)

type Si1TypeDefBitSequence struct {
	BitStoreType big.Int
	BitOrderType big.Int
}

type Si1TypeDefCompact struct {
	Type big.Int
}

type Si1Type struct {
	Path   string
	Params []Si1TypeParameter
	Def    Si1TypeDef
	Docs   []string
}
type Si1TypeDefPrimitive = string

type Si1TypeDefSequence struct {
	Type big.Int
}
type Si0Path = string
type Si1LookupTypeId = big.Int

type Si1Variant struct {
	Name   string
	Fields []Si1Field
	Index  uint8
	Docs   []string
}

type Si1Field struct {
	Name     *string
	Type     big.Int
	TypeName *string
	Docs     []string
}

type Si1TypeParameter struct {
	Name string
	Type *big.Int
}

type Si1TypeDef struct {
	Kind               string
	Composite          *Si1TypeDefComposite
	Variant            *Si1TypeDefVariant
	Sequence           *Si1TypeDefSequence
	Array              *Si1TypeDefArray
	Tuple              *[]big.Int
	Primitive          *string
	Compact            *Si1TypeDefCompact
	BitSequence        *Si1TypeDefBitSequence
	HistoricMetaCompat *string
}

type Si1TypeDefVariant struct {
	Variants []Si1Variant
}
type Si0TypeDefPrimitive = string
type Si1Path = string

type Si1TypeDefArray struct {
	Len  uint32
	Type big.Int
}

type Si1TypeDefComposite struct {
	Fields []Si1Field
}
