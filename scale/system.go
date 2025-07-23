
package scale

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

// Based on https://github.com/polkadot-js/api/blob/master/packages/types/src/interfaces/balances/types.ts
type AccountData struct {
	Free       *big.Int
	Reserved   *big.Int
	MiscFrozen *big.Int
	FeeFrozen  *big.Int
}

func DecodeAccountData(r *Reader) (AccountData, error) {
	var ad AccountData
	// This is a placeholder implementation.
	// A real implementation would decode the 4 u128 values.
	return ad, nil
}

// Based on https://github.com/polkadot-js/api/blob/master/packages/types/src/interfaces/system/types.ts

type RefCount uint32

func DecodeRefCount(r *Reader) (RefCount, error) {
	u, err := DecodeU32(r)
	return RefCount(u), err
}

type AccountInfo AccountInfoWithTripleRefCount

type AccountInfoWithTripleRefCount struct {
	Nonce       uint32
	Consumers   RefCount
	Providers   RefCount
	Sufficients RefCount
	Data        AccountData
}

func DecodeAccountInfoWithTripleRefCount(r *Reader) (AccountInfoWithTripleRefCount, error) {
	var a AccountInfoWithTripleRefCount
	var err error
	a.Nonce, err = DecodeU32(r)
	if err != nil {
		return a, err
	}
	a.Consumers, err = DecodeRefCount(r)
	if err != nil {
		return a, err
	}
	a.Providers, err = DecodeRefCount(r)
	if err != nil {
		return a, err
	}
	a.Sufficients, err = DecodeRefCount(r)
	if err != nil {
		return a, err
	}
	a.Data, err = DecodeAccountData(r)
	return a, err
}

type AccountInfoWithDualRefCount struct {
	Nonce     uint32
	Consumers RefCount
	Providers RefCount
	Data      AccountData
}

func DecodeAccountInfoWithDualRefCount(r *Reader) (AccountInfoWithDualRefCount, error) {
	var a AccountInfoWithDualRefCount
	var err error
	a.Nonce, err = DecodeU32(r)
	if err != nil {
		return a, err
	}
	a.Consumers, err = DecodeRefCount(r)
	if err != nil {
		return a, err
	}
	a.Providers, err = DecodeRefCount(r)
	if err != nil {
		return a, err
	}
	a.Data, err = DecodeAccountData(r)
	return a, err
}

type AccountInfoWithProviders AccountInfoWithDualRefCount

type AccountInfoWithRefCount struct {
	Nonce    uint32
	Refcount RefCount
	Data     AccountData
}

func DecodeAccountInfoWithRefCount(r *Reader) (AccountInfoWithRefCount, error) {
	var a AccountInfoWithRefCount
	var err error
	a.Nonce, err = DecodeU32(r)
	if err != nil {
		return a, err
	}
	a.Refcount, err = DecodeRefCount(r)
	if err != nil {
		return a, err
	}
	a.Data, err = DecodeAccountData(r)
	return a, err
}

type AccountInfoWithRefCountU8 struct {
	Nonce    uint32
	Refcount uint8
	Data     AccountData
}

func DecodeAccountInfoWithRefCountU8(r *Reader) (AccountInfoWithRefCountU8, error) {
	var a AccountInfoWithRefCountU8
	var err error
	a.Nonce, err = DecodeU32(r)
	if err != nil {
		return a, err
	}
	a.Refcount, err = DecodeU8(r)
	if err != nil {
		return a, err
	}
	a.Data, err = DecodeAccountData(r)
	return a, err
}

type ApplyExtrinsicResult struct {
	IsErr bool
	Ok    DispatchOutcome
	Err   TransactionValidityError
}

func DecodeApplyExtrinsicResult(r *Reader) (ApplyExtrinsicResult, error) {
	b, err := r.ReadByte()
	if err != nil {
		return ApplyExtrinsicResult{}, err
	}
	var res ApplyExtrinsicResult
	if b == 0 { // Ok
		res.IsErr = false
		res.Ok, err = DecodeDispatchOutcome(r)
		return res, err
	}
	if b == 1 { // Err
		res.IsErr = true
		res.Err, err = DecodeTransactionValidityError(r)
		return res, err
	}
	return res, fmt.Errorf("invalid ApplyExtrinsicResult variant: %d", b)
}

type ArithmeticErrorKind int

const (
	ArithmeticErrorUnderflow ArithmeticErrorKind = iota
	ArithmeticErrorOverflow
	ArithmeticErrorDivisionByZero
)

type ArithmeticError struct {
	Kind ArithmeticErrorKind
}

