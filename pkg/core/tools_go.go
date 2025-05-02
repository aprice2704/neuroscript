// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 21:04:41 PDT // Register GoImports tool
// filename: pkg/core/tools_go.go

package core

import (
	"fmt"
)

// registerGoTools adds Go toolchain interaction tools to the registry.
func registerGoTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{ // GoBuild Spec...
			Spec: ToolSpec{Name: "GoBuild", Description: "Runs 'go build [target]' within the sandbox.", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional build target relative to sandbox (e.g., './cmd/app', '.'). Defaults to './...'"}}, ReturnType: ArgTypeAny},
			Func: toolGoBuild, // Implementation in tools_go_execution.go
		},
		{ // GoCheck Spec...
			Spec: ToolSpec{Name: "GoCheck", Description: "Checks Go code validity using 'go list -e -json <target>' within the sandbox.", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: true, Description: "Target Go package path or file path relative to sandbox (e.g., './pkg/core', 'main.go')."}}, ReturnType: ArgTypeAny},
			Func: toolGoCheck, // Implementation in tools_go_execution.go
		},
		{ // GoTest Spec...
			Spec: ToolSpec{Name: "GoTest", Description: "Runs 'go test [target]' within the sandbox. Target paths must be relative to the sandbox. Defaults to './...'", Args: []ArgSpec{{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional test target relative to sandbox (e.g., './pkg/core/...', '.'). Defaults to './...'"}}, ReturnType: ArgTypeAny},
			Func: toolGoTest, // Implementation in tools_go_execution.go
		},
		{ // GoFmt Spec...
			Spec: ToolSpec{Name: "GoFmt", Description: "Formats Go source code provided as a string using 'gofmt'. Returns formatted string on success, map with error details on failure.", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true, Description: "Go source code content as a string."}}, ReturnType: ArgTypeAny},
			Func: toolGoFmt, // Implementation in tools_go_fmt.go
		},
		// +++ ADD GoImports Spec +++
		{
			Spec: ToolSpec{
				Name:        "GoImports",
				Description: "Formats Go source code string using 'goimports' logic (adds/removes imports, formats code). Returns formatted string on success, map with error details on failure.",
				Args: []ArgSpec{
					// Consider adding optional 'path' arg later if context is needed for complex imports
					{Name: "content", Type: ArgTypeString, Required: true, Description: "Go source code content as a string."},
				},
				ReturnType: ArgTypeAny, // String on success, map on error
			},
			Func: toolGoImports, // Implementation in tools_go_fmt.go
		},
		// +++ END GoImports Spec +++
		{ // GoModTidy Spec...
			Spec: ToolSpec{Name: "GoModTidy", Description: "Runs 'go mod tidy' within the sandbox.", Args: []ArgSpec{}, ReturnType: ArgTypeAny},
			Func: toolGoModTidy, // Implementation in tools_go_execution.go
		},
		{ // GoListPackages Spec...
			Spec: ToolSpec{Name: "GoListPackages", Description: "Executes 'go list -json <patterns...>' in a specified directory (relative to sandbox) and returns parsed JSON information about the packages found.", Args: []ArgSpec{{Name: "directory", Type: ArgTypeString, Required: false, Description: "Directory relative to sandbox to run 'go list' in. Defaults to sandbox root ('.')."}, {Name: "patterns", Type: ArgTypeSliceString, Required: false, Description: "Go package patterns (e.g., './...', 'github.com/some/pkg'). Defaults to './...'."}}, ReturnType: ArgTypeSliceMap},
			Func: toolGoListPackages, // Implementation in tools_go_execution.go
		},
		{ // GoGetModuleInfo Spec...
			Spec: ToolSpec{Name: "GoGetModuleInfo", Description: "Finds and parses the go.mod file relevant to a directory by searching upwards. Returns a map with module path, go version, root directory, requires, and replaces, or nil if not found.", Args: []ArgSpec{{Name: "directory", Type: ArgTypeString, Required: false, Description: "Directory (relative to sandbox) to start searching upwards for go.mod. Defaults to '.' (sandbox root)."}}, ReturnType: ArgTypeMap},
			Func: toolGoGetModuleInfo, // Implementation in tools_go_mod.go
		},
	}

	// Register all defined tools
	for _, tool := range tools {
		if tool.Func == nil || tool.Spec.Name == "" {
			return fmt.Errorf("internal error: invalid Go tool definition provided for registration (missing Func or Name)")
		}
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register Go tool %q: %w", tool.Spec.Name, err)
		}
	}
	return nil // Success
}
