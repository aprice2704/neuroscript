// NeuroScript Version: 0.6.2
// File version: 2
// Purpose: Post-decode annealing to normalize CallTarget kinds and walk all nodes.
//          Now force step-level call targets to KindCallableExpr (no Unknown diffs).
// Filename: pkg/canon/decoder_part1_anneal.go
// Risk rating: LOW

package canon

import (
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func restoreCallTargetKinds(p *ast.Program) {
	if p == nil {
		return
	}
	procs := make(map[string]struct{}, len(p.Procedures))
	for name := range p.Procedures {
		procs[name] = struct{}{}
	}

	var visitExpr func(ast.Expression)
	var visitStep func(*ast.Step)
	var visitLValue func(*ast.LValueNode)

	setTargetKind := func(call *ast.CallableExprNode) {
		if call == nil {
			return
		}
		// Tools and decoded procedures MUST appear as CallableExpr targets.
		if call.Target.IsTool {
			call.Target.BaseNode.NodeKind = types.KindCallableExpr
			return
		}
		if _, ok := procs[call.Target.Name]; ok {
			call.Target.BaseNode.NodeKind = types.KindCallableExpr
			return
		}
		// Built-ins default to Unknown in expression context.
		call.Target.BaseNode.NodeKind = types.KindUnknown
	}

	visitExpr = func(e ast.Expression) {
		switch n := e.(type) {
		case *ast.CallableExprNode:
			setTargetKind(n)
			for _, a := range n.Arguments {
				visitExpr(a)
			}
		case *ast.BinaryOpNode:
			visitExpr(n.Left)
			visitExpr(n.Right)
		case *ast.UnaryOpNode:
			visitExpr(n.Operand)
		case *ast.ListLiteralNode:
			for _, el := range n.Elements {
				visitExpr(el)
			}
		case *ast.MapLiteralNode:
			for _, me := range n.Entries {
				if me != nil {
					visitExpr(me.Value)
				}
			}
		case *ast.ElementAccessNode:
			visitExpr(n.Collection)
			visitExpr(n.Accessor)
		case *ast.EvalNode:
			visitExpr(n.Argument)
		case *ast.TypeOfNode:
			visitExpr(n.Argument)
		}
	}

	visitLValue = func(lv *ast.LValueNode) {
		if lv == nil {
			return
		}
		for _, acc := range lv.Accessors {
			if acc != nil && acc.Key != nil {
				visitExpr(acc.Key)
			}
		}
	}

	visitStep = func(s *ast.Step) {
		if s == nil {
			return
		}
		switch s.Type {
		case "set", "emit", "return", "fail":
			for _, v := range s.Values {
				visitExpr(v)
			}
			for _, lv := range s.LValues {
				visitLValue(lv)
			}
		case "must":
			visitExpr(s.Cond)
		case "if", "while":
			visitExpr(s.Cond)
			for i := range s.Body {
				visitStep(&s.Body[i])
			}
			if s.Type == "if" {
				for i := range s.ElseBody {
					visitStep(&s.ElseBody[i])
				}
			}
		case "for":
			visitExpr(s.Collection)
			for i := range s.Body {
				visitStep(&s.Body[i])
			}
		case "call":
			// Step-level calls are always operational invocations: force CallableExpr.
			if s.Call != nil {
				s.Call.Target.BaseNode.NodeKind = types.KindCallableExpr
				for _, a := range s.Call.Arguments {
					visitExpr(a)
				}
			}
		case "on_error":
			for i := range s.Body {
				visitStep(&s.Body[i])
			}
		case "ask":
			if s.AskStmt != nil {
				visitExpr(s.AskStmt.AgentModelExpr)
				visitExpr(s.AskStmt.PromptExpr)
				if s.AskStmt.WithOptions != nil {
					visitExpr(s.AskStmt.WithOptions)
				}
				visitLValue(s.AskStmt.IntoTarget)
			}
		case "promptuser":
			if s.PromptUserStmt != nil {
				visitExpr(s.PromptUserStmt.PromptExpr)
				visitLValue(s.PromptUserStmt.IntoTarget)
			}
		}
	}

	for _, proc := range p.Procedures {
		for i := range proc.Steps {
			visitStep(&proc.Steps[i])
		}
	}
	for _, cmd := range p.Commands {
		for i := range cmd.Body {
			visitStep(&cmd.Body[i])
		}
	}
	for _, ev := range p.Events {
		for i := range ev.Body {
			visitStep(&ev.Body[i])
		}
	}
}
