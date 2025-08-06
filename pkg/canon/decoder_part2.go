// NeuroScript Version: 0.6.2
// File version: 35
// Purpose: FIX: Correctly sets the NodeKind for MapEntryNode during deserialization.
// filename: pkg/canon/decoder_part2.go
// nlines: 150
// risk_rating: HIGH

package canon

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// This file continues the implementation from part 1.

func (r *canonReader) readListLiteral() (*ast.ListLiteralNode, error) {
	numElements, err := r.readVarint()
	if err != nil {
		return nil, fmt.Errorf("failed to read list element count: %w", err)
	}
	l := &ast.ListLiteralNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindListLiteral},
	}
	if numElements > 0 {
		l.Elements = make([]ast.Expression, numElements)
		for i := 0; i < int(numElements); i++ {
			elem, err := r.readNode()
			if err != nil {
				return nil, fmt.Errorf("failed to read list element %d: %w", i, err)
			}
			l.Elements[i] = elem.(ast.Expression)
		}
	} else {
		l.Elements = []ast.Expression{}
	}
	return l, nil
}

func (r *canonReader) readElementAccess() (*ast.ElementAccessNode, error) {
	collection, err := r.readNode()
	if err != nil {
		return nil, err
	}
	accessor, err := r.readNode()
	if err != nil {
		return nil, err
	}
	return &ast.ElementAccessNode{
		BaseNode:   ast.BaseNode{NodeKind: types.KindElementAccess},
		Collection: collection.(ast.Expression),
		Accessor:   accessor.(ast.Expression),
	}, nil
}

func (r *canonReader) readSecretRef() (*ast.SecretRef, error) {
	path, err := r.readString()
	if err != nil {
		return nil, err
	}
	return &ast.SecretRef{BaseNode: ast.BaseNode{NodeKind: types.KindSecretRef}, Path: path}, nil
}

func (r *canonReader) readProgram() (*ast.Program, error) {
	prog := ast.NewProgram()
	numMeta, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numMeta > 0 {
		prog.Metadata = make(map[string]string)
		for i := 0; i < int(numMeta); i++ {
			key, err := r.readString()
			if err != nil {
				return nil, err
			}
			val, err := r.readString()
			if err != nil {
				return nil, err
			}
			prog.Metadata[key] = val
		}
	}

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
	if numEvents > 0 {
		prog.Events = make([]*ast.OnEventDecl, numEvents)
		for i := 0; i < int(numEvents); i++ {
			node, err := r.readNode()
			if err != nil {
				return nil, fmt.Errorf("failed to decode event %d: %w", i, err)
			}
			prog.Events[i] = node.(*ast.OnEventDecl)
		}
	} else {
		prog.Events = []*ast.OnEventDecl{}
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
	event.HandlerName, err = r.readString()
	if err != nil {
		return nil, err
	}
	event.EventVarName, err = r.readString()
	if err != nil {
		return nil, err
	}
	numSteps, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numSteps > 0 {
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
	kindVal, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	call.Target.Name = targetName
	call.Target.IsTool = isTool
	call.Target.BaseNode.NodeKind = types.Kind(kindVal)
	numArgs, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if numArgs > 0 {
		call.Arguments = make([]ast.Expression, numArgs)
		for i := 0; i < int(numArgs); i++ {
			argNode, err := r.readNode()
			if err != nil {
				return nil, err
			}
			call.Arguments[i] = argNode.(ast.Expression)
		}
	}
	return call, nil
}

func (r *canonReader) readMapEntry() (*ast.MapEntryNode, error) {
	keyNode, err := r.readNode()
	if err != nil {
		return nil, fmt.Errorf("failed to read map key: %w", err)
	}
	key, ok := keyNode.(*ast.StringLiteralNode)
	if !ok {
		return nil, fmt.Errorf("expected map key to be a string literal, but got %T", keyNode)
	}

	valueNode, err := r.readNode()
	if err != nil {
		return nil, fmt.Errorf("failed to read map value: %w", err)
	}
	value, ok := valueNode.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf("expected map value to be an expression, but got %T", valueNode)
	}

	return &ast.MapEntryNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindMapEntry},
		Key:      key,
		Value:    value,
	}, nil
}
