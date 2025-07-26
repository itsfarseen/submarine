package v14

import (
	"fmt"
	. "submarine/decoder/models"
	. "submarine/scale"
	"submarine/scale/gen/scaleInfo"
	"submarine/scale/gen/v14"
)

// DecodeEvents is the main entry point for decoding the raw bytes from System.Events.
func DecodeEvents(metadata *v14.Metadata, eventBytes []byte) ([]EventRecord, error) {
	r := NewReader(eventBytes)

	// The event bytes are a Vec<EventRecord>. First, decode the length.
	numEvents, err := DecodeCompact(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode event vector length: %w", err)
	}

	records := make([]EventRecord, numEvents.Int64())
	for i := int64(0); i < numEvents.Int64(); i++ {
		record, err := DecodeEventRecord(metadata, r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode event record #%d: %w", i, err)
		}
		records[i] = record
	}

	return records, nil
}

// DecodeEventRecord decodes a single EventRecord from the byte stream.
func DecodeEventRecord(metadata *v14.Metadata, r *Reader) (EventRecord, error) {
	var record EventRecord

	// --- 1. Decode the Phase ---
	phaseIndex, err := r.ReadByte()
	if err != nil {
		return record, fmt.Errorf("failed to read phase index: %w", err)
	}

	switch phaseIndex {
	case 0: // ApplyExtrinsic
		extrinsicIndex, err := DecodeU32(r)
		if err != nil {
			return record, fmt.Errorf("failed to decode extrinsic index for phase: %w", err)
		}
		record.Phase = EventPhase{
			IsApplyExtrinsic: true,
			AsApplyExtrinsic: extrinsicIndex,
		}
	case 1: // Finalization
		record.Phase = EventPhase{IsFinalization: true}
	case 2: // Initialization
		record.Phase = EventPhase{IsInitialization: true}
	default:
		// Default to an empty phase for unknown/unhandled phase indices.
		record.Phase = EventPhase{}
	}

	// --- 2. Decode the Event Payload ---
	// This uses a generalized function that can find variants in either .calls or .events.
	decodedEvent, err := DecodePalletVariant(metadata, r, "events")
	if err != nil {
		return record, fmt.Errorf("failed to decode event payload: %w", err)
	}
	record.Event = DecodedEvent{
		PalletName: decodedEvent.PalletName,
		EventName:  decodedEvent.VariantName,
		Args:       decodedEvent.Args,
	}

	// --- 3. Decode and Skip Topics ---
	// Topics are a Vec<Hash> (Vec<[u8; 32]>).
	// We decode the vector but discard the contents for this example.
	numTopics, err := DecodeCompact(r)
	if err != nil {
		return record, fmt.Errorf("failed to decode topics vector length: %w", err)
	}
	for i := int64(0); i < numTopics.Int64(); i++ {
		_, err := r.ReadBytes(32) // A hash is 32 bytes
		if err != nil {
			return record, fmt.Errorf("failed to read topic #%d: %w", i, err)
		}
	}

	return record, nil
}

// DecodePalletVariant is a generalized function to decode a call or an event.
// It takes a `variantType` string ("calls" or "events") to look in the correct metadata field.
func DecodePalletVariant(metadata *v14.Metadata, r *Reader, variantType string) (*DecodedPalletVariant, error) {
	// The payload starts with the pallet index.
	palletIndex, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read pallet index: %w", err)
	}

	// The next byte is the variant (call or event) index.
	variantIndex, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read variant index: %w", err)
	}

	// --- Find the Pallet Definition ---
	var pallet v14.PalletMetadata
	foundPallet := false
	for _, p := range metadata.Pallets {
		if p.Index == palletIndex {
			pallet = p
			foundPallet = true
			break
		}
	}
	if !foundPallet {
		return nil, fmt.Errorf("pallet with index %d not found", palletIndex)
	}

	// --- Find the Variant (Call/Event) Definition ---
	var variantDefTypeID scaleInfo.Si1LookupTypeId
	switch variantType {
	case "calls":
		if pallet.Calls == nil {
			return nil, fmt.Errorf("pallet '%s' has no calls defined", pallet.Name)
		}
		variantDefTypeID = pallet.Calls.Type
	case "events":
		if pallet.Events == nil {
			return nil, fmt.Errorf("pallet '%s' has no events defined", pallet.Name)
		}
		variantDefTypeID = pallet.Events.Type
	default:
		return nil, fmt.Errorf("invalid variant type: %s", variantType)
	}

	variantTypeInfo, ok := findType(metadata, variantDefTypeID)
	if !ok {
		return nil, fmt.Errorf("%s type definition for pallet '%s' not found", variantType, pallet.Name)
	}

	if variantTypeInfo.Def.Kind != scaleInfo.Si1TypeDefKindVariant {
		return nil, fmt.Errorf("expected %s type to be a variant, but got %T", variantType, variantTypeInfo.Def)
	}
	variantTypeDef := variantTypeInfo.Def.Variant

	var chosenVariant scaleInfo.Si1Variant
	foundVariant := false
	for _, v := range variantTypeDef.Variants {
		if v.Index == variantIndex {
			chosenVariant = v
			foundVariant = true
			break
		}
	}
	if !foundVariant {
		return nil, fmt.Errorf("%s with index %d not found in pallet '%s'", variantType, variantIndex, pallet.Name)
	}

	// --- Decode Arguments ---
	decodedArgs := make([]DecodedArg, len(chosenVariant.Fields))
	for i, field := range chosenVariant.Fields {
		argName := "unnamed"
		if field.Name != nil {
			argName = *field.Name
		}

		argValue, err := DecodeArg(metadata, r, field.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to decode arg '%s' for '%s.%s': %w", argName, pallet.Name, chosenVariant.Name, err)
		}

		decodedArgs[i] = DecodedArg{
			Name:  argName,
			Value: argValue,
		}
	}

	return &DecodedPalletVariant{
		PalletName:  pallet.Name,
		VariantName: chosenVariant.Name,
		Args:        decodedArgs,
	}, nil
}
