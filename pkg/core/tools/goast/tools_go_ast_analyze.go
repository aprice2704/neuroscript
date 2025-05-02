// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 22:35:18 PDT // Use ConvertToIntE for position map
// filename: pkg/core/tools/goast/tools_go_ast_analyze.go
package goast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"reflect"
	"strings" // Need strings for TrimPrefix

	"github.com/aprice2704/neuroscript/pkg/core" // Need core import for error types etc.
	"golang.org/x/tools/go/ast/astutil"          // For PathEnclosingInterval
)

// --- New Tool: GoGetNodeInfo ---

// Helper to convert line/column to offset within a token.File
// --- (getOffsetFromLineCol remains unchanged) ---
func getOffsetFromLineCol(tf *token.File, line, col int) (token.Pos, error) {
	if line <= 0 || col <= 0 {
		return token.NoPos, fmt.Errorf("line (%d) and column (%d) must be positive 1-based integers", line, col)
	}
	if line > tf.LineCount() {
		return token.NoPos, fmt.Errorf("line %d is beyond file line count %d", line, tf.LineCount())
	}
	lineStartPos := tf.LineStart(line)
	if !lineStartPos.IsValid() {
		return token.NoPos, fmt.Errorf("could not determine start position for line %d", line)
	}
	offset := tf.Position(lineStartPos).Offset + (col - 1)
	if offset < 0 || offset >= tf.Size() {
		return token.NoPos, fmt.Errorf("calculated offset %d for line %d, column %d is out of file bounds (size %d)", offset, line, col, tf.Size())
	}
	return tf.Pos(offset), nil
}

// Helper to format position details
// --- (formatPos remains unchanged) ---
func formatPos(fset *token.FileSet, pos token.Pos) map[string]interface{} {
	if !pos.IsValid() {
		return nil
	}
	p := fset.Position(pos)
	return map[string]interface{}{"filename": p.Filename, "line": p.Line, "column": p.Column, "offset": p.Offset}
}

// toolGoGetNodeInfo finds the AST node at a given position and returns info about it.
func toolGoGetNodeInfo(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "GoGetNodeInfo"
	// Validation layer should ensure 2 args: string, map
	handleID := args[0].(string)
	positionArg, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: %s expected position map, got %T", core.ErrValidationTypeMismatch, toolName, args[1])
	}

	// Retrieve AST
	obj, err := interpreter.GetHandleValue(handleID, golangASTTypeTag)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", toolName, err)
	}
	cachedAst, ok := obj.(CachedAst)
	if !ok || cachedAst.File == nil || cachedAst.Fset == nil {
		return nil, fmt.Errorf("%w: %s handle '%s' contains invalid data", core.ErrHandleInvalid, toolName, handleID)
	}
	fset := cachedAst.Fset
	rootNode := cachedAst.File
	tf := fset.File(rootNode.Pos())
	if tf == nil {
		return nil, fmt.Errorf("%w: could not get token.File for AST handle '%s'", core.ErrInternalTool, handleID)
	}

	// Determine target position (token.Pos) from input map
	var targetPos token.Pos
	if offsetVal, okOffset := positionArg["offset"]; okOffset {
		// *** Use new ConvertToIntE helper ***
		offset, okConv := core.ConvertToIntE(offsetVal)
		if !okConv || offset < 0 {
			return nil, fmt.Errorf("%w: %s requires non-negative integer 'offset', got %v", core.ErrValidationArgValue, toolName, offsetVal)
		}
		if offset >= tf.Size() {
			return nil, fmt.Errorf("%w: %s offset %d is out of file bounds (size %d)", core.ErrValidationArgValue, toolName, offset, tf.Size())
		}
		targetPos = tf.Pos(offset)
	} else if lineVal, okLine := positionArg["line"]; okLine {
		if colVal, okCol := positionArg["column"]; okCol {
			// *** Use new ConvertToIntE helper ***
			line, okL := core.ConvertToIntE(lineVal)
			col, okC := core.ConvertToIntE(colVal)
			if !okL || !okC {
				return nil, fmt.Errorf("%w: %s requires integer 'line' and 'column', got line=%v, col=%v", core.ErrValidationArgValue, toolName, lineVal, colVal)
			}
			var posErr error
			targetPos, posErr = getOffsetFromLineCol(tf, line, col)
			if posErr != nil {
				return nil, fmt.Errorf("%w: %s %w", core.ErrValidationArgValue, toolName, posErr)
			}
		} else {
			return nil, fmt.Errorf("%w: %s position map requires 'column' if 'line' is provided", core.ErrValidationArgValue, toolName)
		}
	} else {
		return nil, fmt.Errorf("%w: %s position map requires 'offset' or ('line' and 'column')", core.ErrValidationArgValue, toolName)
	}

	if !targetPos.IsValid() {
		return nil, fmt.Errorf("%w: %s could not determine valid target position from input: %v", core.ErrValidationArgValue, toolName, positionArg)
	}

	// Find the node(s) enclosing the position
	path, exact := astutil.PathEnclosingInterval(rootNode, targetPos, targetPos)
	if len(path) == 0 {
		return nil, fmt.Errorf("%w: no AST node found at the specified position", core.ErrNotFound)
	}
	node := path[len(path)-1] // Innermost node

	// Extract information
	nodeType := reflect.TypeOf(node).String()
	nodeType = strings.TrimPrefix(nodeType, "*")
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, node)
	if err != nil {
		interpreter.Logger().Warn("Failed to print node source text", "error", err)
	}
	nodeText := buf.String()
	posStartMap := formatPos(fset, node.Pos())
	posEndMap := formatPos(fset, node.End())
	nodeValue := interface{}(nil)
	details := make(map[string]interface{})

	switch n := node.(type) {
	case *ast.Ident:
		nodeValue = n.Name
		if n.Obj != nil {
			details["declaration_kind"] = n.Obj.Kind.String()
		}
	case *ast.BasicLit:
		nodeValue = n.Value
		details["kind"] = n.Kind.String()
	case *ast.SelectorExpr:
		if xIdent, ok := n.X.(*ast.Ident); ok {
			details["package"] = xIdent.Name
			details["selector"] = n.Sel.Name
			nodeValue = details["package"].(string) + "." + details["selector"].(string)
		} else {
			details["selector"] = n.Sel.Name
		}
	case *ast.CallExpr:
		var funcBuf bytes.Buffer
		_ = printer.Fprint(&funcBuf, fset, n.Fun)
		details["function"] = funcBuf.String()
		details["arg_count"] = len(n.Args)
	case *ast.FuncDecl:
		if n.Recv != nil && len(n.Recv.List) > 0 {
			var recvBuf bytes.Buffer
			_ = printer.Fprint(&recvBuf, fset, n.Recv.List[0].Type)
			details["receiver_type"] = recvBuf.String()
		}
		if n.Name != nil {
			details["name"] = n.Name.Name
			nodeValue = n.Name.Name
		}
	}

	resultMap := map[string]interface{}{
		"id": int(node.Pos()), "type": nodeType, "text": nodeText, "value": nodeValue,
		"pos_start": posStartMap, "pos_end": posEndMap, "filename": tf.Name(),
		"exact_match": exact, "details": details,
	}
	return resultMap, nil
}
