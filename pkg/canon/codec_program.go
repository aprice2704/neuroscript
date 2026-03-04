// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 6
// :: description: Updated decodeProgram with safe type assertions to prevent panics.
// :: latestChange: Replaced raw type assertions with safe checks.
// :: filename: pkg/canon/codec_program.go
// :: serialization: go

package canon

import (
	"fmt"
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

	v.writeVarint(int64(len(node.Comments)))
	for _, comment := range node.Comments {
		if err := v.visitor(comment); err != nil {
			return err
		}
	}

	return nil
}

func decodeProgram(r *canonReader) (ast.Node, error) {
	prog := ast.NewProgram()
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
		proc, ok := node.(*ast.Procedure)
		if !ok {
			return nil, fmt.Errorf("expected *ast.Procedure in program, got %T", node)
		}
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
		eventNode, ok := node.(*ast.OnEventDecl)
		if !ok {
			return nil, fmt.Errorf("expected *ast.OnEventDecl in program, got %T", node)
		}
		prog.Events[i] = eventNode
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
		cmdNode, ok := node.(*ast.CommandNode)
		if !ok {
			return nil, fmt.Errorf("expected *ast.CommandNode in program, got %T", node)
		}
		prog.Commands[i] = cmdNode
	}

	// Decode Comments
	commentCount, err := r.readVarint()
	if err != nil {
		return nil, err
	}
	if commentCount > 0 {
		prog.Comments = make([]*ast.Comment, commentCount)
		for i := 0; i < int(commentCount); i++ {
			node, err := r.visitor()
			if err != nil {
				return nil, err
			}
			commentNode, ok := node.(*ast.Comment)
			if !ok {
				return nil, fmt.Errorf("expected *ast.Comment in program, got %T", node)
			}
			prog.Comments[i] = commentNode
		}
	}

	return prog, nil
}
