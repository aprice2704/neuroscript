// NeuroScript Version: 0.7.2
// File version: 3
// Purpose: Ensures Metadata is initialized to a non-nil, empty map during decoding.
// filename: pkg/canon/codec_program.go
// nlines: 80
// risk_rating: LOW

package canon

import (
	"sort"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func encodeProgram(v *canonVisitor, n ast.Node) error {
	node := n.(*ast.Program)

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

	// Encode Procedures
	v.writeVarint(int64(len(node.Procedures)))
	procNames := make([]string, 0, len(node.Procedures))
	for name := range node.Procedures {
		procNames = append(procNames, name)
	}
	sort.Strings(procNames)
	for _, name := range procNames {
		if err := v.visitor(node.Procedures[name]); err != nil {
			return err
		}
	}

	// Encode Events
	v.writeVarint(int64(len(node.Events)))
	for _, event := range node.Events {
		if err := v.visitor(event); err != nil {
			return err
		}
	}

	// Encode Commands
	v.writeVarint(int64(len(node.Commands)))
	for _, cmd := range node.Commands {
		if err := v.visitor(cmd); err != nil {
			return err
		}
	}
	return nil
}

func decodeProgram(r *canonReader) (ast.Node, error) {
	prog := ast.NewProgram()
	// FIX: Unconditionally initialize Metadata to guarantee it's not nil.
	prog.Metadata = make(map[string]string)

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
			prog.Metadata[key] = val
		}
	}

	// Decode Procedures
	procCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(procCount); i++ {
		node, err := r.visitor()
		if err != nil {
			return nil, err
		}
		proc := node.(*ast.Procedure)
		prog.Procedures[proc.Name()] = proc
	}

	// Decode Events
	eventCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	prog.Events = make([]*ast.OnEventDecl, eventCount)
	for i := 0; i < int(eventCount); i++ {
		node, err := r.visitor()
		if err != nil {
			return nil, err
		}
		prog.Events[i] = node.(*ast.OnEventDecl)
	}

	// Decode Commands
	cmdCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	prog.Commands = make([]*ast.CommandNode, cmdCount)
	for i := 0; i < int(cmdCount); i++ {
		node, err := r.visitor()
		if err != nil {
			return nil, err
		}
		prog.Commands[i] = node.(*ast.CommandNode)
	}

	return prog, nil
}
