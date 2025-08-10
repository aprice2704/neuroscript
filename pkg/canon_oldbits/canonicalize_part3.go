// NeuroScript Version: 0.6.2
// File version: 25
// Purpose: FIX: visitCallableExpr now ONLY writes the payload (starting with
//          the CE magic header), as the main dispatcher now handles writing
//          the node kind. This resolves the "bad header" serialization bug.
// Filename: pkg/canon/canonicalize_part3.go
// Risk rating: MEDIUM

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
		if s.AskStmt != nil {
			return v.visitAskStmt(s.AskStmt)
		}
	case "promptuser":
		if s.PromptUserStmt != nil {
			return v.visitPromptUserStmt(s.PromptUserStmt)
		}
	case "call":
		if s.Call != nil {
			return v.visit(s.Call)
		}
	case "on_error":
		v.writeVarint(int64(len(s.Body)))
		for i := range s.Body {
			if err := v.visit(&s.Body[i]); err != nil {
				return err
			}
		}
	case "break", "continue", "clear_error":
		// No fields to encode.
	}
	return nil
}

func (v *canonVisitor) visitAskStmt(a *ast.AskStmt) error {
	if err := v.visit(a.AgentModelExpr); err != nil {
		return err
	}
	if err := v.visit(a.PromptExpr); err != nil {
		return err
	}
	hasWithOptions := a.WithOptions != nil
	v.writeBool(hasWithOptions)
	if hasWithOptions {
		if err := v.visit(a.WithOptions); err != nil {
			return err
		}
	}
	hasIntoTarget := a.IntoTarget != nil
	v.writeBool(hasIntoTarget)
	if hasIntoTarget {
		if err := v.visit(a.IntoTarget); err != nil {
			return err
		}
	}
	return nil
}

func (v *canonVisitor) visitPromptUserStmt(p *ast.PromptUserStmt) error {
	if err := v.visit(p.PromptExpr); err != nil {
		return err
	}
	// IntoTarget is mandatory for promptuser
	if err := v.visit(p.IntoTarget); err != nil {
		return err
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

// visitCallableExpr writes ONLY the payload for a CallableExpr. The node kind
// is handled by the main visit() dispatcher.
func (v *canonVisitor) visitCallableExpr(c *ast.CallableExprNode) error {
	// DO NOT write the node kind here. The main dispatcher does that.

	// Header: "CE" + version + layout(header)
	v.write([]byte{CallMagic1, CallMagic2, CallWireVersion, CallLayoutHeader})

	// Payload: bool isTool, string name, argc, args...
	v.writeBool(c.Target.IsTool)
	v.writeString(c.Target.Name)

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
	// Normalize -0.0 to "0"
	f, ok := val.(float64)
	if !ok {
		strVal := fmt.Sprintf("%v", val)
		v.write([]byte{0x01})
		v.writeString(strVal)
		return
	}
	strVal := strconv.FormatFloat(f, 'g', -1, 64)
	v.write([]byte{0x01})
	v.writeString(strVal)
}
