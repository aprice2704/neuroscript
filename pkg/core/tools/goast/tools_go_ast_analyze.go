// NeuroScript Version: 0.3.1
// File version: 0.0.2
// Fix type assertion for n.Obj.Decl.Pos().
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

// Define ToolImplementation
var toolGoGetNodeInfoImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name:        "GoGetNodeInfo",
		Description: "Finds the AST node at a specific position (offset or line/column) within an AST handle and returns information about it.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle for the AST (from GoParseFile)."},
			{Name: "position", Type: core.ArgTypeMap, Required: true, Description: "Map specifying position. Requires either {\"offset\": N} or {\"line\": L, \"column\": C}. Line/Column are 1-based."},
		},
		ReturnType: core.ArgTypeMap, // Returns map describing the node, or nil if not found/error
	},
	Func: toolGoGetNodeInfo,
}

// --- New Tool: GoGetNodeInfo ---

// Helper to convert line/column to offset within a token.File
func getOffsetFromLineCol(tf *token.File, line, col int) (token.Pos, error) {
	if line <= 0 || col <= 0 {
		return token.NoPos, fmt.Errorf("line (%d) and column (%d) must be positive 1-based integers", line, col)
	}
	lineCount := tf.LineCount()
	if line > lineCount {
		return token.NoPos, fmt.Errorf("line %d is beyond file line count %d", line, lineCount)
	}
	lineStartPos := tf.LineStart(line)
	if !lineStartPos.IsValid() {
		return token.NoPos, fmt.Errorf("could not determine start position for valid line %d", line)
	}
	lineOffset := col - 1
	fileBaseOffset := tf.Base()
	// Calculate offset relative to the start of the *file* (base)
	targetOffset := int(lineStartPos) - fileBaseOffset + lineOffset

	if targetOffset < 0 || targetOffset >= tf.Size() {
		return token.NoPos, fmt.Errorf("calculated offset %d for line %d, column %d is out of file bounds (size %d)", targetOffset, line, col, tf.Size())
	}
	// Return the absolute position (token.Pos) based on the file's base offset
	return tf.Pos(targetOffset), nil
}

// Helper to format position details
func formatPos(fset *token.FileSet, pos token.Pos) map[string]interface{} {
	if !pos.IsValid() {
		return nil
	}
	p := fset.Position(pos)
	return map[string]interface{}{
		"filename": p.Filename,
		"line":     int64(p.Line),
		"column":   int64(p.Column),
		"offset":   int64(p.Offset),
	}
}

