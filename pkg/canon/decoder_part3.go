// NeuroScript Version: 0.6.2
// File version: 39
// Purpose: FIX: Restored local nodeToValue helper to resolve import cycle.
// filename: pkg/canon/decoder_part3.go
// nlines: 200+
// risk_rating: HIGH

package canon

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// nodeToValue converts a deserialized ast.Node back into a runtime lang.Value.
func nodeToValue(node ast.Node) (lang.Value, error) {
	if node == nil {
		return lang.NilValue{}, nil
	}
	switch n := node.(type) {
	case *ast.StringLiteralNode:
		return lang.StringValue{Value: n.Value}, nil
	case *ast.NumberLiteralNode:
		// The value is already float64 from the decoder
		return lang.NumberValue{Value: n.Value.(float64)}, nil
	case *ast.BooleanLiteralNode:
		return lang.BoolValue{Value: n.Value}, nil
	case *ast.NilLiteralNode:
		return lang.NilValue{}, nil
	// Note: Complex types would require recursive conversion.
	default:
		return nil, fmt.Errorf("unsupported ast.Node type for lang.Value conversion: %T", n)
	}
}

// This file continues the implementation from part 2.

func (r *canonReader) readCommand() (*ast.CommandNode, error) {
	cmd := &ast.CommandNode{BaseNode: ast.BaseNode{NodeKind: types.KindCommandBlock}}
	numMeta, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numMeta > 0 {
		cmd.Metadata = make(map[string]string)
		for i := 0; i < int(numMeta); i++ {
			key, err := r.readString()
			if err != nil {
				return nil, err
			}
			val, err := r.readString()
			if err != nil {
				return nil, err
			}
			cmd.Metadata[key] = val
		}
	}

	numSteps, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numSteps > 0 {
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
	}

	// FIX: Handle nil vs empty slice for ErrorHandlers
	numHandlers, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numHandlers > 0 {
		cmd.ErrorHandlers = make([]*ast.Step, numHandlers)
		for i := 0; i < int(numHandlers); i++ {
			node, err := r.readNode()
			if err != nil {
				return nil, err
			}
			if step, ok := node.(*ast.Step); ok {
				cmd.ErrorHandlers[i] = step
			} else {
				return nil, fmt.Errorf("expected to decode a *ast.Step for an error handler but got %T", node)
			}
		}
	} else if numHandlers == 0 {
		cmd.ErrorHandlers = []*ast.Step{}
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

	numMeta, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numMeta > 0 {
		proc.Metadata = make(map[string]string)
		for i := 0; i < int(numMeta); i++ {
			key, err := r.readString()
			if err != nil {
				return nil, err
			}
			val, err := r.readString()
			if err != nil {
				return nil, err
			}
			proc.Metadata[key] = val
		}
	}
	numRequired, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numRequired > 0 {
		proc.RequiredParams = make([]string, numRequired)
		for i := 0; i < int(numRequired); i++ {
			proc.RequiredParams[i], err = r.readString()
			if err != nil {
				return nil, err
			}
		}
	}
	numOptional, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numOptional > 0 {
		proc.OptionalParams = make([]*ast.ParamSpec, numOptional)
		for i := 0; i < int(numOptional); i++ {
			paramName, err := r.readString()
			if err != nil {
				return nil, err
			}
			kindVal, err := r.readVarint()
			if err != nil {
				return nil, fmt.Errorf("failed to read optional param kind: %w", err)
			}
			defaultNode, err := r.readNode()
			if err != nil {
				return nil, fmt.Errorf("failed to read optional param default value: %w", err)
			}

			var defaultVal lang.Value
			if defaultNode.Kind() != types.KindNilLiteral {
				defaultVal, err = nodeToValue(defaultNode)
				if err != nil {
					return nil, fmt.Errorf("could not convert decoded node to value: %w", err)
				}
			}

			proc.OptionalParams[i] = &ast.ParamSpec{
				Name:     paramName,
				BaseNode: ast.BaseNode{NodeKind: types.Kind(kindVal)},
				Default:  defaultVal,
			}
		}
	}
	numReturn, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numReturn > 0 {
		proc.ReturnVarNames = make([]string, numReturn)
		for i := 0; i < int(numReturn); i++ {
			proc.ReturnVarNames[i], err = r.readString()
			if err != nil {
				return nil, err
			}
		}
	}
	numHandlers, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numHandlers > 0 {
		proc.ErrorHandlers = make([]*ast.Step, numHandlers)
		for i := 0; i < int(numHandlers); i++ {
			node, err := r.readNode()
			if err != nil {
				return nil, err
			}
			if step, ok := node.(*ast.Step); ok {
				proc.ErrorHandlers[i] = step
			} else {
				return nil, fmt.Errorf("expected to decode a *ast.Step for an error handler but got %T", node)
			}
		}
	}

	numSteps, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numSteps > 0 {
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
	}
	return proc, nil
}
