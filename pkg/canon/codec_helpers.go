// NeuroScript Version: 0.6.3
// File version: 2
// Purpose: Defines shared helper structs and primitive I/O methods, now with robust sentinel error handling.
// filename: pkg/canon/codec_helpers.go
// nlines: 100
// risk_rating: LOW

package canon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// canonVisitor walks the AST and writes its canonical representation.
// It holds the writer, a hasher, and a reference to the main visit function
// for recursive calls within codecs.
type canonVisitor struct {
	w       io.Writer
	hasher  hash.Hash
	visitor func(ast.Node) error
}

// canonReader reads from a canonical binary stream.
type canonReader struct {
	r       *bytes.Reader
	history []string
	visitor func() (ast.Node, error) // For recursive calls
}

// --- Primitive Writers ---

func (v *canonVisitor) write(p []byte) {
	if v.w != nil {
		v.w.Write(p)
	}
	if v.hasher != nil {
		v.hasher.Write(p)
	}
}

func (v *canonVisitor) writeVarint(x int64) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, x)
	v.write(buf[:n])
}

func (v *canonVisitor) writeString(s string) {
	v.writeVarint(int64(len(s)))
	if len(s) > 0 {
		v.write([]byte(s))
	}
}

func (v *canonVisitor) writeBool(b bool) {
	if b {
		v.write([]byte{1})
	} else {
		v.write([]byte{0})
	}
}

func (v *canonVisitor) writeNumber(val interface{}) {
	strVal := fmt.Sprintf("%v", val)
	// Use a type marker for potential future use (e.g., int vs float)
	v.write([]byte{0x01}) // 0x01 for float64/generic
	v.writeString(strVal)
}

// --- Primitive Readers ---

func (r *canonReader) readByte() (byte, error) {
	b, err := r.r.ReadByte()
	if err != nil {
		return 0, ErrTruncatedData
	}
	return b, nil
}

func (r *canonReader) readVarint() (int64, error) {
	val, err := binary.ReadVarint(r.r)
	if err != nil {
		return 0, ErrTruncatedData
	}
	return val, nil
}

func (r *canonReader) readString() (string, error) {
	length, err := r.readVarint()
	if err != nil {
		return "", err // Already converted to ErrTruncatedData by readVarint
	}
	if length < 0 {
		return "", fmt.Errorf("invalid string length: %d", length)
	}
	if length == 0 {
		return "", nil
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(r.r, buf)
	if err != nil {
		return "", ErrTruncatedData
	}
	return string(buf), nil
}

func (r *canonReader) readBool() (bool, error) {
	b, err := r.readByte()
	if err != nil {
		return false, err // Already converted by readByte
	}
	return b == 1, nil
}

func (r *canonReader) readNumber() (interface{}, error) {
	_, err := r.readByte() // Read and discard the type marker
	if err != nil {
		return nil, err // Already converted
	}
	s, err := r.readString()
	if err != nil {
		return nil, err // Already converted
	}
	// Always parse as float64 to match parser behavior
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number format: %w", err)
	}
	return val, nil
}
