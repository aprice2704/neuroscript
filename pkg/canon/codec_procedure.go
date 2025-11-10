// NeuroScript Version: 0.7.2
// File version: 13
// Purpose: Exports ValueToNode and NodeToValue helpers so they can be re-exported by pkg/api.
// filename: pkg/canon/codec_procedure.go
// nlines: 200+

package canon

import (
	"fmt"
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

	// --- FIX: Serialize missing fields ---
	// Comments
	v.writeVarint(int64(len(node.Comments)))
	for _, comment := range node.Comments {
		if err := v.visitor(comment); err != nil {
			return err
		}
	}
	// BlankLinesBefore
	v.writeVarint(int64(node.BlankLinesBefore))
	// --- END FIX ---

	// Params
	v.writeVarint(int64(len(node.RequiredParams)))
	for _, param := range node.RequiredParams {
		v.writeString(param)
	}
	v.writeVarint(int64(len(node.OptionalParams)))
	for _, param := range node.OptionalParams {
		v.writeString(param.Name)
		// --- FIX: Call exported ValueToNode ---
		defaultValueNode, _ := ValueToNode(param.Default)
		if err := v.visitor(defaultValueNode); err != nil {
			return err
		}
	}

	// --- FIX: Serialize missing Variadic fields ---
	v.writeBool(node.Variadic)
	v.writeString(node.VariadicParamName)
	// --- END FIX ---

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
	proc := &ast.Procedure{
		BaseNode:       ast.BaseNode{NodeKind: types.KindProcedureDecl},
		Metadata:       make(map[string]string),
		RequiredParams: make([]string, 0),
		OptionalParams: make([]*ast.ParamSpec, 0),
		ErrorHandlers:  make([]*ast.Step, 0),
		Comments:       make([]*ast.Comment, 0), // Init empty
	}
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

	// --- FIX: Deserialize missing fields ---
	// Comments
	commentCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if commentCount > 0 {
		proc.Comments = make([]*ast.Comment, commentCount)
		for i := 0; i < int(commentCount); i++ {
			node, err := r.visitor()
			if err != nil {
				return nil, err
			}
			proc.Comments[i] = node.(*ast.Comment)
		}
	}

	// BlankLinesBefore
	blankLines, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	proc.BlankLinesBefore = int(blankLines)
	// --- END FIX ---

	// Params
	reqCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if reqCount > 0 {
		proc.RequiredParams = make([]string, reqCount)
		for i := 0; i < int(reqCount); i++ {
			proc.RequiredParams[i], err = r.readString()
			if err != nil {
				return nil, err
			}
		}
	}

	optCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if optCount > 0 {
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
			// --- FIX: Call exported NodeToValue ---
			defaultVal, err = NodeToValue(defaultNode)
			if err != nil {
				return nil, fmt.Errorf("could not decode param %s: %w", paramName, err)
			}

			proc.OptionalParams[i] = &ast.ParamSpec{
				BaseNode: ast.BaseNode{NodeKind: types.KindVariable},
				Name:     paramName,
				Default:  defaultVal,
			}
		}
	}

	// --- FIX: Deserialize missing Variadic fields ---
	proc.Variadic, err = r.readBool()
	if err != nil {
		return nil, err
	}
	proc.VariadicParamName, err = r.readString()
	if err != nil {
		return nil, err
	}
	// --- END FIX ---

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
	if handlerCount > 0 {
		proc.ErrorHandlers = make([]*ast.Step, handlerCount)
		for i := 0; i < int(handlerCount); i++ {
			node, err := r.visitor()
			if err != nil {
				return nil, err
			}
			proc.ErrorHandlers[i] = node.(*ast.Step)
		}
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

// --- FIX: Rename to ValueToNode (exported) ---
// ValueToNode and NodeToValue are helpers to convert between runtime values and AST nodes for default params.
func ValueToNode(val lang.Value) (ast.Node, error) {
	if val == nil {
		return &ast.NilLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindNilLiteral}}, nil
	}
	switch v := val.(type) {
	case lang.StringValue:
		return &ast.StringLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindStringLiteral}, Value: v.Value}, nil
	case lang.NumberValue:
		return &ast.NumberLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindNumberLiteral}, Value: v.Value}, nil
	case lang.BoolValue:
		return &ast.BooleanLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindBooleanLiteral}, Value: v.Value}, nil
	case lang.NilValue:
		return &ast.NilLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindNilLiteral}}, nil
	case lang.ListValue:
		listNode := &ast.ListLiteralNode{
			BaseNode: ast.BaseNode{NodeKind: types.KindListLiteral},
			Elements: make([]ast.Expression, len(v.Value)),
		}
		for i, elemVal := range v.Value {
			elemNode, err := ValueToNode(elemVal) // <<< FIX: Recursive call
			if err != nil {
				return nil, err
			}
			listNode.Elements[i] = elemNode.(ast.Expression)
		}
		return listNode, nil
	case lang.MapValue:
		mapNode := &ast.MapLiteralNode{
			BaseNode: ast.BaseNode{NodeKind: types.KindMapLiteral},
			Entries:  make([]*ast.MapEntryNode, 0, len(v.Value)),
		}
		for key, val := range v.Value {
			keyNode := &ast.StringLiteralNode{BaseNode: ast.BaseNode{NodeKind: types.KindStringLiteral}, Value: key}
			valNode, err := ValueToNode(val) // <<< FIX: Recursive call
			if err != nil {
				return nil, err
			}
			mapNode.Entries = append(mapNode.Entries, &ast.MapEntryNode{
				BaseNode: ast.BaseNode{NodeKind: types.KindMapEntry},
				Key:      keyNode,
				Value:    valNode.(ast.Expression),
			})
		}
		return mapNode, nil
	default:
		return nil, fmt.Errorf("ValueToNode: unsupported lang.Value type %T", v)
	}
}

// --- FIX: Rename to NodeToValue (exported) ---
func NodeToValue(node ast.Node) (lang.Value, error) {
	if node == nil {
		return lang.NilValue{}, nil
	}
	switch n := node.(type) {
	case *ast.StringLiteralNode:
		return lang.StringValue{Value: n.Value}, nil
	case *ast.NumberLiteralNode:
		num, ok := n.Value.(float64)
		if !ok {
			return nil, fmt.Errorf("NodeToValue: NumberLiteralNode value is not float64, but %T", n.Value)
		}
		return lang.NumberValue{Value: num}, nil
	case *ast.BooleanLiteralNode:
		return lang.BoolValue{Value: n.Value}, nil
	case *ast.NilLiteralNode:
		return lang.NilValue{}, nil
	case *ast.ListLiteralNode:
		listVal := lang.ListValue{Value: make([]lang.Value, len(n.Elements))}
		for i, elemNode := range n.Elements {
			elemVal, err := NodeToValue(elemNode) // <<< FIX: Recursive call
			if err != nil {
				return nil, err
			}
			listVal.Value[i] = elemVal
		}
		return listVal, nil
	case *ast.MapLiteralNode:
		mapVal := lang.MapValue{Value: make(map[string]lang.Value, len(n.Entries))}
		for _, entry := range n.Entries {
			key := entry.Key.Value
			val, err := NodeToValue(entry.Value) // <<< FIX: Recursive call
			if err != nil {
				return nil, err
			}
			mapVal.Value[key] = val
		}
		return mapVal, nil
	default:
		return nil, fmt.Errorf("NodeToValue: unsupported ast.Node type %T", n)
	}
}
