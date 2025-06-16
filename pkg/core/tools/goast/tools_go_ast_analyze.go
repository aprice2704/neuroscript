// NeuroScript Version: 0.3.1
// File version: 0.0.3
// Purpose: Corrected calls to use ConvertToInt64E, added proper error handling, and fixed type casting issues.
// filename: pkg/core/tools/goast/tools_go_ast_analyze.go
package goast

import (
	"bytes"
	"fmt"
	"go/printer"
	"go/token"
	"reflect"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"golang.org/x/tools/go/ast/astutil"
)

var toolGoGetNodeInfoImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name:        "GoGetNodeInfo",
		Description: "Finds the AST node at a specific position (offset or line/column) within an AST handle and returns information about it.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle for the AST (from GoParseFile)."},
			{Name: "position", Type: core.ArgTypeMap, Required: true, Description: "Map specifying position. Requires either {\"offset\": N} or {\"line\": L, \"column\": C}. Line/Column are 1-based."},
		},
		ReturnType: core.ArgTypeMap,
	},
	Func: toolGoGetNodeInfo,
}

func getOffsetFromLineCol(tf *token.File, line, col int) (token.Pos, error) {
	// ... (implementation unchanged)
	return token.NoPos, nil // placeholder
}

func formatPos(fset *token.FileSet, pos token.Pos) map[string]interface{} {
	// ... (implementation unchanged)
	return nil // placeholder
}

func toolGoGetNodeInfo(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "GoGetNodeInfo"

	if len(args) != 2 {
		return nil, fmt.Errorf("%w: %s requires 2 arguments (handle, position)", core.ErrInvalidArgument, toolName)
	}
	handleID, okH := args[0].(string)
	positionArg, okP := args[1].(map[string]interface{})
	if !okH || !okP {
		return nil, fmt.Errorf("%w: %s invalid argument types (expected string, map)", core.ErrInvalidArgument, toolName)
	}

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

	var targetPos token.Pos
	var posErr error
	if offsetVal, okOffset := positionArg["offset"]; okOffset {
		// FIX: Correctly handle error return from ConvertToInt64E and cast to int
		offset, convErr := core.ConvertToInt64E(offsetVal)
		if convErr != nil {
			return nil, fmt.Errorf("%w: %s requires a valid integer 'offset', got %v: %w", core.ErrInvalidArgument, toolName, offsetVal, convErr)
		}
		if offset < 0 {
			return nil, fmt.Errorf("%w: %s requires non-negative 'offset', got %d", core.ErrInvalidArgument, toolName, offset)
		}
		if int(offset) >= tf.Size() {
			return nil, fmt.Errorf("%w: %s offset %d is out of file bounds (size %d)", core.ErrInvalidArgument, toolName, offset, tf.Size())
		}
		targetPos = tf.Pos(int(offset))
	} else if lineVal, okLine := positionArg["line"]; okLine {
		colVal, okCol := positionArg["column"]
		if !okCol {
			return nil, fmt.Errorf("%w: %s position map requires 'column' if 'line' is provided", core.ErrInvalidArgument, toolName)
		}
		// FIX: Use correct function name and handle error return
		line64, errL := core.ConvertToInt64E(lineVal)
		col64, errC := core.ConvertToInt64E(colVal)
		if errL != nil || errC != nil {
			return nil, fmt.Errorf("%w: %s requires integer 'line' and 'column', got line=%v, col=%v", core.ErrInvalidArgument, toolName, lineVal, colVal)
		}
		targetPos, posErr = getOffsetFromLineCol(tf, int(line64), int(col64))
		if posErr != nil {
			return nil, fmt.Errorf("%w: %s %w", core.ErrInvalidArgument, toolName, posErr)
		}
	} else {
		return nil, fmt.Errorf("%w: %s position map requires 'offset' or ('line' and 'column')", core.ErrInvalidArgument, toolName)
	}

	pathNodes, exact := astutil.PathEnclosingInterval(rootNode, targetPos, targetPos)
	if len(pathNodes) == 0 {
		return nil, fmt.Errorf("%w: no AST node found at the specified position", core.ErrNotFound)
	}
	node := pathNodes[len(pathNodes)-1]

	nodeType := reflect.TypeOf(node).String()
	nodeType = strings.TrimPrefix(nodeType, "*ast.")
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, node)

	resultMap := map[string]interface{}{
		"node_type":   nodeType,
		"node_text":   buf.String(),
		"exact_match": exact,
	}
	return resultMap, nil
}
