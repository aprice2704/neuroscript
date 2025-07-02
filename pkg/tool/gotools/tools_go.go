// NeuroScript Version: 0.3.0
// File version: 0.0.1 // Register GoFindUsages tool
// filename: pkg/tool/gotools/tools_go.go

package gotools

import (
	"fmt"
	// Keep existing imports: go/token, golang.org/x/tools/go/packages etc. from sub-files assumed available
)

// registerGoTools adds Go toolchain interaction tools to the registry.
// Includes build, format, test, diagnostics, and semantic indexing tools.
func registerGoTools(registry ToolRegistrar) error {	// Changed signature to accept ToolRegistrar interface
	// Note: ToolImplementation structs defined inline here for tools whose impl funcs
	// are in separate files but don't have their own *_impl.go variable defined.
	// Consider defining impl variables in respective files (e.g., toolGoBuildImpl) for consistency later.
	tools := []ToolImplementation{
		// --- Build / Test / Basic Commands ---
		// +++ ADDED: Defined in tools_go_find_usages.go +++

		// --- AST Tools (Registered Separately) ---
		// toolGoParseFileImpl, // Example if registered here
		// toolGoFormatASTImpl,
		// ... etc ...
	}

	// Register all defined tools
	for _, tool := range tools {
		// Basic validation before registration attempt
		if tool.Func == nil || tool.Spec.Name == "" {
			return fmt.Errorf("internal error: invalid Go tool definition provided for registration (tool name: %q, func defined: %t)", tool.Spec.Name, tool.Func != nil)
		}
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register Go tool %q: %w", tool.Spec.Name, err)
		}
	}

	// Register sub-packages of Go tools
	// Example: err := registerGoAstTools(registry)
	// if err != nil { return fmt.Errorf("failed to register Go AST tools: %w", err) }

	return nil	// Success
}

// Define other tool implementation functions (toolGoBuild, toolGoCheck, etc.) here or in separate files...
// Ensure funcs used above (toolGoBuild, toolGoCheck, toolGoTest, toolGoModTidy,
// toolGoListPackages, toolGoGetModuleInfo, toolGoFmt, toolGoImports, toolGoVet,
// toolStaticcheck) are defined in this package or imported.

// Dummy implementations for functions potentially defined elsewhere to allow this file to compile standalone for review
// In the actual project, these would be defined in their respective files (e.g., tools_go_execution.go)
// func toolGoBuild(interpreter *Interpreter, args []interface{}) (interface{}, error)           { return nil, nil }
// func toolGoCheck(interpreter *Interpreter, args []interface{}) (interface{}, error)           { return nil, nil }
// func toolGoTest(interpreter *Interpreter, args []interface{}) (interface{}, error)            { return nil, nil }
// func toolGoModTidy(interpreter *Interpreter, args []interface{}) (interface{}, error)         { return nil, nil }
// func toolGoListPackages(interpreter *Interpreter, args []interface{}) (interface{}, error)    { return nil, nil }
// func toolGoGetModuleInfo(interpreter *Interpreter, args []interface{}) (interface{}, error)   { return nil, nil }
// func toolGoFmt(interpreter *Interpreter, args []interface{}) (interface{}, error)             { return nil, nil }
// func toolGoImports(interpreter *Interpreter, args []interface{}) (interface{}, error)         { return nil, nil }
// func toolGoVet(interpreter *Interpreter, args []interface{}) (interface{}, error)             { return nil, nil }
// func toolStaticcheck(interpreter *Interpreter, args []interface{}) (interface{}, error)       { return nil, nil }