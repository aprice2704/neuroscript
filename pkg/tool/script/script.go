// NeuroScript Version: 0.8.0
// File version: 12
// Purpose: Adds a static type assertion to ensure the interpreter implements the ScriptHost interface.
// filename: pkg/tool/script/script.go
// nlines: 154
// risk_rating: MEDIUM
package script

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// ScriptHost defines the methods the script tools need from the interpreter.
// This is necessary because the standard tool.Runtime is too restrictive
// for tools that need to inspect or modify the program's structure.
type ScriptHost interface {
	tool.Runtime
	// AddProcedure adds a single procedure to the interpreter's registry.
	AddProcedure(proc ast.Procedure) error
	// RegisterEvent registers a single event handler.
	RegisterEvent(decl *ast.OnEventDecl) error
	// KnownProcedures returns the map of all loaded procedures.
	KnownProcedures() map[string]*ast.Procedure
}

// Statically assert that the concrete Interpreter type satisfies the ScriptHost interface.
var _ ScriptHost = (*interpreter.Interpreter)(nil)

// toolLoadScript is the implementation for the `LoadScript` tool. It takes a
// script as a string, parses it, builds an AST, and merges it into the
// interpreter's currently loaded program.
func toolLoadScript(rt tool.Runtime, args []any) (any, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "LoadScript requires exactly one argument", nil)
	}

	scriptContent, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "LoadScript argument must be a string", nil)
	}

	// Return an error if the script content is empty or only whitespace.
	if strings.TrimSpace(scriptContent) == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, "cannot load an empty script", lang.ErrSyntax)
	}

	host, ok := rt.(ScriptHost)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "runtime does not support required script-loading interface", nil)
	}

	logger := rt.GetLogger()
	if logger == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "runtime logger is not available", nil)
	}

	parserAPI := parser.NewParserAPI(logger)
	parseTree, parseErr := parserAPI.Parse(scriptContent)
	if parseErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, fmt.Sprintf("failed to parse script: %v", parseErr), parseErr)
	}

	astBuilder := parser.NewASTBuilder(logger)
	programAST, metadata, buildErr := astBuilder.Build(parseTree)
	if buildErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("failed to build script AST: %v", buildErr), buildErr)
	}

	// Manually merge the new program AST into the interpreter via the ScriptHost interface.
	for _, proc := range programAST.Procedures {
		if err := host.AddProcedure(*proc); err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeExecutionFailed, fmt.Sprintf("failed to load function '%s': %v", proc.Name(), err), err)
		}
	}
	for _, eventDecl := range programAST.Events {
		if err := host.RegisterEvent(eventDecl); err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeExecutionFailed, fmt.Sprintf("failed to load event handler: %v", err), err)
		}
	}

	// Construct the result map according to the tool's specification.
	result := map[string]interface{}{
		"functions_loaded":      len(programAST.Procedures),
		"event_handlers_loaded": len(programAST.Events),
		"metadata":              convertMap(metadata),
	}
	return result, nil
}

// toolScriptListFunctions implements the `Script.ListFunctions` tool. It
// inspects the interpreter's loaded program and returns a map of all
// available function signatures.
func toolScriptListFunctions(rt tool.Runtime, args []any) (any, error) {
	if len(args) != 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Script.ListFunctions requires no arguments", nil)
	}

	host, ok := rt.(ScriptHost)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "runtime does not support required script-listing interface", nil)
	}

	programProcs := host.KnownProcedures()
	if programProcs == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "no program loaded", nil)
	}

	funcSigs := make(map[string]interface{})
	for name, proc := range programProcs {
		var paramInfo []string
		for _, paramName := range proc.RequiredParams {
			paramInfo = append(paramInfo, paramName)
		}
		for _, optParam := range proc.OptionalParams {
			paramInfo = append(paramInfo, fmt.Sprintf("[%s]", optParam.Name))
		}
		funcSigs[name] = fmt.Sprintf("procedure %s(%s)", name, strings.Join(paramInfo, ", "))
	}
	return funcSigs, nil
}

// convertMap converts a map[string]string to map[string]any for compatibility
// with the lang.Wrap function.
func convertMap(in map[string]string) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
