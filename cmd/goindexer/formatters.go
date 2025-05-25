// NeuroScript Go Indexer - Formatters
// File version: 1.0.0
// Purpose: Provides helper functions for formatting AST nodes into strings or structured data.
// filename: cmd/goindexer/formatters.go
package main

import (
	"go/ast"
	"go/printer"
	"go/token"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/goindex" // Ensure this import path is correct for your project
)

// formatNode converts an AST node to its string representation.
func formatNode(fset *token.FileSet, node ast.Node) string {
	if node == nil {
		return ""
	}
	var buf strings.Builder
	err := printer.Fprint(&buf, fset, node)
	if err != nil {
		// Consider logging this error if it occurs, though it's rare with valid AST.
		return "error_formatting_node"
	}
	return buf.String()
}

// formatFieldList converts an *ast.FieldList (used for parameters or results)
// into a slice of goindex.ParamDetail.
func formatFieldList(fset *token.FileSet, list *ast.FieldList) []goindex.ParamDetail {
	var params []goindex.ParamDetail
	if list == nil {
		return params
	}
	for _, field := range list.List {
		typeName := formatNode(fset, field.Type)
		if len(field.Names) > 0 {
			// Named parameters/results
			for _, name := range field.Names {
				params = append(params, goindex.ParamDetail{Name: name.Name, Type: typeName})
			}
		} else {
			// Unnamed parameter/result (e.g., func(int, string) or interface methods)
			params = append(params, goindex.ParamDetail{Name: "", Type: typeName})
		}
	}
	return params
}

// formatReceiver extracts the receiver's variable name and type string.
// It now takes an *ast.Field, which is how receivers are represented.
func formatReceiver(fset *token.FileSet, field *ast.Field) (nameStr string, typeStr string) {
	if field == nil {
		return "", ""
	}
	typeStr = formatNode(fset, field.Type)
	if len(field.Names) > 0 && field.Names[0] != nil {
		nameStr = field.Names[0].Name
	}
	return nameStr, typeStr
}
