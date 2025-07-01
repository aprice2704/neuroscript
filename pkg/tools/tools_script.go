// NeuroScript Version: 0.4.2
// File version: 1.2.2
// Purpose: Implements script-loading tools. Corrects return type of ListFunctions to []interface{}.
// filename: pkg/core/tools_script.go
package core

import (
	"fmt"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// toolLoadScript implements the LoadScript tool.
func toolLoadScript(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "LoadScript requires exactly one argument: script_content (string)", nil)
	}
	scriptContent, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "LoadScript argument must be a string", nil)
	}
	if scriptContent == "" {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "LoadScript argument 'script_content' cannot be an empty string", nil)
	}

	logger := interpreter.Logger()

	// Stage 1: Parse
	parserAPI := NewParserAPI(logger)
	parseTree, parseErr := parserAPI.Parse(scriptContent)
	if parseErr != nil {
		return nil, lang.NewRuntimeError(ErrorCodeSyntax, fmt.Sprintf("parsing script failed: %s", parseErr.Error()), parseErr)
	}
	if parseTree == nil {
		return nil, lang.NewRuntimeError(ErrorCodeInternal, "internal error: parser returned nil ANTLR tree", nil)
	}

	// Stage 2: Build AST and capture file metadata.
	astBuilder := NewASTBuilder(logger)
	programAST, fileMetadata, buildErr := astBuilder.Build(parseTree)
	if buildErr != nil {
		return nil, lang.NewRuntimeError(ErrorCodeSyntax, fmt.Sprintf("building AST for script failed: %s", buildErr.Error()), buildErr)
	}
	if programAST == nil {
		return nil, lang.NewRuntimeError(ErrorCodeInternal, "internal error: AST builder yielded a nil ast.Program", nil)
	}

	// Stage 3: Load into Interpreter
	if err := interpreter.LoadProgram(programAST); err != nil {
		return nil, lang.NewRuntimeError(ErrorCodeExecutionFailed, fmt.Sprintf("failed to load program into interpreter: %s", err.Error()), err)
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
	}

	return resultMap, nil
}

// toolScriptListFunctions implements the Script.ListFunctions tool for introspection.
func toolScriptListFunctions(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	knownProcs := interpreter.KnownProcedures()
	names := make([]string, 0, len(knownProcs))
	for name := range knownProcs {
		names = append(names, name)
	}

	sort.Strings(names)

	// Convert []string to []interface{} to ensure it can be wrapped by the interpreter.
	interfaceSlice := make([]interface{}, len(names))
	for i, v := range names {
		interfaceSlice[i] = v
	}

	return interfaceSlice, nil
}
