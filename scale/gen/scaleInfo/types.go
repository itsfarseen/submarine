package scaleInfo

import (
	"fmt"
	"math/big"
	"submarine/scale"
)

type Si0Path = string

func DecodeSi0Path(reader *scale.Reader) (Si0Path, error) {
	return scale.DecodeText(reader)
}

type Si0TypeDefPrimitive = string

func DecodeSi0TypeDefPrimitive(reader *scale.Reader) (Si0TypeDefPrimitive, error) {
	return scale.DecodeText(reader)
}

type Si1Field struct {
	Name     *string
	Type     big.Int
	TypeName *string
	Docs     []string
}

func DecodeSi1Field(reader *scale.Reader) (Si1Field, error) {
	var t Si1Field
	var err error

	t.Name, err = scale.DecodeOption(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	t.TypeName, err = scale.DecodeOption(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field TypeName: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	return t, nil
}

type Si1LookupTypeId = big.Int

func DecodeSi1LookupTypeId(reader *scale.Reader) (Si1LookupTypeId, error) {
	return scale.DecodeCompact(reader)
}

type Si1Path = string

func DecodeSi1Path(reader *scale.Reader) (Si1Path, error) {
	return DecodeSi0Path(reader)
}

type Si1Type struct {
	Path   string
	Params []Si1TypeParameter
	Def    Si1TypeDef
	Docs   []string
}

func DecodeSi1Type(reader *scale.Reader) (Si1Type, error) {
	var t Si1Type
	var err error

	t.Path, err = DecodeSi1Path(reader)
	if err != nil {
		return t, fmt.Errorf("field Path: %w", err)
	}

	t.Params, err = scale.DecodeVec(reader, func(reader *scale.Reader) (Si1TypeParameter, error) { return DecodeSi1TypeParameter(reader) })
	if err != nil {
		return t, fmt.Errorf("field Params: %w", err)
	}

	t.Def, err = DecodeSi1TypeDef(reader)
	if err != nil {
		return t, fmt.Errorf("field Def: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	return t, nil
}

type Si1TypeDefKind byte

const (
	Si1TypeDefKindComposite          Si1TypeDefKind = 0
	Si1TypeDefKindVariant            Si1TypeDefKind = 1
	Si1TypeDefKindSequence           Si1TypeDefKind = 2
	Si1TypeDefKindArray              Si1TypeDefKind = 3
	Si1TypeDefKindTuple              Si1TypeDefKind = 4
	Si1TypeDefKindPrimitive          Si1TypeDefKind = 5
	Si1TypeDefKindCompact            Si1TypeDefKind = 6
	Si1TypeDefKindBitSequence        Si1TypeDefKind = 7
	Si1TypeDefKindHistoricMetaCompat Si1TypeDefKind = 8
)

type Si1TypeDef struct {
	Kind               Si1TypeDefKind
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

func DecodeSi1TypeDef(reader *scale.Reader) (Si1TypeDef, error) {
	var t Si1TypeDef

	tag, err := reader.ReadByte()
	if err != nil {
		return t, fmt.Errorf("enum tag: %w", err)
	}

	t.Kind = Si1TypeDefKind(tag)
	switch t.Kind {

	case Si1TypeDefKindComposite:
		value, err := DecodeSi1TypeDefComposite(reader)
		if err != nil {
			return t, fmt.Errorf("field Composite: %w", err)
		}
		t.Composite = &value
		return t, nil

	case Si1TypeDefKindVariant:
		value, err := DecodeSi1TypeDefVariant(reader)
		if err != nil {
			return t, fmt.Errorf("field Variant: %w", err)
		}
		t.Variant = &value
		return t, nil

	case Si1TypeDefKindSequence:
		value, err := DecodeSi1TypeDefSequence(reader)
		if err != nil {
			return t, fmt.Errorf("field Sequence: %w", err)
		}
		t.Sequence = &value
		return t, nil

	case Si1TypeDefKindArray:
		value, err := DecodeSi1TypeDefArray(reader)
		if err != nil {
			return t, fmt.Errorf("field Array: %w", err)
		}
		t.Array = &value
		return t, nil

	case Si1TypeDefKindTuple:
		value, err := DecodeSi1TypeDefTuple(reader)
		if err != nil {
			return t, fmt.Errorf("field Tuple: %w", err)
		}
		t.Tuple = &value
		return t, nil

	case Si1TypeDefKindPrimitive:
		value, err := DecodeSi1TypeDefPrimitive(reader)
		if err != nil {
			return t, fmt.Errorf("field Primitive: %w", err)
		}
		t.Primitive = &value
		return t, nil

	case Si1TypeDefKindCompact:
		value, err := DecodeSi1TypeDefCompact(reader)
		if err != nil {
			return t, fmt.Errorf("field Compact: %w", err)
		}
		t.Compact = &value
		return t, nil

	case Si1TypeDefKindBitSequence:
		value, err := DecodeSi1TypeDefBitSequence(reader)
		if err != nil {
			return t, fmt.Errorf("field BitSequence: %w", err)
		}
		t.BitSequence = &value
		return t, nil

	case Si1TypeDefKindHistoricMetaCompat:
		value, err := scale.DecodeText(reader)
		if err != nil {
			return t, fmt.Errorf("field HistoricMetaCompat: %w", err)
		}
		t.HistoricMetaCompat = &value
		return t, nil

	default:
		return t, fmt.Errorf("unknown tag: %d", tag)
	}
}

type Si1TypeDefArray struct {
	Len  uint32
	Type big.Int
}

func DecodeSi1TypeDefArray(reader *scale.Reader) (Si1TypeDefArray, error) {
	var t Si1TypeDefArray
	var err error

	t.Len, err = scale.DecodeU32(reader)
	if err != nil {
		return t, fmt.Errorf("field Len: %w", err)
	}

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	return t, nil
}

type Si1TypeDefBitSequence struct {
	BitStoreType big.Int
	BitOrderType big.Int
}

func DecodeSi1TypeDefBitSequence(reader *scale.Reader) (Si1TypeDefBitSequence, error) {
	var t Si1TypeDefBitSequence
	var err error

	t.BitStoreType, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field BitStoreType: %w", err)
	}

	t.BitOrderType, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field BitOrderType: %w", err)
	}

	return t, nil
}

type Si1TypeDefCompact struct {
	Type big.Int
}

func DecodeSi1TypeDefCompact(reader *scale.Reader) (Si1TypeDefCompact, error) {
	var t Si1TypeDefCompact
	var err error

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	return t, nil
}

type Si1TypeDefComposite struct {
	Fields []Si1Field
}

func DecodeSi1TypeDefComposite(reader *scale.Reader) (Si1TypeDefComposite, error) {
	var t Si1TypeDefComposite
	var err error

	t.Fields, err = scale.DecodeVec(reader, func(reader *scale.Reader) (Si1Field, error) { return DecodeSi1Field(reader) })
	if err != nil {
		return t, fmt.Errorf("field Fields: %w", err)
	}

	return t, nil
}

type Si1TypeDefPrimitive = string

func DecodeSi1TypeDefPrimitive(reader *scale.Reader) (Si1TypeDefPrimitive, error) {
	return DecodeSi0TypeDefPrimitive(reader)
}

type Si1TypeDefSequence struct {
	Type big.Int
}

func DecodeSi1TypeDefSequence(reader *scale.Reader) (Si1TypeDefSequence, error) {
	var t Si1TypeDefSequence
	var err error

	t.Type, err = DecodeSi1LookupTypeId(reader)
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	return t, nil
}

type Si1TypeDefTuple = []big.Int

func DecodeSi1TypeDefTuple(reader *scale.Reader) (Si1TypeDefTuple, error) {
	return scale.DecodeVec(reader, func(reader *scale.Reader) (big.Int, error) { return DecodeSi1LookupTypeId(reader) })
}

type Si1TypeDefVariant struct {
	Variants []Si1Variant
}

func DecodeSi1TypeDefVariant(reader *scale.Reader) (Si1TypeDefVariant, error) {
	var t Si1TypeDefVariant
	var err error

	t.Variants, err = scale.DecodeVec(reader, func(reader *scale.Reader) (Si1Variant, error) { return DecodeSi1Variant(reader) })
	if err != nil {
		return t, fmt.Errorf("field Variants: %w", err)
	}

	return t, nil
}

type Si1TypeParameter struct {
	Name string
	Type *big.Int
}

func DecodeSi1TypeParameter(reader *scale.Reader) (Si1TypeParameter, error) {
	var t Si1TypeParameter
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Type, err = scale.DecodeOption(reader, func(reader *scale.Reader) (big.Int, error) { return DecodeSi1LookupTypeId(reader) })
	if err != nil {
		return t, fmt.Errorf("field Type: %w", err)
	}

	return t, nil
}

type Si1Variant struct {
	Name   string
	Fields []Si1Field
	Index  uint8
	Docs   []string
}

func DecodeSi1Variant(reader *scale.Reader) (Si1Variant, error) {
	var t Si1Variant
	var err error

	t.Name, err = scale.DecodeText(reader)
	if err != nil {
		return t, fmt.Errorf("field Name: %w", err)
	}

	t.Fields, err = scale.DecodeVec(reader, func(reader *scale.Reader) (Si1Field, error) { return DecodeSi1Field(reader) })
	if err != nil {
		return t, fmt.Errorf("field Fields: %w", err)
	}

	t.Index, err = scale.DecodeU8(reader)
	if err != nil {
		return t, fmt.Errorf("field Index: %w", err)
	}

	t.Docs, err = scale.DecodeVec(reader, func(reader *scale.Reader) (string, error) { return scale.DecodeText(reader) })
	if err != nil {
		return t, fmt.Errorf("field Docs: %w", err)
	}

	return t, nil
}
