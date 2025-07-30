package legacy

import (
	"fmt"
	"strconv"
	"strings"
	"submarine/scale"
	"submarine/scale/base"
)

type Event struct {
	Phase      EventPhase
	PalletName string
	EventName  string
	Args       []Arg
}

type Extrinsic struct {
	Address    *base.Address
	Signature  *base.Signature
	PalletName string
	EventName  string
	Args       []Arg
}

type Arg struct {
	Name  string
	Value any
}

type EventPhaseKind int

const (
	EventPhaseApplyExtrinsic EventPhaseKind = iota
	EventPhaseInitialization
	EventPhaseFinalization
)

type EventPhase struct {
	Kind           EventPhaseKind
	ExtrinsicIndex int
}

func DecodeEvents(metadata *Metadata, eventsBytes []byte) ([]Event, error) {
	r := scale.NewReader(eventsBytes)

	events, err := scale.DecodeVec(r, func(r *scale.Reader) (Event, error) {
		return DecodeEvent(metadata, r)
	})
	if err != nil {
		return nil, err
	}

	return events, nil
}

func DecodeEvent(metadata *Metadata, r *scale.Reader) (Event, error) {
	var event Event

	phaseIndex, err := r.ReadByte()
	if err != nil {
		return event, fmt.Errorf("phase index: %w", err)
	}

	switch phaseIndex {
	case 0: // ApplyExtrinsic
		extrinsicIndex, err := scale.DecodeU32(r)
		if err != nil {
			return event, fmt.Errorf("extrinsic index: %w", err)
		}
		event.Phase = EventPhase{
			Kind:           EventPhaseApplyExtrinsic,
			ExtrinsicIndex: int(extrinsicIndex),
		}
	case 1: // Finalization
		event.Phase = EventPhase{Kind: EventPhaseFinalization}
	case 2: // Initialization
		event.Phase = EventPhase{Kind: EventPhaseInitialization}
	default:
		return event, fmt.Errorf("phase: unknown: %d", phaseIndex)
	}

	moduleIndex, err := r.ReadByte()
	if err != nil {
		return event, fmt.Errorf("module index: %w", err)
	}

	eventIndex, err := r.ReadByte()
	if err != nil {
		return event, fmt.Errorf("extrinsic index: %w", err)
	}

	moduleMetadata, err := metadata.GetModuleForEvent(int(moduleIndex))
	if err != nil {
		return event, fmt.Errorf("module: %w", err)
	}

	if int(eventIndex) > len(moduleMetadata.Events) {
		return event, fmt.Errorf("event %d out of bounds", eventIndex)
	}

	eventMetadata := moduleMetadata.Events[eventIndex]
	for i, argType := range eventMetadata.Args {
		var arg Arg
		arg.Name = fmt.Sprintf("arg%d", i)
		value, err := DecodeArgFromTypename(r, argType)
		if err != nil {
			return event, fmt.Errorf("arg %d: %w", i, err)
		}
		arg.Value = value
		event.Args = append(event.Args, arg)
	}

	return event, nil
}

func DecodeExtrinsic(metadata *Metadata, extrinsicBytes []byte) (Extrinsic, error) {
	var extrinsic Extrinsic

	r := scale.NewReader(extrinsicBytes)
	if metadata.Version >= 11 {
		// Skip the compact-encoded length of the extrinsic
		_, err := scale.DecodeCompact(r)
		if err != nil {
			return extrinsic, fmt.Errorf("v11+ extrinsic length prefix: %w", err)
		}
	}

	// The next byte describes the transaction format.
	txFormat, err := r.ReadByte()
	if err != nil {
		return extrinsic, fmt.Errorf("failed to read transaction format byte: %w", err)
	}

	isSigned := (txFormat & 0b10000000) != 0

	if isSigned {
		// 1. Decode the sender's Address.
		address, err := base.DecodeAddress(r)
		if err != nil {
			return extrinsic, fmt.Errorf("failed to decode sender address: %w", err)
		}
		extrinsic.Address = &address

		// 2. Decode the Signature.
		signature, err := base.DecodeSignature(r)
		if err != nil {
			return extrinsic, fmt.Errorf("failed to decode signature: %w", err)
		}
		extrinsic.Signature = &signature

		// // 3. Decode the signed extensions.
		// for _, extension := range metadata.Extrinsic.SignedExtensions {
		// 	_, err := DecodeArgFromString(metadata, r, string(extension.Type))
		// 	if err != nil {
		// 		return nil, fmt.Errorf("failed to decode signed extension '%s': %w", extension.Identifier, err)
		// 	}
		// }
	}

	moduleIndex, err := r.ReadByte()
	if err != nil {
		return extrinsic, fmt.Errorf("module index: %w", err)
	}

	extrinsicIndex, err := r.ReadByte()
	if err != nil {
		return extrinsic, fmt.Errorf("extrinsic index: %w", err)
	}

	moduleMetadata, err := metadata.GetModuleForEvent(int(moduleIndex))
	if err != nil {
		return extrinsic, fmt.Errorf("module: %w", err)
	}

	if int(extrinsicIndex) > len(moduleMetadata.Calls) {
		return extrinsic, fmt.Errorf("extrinsic %d out of bounds", extrinsicIndex)
	}

	extrinsicMetadata := moduleMetadata.Calls[extrinsicIndex]
	for i, argMetadata := range extrinsicMetadata.Args {
		var arg Arg
		arg.Name = argMetadata.Name
		value, err := DecodeArgFromTypename(r, argMetadata.Type)
		if err != nil {
			return extrinsic, fmt.Errorf("arg %d: %w", i, err)
		}
		arg.Value = value
		extrinsic.Args = append(extrinsic.Args, arg)
	}

	return extrinsic, nil
}

