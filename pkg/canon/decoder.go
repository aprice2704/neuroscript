// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Implements the decoding logic for reconstructing an AST from its canonical binary representation.
// filename: pkg/canon/decoder.go
// nlines: 120
// risk_rating: HIGH

package canon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

// Decode reconstructs an AST Tree from its canonical binary representation.
func Decode(blob []byte) (*ast.Tree, error) {
	if len(blob) == 0 {
		return nil, fmt.Errorf("cannot decode an empty blob")
	}

	reader := &canonReader{
		r: bytes.NewReader(blob),
	}

	root, err := reader.readNode()
	if err != nil {
		return nil, fmt.Errorf("failed to decode root node: %w", err)
	}

	program, ok := root.(*ast.Program)
	if !ok {
		return nil, fmt.Errorf("decoded root node is not a *ast.Program, but %T", root)
	}

	return &ast.Tree{Root: program}, nil
}

// canonReader provides safe methods for reading from the canonical byte stream.
type canonReader struct {
	r *bytes.Reader
}

// readNode is the dispatcher that reads a node kind and calls the appropriate
// method to decode the rest of the node.
func (r *canonReader) readNode() (ast.Node, error) {
	kindVal, err := r.readVarint()
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read node kind: %w", err)
	}
	kind := ast.Kind(kindVal)

	switch kind {
	case ast.KindProgram:
		return r.readProgram()
	case ast.KindProcedureDecl:
		return r.readProcedure()
	case ast.KindStep:
		return r.readStep()
	case ast.KindStringLiteral:
		val, err := r.readString()
		if err != nil {
			return nil, err
		}
		return &ast.StringLiteralNode{Value: val}, nil
	case ast.KindNumberLiteral:
		strVal, err := r.readString()
		if err != nil {
			return nil, err
		}
		numVal, err := strconv.ParseFloat(strVal, 64) // For simplicity, decode all numbers as float64
		if err != nil {
			return nil, fmt.Errorf("failed to parse decoded number string '%s': %w", strVal, err)
		}
		return &ast.NumberLiteralNode{Value: numVal}, nil
	case ast.KindNilLiteral:
		return &ast.NilLiteralNode{}, nil
	// Add cases for other node types here...
	default:
		return nil, fmt.Errorf("unhandled node kind for decoding: %v", kind)
	}
}

// --- Specific Node Readers ---

func (r *canonReader) readProgram() (*ast.Program, error) {
	prog := ast.NewProgram()
	numProcs, err := r.readVarint()
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(numProcs); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, fmt.Errorf("failed to decode procedure %d: %w", i, err)
		}
		proc, ok := node.(*ast.Procedure)
		if !ok {
			return nil, fmt.Errorf("expected to decode a *ast.Procedure, but got %T", node)
		}
		prog.Procedures[proc.Name()] = proc
	}
	// Add loops for Events, Commands, etc.
	return prog, nil
}

func (r *canonReader) readProcedure() (*ast.Procedure, error) {
	name, err := r.readString()
	if err != nil {
		return nil, err
	}
	proc := &ast.Procedure{}
	proc.SetName(name)

	numSteps, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	proc.Steps = make([]ast.Step, numSteps)
	for i := 0; i < int(numSteps); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, fmt.Errorf("failed to decode step %d in procedure '%s': %w", i, name, err)
		}
		step, ok := node.(*ast.Step)
		if !ok {
			return nil, fmt.Errorf("expected to decode a *ast.Step, but got %T", node)
		}
		proc.Steps[i] = *step
	}
	return proc, nil
}

func (r *canonReader) readStep() (*ast.Step, error) {
	stepType, err := r.readString()
	if err != nil {
		return nil, err
	}
	step := &ast.Step{Type: stepType}
	// Add logic to read the rest of the step based on its type
	return step, nil
}

// --- Primitive Readers ---

func (r *canonReader) readVarint() (int64, error) {
	val, err := binary.ReadVarint(r.r)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (r *canonReader) readString() (string, error) {
	length, err := r.readVarint()
	if err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}
	if length < 0 {
		return "", fmt.Errorf("invalid string length: %d", length)
	}
	if length == 0 {
		return "", nil
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r.r, buf); err != nil {
		return "", fmt.Errorf("failed to read string content: %w", err)
	}
	return string(buf), nil
}
