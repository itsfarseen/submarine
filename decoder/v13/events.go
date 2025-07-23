package v13

import (
	"fmt"
	. "submarine/decoder/models"
	. "submarine/scale"
	"submarine/scale/v13"
)

// DecodeEvents is the main entry point for decoding the raw bytes from System.Events.
func DecodeEvents(metadata *v13.Metadata, eventBytes []byte) ([]EventRecord, error) {
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
func DecodeEventRecord(metadata *v13.Metadata, r *Reader) (EventRecord, error) {
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
	decodedEvent, err := DecodePalletEvent(metadata, r)
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

// DecodePalletEvent decodes an event.
func DecodePalletEvent(metadata *v13.Metadata, r *Reader) (*DecodedPalletVariant, error) {
	// The payload starts with the pallet index.
	palletIndex, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read pallet index: %w", err)
	}

	// The next byte is the variant (event) index.
	variantIndex, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read variant index: %w", err)
	}

	// --- Find the Pallet Definition ---
	var pallet v13.ModuleMetadata
	foundPallet := false
	for _, p := range metadata.Modules {
		if p.Index == palletIndex {
			pallet = p
			foundPallet = true
			break
		}
	}
	if !foundPallet {
		return nil, fmt.Errorf("pallet with index %d not found", palletIndex)
	}

	if !pallet.Events.HasValue {
		return nil, fmt.Errorf("pallet '%s' has no events defined", pallet.Name)
	}

	if int(variantIndex) >= len(pallet.Events.Value) {
		return nil, fmt.Errorf("event with index %d not found in pallet '%s'", variantIndex, pallet.Name)
	}

	chosenVariant := pallet.Events.Value[variantIndex]

	// --- Decode Arguments ---
	// For v13, we don't have named args in the metadata for events, just a list of type names.
	decodedArgs := make([]DecodedArg, len(chosenVariant.Args))
	for i, argType := range chosenVariant.Args {
		argValue, err := DecodeArgFromString(metadata, r, string(argType))
		if err != nil {
			return nil, fmt.Errorf("failed to decode arg %d for '%s.%s': %w", i, pallet.Name, chosenVariant.Name, err)
		}

		decodedArgs[i] = DecodedArg{
			Name:  fmt.Sprintf("arg%d", i), // No names in v13 event metadata
			Value: argValue,
		}
	}

	return &DecodedPalletVariant{
		PalletName:  string(pallet.Name),
		VariantName: string(chosenVariant.Name),
		Args:        decodedArgs,
	}, nil
}