func DecodeArithmeticError(r *Reader) (ArithmeticError, error) {
	b, err := r.ReadByte()
	if err != nil {
		return ArithmeticError{}, err
	}
	return ArithmeticError{Kind: ArithmeticErrorKind(b)}, nil
}

type PerDispatchClassU32 struct {
	Normal      uint32
	Operational uint32
	Mandatory   uint32
}

func DecodePerDispatchClassU32(r *Reader) (PerDispatchClassU32, error) {
	var p PerDispatchClassU32
	var err error
	p.Normal, err = DecodeU32(r)
	if err != nil {
		return p, err
	}
	p.Operational, err = DecodeU32(r)
	if err != nil {
		return p, err
	}
	p.Mandatory, err = DecodeU32(r)
	return p, err
}

type BlockLength struct {
	Max PerDispatchClassU32
}

func DecodeBlockLength(r *Reader) (BlockLength, error) {
	var b BlockLength
	var err error
	b.Max, err = DecodePerDispatchClassU32(r)
	return b, err
}

type Weight uint64

func DecodeWeight(r *Reader) (Weight, error) {
	b, err := r.ReadBytes(8)
	if err != nil {
		return 0, err
	}
	return Weight(binary.LittleEndian.Uint64(b)), nil
}

type PerDispatchClassWeight struct {
	Normal      Weight
	Operational Weight
	Mandatory   Weight
}

func DecodePerDispatchClassWeight(r *Reader) (PerDispatchClassWeight, error) {
	var p PerDispatchClassWeight
	var err error
	p.Normal, err = DecodeWeight(r)
	if err != nil {
		return p, err
	}
	p.Operational, err = DecodeWeight(r)
	if err != nil {
		return p, err
	}
	p.Mandatory, err = DecodeWeight(r)
	return p, err
}

type WeightPerClass struct {
	BaseExtrinsic Weight
	MaxExtrinsic  Option[Weight]
	MaxTotal      Option[Weight]
	Reserved      Option[Weight]
}

func DecodeWeightPerClass(r *Reader) (WeightPerClass, error) {
	var w WeightPerClass
	var err error
	w.BaseExtrinsic, err = DecodeWeight(r)
	if err != nil {
		return w, err
	}
	w.MaxExtrinsic, err = DecodeOption(r, DecodeWeight)
	if err != nil {
		return w, err
	}
	w.MaxTotal, err = DecodeOption(r, DecodeWeight)
	if err != nil {
		return w, err
	}
	w.Reserved, err = DecodeOption(r, DecodeWeight)
	return w, err
}

type PerDispatchClassWeightsPerClass struct {
	Normal      WeightPerClass
	Operational WeightPerClass
	Mandatory   WeightPerClass
}

func DecodePerDispatchClassWeightsPerClass(r *Reader) (PerDispatchClassWeightsPerClass, error) {
	var p PerDispatchClassWeightsPerClass
	var err error
	p.Normal, err = DecodeWeightPerClass(r)
	if err != nil {
		return p, err
	}
	p.Operational, err = DecodeWeightPerClass(r)
	if err != nil {
		return p, err
	}
	p.Mandatory, err = DecodeWeightPerClass(r)
	return p, err
}

type BlockWeights struct {
	BaseBlock Weight
	MaxBlock  Weight
	PerClass  PerDispatchClassWeightsPerClass
}

func DecodeBlockWeights(r *Reader) (BlockWeights, error) {
	var b BlockWeights
	var err error
	b.BaseBlock, err = DecodeWeight(r)
	if err != nil {
		return b, err
	}
	b.MaxBlock, err = DecodeWeight(r)
	if err != nil {
		return b, err
	}
	b.PerClass, err = DecodePerDispatchClassWeightsPerClass(r)
	return b, err
}

type ChainProperties struct {
	Ss58Format    Option[uint8]
	TokenDecimals Option[[]uint32]
	TokenSymbol   Option[[]Text]
}

func DecodeChainProperties(r *Reader) (ChainProperties, error) {
	var cp ChainProperties
	var err error

	cp.Ss58Format, err = DecodeOption(r, DecodeU8)
	if err != nil {
		return cp, err
	}

	cp.TokenDecimals, err = DecodeOption(r, func(r *Reader) ([]uint32, error) {
		return DecodeVec(r, DecodeU32)
	})
	if err != nil {
		return cp, err
	}

	cp.TokenSymbol, err = DecodeOption(r, func(r *Reader) ([]Text, error) {
		return DecodeVec(r, DecodeText)
	})
	if err != nil {
		return cp, err
	}

	return cp, nil
}

