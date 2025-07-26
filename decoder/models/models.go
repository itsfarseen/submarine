package models

import (
	"submarine/scale/base"
)

// DecodedArg holds the name and decoded value of a single extrinsic argument.
type DecodedArg struct {
	Name  string
	Value any // Using `any` to hold various decoded types.
}

// DecodedCall represents the action part of an extrinsic.
type DecodedCall struct {
	PalletName string
	CallName   string
	Args       []DecodedArg
}

// DecodedExtrinsic represents the full decoded extrinsic.
type DecodedExtrinsic struct {
	Signature base.Signature
	Call      DecodedPalletVariant
}

type EventRecord struct {
	Phase EventPhase
	Event DecodedEvent
	// Topics are skipped for this implementation but would be a [][]byte.
}

type EventPhase struct {
	IsApplyExtrinsic bool
	AsApplyExtrinsic uint32 // The index of the extrinsic
	IsFinalization   bool
	IsInitialization bool
}

type DecodedEvent struct {
	PalletName string
	EventName  string
	Args       []DecodedArg
}

// DecodedPalletVariant is a generic representation of a decoded call or event.
type DecodedPalletVariant struct {
	PalletName  string
	VariantName string
	Args        []DecodedArg
}