func DecodeArgFromTypename(r *scale.Reader, typeName string) (any, error) {
	typeName = strings.TrimSpace(typeName)

	// Handle compact encoding wrapper
	if strings.HasPrefix(typeName, "Compact<") && strings.HasSuffix(typeName, ">") {
		return scale.DecodeCompact(r)
	}

	// Handle vector wrapper
	if strings.HasPrefix(typeName, "Vec<") && strings.HasSuffix(typeName, ">") {
		innerTypeName := typeName[4 : len(typeName)-1]
		// Optimization for Vec<u8> which is decoded as Bytes
		if innerTypeName == "u8" {
			return scale.DecodeBytes(r)
		}
		return scale.DecodeVec(r, func(r *scale.Reader) (any, error) {
			return DecodeArgFromTypename(r, innerTypeName)
		})
	}

	// Handle option wrapper
	if strings.HasPrefix(typeName, "Option<") && strings.HasSuffix(typeName, ">") {
		innerTypeName := typeName[7 : len(typeName)-1]
		return scale.DecodeOption(r, func(r *scale.Reader) (any, error) {
			return DecodeArgFromTypename(r, innerTypeName)
		})
	}

	// Handle tuple wrapper
	if strings.HasPrefix(typeName, "(") && strings.HasSuffix(typeName, ")") {
		innerTypesStr := typeName[1 : len(typeName)-1]
		// This is a simplified tuple parser. It won't handle nested complex types correctly.
		// e.g., (u32, Vec<(u8, u8)>) will fail.
		// But it should work for simple cases like (u32, bool).
		innerTypes := strings.Split(innerTypesStr, ",")
		result := make([]any, len(innerTypes))
		for i, innerType := range innerTypes {
			val, err := DecodeArgFromTypename(r, strings.TrimSpace(innerType))
			if err != nil {
				return nil, fmt.Errorf("failed to decode tuple element %d ('%s'): %w", i, innerType, err)
			}
			result[i] = val
		}
		return result, nil
	}

	// Handle fixed-size array
	if strings.HasPrefix(typeName, "[") && strings.HasSuffix(typeName, "]") {
		parts := strings.Split(strings.Trim(typeName, "[]"), ";")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid array type string: %s", typeName)
		}
		innerTypeName := strings.TrimSpace(parts[0])
		sizeStr := strings.TrimSpace(parts[1])
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid array size '%s': %w", sizeStr, err)
		}

		result := make([]any, size)
		for i := range size {
			val, err := DecodeArgFromTypename(r, innerTypeName)
			if err != nil {
				return nil, fmt.Errorf("failed to decode array element %d ('%s'): %w", i, innerTypeName, err)
			}
			result[i] = val
		}
		return result, nil
	}

	// Handle primitive and common types
	switch typeName {
	case "u8":
		return scale.DecodeU8(r)
	case "u16":
		return scale.DecodeU16(r)
	case "u32":
		return scale.DecodeU32(r)
	case "u64":
		return scale.DecodeU64(r)
	case "u128", "Balance":
		return scale.DecodeU128(r)
	case "bool":
		return scale.DecodeBool(r)
	case "Bytes":
		return scale.DecodeBytes(r)
	case "Text", "String":
		return scale.DecodeText(r)
	case "AccountId":
		return r.ReadBytes(32)
	case "H256", "Hash": // 32-byte hash
		return r.ReadBytes(32)
	// case "AccountInfo":
	// 	return system.DecodeAccountInfoWithTripleRefCount(r)
	// case "DispatchResult":
	// 	return system.DecodeDispatchOutcome(r)
	// case "Weight":
	// 	return system.DecodeWeight(r)
	// case "Phase":
	// 	return system.DecodePhase(r)
	// case "EventRecord":
	// 	return system.DecodeEventRecord(r)
	// case "LastRuntimeUpgradeInfo":
	// 	return system.DecodeLastRuntimeUpgradeInfo(r)
	// case "BlockLength":
	// 	return system.DecodeBlockLength(r)
	// case "BlockWeights":
	// 	return system.DecodeBlockWeights(r)
	// case "DispatchInfo":
	// 	return system.DecodeDispatchInfo(r)
	// case "DispatchError":
	// 	return system.DecodeDispatchError(r)
	default:
		return nil, fmt.Errorf("unsupported type string '%s'", typeName)
	}
}
