// NeuroScript Version: 0.7.2
// File version: 6
// Purpose: Adds wrapper functions to expose Ask/Prompt/Whisper statement codecs to the registry.
// filename: pkg/canon/codec_statements.go
// nlines: 200+
// risk_rating: MEDIUM

package canon

import (
	"sort"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Note: AskStmt, PromptUserStmt, and WhisperStmt are not standalone nodes, so their
// encoders/decoders are helpers called by the Step codec, not registered
// in the main codec registry.

func encodeAskStmt(v *canonVisitor, stmt *ast.AskStmt) error {
	// A nil check is crucial here.
	if stmt == nil {
		v.writeBool(false) // Write a single byte indicating nil
		return nil
	}
	v.writeBool(true) // Indicate that the statement is not nil

	if err := v.visitor(stmt.AgentModelExpr); err != nil {
		return err
	}
	if err := v.visitor(stmt.PromptExpr); err != nil {
		return err
	}

	hasWithOptions := stmt.WithOptions != nil
	v.writeBool(hasWithOptions)
	if hasWithOptions {
		if err := v.visitor(stmt.WithOptions); err != nil {
			return err
		}
	}

	hasIntoTarget := stmt.IntoTarget != nil
	v.writeBool(hasIntoTarget)
	if hasIntoTarget {
		if err := v.visitor(stmt.IntoTarget); err != nil {
			return err
		}
	}
	return nil
}

func decodeAskStmt(r *canonReader) (*ast.AskStmt, error) {
	isNotNil, err := r.readBool()
	if err != nil {
		return nil, err
	}
	if !isNotNil {
		return nil, nil
	}

	stmt := &ast.AskStmt{BaseNode: ast.BaseNode{NodeKind: types.KindAskStmt}}
	agentModel, err := r.visitor()
	if err != nil {
		return nil, err
	}
	stmt.AgentModelExpr = agentModel.(ast.Expression)

	prompt, err := r.visitor()
	if err != nil {
		return nil, err
	}
	stmt.PromptExpr = prompt.(ast.Expression)

	hasWithOptions, err := r.readBool()
	if err != nil {
		return nil, err
	}
	if hasWithOptions {
		withOptions, err := r.visitor()
		if err != nil {
			return nil, err
		}
		stmt.WithOptions = withOptions.(ast.Expression)
	}

	hasIntoTarget, err := r.readBool()
	if err != nil {
		return nil, err
	}
	if hasIntoTarget {
		intoTarget, err := r.visitor()
		if err != nil {
			return nil, err
		}
		stmt.IntoTarget = intoTarget.(*ast.LValueNode)
	}
	return stmt, nil
}

func encodePromptUserStmt(v *canonVisitor, stmt *ast.PromptUserStmt) error {
	if stmt == nil {
		v.writeBool(false)
		return nil
	}
	v.writeBool(true)

	if err := v.visitor(stmt.PromptExpr); err != nil {
		return err
	}
	if err := v.visitor(stmt.IntoTarget); err != nil {
		return err
	}
	return nil
}

func decodePromptUserStmt(r *canonReader) (*ast.PromptUserStmt, error) {
	isNotNil, err := r.readBool()
	if err != nil {
		return nil, err
	}
	if !isNotNil {
		return nil, nil
	}

	stmt := &ast.PromptUserStmt{BaseNode: ast.BaseNode{NodeKind: types.KindPromptUserStmt}}
	prompt, err := r.visitor()
	if err != nil {
		return nil, err
	}
	stmt.PromptExpr = prompt.(ast.Expression)

	intoTarget, err := r.visitor()
	if err != nil {
		return nil, err
	}
	stmt.IntoTarget = intoTarget.(*ast.LValueNode)
	return stmt, nil
}

func encodeWhisperStmt(v *canonVisitor, stmt *ast.WhisperStmt) error {
	if stmt == nil {
		v.writeBool(false)
		return nil
	}
	v.writeBool(true)
	if err := v.visitor(stmt.Handle); err != nil {
		return err
	}
	return v.visitor(stmt.Value)
}

func decodeWhisperStmt(r *canonReader) (*ast.WhisperStmt, error) {
	isNotNil, err := r.readBool()
	if err != nil {
		return nil, err
	}
	if !isNotNil {
		return nil, nil
	}
	stmt := &ast.WhisperStmt{BaseNode: ast.BaseNode{NodeKind: types.KindWhisperStmt}}
	handle, err := r.visitor()
	if err != nil {
		return nil, err
	}
	stmt.Handle = handle.(ast.Expression)
	value, err := r.visitor()
	if err != nil {
		return nil, err
	}
	stmt.Value = value.(ast.Expression)
	return stmt, nil
}

func encodeCommandBlock(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.CommandNode)
	// Encode Metadata
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

	v.writeVarint(int64(len(node.Body)))
	for i := range node.Body {
		if err := v.visitor(&node.Body[i]); err != nil {
			return err
		}
	}
	return nil
}

