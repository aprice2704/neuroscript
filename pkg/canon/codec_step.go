// NeuroScript Version: 0.7.2
// File version: 8
// Purpose: Enforces that ElseBody is initialized to a non-nil empty slice during decoding for consistency.
// filename: pkg/canon/codec_step.go
// nlines: 200+
// risk_rating: HIGH

package canon

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func encodeStep(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.Step)
	v.writeString(node.Type)

	switch node.Type {
	case "set":
		v.writeVarint(int64(len(node.LValues)))
		for _, lval := range node.LValues {
			if err := v.visitor(lval); err != nil {
				return err
			}
		}
		v.writeVarint(int64(len(node.Values)))
		for _, rval := range node.Values {
			if err := v.visitor(rval); err != nil {
				return err
			}
		}
	case "emit", "return", "fail":
		v.writeVarint(int64(len(node.Values)))
		for _, val := range node.Values {
			if err := v.visitor(val); err != nil {
				return err
			}
		}
	case "if", "while", "must":
		if err := v.visitor(node.Cond); err != nil {
			return err
		}
		if node.Type != "must" {
			v.writeVarint(int64(len(node.Body)))
			for i := range node.Body {
				if err := v.visitor(&node.Body[i]); err != nil {
					return err
				}
			}
		}
		if node.Type == "if" {
			v.writeVarint(int64(len(node.ElseBody)))
			for i := range node.ElseBody {
				if err := v.visitor(&node.ElseBody[i]); err != nil {
					return err
				}
			}
		}
	case "for":
		v.writeString(node.LoopVarName)
		if err := v.visitor(node.Collection); err != nil {
			return err
		}
		v.writeVarint(int64(len(node.Body)))
		for i := range node.Body {
			if err := v.visitor(&node.Body[i]); err != nil {
				return err
			}
		}
	case "call":
		if err := v.visitor(node.Call); err != nil {
			return err
		}
	case "on_error":
		v.writeVarint(int64(len(node.Body)))
		for i := range node.Body {
			if err := v.visitor(&node.Body[i]); err != nil {
				return err
			}
		}
	case "ask":
		// Normalize the AST on the fly. The parser puts the data in Values/LValues.
		stmt := node.AskStmt
		if stmt == nil && len(node.Values) > 0 {
			stmt = &ast.AskStmt{
				AgentModelExpr: node.Values[0],
				PromptExpr:     node.Values[1],
			}
			if len(node.LValues) > 0 {
				stmt.IntoTarget = node.LValues[0]
			}
		}
		return encodeAskStmt(v, stmt)
	case "promptuser":
		stmt := node.PromptUserStmt
		if stmt == nil && len(node.Values) > 0 {
			stmt = &ast.PromptUserStmt{
				PromptExpr: node.Values[0],
			}
			if len(node.LValues) > 0 {
				stmt.IntoTarget = node.LValues[0]
			}
		}
		return encodePromptUserStmt(v, stmt)
	case "whisper":
		return encodeWhisperStmt(v, node.WhisperStmt)
	}
	return nil
}

func decodeStep(r *canonReader) (ast.Node, error) {
	step := &ast.Step{BaseNode: ast.BaseNode{NodeKind: types.KindStep}}
	var err error
	step.Type, err = r.readString()
	if err != nil {
		return nil, err
	}

	switch step.Type {
	case "set":
		lvalCount, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.LValues = make([]*ast.LValueNode, lvalCount)
		for i := 0; i < int(lvalCount); i++ {
			node, err := r.visitor()
			if err != nil {
				return nil, err
			}
			step.LValues[i] = node.(*ast.LValueNode)
		}
		valCount, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Values = make([]ast.Expression, valCount)
		for i := 0; i < int(valCount); i++ {
			node, err := r.visitor()
			if err != nil {
				return nil, err
			}
			step.Values[i] = node.(ast.Expression)
		}
	case "emit", "return", "fail":
		valCount, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Values = make([]ast.Expression, valCount)
		for i := 0; i < int(valCount); i++ {
			node, err := r.visitor()
			if err != nil {
				return nil, err
			}
			step.Values[i] = node.(ast.Expression)
		}
	case "if", "while", "must":
		cond, err := r.visitor()
		if err != nil {
			return nil, err
		}
		step.Cond = cond.(ast.Expression)
		if step.Type != "must" {
			bodyCount, err := r.readVarint()
			if err != nil {
				return nil, err
			}
			step.Body = make([]ast.Step, bodyCount)
			for i := 0; i < int(bodyCount); i++ {
				node, err := r.visitor()
				if err != nil {
					return nil, err
				}
				step.Body[i] = *node.(*ast.Step)
			}
		}
		if step.Type == "if" {
			elseCount, err := r.readVarint()
			if err != nil {
				return nil, err
			}
			// FIX: Ensure ElseBody is a non-nil, empty slice if count is zero.
			step.ElseBody = make([]ast.Step, elseCount)
			for i := 0; i < int(elseCount); i++ {
				node, err := r.visitor()
				if err != nil {
					return nil, err
				}
				step.ElseBody[i] = *node.(*ast.Step)
			}
		}
	case "for":
		step.LoopVarName, err = r.readString()
		if err != nil {
			return nil, err
		}
		coll, err := r.visitor()
		if err != nil {
			return nil, err
		}
		step.Collection = coll.(ast.Expression)
		bodyCount, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Body = make([]ast.Step, bodyCount)
		for i := 0; i < int(bodyCount); i++ {
			node, err := r.visitor()
			if err != nil {
				return nil, err
			}
			step.Body[i] = *node.(*ast.Step)
		}
	case "call":
		call, err := r.visitor()
		if err != nil {
			return nil, err
		}
		step.Call = call.(*ast.CallableExprNode)
	case "on_error":
		bodyCount, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		step.Body = make([]ast.Step, bodyCount)
		for i := 0; i < int(bodyCount); i++ {
			node, err := r.visitor()
			if err != nil {
				return nil, err
			}
			step.Body[i] = *node.(*ast.Step)
		}
	case "ask":
		step.AskStmt, err = decodeAskStmt(r)
		if err != nil {
			return nil, err
		}
	case "promptuser":
		step.PromptUserStmt, err = decodePromptUserStmt(r)
		if err != nil {
			return nil, err
		}
	case "whisper":
		step.WhisperStmt, err = decodeWhisperStmt(r)
		if err != nil {
			return nil, err
		}
	}
	return step, nil
}

func encodeLValue(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.LValueNode)
	v.writeString(node.Identifier)
	v.writeVarint(int64(len(node.Accessors)))
	for _, acc := range node.Accessors {
		v.writeVarint(int64(acc.Type))
		if err := v.visitor(acc.Key); err != nil {
			return err
		}
	}
	return nil
}

func decodeLValue(r *canonReader) (ast.Node, error) {
	lval := &ast.LValueNode{BaseNode: ast.BaseNode{NodeKind: types.KindLValue}}
	var err error
	lval.Identifier, err = r.readString()
	if err != nil {
		return nil, err
	}
	accessorCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	lval.Accessors = make([]*ast.AccessorNode, accessorCount)
	for i := 0; i < int(accessorCount); i++ {
		accType, err := r.readVarint()
		if err != nil {
			return nil, err
		}
		key, err := r.visitor()
		if err != nil {
			return nil, err
		}
		lval.Accessors[i] = &ast.AccessorNode{
			BaseNode: ast.BaseNode{NodeKind: types.KindElementAccess}, // Accessors imply element access
			Type:     ast.AccessorType(accType),
			Key:      key.(ast.Expression),
		}
	}
	return lval, nil
}
