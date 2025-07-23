// NeuroScript Version: 0.6.0
// File version: 20.2
// Purpose: Adds a versioned magic number header for integrity checks and fixes MapEntryNode serialization.
// filename: pkg/canon/canonicalize_part2.go
// nlines: 130+
// risk_rating: HIGH

package canon

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

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
	// Note: Enc and Raw are runtime values and not part of the canonical form.
	return nil
}

func (v *canonVisitor) visitUnaryOp(u *ast.UnaryOpNode) error {
	// Handle negative numbers as a special case for more compact representation
	if u.Operator == "-" {
		if num, ok := u.Operand.(*ast.NumberLiteralNode); ok {
			v.writeVarint(int64(types.KindNumberLiteral))
			switch val := num.Value.(type) {
			case int64:
				v.writeNumber(-val)
			case float64:
				v.writeNumber(-val)
			}
			return nil
		}
	}
	v.writeString(u.Operator)
	return v.visit(u.Operand)
}

func (v *canonVisitor) visitProgram(p *ast.Program) error {
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
	v.writeVarint(int64(len(c.Body)))
	for i := range c.Body {
		if err := v.visit(&c.Body[i]); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitProcedure(p *ast.Procedure) error {
	v.writeString(p.Name())
	v.writeVarint(int64(len(p.Steps)))
	for i := range p.Steps {
		if err := v.visit(&p.Steps[i]); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitStep(s *ast.Step) error {
	v.writeString(s.Type)
	switch s.Type {
	case "set":
		v.writeVarint(int64(len(s.LValues)))
		for _, lval := range s.LValues {
			if err := v.visit(lval); err != nil {
				return err
			}
		}
		v.writeVarint(int64(len(s.Values)))
		for _, rval := range s.Values {
			if err := v.visit(rval); err != nil {
				return err
			}
		}
	case "return":
		v.writeVarint(int64(len(s.Values)))
		for _, val := range s.Values {
			if err := v.visit(val); err != nil {
				return err
			}
		}
	case "emit":
		return v.visit(s.Values[0])
	case "call":
		if s.Call != nil {
			return v.visit(s.Call)
		}
		// Handle expression statements that might be wrapped in a step
		if s.ExpressionStmt != nil {
			return v.visit(s.ExpressionStmt)
		}

	case "expression":
		if s.ExpressionStmt != nil {
			return v.visit(s.ExpressionStmt)
		}
	}
	return nil
}

func (v *canonVisitor) visitOnEventDecl(e *ast.OnEventDecl) error {
	if err := v.visit(e.EventNameExpr); err != nil {
		return err
	}
	v.writeVarint(int64(len(e.Body)))
	for i := range e.Body {
		if err := v.visit(&e.Body[i]); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitCallableExpr(c *ast.CallableExprNode) error {
	v.writeString(c.Target.Name)
	v.writeBool(c.Target.IsTool)
	v.writeVarint(int64(len(c.Arguments)))
	for _, arg := range c.Arguments {
		if err := v.visit(arg); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitLValue(lval *ast.LValueNode) error {
	v.writeString(lval.Identifier)
	// Accessors are part of the LValue's structure but are handled by ElementAccessNode.
	// We do not serialize them here to avoid duplication.
	return nil
}

func (v *canonVisitor) visitBinaryOp(b *ast.BinaryOpNode) error {
	v.writeString(b.Operator)
	if err := v.visit(b.Left); err != nil {
		return err
	}
	return v.visit(b.Right)
}

// --- Primitive Writers ---

func (v *canonVisitor) write(p []byte) {
	v.w.Write(p)
	v.hasher.Write(p)
}

func (v *canonVisitor) writeVarint(x int64) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, x)
	v.write(buf[:n])
}

func (v *canonVisitor) writeString(s string) {
	v.writeVarint(int64(len(s)))
	if len(s) > 0 {
		v.write([]byte(s))
	}
}

func (v *canonVisitor) writeBool(b bool) {
	if b {
		v.write([]byte{1})
	} else {
		v.write([]byte{0})
	}
}

func (v *canonVisitor) writeNumber(val interface{}) {
	// Normalize -0.0 to 0.0 for deterministic output
	if f, ok := val.(float64); ok && f == 0 && math.Signbit(f) {
		val = 0.0
	}
	strVal := fmt.Sprintf("%v", val)
	v.writeString(strVal)
}