// toolGoGetNodeInfo finds the AST node at a given position and returns info about it.
func toolGoGetNodeInfo(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "GoGetNodeInfo"
	logger := interpreter.Logger()

	// --- Argument Validation ---
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: %s requires 2 arguments (handle, position)", core.ErrInvalidArgument, toolName)
	}
	handleID, okH := args[0].(string)
	positionArg, okP := args[1].(map[string]interface{})
	if !okH || !okP {
		return nil, fmt.Errorf("%w: %s invalid argument types (expected string, map)", core.ErrInvalidArgument, toolName)
	}
	if handleID == "" {
		return nil, fmt.Errorf("%w: %s handle cannot be empty", core.ErrInvalidArgument, toolName)
	}

	// --- Retrieve AST ---
	obj, err := interpreter.GetHandleValue(handleID, golangASTTypeTag)
	if err != nil {
		return nil, fmt.Errorf("%s get handle failed: %w", toolName, err)
	}
	cachedAst, ok := obj.(CachedAst)
	if !ok || cachedAst.File == nil || cachedAst.Fset == nil {
		return nil, fmt.Errorf("%w: %s handle '%s' contains invalid AST data", core.ErrHandleInvalid, toolName, handleID)
	}
	fset := cachedAst.Fset
	rootNode := cachedAst.File
	tf := fset.File(rootNode.Pos())
	if tf == nil {
		return nil, fmt.Errorf("%w: %s could not get token.File for AST handle '%s'", core.ErrInternalTool, toolName, handleID)
	}
	logger.Debug("GoGetNodeInfo: Retrieved AST", "handle", handleID, "filename", tf.Name())

	// --- Determine target position (token.Pos) ---
	var targetPos token.Pos
	var posErr error
	if offsetVal, okOffset := positionArg["offset"]; okOffset {
		offset, okConv := core.ConvertToIntE(offsetVal)
		if !okConv || offset < 0 {
			return nil, fmt.Errorf("%w: %s requires non-negative integer 'offset', got %v (%T)", core.ErrInvalidArgument, toolName, offsetVal, offsetVal)
		}
		if offset >= tf.Size() {
			return nil, fmt.Errorf("%w: %s offset %d is out of file bounds (size %d)", core.ErrInvalidArgument, toolName, offset, tf.Size())
		}
		targetPos = tf.Pos(offset)
		logger.Debug("GoGetNodeInfo: Using offset", "offset", offset, "targetPos", targetPos)
	} else if lineVal, okLine := positionArg["line"]; okLine {
		colVal, okCol := positionArg["column"]
		if !okCol {
			return nil, fmt.Errorf("%w: %s position map requires 'column' if 'line' is provided", core.ErrInvalidArgument, toolName)
		}
		line, okL := core.ConvertToIntE(lineVal)
		col, okC := core.ConvertToIntE(colVal)
		if !okL || !okC {
			return nil, fmt.Errorf("%w: %s requires integer 'line' and 'column', got line=%v (%T), col=%v (%T)", core.ErrInvalidArgument, toolName, lineVal, lineVal, colVal, colVal)
		}
		targetPos, posErr = getOffsetFromLineCol(tf, line, col)
		if posErr != nil {
			return nil, fmt.Errorf("%w: %s %w", core.ErrInvalidArgument, toolName, posErr)
		}
		logger.Debug("GoGetNodeInfo: Using line/col", "line", line, "col", col, "targetPos", targetPos)
	} else {
		return nil, fmt.Errorf("%w: %s position map requires 'offset' or ('line' and 'column')", core.ErrInvalidArgument, toolName)
	}
	if !targetPos.IsValid() {
		return nil, fmt.Errorf("%w: %s could not determine valid target position from input: %v", core.ErrInvalidArgument, toolName, positionArg)
	}

	// --- Find Node ---
	pathNodes, exact := astutil.PathEnclosingInterval(rootNode, targetPos, targetPos)
	if len(pathNodes) == 0 {
		logger.Warn("GoGetNodeInfo: No AST node found at position", "pos", targetPos, "filename", tf.Name())
		return nil, fmt.Errorf("%w: no AST node found at the specified position", core.ErrNotFound)
	}
	node := pathNodes[len(pathNodes)-1]

	// --- Extract Information ---
	nodeType := reflect.TypeOf(node).String()
	nodeType = strings.TrimPrefix(nodeType, "*ast.")
	var buf bytes.Buffer
	printErr := printer.Fprint(&buf, fset, node)
	if printErr != nil {
		logger.Warn("Failed to print node source text", "error", printErr)
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
			details["obj_kind"] = n.Obj.Kind.String()
			// --- FIXED: Check if Decl implements ast.Node ---
			if n.Obj.Decl != nil {
				if declNode, ok := n.Obj.Decl.(ast.Node); ok { // Type assertion
					if declPos := declNode.Pos(); declPos.IsValid() {
						details["decl_pos"] = formatPos(fset, declPos) // Use helper
					} else {
						logger.Debug("Declaration position is invalid", "ident", n.Name)
					}
				} else {
					logger.Debug("Declaration object does not implement ast.Node", "ident", n.Name, "declType", fmt.Sprintf("%T", n.Obj.Decl))
				}
			}
			// --- End Fix ---
		}
	case *ast.BasicLit:
		nodeValue = n.Value
		details["kind"] = n.Kind.String()
	case *ast.SelectorExpr:
		var xText string
		if xIdent, ok := n.X.(*ast.Ident); ok {
			xText = xIdent.Name
		} else {
			var xBuf bytes.Buffer
			_ = printer.Fprint(&xBuf, fset, n.X)
			xText = xBuf.String()
		}
		selText := n.Sel.Name
		details["selector_x"] = xText
		details["selector_sel"] = selText
		nodeValue = xText + "." + selText
	case *ast.CallExpr:
		var funcBuf bytes.Buffer
		_ = printer.Fprint(&funcBuf, fset, n.Fun)
		details["function"] = funcBuf.String()
		details["arg_count"] = len(n.Args)
		if funIdent, ok := n.Fun.(*ast.Ident); ok {
			nodeValue = funIdent.Name
		} else if funSel, ok := n.Fun.(*ast.SelectorExpr); ok {
			nodeValue = funSel.Sel.Name
		}
	case *ast.FuncDecl:
		if n.Recv != nil && len(n.Recv.List) > 0 && len(n.Recv.List[0].Names) > 0 {
			details["receiver_name"] = n.Recv.List[0].Names[0].Name
		}
		if n.Name != nil {
			details["name"] = n.Name.Name
			nodeValue = n.Name.Name
			details["is_method"] = n.Recv != nil
		}
	default:
		logger.Debug("GoGetNodeInfo: Unhandled node type for detailed info", "type", nodeType)
	}

	// --- Construct Result ---
	resultMap := map[string]interface{}{
		"node_type": nodeType, "node_text": nodeText, "node_value": nodeValue,
		"pos_start": posStartMap, "pos_end": posEndMap, "filename": tf.Name(),
		"exact_match": exact, "details": details,
	}
	logger.Debug("GoGetNodeInfo: Returning node info", "type", nodeType, "text_len", len(nodeText))
	return resultMap, nil
}
