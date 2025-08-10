// NeuroScript Version: 0.6.3
// File version: 6
// Purpose: Implements a comprehensive encoder/decoder for the ProcedureDecl AST node, fixing nil default param handling.
// filename: pkg/canon/codec_procedure.go
// nlines: 120
// risk_rating: HIGH

package canon

import (
	"sort"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func encodeProcedure(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.Procedure)
	v.writeString(node.Name())

	// Metadata
	v.writeVarint(int64(len(node.Metadata)))
	keys := make([]string, 0, len(node.Metadata))
	for k := range node.Metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v.writeString(k)
		v.writeString(node.Metadata[k])
	}

	// Params
	v.writeVarint(int64(len(node.RequiredParams)))
	for _, param := range node.RequiredParams {
		v.writeString(param)
	}
	v.writeVarint(int64(len(node.OptionalParams)))
	for _, param := range node.OptionalParams {
		v.writeString(param.Name)
		defaultValueNode, _ := valueToNode(param.Default)
		if err := v.visitor(defaultValueNode); err != nil {
			return err
		}
	}
	v.writeVarint(int64(len(node.ReturnVarNames)))
	for _, name := range node.ReturnVarNames {
		v.writeString(name)
	}

	// Error Handlers
	v.writeVarint(int64(len(node.ErrorHandlers)))
	for _, handler := range node.ErrorHandlers {
		if err := v.visitor(handler); err != nil {
			return err
		}
	}

	// Steps
	v.writeVarint(int64(len(node.Steps)))
	for i := range node.Steps {
		if err := v.visitor(&node.Steps[i]); err != nil {
			return err
		}
	}
	return nil
}

func decodeProcedure(r *canonReader) (ast.Node, error) {
	proc := &ast.Procedure{BaseNode: ast.BaseNode{NodeKind: types.KindProcedureDecl}}
	name, err := r.readString()
	if err != nil {
		return nil, err
	}
	proc.SetName(name)

	// Metadata
	metaCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if metaCount > 0 {
		proc.Metadata = make(map[string]string, metaCount)
		for i := 0; i < int(metaCount); i++ {
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

	// Params
	reqCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	proc.RequiredParams = make([]string, reqCount)
	for i := 0; i < int(reqCount); i++ {
		proc.RequiredParams[i], err = r.readString()
		if err != nil {
			return nil, err
		}
	}
	optCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	proc.OptionalParams = make([]*ast.ParamSpec, optCount)
	for i := 0; i < int(optCount); i++ {
		paramName, err := r.readString()
		if err != nil {
			return nil, err
		}
		defaultNode, err := r.visitor()
		if err != nil {
			return nil, err
		}

		var defaultVal lang.Value
		// FIX: If the decoded node is a NilLiteral, the default value should be nil.
		if _, ok := defaultNode.(*ast.NilLiteralNode); !ok {
			defaultVal, _ = nodeToValue(defaultNode)
		}

		proc.OptionalParams[i] = &ast.ParamSpec{
			BaseNode: ast.BaseNode{NodeKind: types.KindVariable},
			Name:     paramName,
			Default:  defaultVal,
		}
	}
	retCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	proc.ReturnVarNames = make([]string, retCount)
	for i := 0; i < int(retCount); i++ {
		proc.ReturnVarNames[i], err = r.readString()
		if err != nil {
			return nil, err
		}
	}

	// Error Handlers
	handlerCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	proc.ErrorHandlers = make([]*ast.Step, handlerCount)
	for i := 0; i < int(handlerCount); i++ {
		node, err := r.visitor()
		if err != nil {
			return nil, err
		}
		proc.ErrorHandlers[i] = node.(*ast.Step)
	}

	// Steps
	stepCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	proc.Steps = make([]ast.Step, stepCount)
	for i := 0; i < int(stepCount); i++ {
		node, err := r.visitor()
		if err != nil {
			return nil, err
		}
		proc.Steps[i] = *node.(*ast.Step)
	}
	return proc, nil
}

// valueToNode and nodeToValue are helpers to convert between runtime values and AST nodes for default params.
func valueToNode(val lang.Value) (ast.Node, error) {
	if val == nil {
		return &ast.NilLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindNilLiteral}}, nil
	}
	switch v := val.(type) {
	case lang.StringValue:
		return &ast.StringLiteralNode{Value: v.Value}, nil
	case lang.NumberValue:
		return &ast.NumberLiteralNode{Value: v.Value}, nil
	case lang.BoolValue:
		return &ast.BooleanLiteralNode{Value: v.Value}, nil
	case lang.NilValue:
		return &ast.NilLiteralNode{}, nil
	default:
		return &ast.NilLiteralNode{}, nil // Fallback for complex types
	}
}

func nodeToValue(node ast.Node) (lang.Value, error) {
	if node == nil {
		return lang.NilValue{}, nil
	}
	switch n := node.(type) {
	case *ast.StringLiteralNode:
		return lang.StringValue{Value: n.Value}, nil
	case *ast.NumberLiteralNode:
		return lang.NumberValue{Value: n.Value.(float64)}, nil
	case *ast.BooleanLiteralNode:
		return lang.BoolValue{Value: n.Value}, nil
	case *ast.NilLiteralNode:
		return lang.NilValue{}, nil
	default:
		return lang.NilValue{}, nil
	}
}
