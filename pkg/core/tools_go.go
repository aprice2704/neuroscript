// NeuroScript Version: 0.3.0
// File version: 0.0.1 // Register GoFindUsages tool
// filename: pkg/core/tools_go.go

package core

import (
	"fmt"
	// Keep existing imports: go/token, golang.org/x/tools/go/packages etc. from sub-files assumed available
)

// registerGoTools adds Go toolchain interaction tools to the registry.
// Includes build, format, test, diagnostics, and semantic indexing tools.
func registerGoTools(registry ToolRegistrar) error { // Changed signature to accept ToolRegistrar interface
	// Note: ToolImplementation structs defined inline here for tools whose impl funcs
	// are in separate files but don't have their own *_impl.go variable defined.
	// Consider defining impl variables in respective files (e.g., toolGoBuildImpl) for consistency later.
	tools := []ToolImplementation{
		// --- Build / Test / Basic Commands ---
		{
			Spec: ToolSpec{Name: "GoBuild", Description: "Runs 'go build [target]' within the sandbox.", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional build target relative to sandbox (e.g., './cmd/app', '.'). Defaults to './...'"}}, ReturnType: ArgTypeAny},
			Func: toolGoBuild,
		},
		{
			Spec: ToolSpec{Name: "GoCheck", Description: "Checks Go code validity using 'go list -e -json <target>' within the sandbox.", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: true, Description: "Target Go package path or file path relative to sandbox (e.g., './pkg/core', 'main.go')."}}, ReturnType: ArgTypeAny},
			Func: toolGoCheck,
		},
		{
			Spec: ToolSpec{Name: "GoTest", Description: "Runs 'go test [target]' within the sandbox. Target paths must be relative to the sandbox. Defaults to './...'", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional test target relative to sandbox (e.g., './pkg/core/...', '.'). Defaults to './...'"}}, ReturnType: ArgTypeAny},
			Func: toolGoTest,
		},
		{
			Spec: ToolSpec{Name: "GoModTidy", Description: "Runs 'go mod tidy' within the sandbox.", Args: []ArgSpec{}, ReturnType: ArgTypeAny},
			Func: toolGoModTidy,
		},
		{
			Spec: ToolSpec{Name: "GoListPackages", Description: "Executes 'go list -json <patterns...>' in a specified directory (relative to sandbox) and returns parsed JSON information about the packages found.", Args: []ArgSpec{{Name: "directory", Type: ArgTypeString, Required: false, Description: "Directory relative to sandbox to run 'go list' in. Defaults to sandbox root ('.')."}, {Name: "patterns", Type: ArgTypeSliceString, Required: false, Description: "Go package patterns (e.g., './...', 'github.com/some/pkg'). Defaults to './...'."}}, ReturnType: ArgTypeSliceMap},
			Func: toolGoListPackages,
		},
		{
			Spec: ToolSpec{Name: "GoGetModuleInfo", Description: "Finds and parses the go.mod file relevant to a directory by searching upwards. Returns a map with module path, go version, root directory, requires, and replaces, or nil if not found.", Args: []ArgSpec{{Name: "directory", Type: ArgTypeString, Required: false, Description: "Directory (relative to sandbox) to start searching upwards for go.mod. Defaults to '.' (sandbox root)."}}, ReturnType: ArgTypeMap},
			Func: toolGoGetModuleInfo,
		},

		// --- Formatting ---
		{
			Spec: ToolSpec{Name: "GoFmt", Description: "Formats Go source code provided as a string using 'gofmt'. Returns formatted string on success, map with error details on failure.", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true, Description: "Go source code content as a string."}}, ReturnType: ArgTypeAny},
			Func: toolGoFmt,
		},
		{
			Spec: ToolSpec{Name: "GoImports", Description: "Formats Go source code string using 'goimports' logic (adds/removes imports, formats code). Returns formatted string on success, map with error details on failure.", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true, Description: "Go source code content as a string."}}, ReturnType: ArgTypeAny},
			Func: toolGoImports,
		},

		// --- Diagnostics ---
		{
			Spec: ToolSpec{Name: "GoVet", Description: "Runs 'go vet [target]' within the sandbox to find possible errors and suspicious constructs.", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional target relative to sandbox (e.g., './pkg/core/...', '.'). Defaults to './...'"}}, ReturnType: ArgTypeAny},
			Func: toolGoVet,
		},
		{
			Spec: ToolSpec{Name: "Staticcheck", Description: "Runs 'staticcheck [target]' within the sandbox for advanced static analysis. Assumes 'staticcheck' executable is in PATH.", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional target relative to sandbox (e.g., './pkg/core/...', '.'). Defaults to './...'"}}, ReturnType: ArgTypeAny},
			Func: toolStaticcheck,
		},

		// --- Semantic Indexing & Search ---
		// Note: Using ToolImplementation variables defined in respective files where available
		{ // Defined inline previously, keep for now unless we define toolGoIndexCodeImpl
			Spec: ToolSpec{
				Name:        "GoIndexCode",
				Description: "Loads Go package information for the specified directory using 'go/packages' to build an in-memory semantic index. Returns a handle to the index.",
				Args: []ArgSpec{
					{Name: "directory", Type: ArgTypeString, Required: false, Description: "Directory relative to sandbox to index (packages loaded via './...'). Defaults to sandbox root ('.')."},
				},
				ReturnType: ArgTypeString, // Returns the handle
			},
			Func: toolGoIndexCode, // Implementation in tools_go_semantic.go
		},
		toolGoFindDeclarationsImpl, // Defined in tools_go_find_declarations.go
		toolGoFindUsagesImpl,       // +++ ADDED: Defined in tools_go_find_usages.go +++

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

	return nil // Success
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
