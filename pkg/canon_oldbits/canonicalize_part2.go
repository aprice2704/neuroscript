// NeuroScript Version: 0.6.2
// File version: 41
// Purpose: FIX: Restored local valueToNode helper to resolve import cycle.
// filename: pkg/canon/canonicalize_part2.go
// nlines: 200+
// risk_rating: HIGH

package canon

import (
	"fmt"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// valueToNode converts a runtime lang.Value into a serializable ast.Node.
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
	// Note: Complex types like lists and maps would require recursive conversion.
	default:
		return nil, fmt.Errorf("unsupported lang.Value type for AST conversion: %T", v)
	}
}

// This file continues the implementation from part 1.

func (v *canonVisitor) visitListLiteral(l *ast.ListLiteralNode) error {
	v.writeVarint(int64(len(l.Elements)))
	for _, elem := range l.Elements {
		if err := v.visit(elem); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitElementAccess(e *ast.ElementAccessNode) error {
	if err := v.visit(e.Collection); err != nil {
		return err
	}
	return v.visit(e.Accessor)
}

func (v *canonVisitor) visitSecretRef(s *ast.SecretRef) error {
	v.writeString(s.Path)
	return nil
}

func (v *canonVisitor) visitUnaryOp(u *ast.UnaryOpNode) error {
	v.writeString(u.Operator)
	return v.visit(u.Operand)
}

func (v *canonVisitor) visitProgram(p *ast.Program) error {
	v.writeVarint(int64(len(p.Metadata)))
	keys := make([]string, 0, len(p.Metadata))
	for k := range p.Metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v.writeString(k)
		v.writeString(p.Metadata[k])
	}

	procNames := make([]string, 0, len(p.Procedures))
	for name := range p.Procedures {
		procNames = append(procNames, name)
	}
	sort.Strings(procNames)
	v.writeVarint(int64(len(procNames)))
	for _, name := range procNames {
		if err := v.visit(p.Procedures[name]); err != nil {
			return err
		}
	}

	v.writeVarint(int64(len(p.Events)))
	for _, event := range p.Events {
		if err := v.visit(event); err != nil {
			return err
		}
	}

	v.writeVarint(int64(len(p.Commands)))
	for _, command := range p.Commands {
		if err := v.visit(command); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitCommand(c *ast.CommandNode) error {
	v.writeVarint(int64(len(c.Metadata)))
	keys := make([]string, 0, len(c.Metadata))
	for k := range c.Metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v.writeString(k)
		v.writeString(c.Metadata[k])
	}

	v.writeVarint(int64(len(c.Body)))
	for i := range c.Body {
		if err := v.visit(&c.Body[i]); err != nil {
			return err
		}
	}

	v.writeVarint(int64(len(c.ErrorHandlers)))
	for _, handler := range c.ErrorHandlers {
		if err := v.visit(handler); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitProcedure(p *ast.Procedure) error {
	v.writeString(p.Name())
	v.writeVarint(int64(len(p.Metadata)))
	keys := make([]string, 0, len(p.Metadata))
	for k := range p.Metadata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v.writeString(k)
		v.writeString(p.Metadata[k])
	}

	v.writeVarint(int64(len(p.RequiredParams)))
	for _, param := range p.RequiredParams {
		v.writeString(param)
	}
	v.writeVarint(int64(len(p.OptionalParams)))
	for _, param := range p.OptionalParams {
		v.writeString(param.Name)
		v.writeVarint(int64(param.BaseNode.NodeKind))

		var nodeToVisit ast.Node
		if param.Default != nil {
			var err error
			nodeToVisit, err = valueToNode(param.Default)
			if err != nil {
				return fmt.Errorf("could not convert default parameter value to AST node: %w", err)
			}
		}

		if err := v.visit(nodeToVisit); err != nil {
			return fmt.Errorf("failed to visit default param: %w", err)
		}
	}
	v.writeVarint(int64(len(p.ReturnVarNames)))
	for _, name := range p.ReturnVarNames {
		v.writeString(name)
	}

	v.writeVarint(int64(len(p.ErrorHandlers)))
	for _, handler := range p.ErrorHandlers {
		if err := v.visit(handler); err != nil {
			return err
		}
	}

	v.writeVarint(int64(len(p.Steps)))
	for i := range p.Steps {
		if err := v.visit(&p.Steps[i]); err != nil {
			return err
		}
	}
	return nil
}
