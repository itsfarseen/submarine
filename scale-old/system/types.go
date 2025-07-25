package system

import (
	. "submarine/scale"
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
	var err error
	ad.Free, err = DecodeU128(r)
	if err != nil {
		return ad, err
	}
	ad.Reserved, err = DecodeU128(r)
	if err != nil {
		return ad, err
	}
	ad.MiscFrozen, err = DecodeU128(r)
	if err != nil {
		return ad, err
	}
	ad.FeeFrozen, err = DecodeU128(r)
	return ad, err
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

func (a ArithmeticError) IsUnderflow() bool     { return a.Kind == ArithmeticErrorUnderflow }
func (a ArithmeticError) IsOverflow() bool      { return a.Kind == ArithmeticErrorOverflow }
func (a ArithmeticError) IsDivisionByZero() bool { return a.Kind == ArithmeticErrorDivisionByZero }

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
	DigestItemOther                     DigestItemKind = 0
	DigestItemChangesTrieRoot           DigestItemKind = 2
	DigestItemConsensus                 DigestItemKind = 4
	DigestItemSeal                      DigestItemKind = 5
	DigestItemPreRuntime                DigestItemKind = 6
	DigestItemChangesTrieSignal         DigestItemKind = 7
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

func (d DispatchClass) IsNormal() bool      { return d.Kind == DispatchClassNormal }
func (d DispatchClass) IsOperational() bool { return d.Kind == DispatchClassOperational }
func (d DispatchClass) IsMandatory() bool   { return d.Kind == DispatchClassMandatory }

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

func (p Phase) IsApplyExtrinsic() bool { return p.Kind == PhaseApplyExtrinsic }
func (p Phase) IsFinalization() bool   { return p.Kind == PhaseFinalization }
func (p Phase) IsInitialization() bool { return p.Kind == PhaseInitialization }

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

type GenericEvent struct {
	PalletIndex  byte
	VariantIndex byte
	Data         Bytes
}

type Event = GenericEvent

type EventRecord struct {
	Phase  Phase
	Event  GenericEvent
	Topics [][32]byte
}

func DecodeEventRecord(r *Reader) (EventRecord, error) {
	var er EventRecord
	var err error
	er.Phase, err = DecodePhase(r)
	if err != nil {
		return er, err
	}

	// Event is a variant, but its internal structure depends on metadata.
	// We can decode the pallet and variant index, but the fields must be
	// decoded by a higher-level package with metadata access.
	// For now, we'll just read the indexes and treat the rest as raw bytes.
	// This part is tricky because we don't know the length of the event data.
	// The `decoder` package handles this properly. This implementation is a placeholder.
	er.Event.PalletIndex, err = r.ReadByte()
	if err != nil {
		return er, err
	}
	er.Event.VariantIndex, err = r.ReadByte()
	if err != nil {
		return er, err
	}

	er.Topics, err = DecodeVec(r, func(r *Reader) ([32]byte, error) {
		var h [32]byte
		b, err := r.ReadBytes(32)
		if err != nil {
			return h, err
		}
		copy(h[:], b)
		return h, nil
	})
	return er, err
}

type Health struct {
	Peers           uint64
	IsSyncing       bool
	ShouldHavePeers bool
}

// DecodeHealth is a placeholder as Health is an RPC response type, not SCALE encoded.
func DecodeHealth(r *Reader) (Health, error) {
	return Health{}, fmt.Errorf("DecodeHealth is not implemented for SCALE")
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

// --- Added Missing Types ---

type EventId [32]byte
type EventIndex uint32
type Key Bytes
type RefCountTo259 uint8

type DispatchErrorPre6FirstKind int

const (
	DispatchErrorPre6FirstOther DispatchErrorPre6FirstKind = iota
	DispatchErrorPre6FirstCannotLookup
	DispatchErrorPre6FirstBadOrigin
	DispatchErrorPre6FirstModule
	DispatchErrorPre6FirstConsumerRemaining
	DispatchErrorPre6FirstNoProviders
	DispatchErrorPre6FirstToken
	DispatchErrorPre6FirstArithmetic
	DispatchErrorPre6FirstTransactional
)

type DispatchErrorPre6First struct {
	Kind          DispatchErrorPre6FirstKind
	Module        DispatchErrorModuleU8a
	Token         TokenError
	Arithmetic    ArithmeticError
	Transactional TransactionalError
}

type DispatchResultTo198 struct {
	IsErr bool
	Err   Text
}

type SyncState struct {
	StartingBlock uint32
	CurrentBlock  uint32
	HighestBlock  Option[uint32]
}

func DecodeSyncState(r *Reader) (SyncState, error) {
	var s SyncState
	var err error
	s.StartingBlock, err = DecodeU32(r)
	if err != nil {
		return s, err
	}
	s.CurrentBlock, err = DecodeU32(r)
	if err != nil {
		return s, err
	}
	s.HighestBlock, err = DecodeOption(r, DecodeU32)
	return s, err
}

type PeerInfo struct {
	PeerId          Text
	Roles           Text
	ProtocolVersion uint32
	BestHash        [32]byte
	BestNumber      uint32
}

func DecodePeerInfo(r *Reader) (PeerInfo, error) {
	var p PeerInfo
	var err error
	p.PeerId, err = DecodeText(r)
	if err != nil {
		return p, err
	}
	p.Roles, err = DecodeText(r)
	if err != nil {
		return p, err
	}
	p.ProtocolVersion, err = DecodeU32(r)
	if err != nil {
		return p, err
	}
	b, err := r.ReadBytes(32)
	if err != nil {
		return p, err
	}
	copy(p.BestHash[:], b)
	p.BestNumber, err = DecodeU32(r)
	return p, err
}

// --- Historical and Network Types ---

type DispatchErrorTo198 struct {
	Module Option[uint8]
	Error  uint8
}

func DecodeDispatchErrorTo198(r *Reader) (DispatchErrorTo198, error) {
	var d DispatchErrorTo198
	var err error
	d.Module, err = DecodeOption(r, DecodeU8)
	if err != nil {
		return d, err
	}
	d.Error, err = r.ReadByte()
	return d, err
}

type DispatchInfoTo190 struct {
	Weight Weight
	Class  DispatchClass
}

func DecodeDispatchInfoTo190(r *Reader) (DispatchInfoTo190, error) {
	var d DispatchInfoTo190
	var err error
	d.Weight, err = DecodeWeight(r)
	if err != nil {
		return d, err
	}
	d.Class, err = DecodeDispatchClass(r)
	return d, err
}

type DispatchInfoTo244 struct {
	Weight  Weight
	Class   DispatchClass
	PaysFee bool
}

func DecodeDispatchInfoTo244(r *Reader) (DispatchInfoTo244, error) {
	var d DispatchInfoTo244
	var err error
	d.Weight, err = DecodeWeight(r)
	if err != nil {
		return d, err
	}
	d.Class, err = DecodeDispatchClass(r)
	if err != nil {
		return d, err
	}
	d.PaysFee, err = DecodeBool(r)
	return d, err
}

type ApplyExtrinsicResultPre6 struct {
	IsErr bool
	Ok    DispatchOutcomePre6
	Err   TransactionValidityError
}

type DispatchOutcomePre6 struct {
	IsErr bool
	Err   DispatchErrorPre6
}

type DispatchErrorPre6 struct {
	Kind          DispatchErrorKind // Re-using for simplicity
	Module        DispatchErrorModuleU8a
	Token         TokenError
	Arithmetic    ArithmeticError
	Transactional TransactionalError
}

// NetworkState and related types are for RPC responses, not SCALE encoded.
// Their structs are defined for type safety, but decoders are omitted.
type NetworkState struct {
	PeerId                Text
	ListenedAddresses     []Text
	ExternalAddresses     []Text
	ConnectedPeers        map[Text]Peer
	NotConnectedPeers     map[Text]NotConnectedPeer
	AverageDownloadPerSec uint64
	AverageUploadPerSec   uint64
	Peerset               NetworkStatePeerset
}

type NetworkStatePeerset struct {
	MessageQueue uint64
	Nodes        map[Text]NetworkStatePeersetInfo
}

type NetworkStatePeersetInfo struct {
	Connected  bool
	Reputation int32
}

type NotConnectedPeer struct {
	KnownAddresses []Text
	LatestPingTime Option[PeerPing]
	VersionString  Option[Text]
}

type Peer struct {
	Enabled        bool
	Endpoint       PeerEndpoint
	KnownAddresses []Text
	LatestPingTime PeerPing
	Open           bool
	VersionString  Text
}

type PeerEndpoint struct {
	Listening PeerEndpointAddr
}

type PeerEndpointAddr struct {
	LocalAddr    Text
	SendBackAddr Text
}

type PeerPing struct {
	Nanos uint64
	Secs  uint64
}

// --- Decoders for new types ---

func DecodeEventId(r *Reader) (EventId, error) {
	var h EventId
	b, err := r.ReadBytes(32)
	if err != nil {
		return h, err
	}
	copy(h[:], b)
	return h, nil
}

func DecodeEventIndex(r *Reader) (EventIndex, error) {
	u, err := DecodeU32(r)
	return EventIndex(u), err
}

func DecodeKey(r *Reader) (Key, error) {
	b, err := DecodeBytes(r)
	return Key(b), err
}

func DecodeRefCountTo259(r *Reader) (RefCountTo259, error) {
	u, err := DecodeU8(r)
	return RefCountTo259(u), err
}

func DecodeDispatchErrorPre6First(r *Reader) (DispatchErrorPre6First, error) {
	b, err := r.ReadByte()
	if err != nil {
		return DispatchErrorPre6First{}, err
	}
	var d DispatchErrorPre6First
	d.Kind = DispatchErrorPre6FirstKind(b)
	switch d.Kind {
	case DispatchErrorPre6FirstModule:
		d.Module, err = DecodeDispatchErrorModuleU8a(r)
	case DispatchErrorPre6FirstToken:
		d.Token, err = DecodeTokenError(r)
	case DispatchErrorPre6FirstArithmetic:
		d.Arithmetic, err = DecodeArithmeticError(r)
	case DispatchErrorPre6FirstTransactional:
		d.Transactional, err = DecodeTransactionalError(r)
	}
	return d, err
}

func DecodeDispatchResultTo198(r *Reader) (DispatchResultTo198, error) {
	b, err := r.ReadByte()
	if err != nil {
		return DispatchResultTo198{}, err
	}
	var res DispatchResultTo198
	if b == 0 { // Ok
		res.IsErr = false
		return res, nil
	}
	if b == 1 { // Err
		res.IsErr = true
		res.Err, err = DecodeText(r)
		return res, err
	}
	return res, fmt.Errorf("invalid DispatchResultTo198 variant: %d", b)
}