type ChainTypeKind int

const (
	ChainTypeDevelopment ChainTypeKind = iota
	ChainTypeLocal
	ChainTypeLive
	ChainTypeCustom
)

type ChainType struct {
	Kind   ChainTypeKind
	Custom Text
}

func DecodeChainType(r *Reader) (ChainType, error) {
	b, err := r.ReadByte()
	if err != nil {
		return ChainType{}, err
	}
	var c ChainType
	c.Kind = ChainTypeKind(b)
	if c.Kind == ChainTypeCustom {
		c.Custom, err = DecodeText(r)
	}
	return c, err
}

type ConsumedWeight PerDispatchClassWeight

type DigestItemKind int

const (
	DigestItemOther             DigestItemKind = 0
	DigestItemChangesTrieRoot   DigestItemKind = 2
	DigestItemConsensus         DigestItemKind = 4
	DigestItemSeal              DigestItemKind = 5
	DigestItemPreRuntime        DigestItemKind = 6
	DigestItemChangesTrieSignal DigestItemKind = 7
	DigestItemRuntimeEnvironmentUpdated DigestItemKind = 8
)

type ConsensusEngineId [4]byte

type DigestItem struct {
	Kind  DigestItemKind
	Value any
}

type Digest struct {
	Logs []DigestItem
}

type DigestOf Digest

func DecodeDigest(r *Reader) (Digest, error) {
	logs, err := DecodeVec(r, DecodeDigestItem)
	if err != nil {
		return Digest{}, err
	}
	return Digest{Logs: logs}, nil
}

func DecodeDigestItem(r *Reader) (DigestItem, error) {
	variant, err := r.ReadByte()
	if err != nil {
		return DigestItem{}, err
	}
	var item DigestItem
	item.Kind = DigestItemKind(variant)

	switch item.Kind {
	case DigestItemOther:
		val, err := DecodeBytes(r)
		item.Value = val
		return item, err
	case DigestItemChangesTrieRoot:
		val, err := r.ReadBytes(32) // Hash
		item.Value = val
		return item, err
	case DigestItemConsensus, DigestItemSeal, DigestItemPreRuntime:
		var consensusId ConsensusEngineId
		b, err := r.ReadBytes(4)
		if err != nil {
			return item, err
		}
		copy(consensusId[:], b)
		data, err := DecodeBytes(r)
		if err != nil {
			return item, err
		}
		item.Value = [2]any{consensusId, data}
		return item, nil
	case DigestItemChangesTrieSignal:
		// This is an enum, needs further definition if we want to decode the value
		return item, nil
	case DigestItemRuntimeEnvironmentUpdated:
		// No data associated with this variant
		return item, nil
	default:
		return item, fmt.Errorf("unsupported DigestItem kind: %d", item.Kind)
	}
}

type DispatchClassKind int

const (
	DispatchClassNormal DispatchClassKind = iota
	DispatchClassOperational
	DispatchClassMandatory
)

type DispatchClass struct {
	Kind DispatchClassKind
}

func DecodeDispatchClass(r *Reader) (DispatchClass, error) {
	b, err := r.ReadByte()
	if err != nil {
		return DispatchClass{}, err
	}
	return DispatchClass{Kind: DispatchClassKind(b)}, nil
}

type DispatchErrorModuleU8a struct {
	Index uint8
	Error [8]byte // U8aFixed of length 8, assuming a max error size
}

func DecodeDispatchErrorModuleU8a(r *Reader) (DispatchErrorModuleU8a, error) {
	var d DispatchErrorModuleU8a
	var err error
	d.Index, err = r.ReadByte()
	if err != nil {
		return d, err
	}
	b, err := r.ReadBytes(8)
	if err != nil {
		return d, err
	}
	copy(d.Error[:], b)
	return d, err
}

type TokenErrorKind int

const (
	TokenErrorNoFunds TokenErrorKind = iota
	TokenErrorWouldDie
	TokenErrorBelowMinimum
	TokenErrorCannotCreate
	TokenErrorUnknownAsset
	TokenErrorFrozen
	TokenErrorUnsupported
	TokenErrorUnderflow
	TokenErrorOverflow
)

type TokenError struct {
	Kind TokenErrorKind
}

func DecodeTokenError(r *Reader) (TokenError, error) {
	b, err := r.ReadByte()
	if err != nil {
		return TokenError{}, err
	}
	return TokenError{Kind: TokenErrorKind(b)}, nil
}

type TransactionalErrorKind int

