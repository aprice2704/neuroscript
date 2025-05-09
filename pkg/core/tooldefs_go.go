// NeuroScript Version: 0.3.1
// File version: 0.1.3 // Corrected GoFormat to GoFmt to match test expectations.
// Defines ToolImplementation structs for selected Go language tools.
// filename: pkg/core/tooldefs_go.go

package core

// goToolsToRegister contains ToolImplementation definitions for a subset of Go language tools.
// This array is intended to be concatenated with other similar arrays in a central
// registrar (e.g., zz_core_tools_registrar.go) to be processed by AddToolImplementations.
var goToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "Go.ModTidy",
			Description: "Runs 'go mod tidy' in the sandbox to add missing and remove unused modules. Operates in the sandbox root.",
			Args:        []ArgSpec{},
			ReturnType:  ArgTypeMap,
		},
		Func: toolGoModTidy,
	},
	{
		Spec: ToolSpec{
			Name:        "Go.ListPackages",
			Description: "Runs 'go list -json' for specified patterns in a target directory. Returns a list of maps, each describing a package.",
			Args: []ArgSpec{
				{Name: "target_directory", Type: ArgTypeString, Required: false, Description: "Optional. The directory relative to the sandbox root to run 'go list'. Defaults to '.' (sandbox root)."},
				{Name: "patterns", Type: ArgTypeSliceString, Required: false, Description: "Optional. A list of package patterns (e.g., './...', 'example.com/project/...'). Defaults to ['./...']."},
			},
			ReturnType: ArgTypeSliceMap,
		},
		Func: toolGoListPackages,
	},
	{
		Spec: ToolSpec{
			Name:        "Go.Build",
			Description: "Runs 'go build' for a specified target in the sandbox. Defaults to './...'.",
			Args: []ArgSpec{
				{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional. The build target (e.g., a package path or './...'). Defaults to './...'."},
			},
			ReturnType: ArgTypeMap,
		},
		Func: toolGoBuild,
	},
	{
		Spec: ToolSpec{
			Name:        "Go.Test",
			Description: "Runs 'go test' for a specified target in the sandbox. Defaults to './...'.",
			Args: []ArgSpec{
				{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional. The test target (e.g., a package path or './...'). Defaults to './...'."},
			},
			ReturnType: ArgTypeMap,
		},
		Func: toolGoTest,
	},
	{
		Spec: ToolSpec{
			Name:        "Go.Fmt", // Corrected name to match test expectation for toolGoFmt
			Description: "Formats Go source code using 'go/format.Source'. Returns the formatted code or an error map.",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The Go source code content to format."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolGoFmt,
	},
	{
		Spec: ToolSpec{
			Name:        "Go.Imports",
			Description: "Formats Go source code and adjusts imports using 'golang.org/x/tools/imports'. Returns the processed code or an error map.",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The Go source code content to process."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolGoImports,
	},
	{
		Spec: ToolSpec{
			Name:        "Go.Vet",
			Description: "Runs 'go vet' on the specified target(s) in the sandbox to report likely mistakes in Go source code. Defaults to './...'.",
			Args: []ArgSpec{
				{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional. The target for 'go vet' (e.g., a package path or './...'). Defaults to './...'."},
			},
			ReturnType: ArgTypeMap,
		},
		Func: toolGoVet,
	},
	{
		Spec: ToolSpec{
			Name:        "Staticcheck",
			Description: "Runs 'staticcheck' on the specified target(s) in the sandbox. Reports bugs, stylistic errors, and performance issues. Defaults to './...'. Assumes 'staticcheck' is in PATH.",
			Args: []ArgSpec{
				{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional. The target for 'staticcheck' (e.g., a package path or './...'). Defaults to './...'."},
			},
			ReturnType: ArgTypeMap,
		},
		Func: toolStaticcheck,
	},
	{
		Spec: ToolSpec{
			Name:        "Go.Check",
			Description: "Checks Go code validity using 'go list -e -json <target>' within the sandbox. Returns a map indicating success and error details.",
			Args:        []ArgSpec{{Name: "target", Type: ArgTypeString, Required: true, Description: "Target Go package path or file path relative to sandbox (e.g., './pkg/core', 'main.go')."}},
			ReturnType:  ArgTypeMap,
		},
		Func: toolGoCheck,
	},
	{
		Spec: ToolSpec{
			Name:        "Go.GetModuleInfo",
			Description: "Finds and parses the go.mod file relevant to a directory by searching upwards. Returns a map with module path, go version, root directory, requires, and replaces, or nil if not found.",
			Args:        []ArgSpec{{Name: "directory", Type: ArgTypeString, Required: false, Description: "Directory (relative to sandbox) to start searching upwards for go.mod. Defaults to '.' (sandbox root)."}},
			ReturnType:  ArgTypeMap,
		},
		Func: toolGoGetModuleInfo,
	},
}
