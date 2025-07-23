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

// Si1TypeDefKind enumerates the different kinds of type definitions.
type Si1TypeDefKind byte

const (
	KindSi1TypeDefComposite Si1TypeDefKind = iota
	KindSi1TypeDefVariant
	KindSi1TypeDefSequence
	KindSi1TypeDefArray
	KindSi1TypeDefTuple
	KindSi1TypeDefPrimitive
	KindSi1TypeDefCompact
	KindSi1TypeDefBitSequence
	KindHistoricMetaCompat
)

// Si1TypeDef is a tagged union representing the structure of a type.
type Si1TypeDef struct {
	Kind        Si1TypeDefKind
	Composite   Si1TypeDefComposite
	Variant     Si1TypeDefVariant
	Sequence    Si1TypeDefSequence
	Array       Si1TypeDefArray
	Tuple       Si1TypeDefTuple
	Primitive   Si1TypeDefPrimitive
	Compact     Si1TypeDefCompact
	BitSequence Si1TypeDefBitSequence
	Historic    HistoricMetaCompat
}

type Si1TypeDefComposite struct {
	Fields []Si1Field
}

type Si1TypeDefVariant struct {
	Variants []Si1Variant
}

type Si1TypeDefSequence struct {
	Type SiLookupTypeId
}

type Si1TypeDefArray struct {
	Len  uint32
	Type SiLookupTypeId
}

type Si1TypeDefTuple struct {
	Fields []SiLookupTypeId
}

// NOTE: Si0TypeDefPrimitive is a byte enum.
type Si1TypeDefPrimitive byte

type Si1TypeDefCompact struct {
	Type SiLookupTypeId
}

type Si1TypeDefBitSequence struct {
	BitStoreType SiLookupTypeId
	BitOrderType SiLookupTypeId
}

// Some old variant of Si1TypeDef
type HistoricMetaCompat struct {
	Type Text
}

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
		return Si1TypeDef{}, err
	}

	switch variant {
	case 0: // Composite
		fields, err := DecodeVec(r, DecodeSi1Field)
		if err != nil {
			return Si1TypeDef{}, err
		}
		return Si1TypeDef{Kind: KindSi1TypeDefComposite, Composite: Si1TypeDefComposite{Fields: fields}}, nil
	case 1: // Variant
		variants, err := DecodeVec(r, DecodeSi1Variant)
		if err != nil {
			return Si1TypeDef{}, err
		}
		return Si1TypeDef{Kind: KindSi1TypeDefVariant, Variant: Si1TypeDefVariant{Variants: variants}}, nil
	case 2: // Sequence
		typ, err := DecodeSiLookupTypeId(r)
		if err != nil {
			return Si1TypeDef{}, err
		}
		return Si1TypeDef{Kind: KindSi1TypeDefSequence, Sequence: Si1TypeDefSequence{Type: typ}}, nil
	case 3: // Array
		var result Si1TypeDefArray
		result.Len, err = DecodeU32(r)
		if err != nil {
			return Si1TypeDef{}, err
		}
		result.Type, err = DecodeSiLookupTypeId(r)
		if err != nil {
			return Si1TypeDef{}, err
		}
		return Si1TypeDef{Kind: KindSi1TypeDefArray, Array: result}, nil
	case 4: // Tuple
		fields, err := DecodeVec(r, DecodeSiLookupTypeId)
		if err != nil {
			return Si1TypeDef{}, err
		}
		return Si1TypeDef{Kind: KindSi1TypeDefTuple, Tuple: Si1TypeDefTuple{Fields: fields}}, nil
	case 5: // Primitive
		b, err := r.ReadByte()
		if err != nil {
			return Si1TypeDef{}, err
		}
		return Si1TypeDef{Kind: KindSi1TypeDefPrimitive, Primitive: Si1TypeDefPrimitive(b)}, nil
	case 6: // Compact
		typ, err := DecodeSiLookupTypeId(r)
		if err != nil {
			return Si1TypeDef{}, err
		}
		return Si1TypeDef{Kind: KindSi1TypeDefCompact, Compact: Si1TypeDefCompact{Type: typ}}, nil
	case 7: // BitSequence
		var result Si1TypeDefBitSequence
		result.BitStoreType, err = DecodeSiLookupTypeId(r)
		if err != nil {
			return Si1TypeDef{}, err
		}
		result.BitOrderType, err = DecodeSiLookupTypeId(r)
		if err != nil {
			return Si1TypeDef{}, err
		}
		return Si1TypeDef{Kind: KindSi1TypeDefBitSequence, BitSequence: result}, nil
	case 8: // HistoricMetaCompat
		typ, err := DecodeText(r)
		if err != nil {
			return Si1TypeDef{}, err
		}
		return Si1TypeDef{Kind: KindHistoricMetaCompat, Historic: HistoricMetaCompat{Type: typ}}, nil
	default:
		return Si1TypeDef{}, fmt.Errorf("unknown variant for Si1TypeDef: %d", variant)
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
