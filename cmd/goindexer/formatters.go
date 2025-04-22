package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strings" // Import strings
)

// formatNode uses go/printer to format an AST node into a string.
// This is suitable for printing expressions, types, etc., but not FieldLists directly.
func formatNode(fset *token.FileSet, node ast.Node) string {
	if node == nil {
		return ""
	}
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.RawFormat | printer.UseSpaces, Tabwidth: 8} // Use spaces for clarity
	err := cfg.Fprint(&buf, fset, node)
	if err != nil {
		// Don't log errors repetitively here, let caller handle if needed
		// log.Printf("Error formatting node %T: %v", node, err)
		// Fallback or return error indicator
		// Returning Go syntax representation might be too verbose/confusing
		return fmt.Sprintf("/* error formatting %T */", node)
	}
	return buf.String()
}

// formatFieldList manually builds a string for parameter or result lists.
func formatFieldList(fset *token.FileSet, list *ast.FieldList) string {
	if list == nil || len(list.List) == 0 {
		return "()" // Empty list representation
	}
	var parts []string
	for _, field := range list.List {
		typeStr := formatNode(fset, field.Type)
		if len(field.Names) > 0 {
			// Has names (e.g., "a, b int")
			var names []string
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
			parts = append(parts, fmt.Sprintf("%s %s", strings.Join(names, ", "), typeStr))
		} else {
			// No names (e.g., "error" in results, or unnamed params)
			parts = append(parts, typeStr)
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// formatSignature generates a string representation of a function signature (params + results).
// *** MODIFIED: Use formatFieldList instead of formatNode for Params/Results ***
func formatSignature(fset *token.FileSet, funcType *ast.FuncType) string {
	paramsStr := formatFieldList(fset, funcType.Params) // Use helper for params

	resultsStr := ""
	if funcType.Results != nil {
		resultsStr = formatFieldList(fset, funcType.Results) // Use helper for results
		// If resultsStr is "()", it means no return values, so omit it.
		// If it's "(type1)", format as " type1".
		// If it's "(type1, type2)", format as " (type1, type2)".
		if resultsStr == "()" {
			resultsStr = ""
		} else if !strings.Contains(resultsStr, ",") && len(funcType.Results.List) == 1 && len(funcType.Results.List[0].Names) == 0 {
			// Single, unnamed return value - omit parentheses
			resultsStr = " " + strings.Trim(resultsStr, "()")
		} else {
			resultsStr = " " + resultsStr // Keep parentheses for multiple or named returns
		}
	}

	return fmt.Sprintf("func%s%s", paramsStr, resultsStr)
}

// formatReceiver formats the receiver part of a method declaration.
// Using formatNode should be okay here as it prints the type expression.
func formatReceiver(fset *token.FileSet, fieldType ast.Expr) string {
	return formatNode(fset, fieldType)
}

// determineKind attempts to classify the type declaration.
func determineKind(typeSpec ast.Expr) string {
	switch typeSpec.(type) {
	case *ast.StructType:
		return "struct"
	case *ast.InterfaceType:
		return "interface"
	case *ast.Ident:
		// Could check if it's a builtin type maybe?
		return "alias/basic" // More descriptive
	case *ast.ArrayType:
		return "array"
	case *ast.MapType:
		return "map"
	case *ast.ChanType:
		return "chan"
	case *ast.FuncType:
		return "func"
	case *ast.SelectorExpr:
		return "external_type" // e.g. pkg.Type
	case *ast.StarExpr:
		return "pointer" // Pointer to another type
	default:
		// Use simpler fallback
		rawType := fmt.Sprintf("%T", typeSpec)
		cleanedType := strings.TrimPrefix(rawType, "*ast.")
		return strings.ToLower(cleanedType) // e.g., "sliceexpr" -> "slice" might need more work
	}
}
