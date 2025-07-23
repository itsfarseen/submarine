package scale

import "fmt"

// Types

type Si1Field struct {
	Name     Option[Text]
	Type     SiLookupTypeId
	TypeName Option[Text]
	Docs     []Text
}

type Si1Path []Text

type Si1TypeParameter struct {
	Name Text
	Type Option[SiLookupTypeId]
}

type Si1TypeDef interface {
	isSi1TypeDef()
}

type Si1TypeDefComposite struct {
	Fields []Si1Field
}

func (s Si1TypeDefComposite) isSi1TypeDef() {}

type Si1TypeDefVariant struct {
	Variants []Si1Variant
}

func (s Si1TypeDefVariant) isSi1TypeDef() {}

type Si1TypeDefSequence struct {
	Type SiLookupTypeId
}

func (s Si1TypeDefSequence) isSi1TypeDef() {}

type Si1TypeDefArray struct {
	Len  uint32
	Type SiLookupTypeId
}

func (s Si1TypeDefArray) isSi1TypeDef() {}

type Si1TypeDefTuple struct {
	Fields []SiLookupTypeId
}

func (s Si1TypeDefTuple) isSi1TypeDef() {}

// NOTE: Si0TypeDefPrimitive is a byte enum.
type Si1TypeDefPrimitive byte

func (s Si1TypeDefPrimitive) isSi1TypeDef() {}

type Si1TypeDefCompact struct {
	Type SiLookupTypeId
}

func (s Si1TypeDefCompact) isSi1TypeDef() {}

type Si1TypeDefBitSequence struct {
	BitStoreType SiLookupTypeId
	BitOrderType SiLookupTypeId
}

func (s Si1TypeDefBitSequence) isSi1TypeDef() {}

type HistoricMetaCompat struct {
	Type Text
}

func (s HistoricMetaCompat) isSi1TypeDef() {}

type Si1Type struct {
	Path   Si1Path
	Params []Si1TypeParameter
	Def    Si1TypeDef
	Docs   []Text
}

type Si1Variant struct {
	Name   Text
	Fields []Si1Field
	Index  uint8
	Docs   []Text
}

// Decoders

// DecodeSiLookupTypeId decodes a type ID.
func DecodeSiLookupTypeId(r *Reader) (SiLookupTypeId, error) {
	val, err := DecodeCompact(r)
	if err != nil {
		return 0, err
	}
	return SiLookupTypeId(val.Uint64()), nil
}

func DecodeSi1Field(r *Reader) (Si1Field, error) {
	var result Si1Field
	var err error

	result.Name, err = DecodeOption(r, DecodeText)
	if err != nil {
		return result, err
	}
	result.Type, err = DecodeSiLookupTypeId(r)
	if err != nil {
		return result, err
	}
	result.TypeName, err = DecodeOption(r, DecodeText)
	if err != nil {
		return result, err
	}
	result.Docs, err = DecodeVec(r, DecodeText)
	return result, err
}

func DecodeSi1TypeParameter(r *Reader) (Si1TypeParameter, error) {
	var result Si1TypeParameter
	var err error

	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}
	result.Type, err = DecodeOption(r, DecodeSiLookupTypeId)
	return result, err
}

func DecodeSi1Path(r *Reader) (Si1Path, error) {
	return DecodeVec(r, DecodeText)
}

func DecodeSi1TypeDef(r *Reader) (Si1TypeDef, error) {
	variant, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch variant {
	case 0: // Composite
		fields, err := DecodeVec(r, DecodeSi1Field)
		return Si1TypeDefComposite{Fields: fields}, err
	case 1: // Variant
		// Note: This calls DecodeSi1Variant, which is now updated.
		variants, err := DecodeVec(r, DecodeSi1Variant)
		return Si1TypeDefVariant{Variants: variants}, err
	case 2: // Sequence
		typ, err := DecodeSiLookupTypeId(r)
		return Si1TypeDefSequence{Type: typ}, err
	case 3: // Array
		var result Si1TypeDefArray
		result.Len, err = DecodeU32(r)
		if err != nil {
			return nil, err
		}
		result.Type, err = DecodeSiLookupTypeId(r)
		return result, err
	case 4: // Tuple
		fields, err := DecodeVec(r, DecodeSiLookupTypeId)
		return Si1TypeDefTuple{Fields: fields}, err
	case 5: // Primitive
		b, err := r.ReadByte()
		return Si1TypeDefPrimitive(b), err
	case 6: // Compact
		typ, err := DecodeSiLookupTypeId(r)
		return Si1TypeDefCompact{Type: typ}, err
	case 7: // BitSequence
		var result Si1TypeDefBitSequence
		result.BitStoreType, err = DecodeSiLookupTypeId(r)
		if err != nil {
			return nil, err
		}
		result.BitOrderType, err = DecodeSiLookupTypeId(r)
		return result, err
	case 8: // HistoricMetaCompat
		typ, err := DecodeText(r)
		return HistoricMetaCompat{Type: typ}, err
	default:
		return nil, fmt.Errorf("unknown variant for Si1TypeDef: %d", variant)
	}
}

func DecodeSi1Type(r *Reader) (Si1Type, error) {
	var result Si1Type
	var err error

	result.Path, err = DecodeSi1Path(r)
	if err != nil {
		return result, fmt.Errorf("path: %w", err)

	}
	result.Params, err = DecodeVec(r, DecodeSi1TypeParameter)
	if err != nil {
		return result, fmt.Errorf("params: %w", err)

	}
	result.Def, err = DecodeSi1TypeDef(r)
	if err != nil {
		return result, fmt.Errorf("def: %w", err)

	}
	result.Docs, err = DecodeVec(r, DecodeText)
	if err != nil {
		return result, fmt.Errorf("docs: %w", err)
	}
	return result, nil

}

func DecodeSi1Variant(r *Reader) (Si1Variant, error) {
	var result Si1Variant
	var err error

	result.Name, err = DecodeText(r)
	if err != nil {
		return result, err
	}

	result.Fields, err = DecodeVec(r, DecodeSi1Field)
	if err != nil {
		return result, err
	}
	result.Index, err = DecodeU8(r)
	if err != nil {
		return result, err
	}
	result.Docs, err = DecodeVec(r, DecodeText)
	return result, err
}
