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
			Name:        "GoModTidy",
			Description: "Runs 'go mod tidy' in the sandbox to add missing and remove unused modules. Operates in the sandbox root.",
			Args:        []ArgSpec{},
			ReturnType:  ArgTypeMap,
		},
		Func: toolGoModTidy,
	},
	{
		Spec: ToolSpec{
			Name:        "GoListPackages",
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
			Name:        "GoBuild",
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
			Name:        "GoTest",
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
			Name:        "GoFmt", // Corrected name to match test expectation for toolGoFmt
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
			Name:        "GoImports",
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
			Name:        "GoVet",
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
			Name:        "GoFindDeclarations",
			Description: "Finds the declaration location of the Go symbol at the specified file position using a semantic index handle.",
			Args: []ArgSpec{
				{Name: "index_handle", Type: ArgTypeString, Required: true, Description: "Handle returned by GoIndexCode."},
				{Name: "path", Type: ArgTypeString, Required: true, Description: "File path relative to the indexed directory root."},
				{Name: "line", Type: ArgTypeInt, Required: true, Description: "1-based line number of the symbol."},
				{Name: "column", Type: ArgTypeInt, Required: true, Description: "1-based column number of the symbol."},
			},
			ReturnType: ArgTypeMap,
		},
		Func: toolGoFindDeclarations,
	},
	{
		Spec: ToolSpec{
			Name: "GoFindUsages",
			Description: "Finds all usage locations of a Go symbol given its definition site or any usage site. " +
				"Requires a semantic index handle created by GoIndexCode.",
			Args: []ArgSpec{
				{Name: "handle", Type: ArgTypeString, Required: true, Description: "Handle to the semantic index (from GoIndexCode)."},
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path within the indexed directory to the file containing the symbol identifier."},
				{Name: "line", Type: ArgTypeInt, Required: true, Description: "1-based line number of the symbol identifier."},
				{Name: "column", Type: ArgTypeInt, Required: true, Description: "1-based column number of the symbol identifier."},
			},
			ReturnType: ArgTypeSliceAny,
		},
		Func: toolGoFindUsages,
	},
	{
		Spec: ToolSpec{
			Name:        "GoIndexCode",
			Description: "Loads Go package information for the specified directory using 'go/packages' to build an in-memory semantic index. Returns a handle to the index.",
			Args: []ArgSpec{
				{Name: "directory", Type: ArgTypeString, Required: false, Description: "Directory relative to sandbox to index (packages loaded via './...'). Defaults to sandbox root ('.')."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolGoIndexCode,
	},
	{
		Spec: ToolSpec{
			Name:        "GoCheck",
			Description: "Checks Go code validity using 'go list -e -json <target>' within the sandbox. Returns a map indicating success and error details.",
			Args:        []ArgSpec{{Name: "target", Type: ArgTypeString, Required: true, Description: "Target Go package path or file path relative to sandbox (e.g., './pkg/core', 'main.go')."}},
			ReturnType:  ArgTypeMap,
		},
		Func: toolGoCheck,
	},
	{
		Spec: ToolSpec{
			Name:        "GoGetModuleInfo",
			Description: "Finds and parses the go.mod file relevant to a directory by searching upwards. Returns a map with module path, go version, root directory, requires, and replaces, or nil if not found.",
			Args:        []ArgSpec{{Name: "directory", Type: ArgTypeString, Required: false, Description: "Directory (relative to sandbox) to start searching upwards for go.mod. Defaults to '.' (sandbox root)."}},
			ReturnType:  ArgTypeMap,
		},
		Func: toolGoGetModuleInfo,
	},
}
