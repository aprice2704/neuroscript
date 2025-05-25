// NeuroScript Go Indexer - Data Structures
// File version: 2.0.0 (Plan-aligned version)
// Purpose: Defines the Go structs for the code index, aligning with goindexer_plan_002.md.
// filename: pkg/goindex/types.go
package goindex

// ParamDetail holds information about a function/method parameter.
type ParamDetail struct {
	Name string `json:"name"`
	Type string `json:"type"` // Fully qualified type string
}

// FunctionDetail holds information about an exported function.
type FunctionDetail struct {
	Name       string        `json:"name"`       // Fully qualified name (e.g., "github.com/user/repo/pkg.FuncName")
	SourceFile string        `json:"sourceFile"` // Relative to component root
	Parameters []ParamDetail `json:"parameters,omitempty"`
	Returns    []string      `json:"returns,omitempty"` // List of type strings
}

// MethodDetail holds information about an exported method.
// Also used for methods within InterfaceDetail.
type MethodDetail struct {
	ReceiverName string        `json:"receiverName,omitempty"` // e.g., "MyType" or "*MyType" (empty for interface methods)
	ReceiverType string        `json:"receiverType,omitempty"` // Fully qualified type string (empty for interface methods)
	Name         string        `json:"name"`
	SourceFile   string        `json:"sourceFile"` // Relative to component root (for concrete methods)
	Parameters   []ParamDetail `json:"parameters,omitempty"`
	Returns      []string      `json:"returns,omitempty"`
	// Calls        []string      `json:"calls,omitempty"` // Deferred to a later phase
}

// FieldDetail holds information about an exported struct field.
type FieldDetail struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // Fully qualified type string
	Tags     string `json:"tags,omitempty"`
	Exported bool   `json:"exported"` // Should always be true if indexed based on plan
}

// StructDetail holds information about an exported struct.
type StructDetail struct {
	Name       string        `json:"name"`
	SourceFile string        `json:"sourceFile"` // Relative to component root
	Fields     []FieldDetail `json:"fields,omitempty"`
}

// InterfaceDetail holds information about an exported interface.
type InterfaceDetail struct {
	Name       string         `json:"name"`
	SourceFile string         `json:"sourceFile"`        // Relative to component root
	Methods    []MethodDetail `json:"methods,omitempty"` // Method set for the interface
}

// GlobalVarDetail holds information about an exported global variable.
type GlobalVarDetail struct {
	Name       string `json:"name"`
	Type       string `json:"type"`            // Fully qualified type string
	SourceFile string `json:"sourceFile"`      // Relative to component root
	Value      string `json:"value,omitempty"` // Value if it's a simple literal
}

// GlobalConstDetail holds information about an exported global constant.
type GlobalConstDetail struct {
	Name       string `json:"name"`
	Type       string `json:"type"`       // Fully qualified type string
	SourceFile string `json:"sourceFile"` // Relative to component root
	Value      string `json:"value"`      // Value (constants usually have one)
}

// TypeAliasDetail holds information about an exported type alias or definition.
type TypeAliasDetail struct {
	Name           string `json:"name"`
	UnderlyingType string `json:"underlyingType"` // Fully qualified type string
	SourceFile     string `json:"sourceFile"`     // Relative to component root
}

// PackageDetail aggregates all exported entities within a Go package.
type PackageDetail struct {
	PackagePath  string              `json:"packagePath"` // Fully qualified Go package path (e.g., "github.com/aprice2704/neuroscript/pkg/core")
	PackageName  string              `json:"packageName"` // Short package name (e.g., "core")
	Functions    []FunctionDetail    `json:"functions,omitempty"`
	Methods      []MethodDetail      `json:"methods,omitempty"` // Flat list of all methods in the package
	Structs      []StructDetail      `json:"structs,omitempty"`
	Interfaces   []InterfaceDetail   `json:"interfaces,omitempty"`
	GlobalVars   []GlobalVarDetail   `json:"globalVars,omitempty"`
	GlobalConsts []GlobalConstDetail `json:"globalConsts,omitempty"`
	TypeAliases  []TypeAliasDetail   `json:"typeAliases,omitempty"`
}

