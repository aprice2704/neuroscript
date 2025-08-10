// NeuroScript Version: 0.6.3
// File version: 2
// Purpose: Implements the new registry-based AST canonicalization process.
// filename: pkg/canon/canonicalize_registry.go
// nlines: 60
// risk_rating: MEDIUM

package canon

import (
	"bytes"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
	"golang.org/x/crypto/blake2b"
)

// CanonicaliseWithRegistry produces a deterministic binary representation of an AST
// using the new registry-based codec system.
func CanonicaliseWithRegistry(tree *ast.Tree) ([]byte, [32]byte, error) {
	if tree == nil || tree.Root == nil {
		return nil, [32]byte{}, fmt.Errorf("cannot canonicalize a nil tree or a tree with a nil root")
	}

	var buf bytes.Buffer
	hasher, _ := blake2b.New256(nil)
	visitor := &canonVisitor{
		w:      &buf,
		hasher: hasher,
	}
	// The visitor needs a reference to its own visit function for recursion.
	visitor.visitor = visitor.visitWithRegistry

	// Write the magic number header first.
	visitor.write(magicNumber)

	err := visitor.visitor(tree.Root)
	if err != nil {
		return nil, [32]byte{}, err
	}

	var sum [32]byte
	hasher.Sum(sum[:0])
	return buf.Bytes(), sum, nil
}

// visitWithRegistry is the new dispatcher that uses the CodecRegistry.
func (v *canonVisitor) visitWithRegistry(node ast.Node) error {
	if node == nil {
		v.writeVarint(int64(types.KindNilLiteral))
		return nil
	}

	kind := node.Kind()
	v.writeVarint(int64(kind))

	codec, ok := CodecRegistry[kind]
	if !ok {
		return fmt.Errorf("no codec registered for node kind: %v", kind)
	}

	return codec.EncodeFunc(v, node)
}
