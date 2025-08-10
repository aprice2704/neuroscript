// NeuroScript Version: 0.6.3
// File version: 2
// Purpose: Provides a post-decode "annealing" pass to normalize CallTarget kinds across the entire AST.
// filename: pkg/canon/codec_anneal.go
// nlines: 120
// risk_rating: MEDIUM

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

	setTargetKind := func(call *ast.CallableExprNode) {
		if call == nil {
			return
		}
		if call.Target.IsTool {
			call.Target.BaseNode.NodeKind = types.KindCallableExpr
			return
		}
		if _, ok := procs[call.Target.Name]; ok {
			call.Target.BaseNode.NodeKind = types.KindCallableExpr
			return
		}
		// Default to Unknown for built-ins or unresolved functions
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
		case "must", "if", "while":
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
			if s.Call != nil {
				// A step-level call is an operational invocation, force its kind.
				s.Call.Target.BaseNode.NodeKind = types.KindCallableExpr
				for _, a := range s.Call.Arguments {
					visitExpr(a)
				}
			}
		case "ask":
			if s.AskStmt != nil {
				visitExpr(s.AskStmt.AgentModelExpr)
				visitExpr(s.AskStmt.PromptExpr)
				if s.AskStmt.WithOptions != nil {
					visitExpr(s.AskStmt.WithOptions)
				}
			}
		case "promptuser":
			if s.PromptUserStmt != nil {
				visitExpr(s.PromptUserStmt.PromptExpr)
			}
		}
	}

	for _, proc := range p.Procedures {
		for _, opt := range proc.OptionalParams {
			if opt.Default != nil {
				// This is a bit tricky, we need to get the AST node for the default value
				// but for now, we assume it doesn't contain call expressions.
			}
		}
		for i := range proc.Steps {
			visitStep(&proc.Steps[i])
		}
		for _, handler := range proc.ErrorHandlers {
			visitStep(handler)
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