func decodeCommandBlock(r *canonReader) (ast.Node, error) {
	node := &ast.CommandNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindCommandBlock},
		Metadata: make(map[string]string),
	}
	// Decode Metadata
	metaCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if metaCount > 0 {
		for i := 0; i < int(metaCount); i++ {
			key, err := r.readString()
			if err != nil {
				return nil, err
			}
			val, err := r.readString()
			if err != nil {
				return nil, err
			}
			node.Metadata[key] = val
		}
	}

	bodyCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	node.Body = make([]ast.Step, bodyCount)
	for i := 0; i < int(bodyCount); i++ {
		step, err := r.visitor()
		if err != nil {
			return nil, err
		}
		node.Body[i] = *step.(*ast.Step)
	}
	return node, nil
}

func encodeOnEventDecl(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.OnEventDecl)

	// FIX: Add Metadata encoding.
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

	if err := v.visitor(node.EventNameExpr); err != nil {
		return err
	}
	v.writeString(node.HandlerName)
	v.writeString(node.EventVarName)
	v.writeVarint(int64(len(node.Body)))
	for i := range node.Body {
		if err := v.visitor(&node.Body[i]); err != nil {
			return err
		}
	}
	return nil
}

func decodeOnEventDecl(r *canonReader) (ast.Node, error) {
	node := &ast.OnEventDecl{
		BaseNode: ast.BaseNode{NodeKind: types.KindOnEventDecl},
		Metadata: make(map[string]string),
		Comments: make([]*ast.Comment, 0), // Comments are not yet serialized, but init for consistency.
	}

	// FIX: Add Metadata decoding.
	metaCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if metaCount > 0 {
		for i := 0; i < int(metaCount); i++ {
			key, err := r.readString()
			if err != nil {
				return nil, err
			}
			val, err := r.readString()
			if err != nil {
				return nil, err
			}
			node.Metadata[key] = val
		}
	}

	eventName, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.EventNameExpr = eventName.(ast.Expression)
	node.HandlerName, err = r.readString()
	if err != nil {
		return nil, err
	}
	node.EventVarName, err = r.readString()
	if err != nil {
		return nil, err
	}
	bodyCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	node.Body = make([]ast.Step, bodyCount)
	for i := 0; i < int(bodyCount); i++ {
		step, err := r.visitor()
		if err != nil {
			return nil, err
		}
		node.Body[i] = *step.(*ast.Step)
	}
	return node, nil
}

func encodeExpressionStmt(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.ExpressionStatementNode)
	return v.visitor(node.Expression)
}

func decodeExpressionStmt(r *canonReader) (ast.Node, error) {
	node := &ast.ExpressionStatementNode{BaseNode: ast.BaseNode{NodeKind: types.KindExpressionStmt}}
	expr, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.Expression = expr.(ast.Expression)
	return node, nil
}

func encodeTypeOf(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.TypeOfNode)
	return v.visitor(node.Argument)
}

func decodeTypeOf(r *canonReader) (ast.Node, error) {
	node := &ast.TypeOfNode{BaseNode: ast.BaseNode{NodeKind: types.KindTypeOfExpr}}
	arg, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.Argument = arg.(ast.Expression)
	return node, nil
}

func encodeEval(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.EvalNode)
	return v.visitor(node.Argument)
}

func decodeEval(r *canonReader) (ast.Node, error) {
	node := &ast.EvalNode{BaseNode: ast.BaseNode{NodeKind: types.KindEvalExpr}}
	arg, err := r.visitor()
	if err != nil {
		return nil, err
	}
	node.Argument = arg.(ast.Expression)
	return node, nil
}

func encodeLast(v *canonVisitor, n ast.Node) error {
	// No payload
	return nil
}

func decodeLast(r *canonReader) (ast.Node, error) {
	return &ast.LastNode{BaseNode: ast.BaseNode{NodeKind: types.KindLastResult}}, nil
}

func encodePlaceholder(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.PlaceholderNode)
	v.writeString(node.Name)
	return nil
}

func decodePlaceholder(r *canonReader) (ast.Node, error) {
	node := &ast.PlaceholderNode{BaseNode: ast.BaseNode{NodeKind: types.KindPlaceholder}}
	var err error
	node.Name, err = r.readString()
	if err != nil {
		return nil, err
	}
	return node, nil
}

func encodeSecretRef(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.SecretRef)
	v.writeString(node.Path)
	return nil
}

func decodeSecretRef(r *canonReader) (ast.Node, error) {
	node := &ast.SecretRef{BaseNode: ast.BaseNode{NodeKind: types.KindSecretRef}}
	var err error
	node.Path, err = r.readString()
	if err != nil {
		return nil, err
	}
	return node, nil
}

// --- START: Wrapper functions for registry ---

func encodeAskStmt_wrapper(v *canonVisitor, n ast.Node) error {
	return encodeAskStmt(v, n.(*ast.AskStmt))
}
func decodeAskStmt_wrapper(r *canonReader) (ast.Node, error) {
	return decodeAskStmt(r)
}

func encodePromptUserStmt_wrapper(v *canonVisitor, n ast.Node) error {
	return encodePromptUserStmt(v, n.(*ast.PromptUserStmt))
}
func decodePromptUserStmt_wrapper(r *canonReader) (ast.Node, error) {
	return decodePromptUserStmt(r)
}

func encodeWhisperStmt_wrapper(v *canonVisitor, n ast.Node) error {
	return encodeWhisperStmt(v, n.(*ast.WhisperStmt))
}
func decodeWhisperStmt_wrapper(r *canonReader) (ast.Node, error) {
	return decodeWhisperStmt(r)
}

// --- END: Wrapper functions for registry ---
