package polkadot_scale_schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	s "submarine/scale"
)

//go:embed schema.json
var schemaData []byte

func LoadPolkadotSchema(r *Registry) error {
	var schema map[string]map[string]any
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	return LoadFromSchema(r, schema)
}

func LoadFromSchema(r *Registry, schema map[string]map[string]any) error {
	for moduleName, moduleTypes := range schema {
		for typeName, typeDef := range moduleTypes {
			scaleType, err := parseTypeDef(typeDef)
			if err != nil {
				return fmt.Errorf("failed to parse type %s::%s: %w", moduleName, typeName, err)
			}

			entry := TypeEntry{
				Module: moduleName,
				Type:   scaleType,
			}

			r.types[typeName] = append(r.types[typeName], entry)
		}
	}

	return nil
}

func parseTypeDef(typeDef any) (*LazyType, error) {
	switch def := typeDef.(type) {
	case string:
		return NewLazyType(def), nil

	case map[string]any:
		if enumDef, hasEnum := def["_enum"]; hasEnum {
			return parseEnum(enumDef)
		}

		if setDef, hasSet := def["_set"]; hasSet {
			bitLength := 8
			if bitLengthVal, hasBitLength := def["_bitLength"]; hasBitLength {
				if bl, ok := bitLengthVal.(float64); ok {
					bitLength = int(bl)
				}
			}
			return parseSet(setDef, bitLength)
		}

		return parseStruct(def)

	default:
		return nil, fmt.Errorf("unsupported type definition: %T", typeDef)
	}
}

func parseEnum(enumDef any) (*LazyType, error) {
	switch enum := enumDef.(type) {
	case []any:
		variants := make([]string, len(enum))
		for i, variant := range enum {
			if variantStr, ok := variant.(string); ok {
				variants[i] = variantStr
			} else {
				return nil, fmt.Errorf("enum variant must be string, got %T", variant)
			}
		}

		scaleType := s.Type{
			Kind: s.KindEnumSimple,
			EnumSimple: &s.EnumSimple{
				Variants: variants,
			},
		}

		lazyType := &LazyType{parsed: &scaleType}
		return lazyType, nil

	case map[string]any:
		var variants []s.NamedMember
		for variantName, variantType := range enum {
			if variantType == nil {
				variants = append(variants, s.NamedMember{
					Name: variantName,
					Type: nil,
				})
			} else {
				variantLazyType, err := parseTypeDef(variantType)
				if err != nil {
					return nil, fmt.Errorf("failed to parse variant %s: %w", variantName, err)
				}

				variantScaleType, err := variantLazyType.ToScaleType()
				if err != nil {
					return nil, fmt.Errorf("failed to convert variant %s to scale type: %w", variantName, err)
				}

				variants = append(variants, s.NamedMember{
					Name: variantName,
					Type: variantScaleType,
				})
			}
		}

		scaleType := s.Type{
			Kind: s.KindEnumComplex,
			EnumComplex: &s.EnumComplex{
				Variants: variants,
			},
		}

		lazyType := &LazyType{parsed: &scaleType}
		return lazyType, nil

	default:
		return nil, fmt.Errorf("unsupported _enum type: %T", enumDef)
	}
}

func parseSet(setDef any, bitLength int) (*LazyType, error) {
	setMap, ok := setDef.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("_set must be an object, got %T", setDef)
	}

	var flags []s.BitFlag

	for flagName, flagValue := range setMap {
		var bitValue uint64

		switch v := flagValue.(type) {
		case int:
			bitValue = uint64(v)
		case int64:
			bitValue = uint64(v)
		case uint64:
			bitValue = uint64(v)
		case float32:
			bitValue = uint64(v)
		case float64:
			bitValue = uint64(v)
		default:
			return nil, fmt.Errorf("bit flag value must be int, uint64, float32 or float64, got %T", flagValue)
		}

		flags = append(flags, s.BitFlag{
			Name:  flagName,
			Value: bitValue,
		})
	}

	scaleType := s.Type{
		Kind: s.KindBitFlags,
		BitFlags: &s.BitFlags{
			BitLength: bitLength,
			Flags:     flags,
		},
	}

	lazyType := &LazyType{parsed: &scaleType}
	return lazyType, nil
}

func parseStruct(structDef map[string]any) (*LazyType, error) {
	var fields []s.NamedMember

	for fieldName, fieldType := range structDef {
		if fieldName == "_fallback" {
			continue
		}

		fieldLazyType, err := parseTypeDef(fieldType)
		if err != nil {
			return nil, fmt.Errorf("failed to parse field %s: %w", fieldName, err)
		}

		fieldScaleType, err := fieldLazyType.ToScaleType()
		if err != nil {
			return nil, fmt.Errorf("failed to convert field %s to scale type: %w", fieldName, err)
		}

		fields = append(fields, s.NamedMember{
			Name: fieldName,
			Type: fieldScaleType,
		})
	}

	scaleType := s.Type{
		Kind: s.KindStruct,
		Struct: &s.Struct{
			Fields: fields,
		},
	}

	lazyType := &LazyType{parsed: &scaleType}
	return lazyType, nil
}
