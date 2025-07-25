package scale

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

// Reader helps to decode SCALE types from a byte slice.
type Reader struct {
	data []byte
	pos  int
}

// NewReader creates a new reader instance.
func NewReader(data []byte) *Reader {
	return &Reader{data: data, pos: 0}
}

// ReadByte reads a single byte and advances the position.
func (r *Reader) ReadByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("reader: out of bounds")
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

// ReadBytes reads n bytes and advances the position.
func (r *Reader) ReadBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.data) {
		return nil, fmt.Errorf("reader: out of bounds for %d bytes", n)
	}
	bytes := r.data[r.pos : r.pos+n]
	r.pos += n
	return bytes, nil
}

func (r *Reader) Pos() int {
	return r.pos
}

func reverseBytes(data []byte) []byte {
	reversed := make([]byte, len(data))
	for i, b := range data {
		reversed[len(data)-1-i] = b
	}
	return reversed
}

// DecodeCompact decodes a SCALE compact-encoded integer.
func DecodeCompact(r *Reader) (*big.Int, error) {
	firstByte, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	mode := firstByte & 0b11
	switch mode {
	case 0:
		return big.NewInt(int64(firstByte >> 2)), nil
	case 1:
		secondByte, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("compact[1]: %w", err)
		}
		val := uint16(firstByte>>2) | uint16(secondByte)<<6
		return big.NewInt(int64(val)), nil
	case 2:
		bytes, err := r.ReadBytes(3)
		if err != nil {
			return nil, fmt.Errorf("compact[2]: %w", err)
		}
		val := uint32(firstByte>>2) | uint32(bytes[0])<<6 | uint32(bytes[1])<<14 | uint32(bytes[2])<<22
		return big.NewInt(int64(val)), nil
	case 3:
		length := int((firstByte >> 2) + 4)
		bytes, err := r.ReadBytes(length)
		if err != nil {
			return nil, fmt.Errorf("compact[3]: %w", err)
		}
		bytesLE := reverseBytes(bytes)
		return new(big.Int).SetBytes(bytesLE), nil
	default:
		return nil, fmt.Errorf("compact[?]: %w", err)
	}
}

func DecodeU8(r *Reader) (uint8, error) {
	return r.ReadByte()
}

func DecodeU16(r *Reader) (uint16, error) {
	bytes, err := r.ReadBytes(2)
	if err != nil {
		return 0, fmt.Errorf("u16: %w", err)
	}
	return binary.LittleEndian.Uint16(bytes), nil
}

func DecodeU32(r *Reader) (uint32, error) {
	bytes, err := r.ReadBytes(4)
	if err != nil {
		return 0, fmt.Errorf("u32: %w", err)
	}
	return binary.LittleEndian.Uint32(bytes), nil
}

func DecodeU128(r *Reader) (*big.Int, error) {
	b, err := r.ReadBytes(16)
	if err != nil {
		return nil, err
	}
	// Reverse for big.Int which expects big-endian bytes
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return new(big.Int).SetBytes(b), nil
}

func DecodeBool(r *Reader) (bool, error) {
	b, err := r.ReadByte()
	if err != nil {
		return false, err
	}
	if b == 0x00 {
		return false, nil
	}
	if b == 0x01 {
		return true, nil
	}
	return false, fmt.Errorf("bool? %x", b)
}

func DecodeText(r *Reader) (string, error) {
	length, err := DecodeCompact(r)
	if err != nil {
		return "", fmt.Errorf("text.len: %w", err)
	}
	bytes, err := r.ReadBytes(int(length.Int64()))
	if err != nil {
		return "", fmt.Errorf("text: %w", err)
	}
	return string(bytes), nil
}

func DecodeBytes(r *Reader) ([]byte, error) {
	length, err := DecodeCompact(r)
	if err != nil {
		return nil, fmt.Errorf("bytes.len: %w", err)
	}
	bytes, err := r.ReadBytes(int(length.Int64()))
	if err != nil {
		return nil, fmt.Errorf("bytes: %w", err)
	}
	return bytes, nil
}

func DecodeVec[T any](r *Reader, decoder func(*Reader) (T, error)) ([]T, error) {
	length, err := DecodeCompact(r)
	if err != nil {
		return nil, fmt.Errorf("vec.len: %w", err)
	}
	len64 := length.Int64()

	vec := make([]T, len64)
	for i := range len64 {
		item, err := decoder(r)
		if err != nil {
			return nil, fmt.Errorf("vec[%d]: %w", i, err)
		}
		vec[i] = item
	}
	return vec, nil
}

// Returns nil if the Option doesn't have a value
func DecodeOption[T any](r *Reader, decoder func(*Reader) (T, error)) (*T, error) {
	hasValue, err := DecodeBool(r)
	if err != nil {
		return nil, fmt.Errorf("option.flag: %w", err)
	}

	if !hasValue {
		return nil, nil
	}

	value, err := decoder(r)
	if err != nil {
		return nil, fmt.Errorf("option.value: %w", err)
	}

	return &value, nil
}
