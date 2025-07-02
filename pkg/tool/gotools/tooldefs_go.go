// NeuroScript Version: 0.3.1
// File version: 0.1.4 // Populated Category, Example, ReturnHelp, ErrorConditions for ToolSpecs.
// Defines ToolImplementation structs for selected Go language tools.
// filename: pkg/tool/gotools/tooldefs_go.go
// nlines: 200 // Approximate
// risk_rating: MEDIUM // Executes Go commands, can interact with file system and network.

package gotools

// goToolsToRegister contains ToolImplementation definitions for a subset of Go language tools.
// This array is intended to be concatenated with other similar arrays in a central
// registrar (e.g., zz_core_tools_registrar.go) to be processed by AddToolImplementations.
var goToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:			"Go.ModTidy",
			Description:		"Runs 'go mod tidy' in the sandbox to add missing and remove unused modules. Operates in the sandbox root.",
			Category:		"Go Build Tools",
			Args:			[]ArgSpec{},
			ReturnType:		ArgTypeMap,
			ReturnHelp:		"Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go mod tidy' command execution.",
			Example:		`TOOL.Go.ModTidy()`,
			ErrorConditions:	"ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution failures are reported within the returned map's 'success', 'stderr', and 'exit_code' fields.",
		},
		Func:	toolGoModTidy,	//
	},
	{
		Spec: ToolSpec{
			Name:		"Go.ListPackages",
			Description:	"Runs 'go list -json' for specified patterns in a target directory. Returns a list of maps, each describing a package.",
			Category:	"Go Build Tools",
			Args: []ArgSpec{
				{Name: "target_directory", Type: ArgTypeString, Required: false, Description: "Optional. The directory relative to the sandbox root to run 'go list'. Defaults to '.' (sandbox root)."},
				{Name: "patterns", Type: ArgTypeSliceString, Required: false, Description: "Optional. A list of package patterns (e.g., './...', 'example.com/project/...'). Defaults to ['./...']."},
			},
			ReturnType:		ArgTypeSliceMap,
			ReturnHelp:		"Returns a slice of maps, where each map is a JSON object representing a Go package as output by 'go list -json'. Returns an empty slice on command failure or if JSON decoding fails.",
			Example:		`TOOL.Go.ListPackages(target_directory: "pkg/core", patterns: ["./..."])`,
			ErrorConditions:	"ErrValidationTypeMismatch if patterns arg contains non-string elements; ErrInternalTool if execution helper fails internally or JSON decoding fails; ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. 'go list' command failures are reported in its output map rather than a Go error from the tool.",
		},
		Func:	toolGoListPackages,	//
	},
	{
		Spec: ToolSpec{
			Name:		"Go.Build",
			Description:	"Runs 'go build' for a specified target in the sandbox. Defaults to './...'.",
			Category:	"Go Build Tools",
			Args: []ArgSpec{
				{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional. The build target (e.g., a package path or './...'). Defaults to './...'."},
			},
			ReturnType:		ArgTypeMap,
			ReturnHelp:		"Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go build <target>' command.",
			Example:		`TOOL.Go.Build(target: "./cmd/mytool")`,
			ErrorConditions:	"ErrInvalidArgument if optional target is not a string; ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution failures are reported within the returned map.",
		},
		Func:	toolGoBuild,	//
	},
	{
		Spec: ToolSpec{
			Name:		"Go.Test",
			Description:	"Runs 'go test' for a specified target in the sandbox. Defaults to './...'.",
			Category:	"Go Build Tools",
			Args: []ArgSpec{
				{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional. The test target (e.g., a package path or './...'). Defaults to './...'."},
			},
			ReturnType:		ArgTypeMap,
			ReturnHelp:		"Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go test <target>' command.",
			Example:		`TOOL.Go.Test(target: "./pkg/feature")`,
			ErrorConditions:	"ErrInvalidArgument if optional target is not a string; ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution failures are reported within the returned map.",
		},
		Func:	toolGoTest,	//
	},
	{
		Spec: ToolSpec{
			Name:		"Go.Fmt",
			Description:	"Formats Go source code using 'go/format.Source'. Returns the formatted code or an error map.",
			Category:	"Go Formatting",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The Go source code content to format."},
			},
			ReturnType:		ArgTypeString,	// Returns formatted string on success, or map on error.
			ReturnHelp:		"Returns the formatted Go source code as a string. If formatting fails (e.g., syntax error), returns a map {'formatted_content': <original_content>, 'error': <error_string>, 'success': false} and a Go-level error.",
			Example:		`TOOL.Go.Fmt(content: "package main\nfunc main(){}")`,
			ErrorConditions:	"ErrInternalTool if formatting fails internally, wrapping the original Go error from format.Source. The specific formatting error (e.g. syntax error) is in the 'error' field of the returned map if applicable.",
		},
		Func:	toolGoFmt,	//
	},
	{
		Spec: ToolSpec{
			Name:		"Go.Imports",
			Description:	"Formats Go source code and adjusts imports using 'golang.org/x/tools/imports'. Returns the processed code or an error map.",
			Category:	"Go Formatting",
			Args: []ArgSpec{
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The Go source code content to process."},
			},
			ReturnType:		ArgTypeString,	// Returns formatted string on success, or map on error.
			ReturnHelp:		"Returns the processed Go source code (formatted and with adjusted imports) as a string. If processing fails, returns a map {'formatted_content': <original_content>, 'error': <error_string>, 'success': false} and a Go-level error.",
			Example:		`TOOL.Go.Imports(content: "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"hello\")}")`,
			ErrorConditions:	"ErrInternalTool if goimports processing fails, wrapping the original error from imports.Process. The specific processing error is in the 'error' field of the returned map if applicable.",
		},
		Func:	toolGoImports,	//
	},
	{
		Spec: ToolSpec{
			Name:		"Go.Vet",
			Description:	"Runs 'go vet' on the specified target(s) in the sandbox to report likely mistakes in Go source code. Defaults to './...'.",
			Category:	"Go Diagnostics",
			Args: []ArgSpec{
				{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional. The target for 'go vet' (e.g., a package path or './...'). Defaults to './...'."},
			},
			ReturnType:		ArgTypeMap,
			ReturnHelp:		"Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go vet <target>' command. 'stderr' usually contains the vet diagnostics.",
			Example:		`TOOL.Go.Vet(target: "./pkg/core")`,
			ErrorConditions:	"ErrInvalidArgument if optional target is not a string; ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution failures are reported within the returned map.",
		},
		Func:	toolGoVet,	//
	},
	{
		Spec: ToolSpec{
			Name:		"Staticcheck",
			Description:	"Runs 'staticcheck' on the specified target(s) in the sandbox. Reports bugs, stylistic errors, and performance issues. Defaults to './...'. Assumes 'staticcheck' is in PATH.",
			Category:	"Go Diagnostics",
			Args: []ArgSpec{
				{Name: "target", Type: ArgTypeString, Required: false, Description: "Optional. The target for 'staticcheck' (e.g., a package path or './...'). Defaults to './...'."},
			},
			ReturnType:		ArgTypeMap,
			ReturnHelp:		"Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'staticcheck <target>' command. 'stdout' usually contains the diagnostics.",
			Example:		`TOOL.Staticcheck(target: "./...")`,
			ErrorConditions:	"ErrInvalidArgument if optional target is not a string; ErrToolExecutionFailed if 'staticcheck' command fails (e.g. not found, or internal error), reported via the toolExecuteCommand structure.",
		},
		Func:	toolStaticcheck,	//
	},
	{
		Spec: ToolSpec{
			Name:			"Go.Check",
			Description:		"Checks Go code validity using 'go list -e -json <target>' within the sandbox. Returns a map indicating success and error details.",
			Category:		"Go Diagnostics",
			Args:			[]ArgSpec{{Name: "target", Type: ArgTypeString, Required: true, Description: "Target Go package path or file path relative to sandbox (e.g., './pkg/core', 'main.go')."}},
			ReturnType:		ArgTypeMap,
			ReturnHelp:		"Returns a map with 'check_success' (bool) and 'error_details' (string). 'check_success' is true if 'go list -e -json' finds no errors in the target's JSON output. 'error_details' contains messages if errors are found or if the command fails.",
			Example:		`TOOL.Go.Check(target: "./pkg/core")`,
			ErrorConditions:	"ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution issues or JSON parsing errors result in 'check_success':false and details in 'error_details'.",
		},
		Func:	toolGoCheck,	//
	},
	{
		Spec: ToolSpec{
			Name:			"Go.GetModuleInfo",
			Description:		"Finds and parses the go.mod file relevant to a directory by searching upwards. Returns a map with module path, go version, root directory, requires, and replaces, or nil if not found.",
			Category:		"Go Build Tools",
			Args:			[]ArgSpec{{Name: "directory", Type: ArgTypeString, Required: false, Description: "Directory (relative to sandbox) to start searching upwards for go.mod. Defaults to '.' (sandbox root)."}},
			ReturnType:		ArgTypeMap,	// Returns nil if go.mod not found, or map on success.
			ReturnHelp:		"Returns a map containing 'modulePath', 'goVersion', 'rootDir' (absolute path to module root), 'requires' (list of maps), and 'replaces' (list of maps). Returns nil if no go.mod is found.",
			Example:		`TOOL.Go.GetModuleInfo(directory: "cmd/mytool")`,
			ErrorConditions:	"ErrValidationTypeMismatch if directory arg is not a string; ErrInternalSecurity if sandbox is not set or for path validation errors; ErrInternalTool if FindAndParseGoMod fails for reasons other than os.ErrNotExist (e.g., parsing error, file read error). If go.mod is not found, returns nil result and nil error (not a Go-level tool error).",
		},
		Func:	toolGoGetModuleInfo,	//
	},
}