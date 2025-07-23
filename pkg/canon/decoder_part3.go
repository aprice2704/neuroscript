// NeuroScript Version: 0.6.0
// File version: 21.3
// Purpose: Adds all missing Kind cases to the decoder's switch statement to satisfy the coverage test.
// filename: pkg/canon/decoder_part3.go
// nlines: 150+
// risk_rating: HIGH

package canon

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// This file continues the implementation from part 2.

func (r *canonReader) readCommand() (*ast.CommandNode, error) {
	cmd := &ast.CommandNode{BaseNode: ast.BaseNode{NodeKind: types.KindCommandBlock}}
	numSteps, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	cmd.Body = make([]ast.Step, int(numSteps))
	for i := 0; i < int(numSteps); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		if step, ok := node.(*ast.Step); ok {
			cmd.Body[i] = *step
		} else {
			return nil, fmt.Errorf("expected to decode a *ast.Step but got %T", node)
		}
	}
	return cmd, nil
}

func (r *canonReader) readProcedure() (*ast.Procedure, error) {
	proc := &ast.Procedure{BaseNode: ast.BaseNode{NodeKind: types.KindProcedureDecl}}
	name, err := r.readString()
	if err != nil {
		return nil, err
	}
	proc.SetName(name)
	numSteps, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	proc.Steps = make([]ast.Step, numSteps)
	for i := 0; i < int(numSteps); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		if step, ok := node.(*ast.Step); ok {
			proc.Steps[i] = *step
		} else {
			return nil, fmt.Errorf("expected to decode a *ast.Step but got %T", node)
		}
	}
	return proc, nil
}

func (r *canonReader) readStep() (*ast.Step, error) {
	stepType, err := r.readString()
	if err != nil {
		return nil, err
	}
	step := &ast.Step{BaseNode: ast.BaseNode{NodeKind: types.KindStep}, Type: stepType}
	switch stepType {
	case "emit":
		val, err := r.readNode()
		if err != nil {
			return nil, err
		}
		step.Values = []ast.Expression{val.(ast.Expression)}
	case "set":
		numLValues, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.LValues = make([]*ast.LValueNode, numLValues)
		for i := 0; i < int(numLValues); i++ {
			node, err := r.readNode()
			if err != nil {
				return nil, err
			}
			step.LValues[i] = node.(*ast.LValueNode)
		}
		numRValues, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Values = make([]ast.Expression, numRValues)
		for i := 0; i < int(numRValues); i++ {
			node, err := r.readNode()
			if err != nil {
				return nil, err
			}
			step.Values[i] = node.(ast.Expression)
		}
	case "return":
		numValues, err := r.readVarint()
		if err != nil {
			return nil, fmt.Errorf("failed to read return value count: %w", err)
		}
		step.Values = make([]ast.Expression, numValues)
		for i := 0; i < int(numValues); i++ {
			node, err := r.readNode()
			if err != nil {
				return nil, fmt.Errorf("failed to read return value %d: %w", i, err)
			}
			step.Values[i] = node.(ast.Expression)
		}
	case "call", "expression":
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		if call, ok := node.(*ast.CallableExprNode); ok {
			step.Call = call
		}
		if exprStmt, ok := node.(*ast.ExpressionStatementNode); ok {
			step.ExpressionStmt = exprStmt
		}
	}
	return step, nil
}

func (r *canonReader) readLValue() (*ast.LValueNode, error) {
	identifier, err := r.readString()
	if err != nil {
		return nil, err
	}
	return &ast.LValueNode{BaseNode: ast.BaseNode{NodeKind: types.KindLValue}, Identifier: identifier}, nil
}

func (r *canonReader) readBinaryOp() (*ast.BinaryOpNode, error) {
	op, err := r.readString()
	if err != nil {
		return nil, err
	}
	left, err := r.readNode()
	if err != nil {
		return nil, err
	}
	right, err := r.readNode()
	if err != nil {
		return nil, err
	}
	return &ast.BinaryOpNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindBinaryOp},
		Operator: op,
		Left:     left.(ast.Expression),
		Right:    right.(ast.Expression),
	}, nil
}

func (r *canonReader) readUnaryOp() (*ast.UnaryOpNode, error) {
	op, err := r.readString()
	if err != nil {
		return nil, err
	}
	operand, err := r.readNode()
	if err != nil {
		return nil, err
	}
	return &ast.UnaryOpNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindUnaryOp},
		Operator: op,
		Operand:  operand.(ast.Expression),
	}, nil
}

// --- Primitive Readers ---

func (r *canonReader) readVarint() (int64, error) {
	return binary.ReadVarint(r.r)
}

func (r *canonReader) readString() (string, error) {
	length, err := r.readVarint()
	if err != nil {
		return "", err
	}
	if length < 0 {
		return "", fmt.Errorf("invalid string length: %d", length)
	}
	if length == 0 {
		return "", nil
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(r.r, buf)
	return string(buf), err
}

func (r *canonReader) readBool() (bool, error) {
	b, err := r.r.ReadByte()
	if err != nil {
		return false, err
	}
	return b == 1, nil
}
