// NeuroScript Version: 0.6.2
// File version: 1
// Purpose: Decoder helpers for core program, event, and secret node types.
// Filename: pkg/canon/decoder_part2_core.go
// Risk rating: LOW

package canon

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func (r *canonReader) readSecretRef() (*ast.SecretRef, error) {
	path, err := r.readString()
	if err != nil {
		return nil, err
	}
	return &ast.SecretRef{
		BaseNode: ast.BaseNode{NodeKind: types.KindSecretRef},
		Path:     path,
	}, nil
}

func (r *canonReader) readProgram() (*ast.Program, error) {
	prog := ast.NewProgram()

	// Metadata
	nm, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if nm > 0 {
		prog.Metadata = make(map[string]string, nm)
		for i := 0; i < int(nm); i++ {
			k, err := r.readString()
			if err != nil {
				return nil, err
			}
			v, err := r.readString()
			if err != nil {
				return nil, err
			}
			prog.Metadata[k] = v
		}
	}

	// Procedures
	np, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(np); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		proc, ok := node.(*ast.Procedure)
		if !ok {
			return nil, fmt.Errorf("program: expected *ast.Procedure, got %T", node)
		}
		prog.Procedures[proc.Name()] = proc
	}

	// Events
	ne, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if ne > 0 {
		prog.Events = make([]*ast.OnEventDecl, ne)
		for i := 0; i < int(ne); i++ {
			node, err := r.readNode()
			if err != nil {
				return nil, fmt.Errorf("program: event[%d]: %w", i, err)
			}
			ev, ok := node.(*ast.OnEventDecl)
			if !ok {
				return nil, fmt.Errorf("program: event[%d]: expected *ast.OnEventDecl, got %T", i, node)
			}
			prog.Events[i] = ev
		}
	} else {
		prog.Events = []*ast.OnEventDecl{}
	}

	// Commands
	nc, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(nc); i++ {
		node, err := r.readNode()
		if err != nil {
			return nil, err
		}
		cmd, ok := node.(*ast.CommandNode)
		if !ok {
			return nil, fmt.Errorf("program: command[%d]: expected *ast.CommandNode, got %T", i, node)
		}
		prog.Commands = append(prog.Commands, cmd)
	}

	return prog, nil
}

func (r *canonReader) readOnEventDecl() (*ast.OnEventDecl, error) {
	ev := &ast.OnEventDecl{BaseNode: ast.BaseNode{NodeKind: types.KindOnEventDecl}}

	// Event name expression
	node, err := r.readNode()
	if err != nil {
		return nil, err
	}
	expr, ok := node.(ast.Expression)
	if !ok {
		return nil, fmt.Errorf("onevent: event name must be Expression, got %T", node)
	}
	ev.EventNameExpr = expr

	// Handler name
	ev.HandlerName, err = r.readString()
	if err != nil {
		return nil, err
	}
	// Optional event var name
	ev.EventVarName, err = r.readString()
	if err != nil {
		return nil, err
	}

	// Body
	ns, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if ns > 0 {
		ev.Body = make([]ast.Step, ns)
		for i := 0; i < int(ns); i++ {
			sn, err := r.readNode()
			if err != nil {
				return nil, err
			}
			st, ok := sn.(*ast.Step)
			if !ok {
				return nil, fmt.Errorf("onevent: step[%d]: expected *ast.Step, got %T", i, sn)
			}
			ev.Body[i] = *st
		}
	}

	return ev, nil
}
