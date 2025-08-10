// NeuroScript Version: 0.6.3
// File version: 5
// Purpose: Implements the new registry-based AST decoding process, now with simplified and robust sentinel error handling.
// filename: pkg/canon/decode_registry.go
// nlines: 65
// risk_rating: MEDIUM

package canon

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// DecodeWithRegistry reconstructs an AST Tree from its binary representation
// using the new registry-based codec system.
func DecodeWithRegistry(blob []byte) (*ast.Tree, error) {
	if blob == nil || !bytes.HasPrefix(blob, magicNumber) {
		return nil, ErrInvalidMagic
	}

	reader := &canonReader{r: bytes.NewReader(blob[len(magicNumber):])}
	reader.visitor = reader.readNodeWithRegistry

	root, err := reader.visitor()
	if err != nil {
		// The visitor now returns the correct sentinel error directly.
		return nil, err
	}
	prog, ok := root.(*ast.Program)
	if !ok {
		return nil, fmt.Errorf("decoded root node is not *ast.Program but %T", root)
	}

	// Post-pass: restore CallTarget.BaseNode.NodeKind deterministically.
	restoreCallTargetKinds(prog)

	return &ast.Tree{Root: prog}, nil
}

// readNodeWithRegistry is the new dispatcher that uses the CodecRegistry.
func (r *canonReader) readNodeWithRegistry() (ast.Node, error) {
	offset := r.r.Size() - int64(r.r.Len())
	kindVal, err := r.readVarint()
	if err != nil {
		// Any error reading the kind is a sign of truncation.
		return nil, ErrTruncatedData
	}
	kind := types.Kind(kindVal)
	r.history = append(r.history, fmt.Sprintf("%v", kind))

	codec, ok := CodecRegistry[kind]
	if !ok {
		return nil, fmt.Errorf("%w: %v (%d) at byte offset %d. History: [%s]", ErrUnknownCodec, kind, kind, offset, strings.Join(r.history, ", "))
	}

	node, err := codec.DecodeFunc(r)
	if err != nil {
		// Propagate sentinel errors from deeper in the call stack.
		if errors.Is(err, ErrTruncatedData) {
			return nil, ErrTruncatedData
		}
		// Wrap other, unexpected errors.
		return nil, fmt.Errorf("codec for kind %v failed: %w", kind, err)
	}
	return node, nil
}
