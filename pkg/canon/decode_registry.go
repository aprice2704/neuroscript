// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 8
// :: description: Implements the registry-based AST decoding process.
// :: latestChange: Passing the magic version byte down to the canonReader for backward compatibility.
// :: filename: pkg/canon/decode_registry.go
// :: serialization: go

package canon

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// isValidMagic checks if the blob starts with the stable magic number
// or a known legacy magic number (which used volatile types.KindMarker).
func isValidMagic(blob []byte) bool {
	if len(blob) < 4 {
		return false
	}
	if blob[0] == 'N' && blob[1] == 'S' && blob[2] == 'C' {
		// Stable version
		if blob[3] == magicNumber[3] {
			return true
		}
		// Legacy types.KindMarker values (historical range for backward compatibility)
		if blob[3] >= 20 && blob[3] <= 50 {
			return true
		}
	}
	return false
}

// DecodeWithRegistry reconstructs an AST Tree from its binary representation
// using the new registry-based codec system. It requires the root node
// to be an *ast.Program.
func DecodeWithRegistry(blob []byte) (*ast.Tree, error) {
	if !isValidMagic(blob) {
		return nil, ErrInvalidMagic
	}

	reader := &canonReader{
		r:       bytes.NewReader(blob[4:]),
		version: blob[3],
	}
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

// DecodeNode reconstructs a single AST node from its binary representation.
// This is the correct method for reading minimal AST fragments (e.g.,
// a persisted *ast.Procedure or *ast.StringLiteralNode).
func DecodeNode(blob []byte) (ast.Node, error) {
	if !isValidMagic(blob) {
		return nil, ErrInvalidMagic
	}

	reader := &canonReader{
		r:       bytes.NewReader(blob[4:]),
		version: blob[3],
	}
	reader.visitor = reader.readNodeWithRegistry

	node, err := reader.visitor()
	if err != nil {
		return nil, err
	}

	// Note: We skip the restoreCallTargetKinds pass here, as we do not
	// have the context of the full program. This is correct, as the
	// node is being loaded for data, not direct execution.
	return node, nil
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