const (
	TransactionalErrorLimitReached TransactionalErrorKind = iota
	TransactionalErrorNoLayer
)

type TransactionalError struct {
	Kind TransactionalErrorKind
}

func DecodeTransactionalError(r *Reader) (TransactionalError, error) {
	b, err := r.ReadByte()
	if err != nil {
		return TransactionalError{}, err
	}
	return TransactionalError{Kind: TransactionalErrorKind(b)}, nil
}

type DispatchErrorKind int

const (
	DispatchErrorOther DispatchErrorKind = iota
	DispatchErrorCannotLookup
	DispatchErrorBadOrigin
	DispatchErrorModule
	DispatchErrorConsumerRemaining
	DispatchErrorNoProviders
	DispatchErrorTooManyConsumers
	DispatchErrorToken
	DispatchErrorArithmetic
	DispatchErrorTransactional
	DispatchErrorExhausted
	DispatchErrorCorruption
	DispatchErrorUnavailable
)

type DispatchError struct {
	Kind          DispatchErrorKind
	Module        DispatchErrorModuleU8a
	Token         TokenError
	Arithmetic    ArithmeticError
	Transactional TransactionalError
}

func DecodeDispatchError(r *Reader) (DispatchError, error) {
	b, err := r.ReadByte()
	if err != nil {
		return DispatchError{}, err
	}
	var d DispatchError
	d.Kind = DispatchErrorKind(b)
	switch d.Kind {
	case DispatchErrorModule:
		d.Module, err = DecodeDispatchErrorModuleU8a(r)
	case DispatchErrorToken:
		d.Token, err = DecodeTokenError(r)
	case DispatchErrorArithmetic:
		d.Arithmetic, err = DecodeArithmeticError(r)
	case DispatchErrorTransactional:
		d.Transactional, err = DecodeTransactionalError(r)
	}
	return d, err
}

type PaysFeeKind int

const (
	PaysFeeYes PaysFeeKind = iota
	PaysFeeNo
)

type PaysFee struct {
	Kind PaysFeeKind
}

func DecodePaysFee(r *Reader) (PaysFee, error) {
	b, err := r.ReadByte()
	if err != nil {
		return PaysFee{}, err
	}
	return PaysFee{Kind: PaysFeeKind(b)}, nil
}

type DispatchInfo struct {
	Weight  Weight
	Class   DispatchClass
	PaysFee PaysFee
}

func DecodeDispatchInfo(r *Reader) (DispatchInfo, error) {
	var d DispatchInfo
	var err error
	d.Weight, err = DecodeWeight(r)
	if err != nil {
		return d, err
	}
	d.Class, err = DecodeDispatchClass(r)
	if err != nil {
		return d, err
	}
	d.PaysFee, err = DecodePaysFee(r)
	return d, err
}

type DispatchOutcome struct {
	IsErr bool
	Err   DispatchError
}

func DecodeDispatchOutcome(r *Reader) (DispatchOutcome, error) {
	b, err := r.ReadByte()
	if err != nil {
		return DispatchOutcome{}, err
	}
	var o DispatchOutcome
	if b == 0 { // Ok
		o.IsErr = false
		return o, nil
	}
	if b == 1 { // Err
		o.IsErr = true
		o.Err, err = DecodeDispatchError(r)
		return o, err
	}
	return o, fmt.Errorf("invalid DispatchOutcome variant: %d", b)
}

type DispatchResult DispatchOutcome
type DispatchResultOf DispatchResult

type PhaseKind int

const (
	PhaseApplyExtrinsic PhaseKind = iota
	PhaseFinalization
	PhaseInitialization
)

type Phase struct {
	Kind              PhaseKind
	ApplyExtrinsicU32 uint32
}

func DecodePhase(r *Reader) (Phase, error) {
	b, err := r.ReadByte()
	if err != nil {
		return Phase{}, err
	}
	var p Phase
	p.Kind = PhaseKind(b)
	if p.Kind == PhaseApplyExtrinsic {
		p.ApplyExtrinsicU32, err = DecodeU32(r)
	}
	return p, err
}

type EventRecord struct {
	Phase  Phase
	Event  any // Generic event, to be defined elsewhere
	Topics [][32]byte
}

// Skipping EventRecord decoder as it depends on a generic Event type that requires metadata.

type Health struct {
	Peers           uint64
	IsSyncing       bool
	ShouldHavePeers bool
}

func DecodeHealth(r *Reader) (Health, error) {
	var h Health
	// This is a placeholder, actual decoding depends on the wire format.
	return h, nil
}

type InvalidTransactionKind int

