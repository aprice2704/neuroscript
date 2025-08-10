// NeuroScript Version: 0.6.2
// File version: 9
// Purpose: FIX: Reverted 'call' case in readStep to use the main readNode dispatcher, ensuring consistent deserialization.
// filename: pkg/canon/decoder_part4.go
// nlines: 200+
// risk_rating: HIGH

package canon

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (r *canonReader) readStep() (*ast.Step, error) {
	stepType, err := r.readString()
	if err != nil {
		return nil, fmt.Errorf("failed to read step type: %w", err)
	}
	step := &ast.Step{BaseNode: ast.BaseNode{NodeKind: types.KindStep}, Type: stepType}

	switch stepType {
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
		numValues, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Values = make([]ast.Expression, numValues)
		for i := 0; i < int(numValues); i++ {
			node, err := r.readNode()
			if err != nil {
				return nil, err
			}
			step.Values[i] = node.(ast.Expression)
		}
	case "return", "fail", "emit":
		numValues, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Values = make([]ast.Expression, numValues)
		for i := 0; i < int(numValues); i++ {
			node, err := r.readNode()
			if err != nil {
				return nil, err
			}
			step.Values[i] = node.(ast.Expression)
		}
	case "must":
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		step.Cond = node.(ast.Expression)

	case "for":
		step.LoopVarName, err = r.readString()
		if err != nil {
			return nil, err
		}
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		step.Collection = node.(ast.Expression)

		numBody, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Body = make([]ast.Step, numBody)
		for i := 0; i < int(numBody); i++ {
			bNode, err := r.readNode()
			if err != nil {
				return nil, err
			}
			step.Body[i] = *bNode.(*ast.Step)
		}
	case "if", "while":
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		step.Cond = node.(ast.Expression)
		numBody, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Body = make([]ast.Step, numBody)
		for i := 0; i < int(numBody); i++ {
			bNode, err := r.readNode()
			if err != nil {
				return nil, err
			}
			step.Body[i] = *bNode.(*ast.Step)
		}
		if stepType == "if" {
			numElse, err := r.readVarint()
			if err != nil {
				return nil, err
			}
			if numElse > 0 {
				step.ElseBody = make([]ast.Step, numElse)
				for i := 0; i < int(numElse); i++ {
					eNode, err := r.readNode()
					if err != nil {
						return nil, err
					}
					step.ElseBody[i] = *eNode.(*ast.Step)
				}
			} else if numElse == 0 {
				step.ElseBody = []ast.Step{}
			}
		}
	case "ask":
		step.AskStmt, err = r.readAskStmt()
		if err != nil {
			return nil, err
		}
	case "promptuser":
		step.PromptUserStmt, err = r.readPromptUserStmt()
		if err != nil {
			return nil, err
		}
	case "call":
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		if call, ok := node.(*ast.CallableExprNode); ok {
			step.Call = call
		}
	case "on_error":
		numBody, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Body = make([]ast.Step, numBody)
		for i := 0; i < int(numBody); i++ {
			bNode, err := r.readNode()
			if err != nil {
				return nil, err
			}
			step.Body[i] = *bNode.(*ast.Step)
		}
	case "break", "continue", "clear_error":
		// These types have no fields to decode.
	}

	return step, nil
}

func (r *canonReader) readAskStmt() (*ast.AskStmt, error) {
	ask := &ast.AskStmt{BaseNode: ast.BaseNode{NodeKind: types.KindAskStmt}}
	var err error
	var node ast.Node

	node, err = r.readNode()
	if err != nil {
		return nil, err
	}
	ask.AgentModelExpr = node.(ast.Expression)

	node, err = r.readNode()
	if err != nil {
		return nil, err
	}
	ask.PromptExpr = node.(ast.Expression)

	hasWithOptions, err := r.readBool()
	if err != nil {
		return nil, err
	}
	if hasWithOptions {
		node, err = r.readNode()
		if err != nil {
			return nil, err
		}
		ask.WithOptions = node.(ast.Expression)
	}

	hasIntoTarget, err := r.readBool()
	if err != nil {
		return nil, err
	}
	if hasIntoTarget {
		node, err = r.readNode()
		if err != nil {
			return nil, err
		}
		ask.IntoTarget = node.(*ast.LValueNode)
	}

	return ask, nil
}

func (r *canonReader) readPromptUserStmt() (*ast.PromptUserStmt, error) {
	prompt := &ast.PromptUserStmt{BaseNode: ast.BaseNode{NodeKind: types.KindPromptUserStmt}}
	var err error
	var node ast.Node

	node, err = r.readNode()
	if err != nil {
		return nil, err
	}
	prompt.PromptExpr = node.(ast.Expression)

	node, err = r.readNode()
	if err != nil {
		return nil, err
	}
	prompt.IntoTarget = node.(*ast.LValueNode)

	return prompt, nil
}

func (r *canonReader) readLValue() (*ast.LValueNode, error) {
	identifier, err := r.readString()
	if err != nil {
		return nil, err
	}
	lval := &ast.LValueNode{BaseNode: ast.BaseNode{NodeKind: types.KindLValue}, Identifier: identifier}

	numAccessors, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numAccessors > 0 {
		lval.Accessors = make([]*ast.AccessorNode, numAccessors)
		for i := 0; i < int(numAccessors); i++ {
			kindVal, err := r.readVarint()
			if err != nil {
				return nil, fmt.Errorf("failed to read accessor kind: %w", err)
			}
			accType, err := r.readVarint()
			if err != nil {
				return nil, err
			}
			keyNode, err := r.readNode()
			if err != nil {
				return nil, err
			}
			lval.Accessors[i] = &ast.AccessorNode{
				BaseNode: ast.BaseNode{NodeKind: types.Kind(kindVal)},
				Type:     ast.AccessorType(accType),
				Key:      keyNode.(ast.Expression),
			}
		}
	} else if numAccessors == 0 {
		lval.Accessors = []*ast.AccessorNode{}
	}
	return lval, nil
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
func (r *canonReader) readNumber() (interface{}, error) {
	_, err := r.r.ReadByte() // Read and discard the type marker
	if err != nil {
		return nil, fmt.Errorf("failed to read number type marker: %w", err)
	}

	s, err := r.readString()
	if err != nil {
		return nil, fmt.Errorf("failed to read number string value: %w", err)
	}

	// Always parse as float64 to match parser behavior
	return strconv.ParseFloat(s, 64)
}
