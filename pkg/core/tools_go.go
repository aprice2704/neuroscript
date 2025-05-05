// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 22:38:00 PDT // Add GoFindDeclarations tool registration
// filename: pkg/core/tools_go.go

package core

import (
	"fmt"
	// Keep existing imports
)

// registerGoTools adds Go toolchain interaction tools to the registry.
// Includes build, format, test, diagnostics, and semantic indexing tools.
func registerGoTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		// --- Build / Test / Basic Commands ---
		// ... (GoBuild, GoCheck, GoTest, GoModTidy, GoListPackages, GoGetModuleInfo specs unchanged) ...
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
		// ... (GoFmt, GoImports specs unchanged) ...
		{
			Spec: ToolSpec{Name: "GoFmt", Description: "Formats Go source code provided as a string using 'gofmt'. Returns formatted string on success, map with error details on failure.", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true, Description: "Go source code content as a string."}}, ReturnType: ArgTypeAny},
			Func: toolGoFmt,
		},
		{
			Spec: ToolSpec{Name: "GoImports", Description: "Formats Go source code string using 'goimports' logic (adds/removes imports, formats code). Returns formatted string on success, map with error details on failure.", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true, Description: "Go source code content as a string."}}, ReturnType: ArgTypeAny},
			Func: toolGoImports,
		},

		// --- Diagnostics ---
		// ... (GoVet, Staticcheck specs unchanged) ...
		{
			Spec: ToolSpec{Name: "GoVet", Description: "Runs 'go vet [target]' within the sandbox to find possible errors and suspicious constructs.", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional target relative to sandbox (e.g., './pkg/core/...', '.'). Defaults to './...'"}}, ReturnType: ArgTypeAny},
			Func: toolGoVet,
		},
		{
			Spec: ToolSpec{Name: "Staticcheck", Description: "Runs 'staticcheck [target]' within the sandbox for advanced static analysis. Assumes 'staticcheck' executable is in PATH.", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional target relative to sandbox (e.g., './pkg/core/...', '.'). Defaults to './...'"}}, ReturnType: ArgTypeAny},
			Func: toolStaticcheck,
		},

		// --- Semantic Indexing & Search ---
		{
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
		{ // +++ NEW GoFindDeclarations Spec +++
			Spec: ToolSpec{
				Name:        "GoFindDeclarations",
				Description: "Finds the declaration location of the Go symbol at the specified file position using a semantic index handle.",
				Args: []ArgSpec{
					{Name: "index_handle", Type: ArgTypeString, Required: true, Description: "Handle returned by GoIndexCode."},
					{Name: "path", Type: ArgTypeString, Required: true, Description: "File path relative to the indexed directory root."},
					{Name: "line", Type: ArgTypeInt, Required: true, Description: "1-based line number of the symbol."},
					{Name: "column", Type: ArgTypeInt, Required: true, Description: "1-based column number of the symbol."},
				},
				ReturnType: ArgTypeMap, // Returns map {path, line, column, name, kind} or nil
			},
			Func: toolGoFindDeclarations, // Implementation in tools_go_semantic.go
		},
		// Add GoFindUsages here later...
	}

	// Register all defined tools
	for _, tool := range tools {
		if tool.Func == nil || tool.Spec.Name == "" {
			return fmt.Errorf("internal error: invalid Go tool definition provided for registration (tool name: %q, func defined: %t)", tool.Spec.Name, tool.Func != nil)
		}
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register Go tool %q: %w", tool.Spec.Name, err)
		}
	}
	return nil // Success
}