const (
	InvalidTransactionCall InvalidTransactionKind = iota
	InvalidTransactionPayment
	InvalidTransactionFuture
	InvalidTransactionStale
	InvalidTransactionBadProof
	InvalidTransactionAncientBirthBlock
	InvalidTransactionExhaustsResources
	InvalidTransactionCustom
	InvalidTransactionBadMandatory
	InvalidTransactionMandatoryDispatch
	InvalidTransactionBadSigner
)

type InvalidTransaction struct {
	Kind   InvalidTransactionKind
	Custom uint8
}

func DecodeInvalidTransaction(r *Reader) (InvalidTransaction, error) {
	b, err := r.ReadByte()
	if err != nil {
		return InvalidTransaction{}, err
	}
	var i InvalidTransaction
	i.Kind = InvalidTransactionKind(b)
	if i.Kind == InvalidTransactionCustom {
		i.Custom, err = r.ReadByte()
	}
	return i, err
}

type LastRuntimeUpgradeInfo struct {
	SpecVersion uint32
	SpecName    Text
}

func DecodeLastRuntimeUpgradeInfo(r *Reader) (LastRuntimeUpgradeInfo, error) {
	var l LastRuntimeUpgradeInfo
	var err error
	compactVersion, err := DecodeCompact(r)
	if err != nil {
		return l, err
	}
	l.SpecVersion = uint32(compactVersion.Uint64())
	l.SpecName, err = DecodeText(r)
	return l, err
}

type NodeRoleKind int

const (
	NodeRoleFull NodeRoleKind = iota
	NodeRoleLightClient
	NodeRoleAuthority
	NodeRoleUnknownRole
)

type NodeRole struct {
	Kind        NodeRoleKind
	UnknownRole uint8
}

func DecodeNodeRole(r *Reader) (NodeRole, error) {
	b, err := r.ReadByte()
	if err != nil {
		return NodeRole{}, err
	}
	var n NodeRole
	n.Kind = NodeRoleKind(b)
	if n.Kind == NodeRoleUnknownRole {
		n.UnknownRole, err = r.ReadByte()
	}
	return n, err
}

type RawOriginKind int

const (
	RawOriginRoot RawOriginKind = iota
	RawOriginSigned
	RawOriginNone
)

type RawOrigin struct {
	Kind   RawOriginKind
	Signed [32]byte // AccountId
}

func DecodeRawOrigin(r *Reader) (RawOrigin, error) {
	b, err := r.ReadByte()
	if err != nil {
		return RawOrigin{}, err
	}
	var o RawOrigin
	o.Kind = RawOriginKind(b)
	if o.Kind == RawOriginSigned {
		bytes, err := r.ReadBytes(32)
		if err != nil {
			return o, err
		}
		copy(o.Signed[:], bytes)
	}
	return o, err
}

type SystemOrigin RawOrigin

func DecodeSystemOrigin(r *Reader) (SystemOrigin, error) {
	ro, err := DecodeRawOrigin(r)
	return SystemOrigin(ro), err
}

type UnknownTransactionKind int

const (
	UnknownTransactionCannotLookup UnknownTransactionKind = iota
	UnknownTransactionNoUnsignedValidator
	UnknownTransactionCustom
)

type UnknownTransaction struct {
	Kind   UnknownTransactionKind
	Custom uint8
}

func DecodeUnknownTransaction(r *Reader) (UnknownTransaction, error) {
	b, err := r.ReadByte()
	if err != nil {
		return UnknownTransaction{}, err
	}
	var u UnknownTransaction
	u.Kind = UnknownTransactionKind(b)
	if u.Kind == UnknownTransactionCustom {
		u.Custom, err = r.ReadByte()
	}
	return u, err
}

type TransactionValidityErrorKind int

const (
	TransactionValidityErrorInvalid TransactionValidityErrorKind = iota
	TransactionValidityErrorUnknown
)

type TransactionValidityError struct {
	Kind    TransactionValidityErrorKind
	Invalid InvalidTransaction
	Unknown UnknownTransaction
}

func DecodeTransactionValidityError(r *Reader) (TransactionValidityError, error) {
	b, err := r.ReadByte()
	if err != nil {
		return TransactionValidityError{}, err
	}
	var t TransactionValidityError
	t.Kind = TransactionValidityErrorKind(b)
	switch t.Kind {
	case TransactionValidityErrorInvalid:
		t.Invalid, err = DecodeInvalidTransaction(r)
	case TransactionValidityErrorUnknown:
		t.Unknown, err = DecodeUnknownTransaction(r)
	}
	return t, err
}
