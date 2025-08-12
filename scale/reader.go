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
