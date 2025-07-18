// NeuroScript Version: 0.6.0
// File version: 13
// Purpose: Correctly reads the length prefix for 'return' statement values, fixing a deserialization bug.
// filename: pkg/canon/decoder.go
// nlines: 260
// risk_rating: HIGH

package canon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Decode reconstructs an AST Tree from its canonical binary representation.
func Decode(blob []byte) (*ast.Tree, error) {
	if len(blob) == 0 {
		return nil, fmt.Errorf("cannot decode an empty blob")
	}
	reader := &canonReader{r: bytes.NewReader(blob)}
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

type canonReader struct{ r *bytes.Reader }

func (r *canonReader) readNode() (ast.Node, error) {
	kindVal, err := r.readVarint()
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("failed to read node kind: %w", err)
	}
	kind := types.Kind(kindVal)

	switch kind {
	case types.KindProgram:
		return r.readProgram()
	case types.KindProcedureDecl:
		return r.readProcedure()
	case types.KindStep:
		return r.readStep()
	case types.KindCommandBlock:
		return r.readCommand()
	case types.KindOnEventDecl:
		return r.readOnEventDecl()
	case types.KindCallableExpr:
		return r.readCallableExpr()
	case types.KindVariable:
		name, err := r.readString()
		if err != nil {
			return nil, err
		}
		return &ast.VariableNode{BaseNode: ast.BaseNode{NodeKind: kind}, Name: name}, nil
	case types.KindStringLiteral:
		val, err := r.readString()
		if err != nil {
			return nil, err
		}
		return &ast.StringLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}, Value: val}, nil
	case types.KindNumberLiteral:
		strVal, err := r.readString()
		if err != nil {
			return nil, err
		}
		if i, err := strconv.ParseInt(strVal, 10, 64); err == nil {
			return &ast.NumberLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}, Value: i}, nil
		}
		f, _ := strconv.ParseFloat(strVal, 64)
		return &ast.NumberLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}, Value: f}, nil
	case types.KindBooleanLiteral:
		b, err := r.readBool()
		if err != nil {
			return nil, err
		}
		return &ast.BooleanLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}, Value: b}, nil
	case types.KindNilLiteral:
		return &ast.NilLiteralNode{BaseNode: ast.BaseNode{NodeKind: kind}}, nil
	case types.KindLValue:
		return r.readLValue()
	case types.KindBinaryOp:
		return r.readBinaryOp()
	case types.KindUnaryOp:
		return r.readUnaryOp()
	default:
		return nil, fmt.Errorf("unhandled node kind for decoding: %v (%d)", kind, kind)
	}
}

func (r *canonReader) readProgram() (*ast.Program, error) {
	prog := ast.NewProgram()
	numProcs, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(numProcs); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		proc := node.(*ast.Procedure)
		prog.Procedures[proc.Name()] = proc
	}

	numEvents, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(numEvents); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, fmt.Errorf("failed to decode event %d: %w", i, err)
		}
		prog.Events = append(prog.Events, node.(*ast.OnEventDecl))
	}

	numCommands, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(numCommands); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		prog.Commands = append(prog.Commands, node.(*ast.CommandNode))
	}
	return prog, nil
}

func (r *canonReader) readOnEventDecl() (*ast.OnEventDecl, error) {
	event := &ast.OnEventDecl{BaseNode: ast.BaseNode{NodeKind: types.KindOnEventDecl}}
	node, err := r.readNode()
	if err != nil {
		return nil, err
	}
	event.EventNameExpr = node.(ast.Expression)
	numSteps, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	event.Body = make([]ast.Step, numSteps)
	for i := 0; i < int(numSteps); i++ {
		sNode, err := r.readNode()
		if err != nil {
			return nil, err
		}
		if step, ok := sNode.(*ast.Step); ok {
			event.Body[i] = *step
		} else {
			return nil, fmt.Errorf("expected to decode a *ast.Step but got %T", sNode)
		}
	}
	return event, nil
}

func (r *canonReader) readCallableExpr() (*ast.CallableExprNode, error) {
	call := &ast.CallableExprNode{BaseNode: ast.BaseNode{NodeKind: types.KindCallableExpr}}
	targetName, err := r.readString()
	if err != nil {
		return nil, err
	}
	isTool, err := r.readBool()
	if err != nil {
		return nil, err
	}
	call.Target.Name = targetName
	call.Target.IsTool = isTool
	numArgs, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	call.Arguments = make([]ast.Expression, numArgs)
	for i := 0; i < int(numArgs); i++ {
		argNode, err := r.readNode()
		if err != nil {
			return nil, err
		}
		call.Arguments[i] = argNode.(ast.Expression)
	}
	return call, nil
}

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
		// **FIX:** Read the number of values before trying to decode them.
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
	case "call":
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		step.Call = node.(*ast.CallableExprNode)
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
