// NeuroScript Version: 0.4.2
// File version: 1.2.2
// Purpose: Implements script-loading tools. Corrects return type of ListFunctions to []interface{}.
// filename: pkg/tool/script/tools_script.go
package script

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolLoadScript implements the LoadScript tool.
func toolLoadScript(interp tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "LoadScript requires exactly one argument: script_content (string)", nil)
	}
	scriptContent, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "LoadScript argument must be a string", nil)
	}
	if scriptContent == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "LoadScript argument 'script_content' cannot be an empty string", nil)
	}

	logger := interp.GetLogger()

	// Stage 1: Parse
	parserAPI := parser.NewParserAPI(logger)
	parseTree, parseErr := parserAPI.Parse(scriptContent)
	if parseErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("parsing script failed: %s", parseErr.Error()), parseErr)
	}
	if parseTree == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "internal error: parser returned nil ANTLR tree", nil)
	}

	// Stage 2: Build AST and capture file metadata.
	astBuilder := parser.NewASTBuilder(logger)
	programAST, fileMetadata, buildErr := astBuilder.Build(parseTree)
	if buildErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("building AST for script failed: %s", buildErr.Error()), buildErr)
	}
	if programAST == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "internal error: AST builder yielded a nil ast.Program", nil)
	}

	// Convert map[string]string to map[string]interface{} for compatibility with Wrap.
	metadataInterface := make(map[string]interface{}, len(fileMetadata))
	for k, v := range fileMetadata {
		metadataInterface[k] = v
	}

	// Stage 4: Return result, now including the compatible metadata map.
	resultMap := map[string]interface{}{
		"functions_loaded":      len(programAST.Procedures),
		"event_handlers_loaded": len(programAST.Events),
		"metadata":              metadataInterface,
		"ast":                   programAST,
	}

	return resultMap, nil
}

// toolScriptListFunctions implements the Script.ListFunctions tool for introspection.
func toolScriptListFunctions(interp tool.Runtime, args []interface{}) (interface{}, error) {
	// Use the Meta.ListTools tool to get the list of all tools.
	listTools, err := interp.CallTool("Meta.ListTools", []interface{}{})
	if err != nil {
		return nil, err
	}

	toolsString, ok := listTools.(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "Meta.ListTools did not return a string", nil)
	}

	// Filter the list of tools to get only the functions.
	var functions []interface{}
	for _, line := range strings.Split(toolsString, "\n") {
		if strings.HasPrefix(line, "func ") {
			functions = append(functions, strings.Split(line, "(")[0])
		}
	}

	sort.Slice(functions, func(i, j int) bool {
		return functions[i].(string) < functions[j].(string)
	})

	return functions, nil
}
