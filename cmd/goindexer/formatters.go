// NeuroScript Go Indexer - Formatters
// File version: 1.0.0 // Initial robust version using go/printer
// Purpose: Provides functions to format AST nodes into strings, especially for types.
// filename: cmd/goindexer/formatters.go
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"log"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/goindex" // For goindex.ParamDetail
)

// formatNode converts an AST node (expected to be an expression, typically a type)
// into its string representation using go/printer.
func formatNode(fset *token.FileSet, node ast.Node) string {
	if node == nil {
		return ""
	}
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}
	err := cfg.Fprint(&buf, fset, node)
	if err != nil {
		// Log the error and return a fallback representation
		log.Printf("Error formatting node (%T): %v. Fallback string: %s", node, err, fmt.Sprintf("ERR_FORMATTING_NODE_%T", node))
		return fmt.Sprintf("ERR_FORMATTING_NODE_%T", node)
	}
	return buf.String()
}

// formatFieldList converts an AST FieldList (like parameters or results) to a slice of ParamDetail.
func formatFieldList(fset *token.FileSet, list *ast.FieldList) []goindex.ParamDetail {
	var details []goindex.ParamDetail
	if list == nil {
		return details
	}
	for _, field := range list.List {
		typeName := formatNode(fset, field.Type)
		if len(field.Names) > 0 { // Named parameters/results
			for _, name := range field.Names {
				if name.Name == "_" { // Skip blank identifiers as parameter names
					details = append(details, goindex.ParamDetail{Type: typeName})
				} else {
					details = append(details, goindex.ParamDetail{Name: name.Name, Type: typeName})
				}
			}
		} else { // Unnamed parameter/result (e.g., return types, or params in interface methods)
			details = append(details, goindex.ParamDetail{Type: typeName})
		}
	}
	return details
}

// formatReceiver formats the receiver of a method.
// Returns (receiverVarName string, receiverTypeString string) e.g. ("r", "*MyType") or ("", "MyType")
func formatReceiver(fset *token.FileSet, field *ast.Field) (name string, typeString string) {
	if field == nil {
		return "", ""
	}
	typeString = formatNode(fset, field.Type) // This should give "MyType" or "*MyType" etc.
	if len(field.Names) > 0 && field.Names[0] != nil {
		name = field.Names[0].Name
	}
	return name, typeString
}

// formatReceiverName extracts just the name (identifier) of the receiver variable.
// This might be less used if formatReceiver provides both name and type.
func formatReceiverName(field *ast.Field) string {
	if field != nil && len(field.Names) > 0 && field.Names[0] != nil {
		return field.Names[0].Name
	}
	return ""
}

// getBaseTypeName extracts the base type name from a potentially complex type string
// (e.g., "*pkg.MyType" -> "MyType", "[]some.OtherType" -> "OtherType", "map[string]foo.Bar" -> "Bar").
// This is a utility primarily for finding the "simple" name of a type, often used for embedded fields.
func getBaseTypeName(typeString string) string {
	name := typeString

	// Remove common prefixes and suffixes that are not part of the core type name.
	// Order can matter here.
	name = strings.TrimPrefix(name, "*")
	name = strings.TrimPrefix(name, "[]")
	name = strings.TrimPrefix(name, "...") // Variadic

	// For maps, try to get the value type's base name if complex
	if strings.HasPrefix(name, "map[") && strings.Contains(name, "]") {
		idx := strings.Index(name, "]")
		if idx != -1 && idx+1 < len(name) {
			name = name[idx+1:] // Get the part after "map[keyType]"
			// Recursively clean this part too
			name = getBaseTypeName(name)
		}
	}

	// For channels
	if strings.HasPrefix(name, "chan ") {
		name = strings.TrimPrefix(name, "chan ")
		name = getBaseTypeName(name)
	} else if strings.HasPrefix(name, "<-chan ") {
		name = strings.TrimPrefix(name, "<-chan ")
		name = getBaseTypeName(name)
	}

	// After stripping common prefixes, get the last part of a qualified name.
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return name
}
