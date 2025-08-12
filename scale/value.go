package scale

import "math/big"

type ValueKind int

const (
	ValueKindNull ValueKind = iota
	ValueKindInt
	ValueKindBool
	ValueKindBytes
	ValueKindText
	ValueKindList
	ValueKindStruct
)

type Value struct {
	Kind   ValueKind
	Int    *big.Int
	Bool   bool
	Bytes  []byte
	Text   string
	List   []Value
	Struct map[string]Value
}

// Constructors

func VNull() Value {
	return Value{
		Kind: ValueKindNull,
	}
}

func VInt(i *big.Int) Value {
	return Value{
		Kind: ValueKindInt,
		Int:  i,
	}
}

func VIntFromInt64(i int64) Value {
	return Value{
		Kind: ValueKindInt,
		Int:  big.NewInt(i),
	}
}

func VBool(b bool) Value {
	return Value{
		Kind: ValueKindBool,
		Bool: b,
	}
}

func VBytes(b []byte) Value {
	return Value{
		Kind:  ValueKindBytes,
		Bytes: b,
	}
}

func VText(s string) Value {
	return Value{
		Kind: ValueKindText,
		Text: s,
	}
}

func VList(list []Value) Value {
	return Value{
		Kind: ValueKindList,
		List: list,
	}
}

func VStruct(m map[string]Value) Value {
	return Value{
		Kind:   ValueKindStruct,
		Struct: m,
	}
}
