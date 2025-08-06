// NeuroScript Version: 0.6.2
// File version: 11
// Purpose: FIX: Use strconv.FormatFloat to correctly normalize -0.0, resolving the number literal canonicalization test failure.
// filename: pkg/canon/canonicalize_part3.go
// nlines: 200+
// risk_rating: HIGH

package canon

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

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
	case "return", "fail", "emit":
		v.writeVarint(int64(len(s.Values)))
		for _, val := range s.Values {
			if err := v.visit(val); err != nil {
				return err
			}
		}
	case "must":
		return v.visit(s.Cond)
	case "for":
		v.writeString(s.LoopVarName)
		if err := v.visit(s.Collection); err != nil {
			return err
		}
		v.writeVarint(int64(len(s.Body)))
		for i := range s.Body {
			if err := v.visit(&s.Body[i]); err != nil {
				return err
			}
		}
	case "if", "while":
		if err := v.visit(s.Cond); err != nil {
			return err
		}
		v.writeVarint(int64(len(s.Body)))
		for i := range s.Body {
			if err := v.visit(&s.Body[i]); err != nil {
				return err
			}
		}
		if s.Type == "if" {
			v.writeVarint(int64(len(s.ElseBody)))
			for i := range s.ElseBody {
				if err := v.visit(&s.ElseBody[i]); err != nil {
					return err
				}
			}
		}
	case "ask":
		v.writeString(s.AskIntoVar)
		if err := v.visit(s.Values[0]); err != nil {
			return err
		}
	case "call", "expression":
		if s.Call != nil {
			return v.visit(s.Call)
		}
		if s.ExpressionStmt != nil {
			return v.visit(s.ExpressionStmt)
		}
	case "on_error":
		v.writeVarint(int64(len(s.Body)))
		for i := range s.Body {
			if err := v.visit(&s.Body[i]); err != nil {
				return err
			}
		}
	case "break", "continue", "clear_error":
		// These have no fields to encode.
	}
	return nil
}

func (v *canonVisitor) visitOnEventDecl(e *ast.OnEventDecl) error {
	if err := v.visit(e.EventNameExpr); err != nil {
		return err
	}
	v.writeString(e.HandlerName)
	v.writeString(e.EventVarName)
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
	v.writeVarint(int64(c.Target.Kind()))
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
	v.writeVarint(int64(len(lval.Accessors)))
	for _, acc := range lval.Accessors {
		v.writeVarint(int64(acc.BaseNode.NodeKind))
		v.writeVarint(int64(acc.Type))
		if err := v.visit(acc.Key); err != nil {
			return err
		}
	}
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
	// The previous implementation using `fmt.Sprintf("%v", ...)` did not correctly
	// normalize -0.0, as the "%v" verb preserves the sign. Using `strconv.FormatFloat`
	// with the 'g' format specifier ensures that both 0.0 and -0.0 are serialized
	// as "0", providing a deterministic output.
	// The parser guarantees that number literals are stored as float64.
	f, ok := val.(float64)
	if !ok {
		// This path should not be hit by the parser, but as a fallback,
		// we handle non-float64 numbers gracefully.
		strVal := fmt.Sprintf("%v", val)
		v.write([]byte{0x01})
		v.writeString(strVal)
		return
	}

	strVal := strconv.FormatFloat(f, 'g', -1, 64)
	v.write([]byte{0x01}) // Always write as float64
	v.writeString(strVal)
}