// --- Structs for NeuroScript Tool Indexing (Phase 3) ---

// NeuroScriptArgDetail mirrors core.ArgSpec for the index.
type NeuroScriptArgDetail struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"` // String representation from core.ArgType
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"defaultValue,omitempty"` // From core.ArgSpec
}

// NeuroScriptToolDetail holds all information goindexer collects about a single NeuroScript tool.
type NeuroScriptToolDetail struct {
	// Fields from core.ToolSpec (mirrored based on enhanced version)
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category,omitempty"`
	Args            []NeuroScriptArgDetail `json:"args,omitempty"`
	ReturnType      string                 `json:"returnType"` // String representation from core.ArgType
	ReturnHelp      string                 `json:"returnHelp,omitempty"`
	Variadic        bool                   `json:"variadic,omitempty"`
	Example         string                 `json:"example,omitempty"`
	ErrorConditions string                 `json:"errorConditions,omitempty"`

	// Linkage to Go implementation, identified by goindexer
	ImplementingGoFunctionFullName     string `json:"implementingGoFunctionFullName,omitempty"`   // Fully qualified Go function name
	ImplementingGoFunctionSourceFile   string `json:"implementingGoFunctionSourceFile,omitempty"` // Relative to its component root
	GoImplementingFunctionReturnsError bool   `json:"goImplementingFunctionReturnsError"`         // True if the Go func signature includes 'error'
}

// --- Top-Level Index Structures (Phase 2) ---

// ComponentIndex holds all indexed information for a specific component.
type ComponentIndex struct {
	ComponentName      string                    `json:"componentName"`      // e.g., "core"
	ComponentPath      string                    `json:"componentPath"`      // Path to the component relative to project root (e.g., "pkg/core")
	IndexSchemaVersion string                    `json:"indexSchemaVersion"` // e.g., "component_index_v2.0"
	Packages           map[string]*PackageDetail `json:"packages"`           // Keyed by fully qualified Go package path
	NeuroScriptTools   []NeuroScriptToolDetail   `json:"neuroscriptTools,omitempty"`
	LastIndexed        string                    `json:"lastIndexed,omitempty"`   // Timestamp of when indexing occurred for this component
	GitBranch          string                    `json:"gitBranch,omitempty"`     // Git branch at time of indexing
	GitCommitHash      string                    `json:"gitCommitHash,omitempty"` // Git commit hash at time of indexing
}

// ProjectIndex is the top-level structure for the `project_index.json` file.
type ProjectIndex struct {
	ProjectRootModulePath string                             `json:"projectRootModulePath"`   // e.g., "github.com/aprice2704/neuroscript"
	IndexSchemaVersion    string                             `json:"indexSchemaVersion"`      // e.g., "project_index_v2.0"
	LastIndexedTimestamp  string                             `json:"lastIndexedTimestamp"`    // Timestamp of when project indexing finished
	Components            map[string]ComponentIndexFileEntry `json:"components"`              // Keyed by component name
	GitBranch             string                             `json:"gitBranch,omitempty"`     // Git branch at time of indexing
	GitCommitHash         string                             `json:"gitCommitHash,omitempty"` // Git commit hash at time of indexing
}

// ComponentIndexFileEntry describes a component within the ProjectIndex.
type ComponentIndexFileEntry struct {
	Name        string `json:"name"`                  // e.g., "core" (used as key in ProjectIndex.Components map)
	Path        string `json:"path"`                  // Path to the component relative to project root (e.g., "pkg/core")
	IndexFile   string `json:"indexFile"`             // Filename of the component's detailed index (e.g., "core_index.json")
	Description string `json:"description,omitempty"` // Optional description of the component
}
