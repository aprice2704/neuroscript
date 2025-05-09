// NeuroScript Version: 0.3.1
// File version: 0.1.2
// Registers the GoSemantic toolset with the NeuroScript toolset registry.
// filename: pkg/core/tools/gosemantic/register.go

package gosemantic

import (
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

const toolset = "GoSemantic"

func init() {
	var goSemanticToolsToRegister = []core.ToolImplementation{
		{
			Spec: core.ToolSpec{ // Using core.ToolSpec
				Name:        "Go.FindDeclarations",
				Description: "Finds the declaration location of the Go symbol at the specified file position using a semantic index handle.",
				Args: []core.ArgSpec{ // Using core.ArgSpec
					{Name: "index_handle", Type: core.ArgTypeString, Required: true, Description: "Handle returned by GoIndexCode."},
					{Name: "path", Type: core.ArgTypeString, Required: true, Description: "File path relative to the indexed directory root."},
					{Name: "line", Type: core.ArgTypeInt, Required: true, Description: "1-based line number of the symbol."},
					{Name: "column", Type: core.ArgTypeInt, Required: true, Description: "1-based column number of the symbol."},
				},
				ReturnType: core.ArgTypeMap,
			},
			Func: toolGoFindDeclarations, // Assumes toolGoFindDeclarations is in 'package gosemantic'
		},
		{
			Spec: core.ToolSpec{
				Name: "Go.FindUsages",
				Description: "Finds all usage locations of a Go symbol identified by a semantic query string. " +
					"Requires a semantic index handle created by Go.IndexCode.",
				Args: []core.ArgSpec{
					{
						Name:        "index_handle",
						Type:        core.ArgTypeString,
						Required:    true,
						Description: "Handle to the semantic index (from Go.IndexCode).",
					},
					{
						Name:        "query",
						Type:        core.ArgTypeString,
						Required:    true,
						Description: "Semantic query string identifying the symbol (e.g., 'package:my/pkg; function:MyFunc'). Same format as Go.GetDeclarationOfSymbol.",
					},
				},
				// Returns a list of maps: [{"path": string, "line": int64, "column": int64, "name": string, "kind": string}]
				ReturnType: core.ArgTypeSliceMap,
			},
			Func: toolGoFindUsages,
		},
		{
			Spec: core.ToolSpec{
				Name:        "Go.IndexCode",
				Description: "Loads Go package information for the specified directory using 'go/packages' to build an in-memory semantic index. Returns a handle to the index.",
				Args: []core.ArgSpec{
					{Name: "directory", Type: core.ArgTypeString, Required: false, Description: "Directory relative to sandbox to index (packages loaded via './...'). Defaults to sandbox root ('.')."},
				},
				ReturnType: core.ArgTypeString,
			},
			Func: toolGoIndexCode, // Assumes toolGoIndexCode is in 'package gosemantic'
		},
		{
			Spec: core.ToolSpec{
				Name: "Go.RenameSymbol",
				Description: "Finds a Go symbol via semantic query and generates a list of specific text replacement operations needed to rename it and all its usages within the indexed scope.\n" +
					"Does not modify files directly. Returns a list of patch operations.",
				Args: []core.ArgSpec{
					{Name: "index_handle", Type: core.ArgTypeString, Required: true, Description: "Handle returned by GoIndexCode."},
					{Name: "query", Type: core.ArgTypeString, Required: true, Description: "Semantic query string identifying the symbol to rename (e.g., 'package:my/pkg; function:MyFunc')."},
					{Name: "new_name", Type: core.ArgTypeString, Required: true, Description: "The desired new name for the symbol. Must be a valid Go identifier."},
				},
				// Returns a list of maps, each representing a patch operation:
				// {"path": string, "offset_start": int64, "offset_end": int64, "original_text": string, "new_text": string}
				ReturnType: core.ArgTypeSliceMap, // Confirmed from tools_types.go
			},
			Func: toolGoRenameSymbol,
		},
		{
			Spec: core.ToolSpec{
				Name: "Go.GetDeclarationOfSymbol",
				Description: "Finds the declaration location of a Go symbol using a semantic query string within an indexed codebase.\n" +
					"The query should be a semicolon-separated string of key:value pairs identifying the symbol.\n" +
					"Required Key: 'package' (e.g., 'package:github.com/example/pkg').\n" +
					"Symbol Keys: 'type', 'interface', 'method', 'function', 'var', 'const'. Use only one.\n" +
					"Context Keys (optional): 'receiver' (for methods, e.g., 'receiver:MyStruct' or 'receiver:*MyStruct'), 'field' (for fields within structs, same as using 'var' within 'type').\n" +
					"Examples:\n" +
					"  'package:github.com/example/pkg; function:ProcessData'\n" +
					"  'package:github.com/example/pkg; type:MyStruct'\n" +
					"  'package:github.com/example/pkg; type:MyStruct; method:DoThing'\n" +
					"  'package:github.com/example/pkg; type:MyStruct; field:counter' (or 'var:counter')\n" +
					"  'package:github.com/example/pkg; var:globalVar'",
				Args: []core.ArgSpec{
					{Name: "index_handle", Type: core.ArgTypeString, Required: true, Description: "Handle returned by GoIndexCode."},
					{Name: "query", Type: core.ArgTypeString, Required: true, Description: "Semantic query string identifying the symbol."},
				},
				ReturnType: core.ArgTypeMap,
			},
			Func: toolGoGetDeclarationOfSymbol,
		},
	}

	toolsets.AddToolsetRegistration(toolset,
		toolsets.CreateRegistrationFunc(toolset, goSemanticToolsToRegister))
}
